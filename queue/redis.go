package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/redis/go-redis/v9"
)

const jobTTL = 24 * time.Hour
const enqueueMetricTTL = 48 * time.Hour

// EnqueueMeta carries best-effort attribution for who/what enqueued a job.
// It is stored on the Job and can be surfaced by admin/debug endpoints.
type EnqueueMeta struct {
	Actor     string `json:"actor,omitempty"` // e.g. "http", "scheduler", "continuous_sweep"
	RequestID string `json:"request_id,omitempty"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	RemoteIP  string `json:"remote_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

func (m EnqueueMeta) isEmpty() bool {
	return m.Actor == "" && m.RequestID == "" && m.Method == "" && m.Path == "" && m.RemoteIP == "" && m.UserAgent == ""
}

type enqueueMetaKey struct{}

// WithEnqueueMeta attaches enqueue attribution to the provided context.
func WithEnqueueMeta(ctx context.Context, meta EnqueueMeta) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if meta.isEmpty() {
		return ctx
	}
	return context.WithValue(ctx, enqueueMetaKey{}, meta)
}

// EnqueueMetaFromContext returns enqueue attribution stored on the context, if present.
func EnqueueMetaFromContext(ctx context.Context) EnqueueMeta {
	if ctx == nil {
		return EnqueueMeta{}
	}
	if v := ctx.Value(enqueueMetaKey{}); v != nil {
		if meta, ok := v.(EnqueueMeta); ok {
			return meta
		}
	}
	return EnqueueMeta{}
}

// Job represents a flight search job
type Job struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	Status      string          `json:"status"`
	StreamID    string          `json:"stream_id,omitempty"`
	EnqueueMeta *EnqueueMeta    `json:"enqueue_meta,omitempty"`
}

// Queue defines the interface for a job queue
type Queue interface {
	Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error)
	Dequeue(ctx context.Context, queueName string) (*Job, error)
	Ack(ctx context.Context, queueName, jobID string) error
	Nack(ctx context.Context, queueName, jobID string) error
	GetJobStatus(ctx context.Context, jobID string) (string, error)
	GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error)
	// CancelJob requests cancellation for a job. Workers may stop immediately or on their next cancellation check.
	CancelJob(ctx context.Context, queueName, jobID string) error
	// IsJobCanceled returns whether a cancellation has been requested for the job.
	IsJobCanceled(ctx context.Context, jobID string) (bool, error)
	// GetJob fetches persisted job details by job ID.
	GetJob(ctx context.Context, jobID string) (*Job, error)
	// ListJobs lists jobs from the status set (pending/processing/completed/failed).
	ListJobs(ctx context.Context, queueName, state string, limit, offset int) ([]*Job, error)
	// GetBacklog returns the most recent unacked stream entries for the queue.
	GetBacklog(ctx context.Context, queueName string, limit int) ([]*Job, error)
	// GetEnqueueMetrics aggregates enqueue sources over the last N minutes.
	GetEnqueueMetrics(ctx context.Context, queueName string, minutes int) (map[string]int64, error)
	// ClearFailed removes all failed jobs for the named queue.
	ClearFailed(ctx context.Context, queueName string) (cleared int64, err error)
	// ClearProcessing removes all "processing" jobs for the named queue and acks/dels their stream entries when possible.
	ClearProcessing(ctx context.Context, queueName string) (cleared int64, err error)
	// RetryFailed moves up to limit failed jobs back into pending by re-enqueueing them.
	RetryFailed(ctx context.Context, queueName string, limit int) (retried int64, err error)
	// ClearQueue removes all pending jobs from the named queue without touching in-flight processing jobs.
	// It is intended for admin/debug use when a backlog needs to be drained safely.
	ClearQueue(ctx context.Context, queueName string) (cleared int64, err error)
}

// RedisQueue implements the Queue interface using Redis Streams
type RedisQueue struct {
	client          *redis.Client
	cfg             config.RedisConfig
	consumerName    string
	mu              sync.Mutex
	ensuredStreams  map[string]struct{}
	lastAutoClaimID map[string]string
}

