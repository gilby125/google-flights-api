package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/redis/go-redis/v9"
)

const jobTTL = 24 * time.Hour

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
}

// Queue defines the interface for a job queue
type Queue interface {
	Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error)
	Dequeue(ctx context.Context, queueName string) (*Job, error)
	Ack(ctx context.Context, queueName, jobID string) error
	Nack(ctx context.Context, queueName, jobID string) error
	GetJobStatus(ctx context.Context, jobID string) (string, error)
	GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error)
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

// GetClient returns the underlying Redis client for advanced operations like distributed locking
func (q *RedisQueue) GetClient() *redis.Client {
	return q.client
}
