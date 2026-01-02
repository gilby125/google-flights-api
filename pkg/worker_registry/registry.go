package worker_registry

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type WorkerHeartbeat struct {
	ID            string
	Hostname      string
	Status        string
	CurrentJob    string
	ProcessedJobs int
	Concurrency   int
	StartedAt     time.Time
	LastHeartbeat time.Time
	Version       string
}

type Registry struct {
	redisClient *redis.Client
	namespace   string
}

func New(redisClient *redis.Client, namespace string) *Registry {
	return &Registry{
		redisClient: redisClient,
		namespace:   namespace,
	}
}

func (r *Registry) heartbeatsKey() string {
	return fmt.Sprintf("worker_registry:%s:heartbeats", r.namespace)
}

func (r *Registry) metaKey(workerID string) string {
	return fmt.Sprintf("worker_registry:%s:worker:%s", r.namespace, workerID)
}

func (r *Registry) Publish(ctx context.Context, hb WorkerHeartbeat, ttl time.Duration) error {
	if r == nil || r.redisClient == nil {
		return nil
	}
	if hb.ID == "" {
		return fmt.Errorf("worker id is required")
	}
	if ttl <= 0 {
		ttl = 45 * time.Second
	}

	now := time.Now().UTC()
	if hb.StartedAt.IsZero() {
		hb.StartedAt = now
	}
	hb.StartedAt = hb.StartedAt.UTC()
	if hb.LastHeartbeat.IsZero() {
		hb.LastHeartbeat = now
	}
	hb.LastHeartbeat = hb.LastHeartbeat.UTC()

	pipe := r.redisClient.Pipeline()
	pipe.ZAdd(ctx, r.heartbeatsKey(), redis.Z{
		Score:  float64(hb.LastHeartbeat.Unix()),
		Member: hb.ID,
	})

	meta := map[string]any{
		"id":             hb.ID,
		"hostname":       hb.Hostname,
		"status":         hb.Status,
		"current_job":    hb.CurrentJob,
		"processed_jobs": hb.ProcessedJobs,
		"concurrency":    hb.Concurrency,
		"started_at":     hb.StartedAt.Unix(),
		"last_heartbeat": hb.LastHeartbeat.Unix(),
		"version":        hb.Version,
	}
	pipe.HSet(ctx, r.metaKey(hb.ID), meta)
	pipe.Expire(ctx, r.metaKey(hb.ID), ttl*3)
	pipe.ZRemRangeByScore(ctx, r.heartbeatsKey(), "0", strconv.FormatInt(now.Add(-ttl*10).Unix(), 10))
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (r *Registry) ListActive(ctx context.Context, within time.Duration, limit int64) ([]WorkerHeartbeat, error) {
	if r == nil || r.redisClient == nil {
		return []WorkerHeartbeat{}, nil
	}
	if within <= 0 {
		within = 45 * time.Second
	}
	if limit <= 0 {
		limit = 100
	}

	now := time.Now().UTC()
	ids, err := r.redisClient.ZRevRangeByScore(ctx, r.heartbeatsKey(), &redis.ZRangeBy{
		Max:    strconv.FormatInt(now.Unix(), 10),
		Min:    strconv.FormatInt(now.Add(-within).Unix(), 10),
		Offset: 0,
		Count:  limit,
	}).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []WorkerHeartbeat{}, nil
	}

	pipe := r.redisClient.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(ids))
	for _, id := range ids {
		cmds = append(cmds, pipe.HGetAll(ctx, r.metaKey(id)))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	out := make([]WorkerHeartbeat, 0, len(ids))
	for i, id := range ids {
		m := cmds[i].Val()
		hb := WorkerHeartbeat{
			ID:       id,
			Hostname: m["hostname"],
			Status:   m["status"],
			Version:  m["version"],
		}

		hb.CurrentJob = m["current_job"]
		if v, err := strconv.Atoi(m["processed_jobs"]); err == nil {
			hb.ProcessedJobs = v
		}
		if v, err := strconv.Atoi(m["concurrency"]); err == nil {
			hb.Concurrency = v
		}
		if v, err := strconv.ParseInt(m["started_at"], 10, 64); err == nil {
			hb.StartedAt = time.Unix(v, 0).UTC()
		}
		if v, err := strconv.ParseInt(m["last_heartbeat"], 10, 64); err == nil {
			hb.LastHeartbeat = time.Unix(v, 0).UTC()
		}

		out = append(out, hb)
	}

	return out, nil
}