// NewRedisQueue creates a new Redis-backed queue
func NewRedisQueue(cfg config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "worker"
	}
	consumerName := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())

	return &RedisQueue{
		client:          client,
		cfg:             cfg,
		consumerName:    consumerName,
		ensuredStreams:  make(map[string]struct{}),
		lastAutoClaimID: make(map[string]string),
	}, nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error) {
	if err := q.ensureStream(ctx, jobType); err != nil {
		return "", err
	}

	jobID := fmt.Sprintf("%s-%d", jobType, time.Now().UnixNano())

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job payload: %w", err)
	}

	job := &Job{
		ID:          jobID,
		Type:        jobType,
		Payload:     payloadBytes,
		CreatedAt:   time.Now().UTC(),
		Attempts:    0,
		MaxAttempts: 3,
		Status:      "pending",
	}
	if meta := EnqueueMetaFromContext(ctx); !meta.isEmpty() {
		job.EnqueueMeta = &meta
	}

	enqueueBytes, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	stream := q.streamName(jobType)
	msgID, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"job": enqueueBytes,
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("failed to add job to stream: %w", err)
	}

	job.StreamID = msgID
	if err := q.persistJob(ctx, job); err != nil {
		return "", err
	}

	if err := q.client.SAdd(ctx, q.pendingKey(jobType), jobID).Err(); err != nil {
		return "", fmt.Errorf("failed to record pending job: %w", err)
	}

	// Best-effort enqueue metrics (do not fail enqueue if metrics cannot be recorded).
	_ = q.recordEnqueueMetric(ctx, jobType, job.EnqueueMeta)

	return jobID, nil
}

// Dequeue retrieves a job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*Job, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return nil, err
	}

	// First attempt to reclaim stale messages
	if job, err := q.claimStale(ctx, queueName); err != nil {
		return nil, err
	} else if job != nil {
		return job, nil
	}

	stream := q.streamName(queueName)
	cmd := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.cfg.QueueGroup,
		Consumer: q.consumerName,
		Streams:  []string{stream, ">"},
		Count:    1,
		Block:    q.cfg.QueueBlockTimeout,
	})
	res, err := cmd.Result()
	if errors.Is(err, redis.Nil) || len(res) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(res[0].Messages) == 0 {
		return nil, nil
	}

	job, err := q.prepareMessage(ctx, queueName, res[0].Messages[0])
	if err != nil {
		return nil, err
	}

	return job, nil
}

// Ack acknowledges a job as completed
func (q *RedisQueue) Ack(ctx context.Context, queueName, jobID string) error {
	job, jobKey, err := q.getStoredJob(ctx, jobID)
	if err != nil {
		return err
	}

	stream := q.streamName(queueName)

	job.Status = "completed"
	if err := q.persistJob(ctx, job); err != nil {
		return err
	}

	if job.StreamID != "" {
		if err := q.client.XAck(ctx, stream, q.cfg.QueueGroup, job.StreamID).Err(); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
		// Trim acknowledged entry
		_ = q.client.XDel(ctx, stream, job.StreamID).Err()
	}

	if err := q.client.SRem(ctx, q.processingKey(queueName), jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove job from processing set: %w", err)
	}
	if err := q.client.SAdd(ctx, q.completedKey(queueName), jobID).Err(); err != nil {
		return fmt.Errorf("failed to add job to completed set: %w", err)
	}

	_ = q.client.Expire(ctx, jobKey, jobTTL).Err()

	return nil
}

