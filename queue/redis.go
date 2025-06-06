package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/redis/go-redis/v9"
)

// Job represents a flight search job
type Job struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	Status      string          `json:"status"`
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

// RedisQueue implements the Queue interface using Redis
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a new Redis-backed queue
func NewRedisQueue(cfg config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{client: client}, nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error) {
	// Generate a unique job ID
	jobID := fmt.Sprintf("%s-%d", jobType, time.Now().UnixNano())

	// Serialize the payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job payload: %w", err)
	}

	// Create the job
	job := &Job{
		ID:          jobID,
		Type:        jobType,
		Payload:     payloadBytes,
		CreatedAt:   time.Now(),
		Attempts:    0,
		MaxAttempts: 3,
		Status:      "pending",
	}

	// Serialize the job
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add the job to the queue
	queueName := fmt.Sprintf("queue:%s", jobType)
	if err := q.client.LPush(ctx, queueName, jobBytes).Err(); err != nil {
		return "", fmt.Errorf("failed to push job to queue: %w", err)
	}

	// Store the job details
	jobKey := fmt.Sprintf("job:%s", jobID)
	if err := q.client.Set(ctx, jobKey, jobBytes, 24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to store job details: %w", err)
	}

	return jobID, nil
}

// Dequeue retrieves a job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*Job, error) {
	// Get a job from the queue
	result, err := q.client.BRPop(ctx, 5*time.Second, fmt.Sprintf("queue:%s", queueName)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to pop job from queue: %w", err)
	}

	// Parse the job
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job status and attempts
	job.Attempts++
	job.Status = "processing"

	// Serialize the updated job
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated job: %w", err)
	}

	// Store the updated job details
	jobKey := fmt.Sprintf("job:%s", job.ID)
	if err := q.client.Set(ctx, jobKey, jobBytes, 24*time.Hour).Err(); err != nil {
		return nil, fmt.Errorf("failed to update job details: %w", err)
	}

	// Add the job to the processing set
	processingKey := fmt.Sprintf("processing:%s", queueName)
	if err := q.client.SAdd(ctx, processingKey, job.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to add job to processing set: %w", err)
	}

	return &job, nil
}

// Ack acknowledges a job as completed
func (q *RedisQueue) Ack(ctx context.Context, queueName, jobID string) error {
	// Get the job details
	jobKey := fmt.Sprintf("job:%s", jobID)
	jobBytes, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get job details: %w", err)
	}

	// Parse the job
	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job status
	job.Status = "completed"

	// Serialize the updated job
	jobBytes, err = json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal updated job: %w", err)
	}

	// Store the updated job details
	if err := q.client.Set(ctx, jobKey, jobBytes, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update job details: %w", err)
	}

	// Remove the job from the processing set
	processingKey := fmt.Sprintf("processing:%s", queueName)
	if err := q.client.SRem(ctx, processingKey, jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove job from processing set: %w", err)
	}

	// Add the job to the completed set
	completedKey := fmt.Sprintf("completed:%s", queueName)
	if err := q.client.SAdd(ctx, completedKey, jobID).Err(); err != nil {
		return fmt.Errorf("failed to add job to completed set: %w", err)
	}

	return nil
}

// Nack marks a job as failed
func (q *RedisQueue) Nack(ctx context.Context, queueName, jobID string) error {
	// Get the job details
	jobKey := fmt.Sprintf("job:%s", jobID)
	jobBytes, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get job details: %w", err)
	}

	// Parse the job
	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Check if the job should be retried
	if job.Attempts < job.MaxAttempts {
		// Update job status
		job.Status = "pending"

		// Serialize the updated job
		jobBytes, err = json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal updated job: %w", err)
		}

		// Store the updated job details
		if err := q.client.Set(ctx, jobKey, jobBytes, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to update job details: %w", err)
		}

		// Add the job back to the queue
		queueKey := fmt.Sprintf("queue:%s", queueName)
		if err := q.client.LPush(ctx, queueKey, jobBytes).Err(); err != nil {
			return fmt.Errorf("failed to push job back to queue: %w", err)
		}
	} else {
		// Update job status
		job.Status = "failed"

		// Serialize the updated job
		jobBytes, err = json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal updated job: %w", err)
		}

		// Store the updated job details
		if err := q.client.Set(ctx, jobKey, jobBytes, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to update job details: %w", err)
		}

		// Add the job to the failed set
		failedKey := fmt.Sprintf("failed:%s", queueName)
		if err := q.client.SAdd(ctx, failedKey, jobID).Err(); err != nil {
			return fmt.Errorf("failed to add job to failed set: %w", err)
		}
	}

	// Remove the job from the processing set
	processingKey := fmt.Sprintf("processing:%s", queueName)
	if err := q.client.SRem(ctx, processingKey, jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove job from processing set: %w", err)
	}

	return nil
}

// GetJobStatus gets the status of a job
func (q *RedisQueue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	// Get the job details
	jobKey := fmt.Sprintf("job:%s", jobID)
	jobBytes, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return "", fmt.Errorf("failed to get job details: %w", err)
	}

	// Parse the job
	var job Job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		return "", fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return job.Status, nil
}

// GetQueueStats gets statistics for a queue
func (q *RedisQueue) GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error) {
	// Get queue length
	queueKey := fmt.Sprintf("queue:%s", queueName)
	queueLen, err := q.client.LLen(ctx, queueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue length: %w", err)
	}

	// Get processing count
	processingKey := fmt.Sprintf("processing:%s", queueName)
	processingCount, err := q.client.SCard(ctx, processingKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing count: %w", err)
	}

	// Get completed count
	completedKey := fmt.Sprintf("completed:%s", queueName)
	completedCount, err := q.client.SCard(ctx, completedKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}

	// Get failed count
	failedKey := fmt.Sprintf("failed:%s", queueName)
	failedCount, err := q.client.SCard(ctx, failedKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	return map[string]int64{
		"pending":    queueLen,
		"processing": processingCount,
		"completed":  completedCount,
		"failed":     failedCount,
	}, nil
}