// Nack marks a job as failed or requeues it
func (q *RedisQueue) Nack(ctx context.Context, queueName, jobID string) error {
	job, jobKey, err := q.getStoredJob(ctx, jobID)
	if err != nil {
		return err
	}

	stream := q.streamName(queueName)

	if job.StreamID != "" {
		if err := q.client.XAck(ctx, stream, q.cfg.QueueGroup, job.StreamID).Err(); err != nil {
			return fmt.Errorf("failed to ack message before retry: %w", err)
		}
		_ = q.client.XDel(ctx, stream, job.StreamID).Err()
	}

	if job.Attempts < job.MaxAttempts {
		job.Status = "pending"
		job.StreamID = ""
		if err := q.persistJob(ctx, job); err != nil {
			return err
		}

		requeuePayload, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job for requeue: %w", err)
		}

		msgID, err := q.client.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]interface{}{
				"job": requeuePayload,
			},
		}).Result()
		if err != nil {
			return fmt.Errorf("failed to requeue job: %w", err)
		}

		job.StreamID = msgID
		if err := q.persistJob(ctx, job); err != nil {
			return err
		}

		if err := q.client.SAdd(ctx, q.pendingKey(queueName), jobID).Err(); err != nil {
			return fmt.Errorf("failed to mark job pending: %w", err)
		}
		if err := q.client.SRem(ctx, q.processingKey(queueName), jobID).Err(); err != nil {
			return fmt.Errorf("failed to clear processing flag: %w", err)
		}
	} else {
		job.Status = "failed"
		if err := q.persistJob(ctx, job); err != nil {
			return err
		}

		if err := q.client.SRem(ctx, q.processingKey(queueName), jobID).Err(); err != nil {
			return fmt.Errorf("failed to remove job from processing set: %w", err)
		}
		if err := q.client.SAdd(ctx, q.failedKey(queueName), jobID).Err(); err != nil {
			return fmt.Errorf("failed to add job to failed set: %w", err)
		}
	}

	_ = q.client.Expire(ctx, jobKey, jobTTL).Err()

	return nil
}

// GetJobStatus gets the status of a job
func (q *RedisQueue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	jobKey := q.jobKey(jobID)
	jobBytes, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return "", fmt.Errorf("failed to get job details: %w", err)
	}

	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return "", fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return job.Status, nil
}

// GetQueueStats gets statistics for a queue
func (q *RedisQueue) GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error) {
	stats := make(map[string]int64)

	pendingCount, err := q.client.SCard(ctx, q.pendingKey(queueName)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}
	processingCount, err := q.client.SCard(ctx, q.processingKey(queueName)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing count: %w", err)
	}
	completedCount, err := q.client.SCard(ctx, q.completedKey(queueName)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}
	failedCount, err := q.client.SCard(ctx, q.failedKey(queueName)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	stats["pending"] = pendingCount
	stats["processing"] = processingCount
	stats["completed"] = completedCount
	stats["failed"] = failedCount

	return stats, nil
}

func (q *RedisQueue) CancelJob(ctx context.Context, queueName, jobID string) error {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return err
	}
	if jobID == "" {
		return fmt.Errorf("missing job id")
	}

	// Always set the cancel flag (even if the job record was cleared), so workers can observe it.
	if err := q.client.Set(ctx, q.cancelKey(jobID), "1", jobTTL).Err(); err != nil {
		return fmt.Errorf("failed to set cancel flag: %w", err)
	}

	// Best-effort: mark stored job canceled (do not delete; keep for debugging).
	if job, _, err := q.getStoredJob(ctx, jobID); err == nil && job != nil {
		job.Status = "canceled"
		_ = q.persistJob(ctx, job)
	}

	// Best-effort: remove from pending set so it doesn't start later.
	_ = q.client.SRem(ctx, q.pendingKey(queueName), jobID).Err()

	// Best-effort: if we still have stream bookkeeping, ack/del it so it won't get re-delivered.
	stream := q.streamName(queueName)
	if job, _, err := q.getStoredJob(ctx, jobID); err == nil && job != nil && job.StreamID != "" {
		_ = q.client.XAck(ctx, stream, q.cfg.QueueGroup, job.StreamID).Err()
		_ = q.client.XDel(ctx, stream, job.StreamID).Err()
	}

	return nil
}

func (q *RedisQueue) IsJobCanceled(ctx context.Context, jobID string) (bool, error) {
	if jobID == "" {
		return false, nil
	}
	n, err := q.client.Exists(ctx, q.cancelKey(jobID)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cancel flag: %w", err)
	}
	return n > 0, nil
}

func (q *RedisQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	job, _, err := q.getStoredJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (q *RedisQueue) ListJobs(ctx context.Context, queueName, state string, limit, offset int) ([]*Job, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	var setKey string
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "pending":
		setKey = q.pendingKey(queueName)
	case "processing":
		setKey = q.processingKey(queueName)
	case "completed":
		setKey = q.completedKey(queueName)
	case "failed":
		setKey = q.failedKey(queueName)
	default:
		return nil, fmt.Errorf("invalid state %q", state)
	}

	jobIDs, err := q.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list %s jobs: %w", state, err)
	}

	jobs := make([]*Job, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		job, err := q.GetJob(ctx, jobID)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return nil, err
		}
		jobs = append(jobs, job)
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	if offset >= len(jobs) {
		return []*Job{}, nil
	}
	end := offset + limit
	if end > len(jobs) {
		end = len(jobs)
	}
	return jobs[offset:end], nil
}

func (q *RedisQueue) GetBacklog(ctx context.Context, queueName string, limit int) ([]*Job, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}

	stream := q.streamName(queueName)
	msgs, err := q.client.XRevRangeN(ctx, stream, "+", "-", int64(limit)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []*Job{}, nil
		}
		return nil, fmt.Errorf("failed to read queue backlog: %w", err)
	}

	out := make([]*Job, 0, len(msgs))
	for _, msg := range msgs {
		rawJob, ok := msg.Values["job"]
		if !ok {
			continue
		}

		var jobID string
		switch v := rawJob.(type) {
		case string:
			var tmp Job
			if err := json.Unmarshal([]byte(v), &tmp); err == nil {
				jobID = tmp.ID
			}
		case []byte:
			var tmp Job
			if err := json.Unmarshal(v, &tmp); err == nil {
				jobID = tmp.ID
			}
		}

		var job *Job
		if jobID != "" {
			stored, err := q.GetJob(ctx, jobID)
			if err == nil && stored != nil {
				job = stored
			}
		}
		if job == nil {
			var tmp Job
			var jobBytes []byte
			switch v := rawJob.(type) {
			case string:
				jobBytes = []byte(v)
			case []byte:
				jobBytes = v
			}
			if err := json.Unmarshal(jobBytes, &tmp); err != nil {
				continue
			}
			tmp.StreamID = msg.ID
			job = &tmp
		} else {
			job.StreamID = msg.ID
		}

		if job.Status == "" {
			job.Status = "pending"
		}

		out = append(out, job)
	}

	return out, nil
}

func (q *RedisQueue) GetEnqueueMetrics(ctx context.Context, queueName string, minutes int) (map[string]int64, error) {
	if minutes <= 0 {
		minutes = 60
	}
	if minutes > 24*60 {
		minutes = 24 * 60
	}

	now := time.Now().UTC().Truncate(time.Minute)
	keys := make([]string, 0, minutes)
	for i := 0; i < minutes; i++ {
		keys = append(keys, q.enqueueMetricKey(queueName, now.Add(-time.Duration(i)*time.Minute)))
	}

	pipe := q.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(keys))
	for _, key := range keys {
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to read enqueue metrics: %w", err)
	}

	out := make(map[string]int64)
	for _, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("failed to read enqueue metrics: %w", err)
		}
		for source, countStr := range m {
			var n int64
			_, _ = fmt.Sscan(countStr, &n)
			out[source] += n
		}
	}
	return out, nil
}

func (q *RedisQueue) ClearQueue(ctx context.Context, queueName string) (int64, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return 0, err
	}

	stream := q.streamName(queueName)
	jobIDs, err := q.client.SMembers(ctx, q.pendingKey(queueName)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to list pending jobs: %w", err)
	}

	var cleared int64
	for _, jobID := range jobIDs {
		jobKey := q.jobKey(jobID)
		jobBytes, getErr := q.client.Get(ctx, jobKey).Bytes()
		if getErr != nil && !errors.Is(getErr, redis.Nil) {
			return cleared, fmt.Errorf("failed to get job details for %s: %w", jobID, getErr)
		}

		if getErr == nil {
			var job Job
			if err := json.Unmarshal(jobBytes, &job); err != nil {
				return cleared, fmt.Errorf("failed to unmarshal job %s: %w", jobID, err)
			}

			if job.StreamID != "" {
				_ = q.client.XDel(ctx, stream, job.StreamID).Err()
			}
			_ = q.client.Del(ctx, jobKey).Err()
		}

		_ = q.client.SRem(ctx, q.pendingKey(queueName), jobID).Err()
		cleared++
	}

	return cleared, nil
}

func (q *RedisQueue) ClearFailed(ctx context.Context, queueName string) (int64, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return 0, err
	}

	jobIDs, err := q.client.SMembers(ctx, q.failedKey(queueName)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to list failed jobs: %w", err)
	}

	var cleared int64
	for _, jobID := range jobIDs {
		_ = q.client.Del(ctx, q.jobKey(jobID)).Err()
		_ = q.client.SRem(ctx, q.failedKey(queueName), jobID).Err()
		cleared++
	}
	return cleared, nil
}

func (q *RedisQueue) ClearProcessing(ctx context.Context, queueName string) (int64, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return 0, err
	}

	stream := q.streamName(queueName)
	jobIDs, err := q.client.SMembers(ctx, q.processingKey(queueName)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to list processing jobs: %w", err)
	}

	var cleared int64
	for _, jobID := range jobIDs {
		job, _, getErr := q.getStoredJob(ctx, jobID)
		if getErr == nil && job != nil && job.StreamID != "" {
			_ = q.client.XAck(ctx, stream, q.cfg.QueueGroup, job.StreamID).Err()
			_ = q.client.XDel(ctx, stream, job.StreamID).Err()
		}
		_ = q.client.Del(ctx, q.jobKey(jobID)).Err()
		_ = q.client.SRem(ctx, q.processingKey(queueName), jobID).Err()
		cleared++
	}

	return cleared, nil
}

func (q *RedisQueue) RetryFailed(ctx context.Context, queueName string, limit int) (int64, error) {
	if err := q.ensureStream(ctx, queueName); err != nil {
		return 0, err
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	stream := q.streamName(queueName)
	jobIDs, err := q.client.SMembers(ctx, q.failedKey(queueName)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to list failed jobs: %w", err)
	}

	var retried int64
	for _, jobID := range jobIDs {
		if retried >= int64(limit) {
			break
		}

		job, _, err := q.getStoredJob(ctx, jobID)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				_ = q.client.SRem(ctx, q.failedKey(queueName), jobID).Err()
				continue
			}
			return retried, err
		}

		job.Attempts = 0
		job.Status = "pending"
		job.StreamID = ""

		requeuePayload, err := json.Marshal(job)
		if err != nil {
			return retried, fmt.Errorf("failed to marshal job for retry: %w", err)
		}

		msgID, err := q.client.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]interface{}{"job": requeuePayload},
		}).Result()
		if err != nil {
			return retried, fmt.Errorf("failed to requeue failed job: %w", err)
		}

		job.StreamID = msgID
		if err := q.persistJob(ctx, job); err != nil {
			return retried, err
		}

		_ = q.client.SRem(ctx, q.failedKey(queueName), jobID).Err()
		_ = q.client.SAdd(ctx, q.pendingKey(queueName), jobID).Err()
		retried++
	}

	return retried, nil
}

func (q *RedisQueue) ensureStream(ctx context.Context, queueName string) error {
	stream := q.streamName(queueName)

	q.mu.Lock()
	if _, ok := q.ensuredStreams[stream]; ok {
		q.mu.Unlock()
		return nil
	}
	q.mu.Unlock()

	err := q.client.XGroupCreateMkStream(ctx, stream, q.cfg.QueueGroup, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	q.mu.Lock()
	q.ensuredStreams[stream] = struct{}{}
	q.mu.Unlock()
	return nil
}

func (q *RedisQueue) claimStale(ctx context.Context, queueName string) (*Job, error) {
	stream := q.streamName(queueName)

	q.mu.Lock()
	startID := q.lastAutoClaimID[stream]
	if startID == "" {
		startID = "0-0"
	}
	q.mu.Unlock()

	messages, nextID, err := q.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   stream,
		Group:    q.cfg.QueueGroup,
		Consumer: q.consumerName,
		MinIdle:  q.cfg.QueueVisibilityTimeout,
		Start:    startID,
		Count:    1,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to auto-claim messages: %w", err)
	}

	q.mu.Lock()
	q.lastAutoClaimID[stream] = nextID
	q.mu.Unlock()

	if len(messages) == 0 {
		return nil, nil
	}

	job, err := q.prepareMessage(ctx, queueName, messages[0])
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (q *RedisQueue) prepareMessage(ctx context.Context, queueName string, msg redis.XMessage) (*Job, error) {
	rawJob, ok := msg.Values["job"]
	if !ok {
		return nil, fmt.Errorf("stream message missing job payload")
	}

	var jobBytes []byte
	switch v := rawJob.(type) {
	case string:
		jobBytes = []byte(v)
	case []byte:
		jobBytes = v
	default:
		return nil, fmt.Errorf("unexpected job payload type %T", v)
	}

	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job payload: %w", err)
	}

	job.StreamID = msg.ID
	if job.Type == "" {
		job.Type = queueName
	}
	job.Attempts++
	job.Status = "processing"

	if err := q.persistJob(ctx, &job); err != nil {
		return nil, err
	}

	if err := q.client.SAdd(ctx, q.processingKey(queueName), job.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to mark job processing: %w", err)
	}
	if err := q.client.SRem(ctx, q.pendingKey(queueName), job.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to remove job from pending: %w", err)
	}

	return &job, nil
}

func (q *RedisQueue) persistJob(ctx context.Context, job *Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job for storage: %w", err)
	}

	if err := q.client.Set(ctx, q.jobKey(job.ID), jobBytes, jobTTL).Err(); err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}
	return nil
}

func (q *RedisQueue) getStoredJob(ctx context.Context, jobID string) (*Job, string, error) {
	jobKey := q.jobKey(jobID)
	jobBytes, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get job details: %w", err)
	}

	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, jobKey, nil
}

func (q *RedisQueue) streamName(jobType string) string {
	return fmt.Sprintf("%s:%s", q.cfg.QueueStreamPrefix, jobType)
}

func (q *RedisQueue) jobKey(jobID string) string {
	return fmt.Sprintf("job:%s", jobID)
}

func (q *RedisQueue) cancelKey(jobID string) string {
	return fmt.Sprintf("job:%s:cancel", jobID)
}

func (q *RedisQueue) pendingKey(queueName string) string {
	return fmt.Sprintf("queue:%s:pending", queueName)
}

func (q *RedisQueue) processingKey(queueName string) string {
	return fmt.Sprintf("queue:%s:processing", queueName)
}

func (q *RedisQueue) completedKey(queueName string) string {
	return fmt.Sprintf("queue:%s:completed", queueName)
}

func (q *RedisQueue) failedKey(queueName string) string {
	return fmt.Sprintf("queue:%s:failed", queueName)
}

func (q *RedisQueue) enqueueMetricKey(queueName string, ts time.Time) string {
	// minute bucket key, e.g. queue:flight_search:enqueues:202601261505
	return fmt.Sprintf("queue:%s:enqueues:%s", queueName, ts.Format("200601021504"))
}

func (q *RedisQueue) recordEnqueueMetric(ctx context.Context, queueName string, meta *EnqueueMeta) error {
	source := "unknown"
	if meta != nil {
		if meta.Actor != "" && meta.Method != "" && meta.Path != "" {
			source = fmt.Sprintf("%s %s %s", meta.Actor, meta.Method, meta.Path)
		} else if meta.Method != "" && meta.Path != "" {
			source = fmt.Sprintf("%s %s", meta.Method, meta.Path)
		} else if meta.Actor != "" {
			source = meta.Actor
		} else if meta.RequestID != "" {
			source = fmt.Sprintf("request_id:%s", meta.RequestID)
		}
	}

	key := q.enqueueMetricKey(queueName, time.Now().UTC().Truncate(time.Minute))
	pipe := q.client.Pipeline()
	pipe.HIncrBy(ctx, key, source, 1)
	pipe.Expire(ctx, key, enqueueMetricTTL)
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

// GetClient returns the underlying Redis client for advanced operations like distributed locking
func (q *RedisQueue) GetClient() *redis.Client {
	return q.client
}
