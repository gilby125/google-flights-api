package health

import (
	"context"
	"fmt"
	"time"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/redis/go-redis/v9"
)

// Status represents the health status of a component
type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

// Check represents a single health check
type Check struct {
	Name      string            `json:"name"`
	Status    Status            `json:"status"`
	Message   string            `json:"message,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Timestamp time.Time         `json:"timestamp"`
}

// HealthReport represents the overall health of the application
type HealthReport struct {
	Status    Status           `json:"status"`
	Version   string           `json:"version"`
	Timestamp time.Time        `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
	Uptime    time.Duration    `json:"uptime"`
}

// Checker defines the interface for health checks
type Checker interface {
	Check(ctx context.Context) Check
}

// PostgresChecker checks PostgreSQL connectivity
type PostgresChecker struct {
	DB   db.PostgresDB
	Name string
}

func (c *PostgresChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      c.Name,
		Timestamp: start,
		Details:   make(map[string]string),
	}

	// Test basic connectivity with a cheap query.
	rows, err := c.DB.Search(ctx, "SELECT 1")
	if rows != nil {
		_ = rows.Close()
	}
	duration := time.Since(start)
	check.Duration = duration

	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Database connection failed: %v", err)
		check.Details["error"] = err.Error()
	} else {
		check.Status = StatusUp
		check.Message = "Database connection successful"
		check.Details["response_time"] = duration.String()
	}

	return check
}

// Neo4jChecker checks Neo4j connectivity
type Neo4jChecker struct {
	DB   *db.Neo4jDB
	Name string
}

func (c *Neo4jChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      c.Name,
		Timestamp: start,
		Details:   make(map[string]string),
	}

	// Test Neo4j connectivity with a simple query
	_, err := c.DB.ExecuteReadQuery(ctx, "RETURN 1 as test", nil)
	duration := time.Since(start)
	check.Duration = duration

	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Neo4j connection failed: %v", err)
		check.Details["error"] = err.Error()
	} else {
		check.Status = StatusUp
		check.Message = "Neo4j connection successful"
		check.Details["response_time"] = duration.String()
	}

	return check
}

// RedisChecker checks Redis connectivity
type RedisChecker struct {
	Client *redis.Client
	Name   string
}

func (c *RedisChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      c.Name,
		Timestamp: start,
		Details:   make(map[string]string),
	}

	// Test Redis connectivity with ping
	pong, err := c.Client.Ping(ctx).Result()
	duration := time.Since(start)
	check.Duration = duration

	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Redis connection failed: %v", err)
		check.Details["error"] = err.Error()
	} else {
		check.Status = StatusUp
		check.Message = "Redis connection successful"
		check.Details["response_time"] = duration.String()
		check.Details["ping_response"] = pong
	}

	return check
}

// QueueChecker checks queue connectivity and status
type QueueChecker struct {
	Queue queue.Queue
	Name  string
}

func (c *QueueChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      c.Name,
		Timestamp: start,
		Details:   make(map[string]string),
	}

	// Test queue connectivity by getting queue stats for a common queue
	stats, err := c.Queue.GetQueueStats(ctx, "flight_search")
	duration := time.Since(start)
	check.Duration = duration

	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Queue connectivity check failed: %v", err)
		check.Details["error"] = err.Error()
	} else {
		check.Status = StatusUp
		check.Message = "Queue is operational"
		check.Details["response_time"] = duration.String()
		if pending, ok := stats["pending"]; ok {
			check.Details["pending_jobs"] = fmt.Sprintf("%d", pending)
		}
		if processing, ok := stats["processing"]; ok {
			check.Details["processing_jobs"] = fmt.Sprintf("%d", processing)
		}
	}

	return check
}

// WorkerChecker checks worker manager status (basic check)
type WorkerChecker struct {
	Manager *worker.Manager
	Name    string
}

func (c *WorkerChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      c.Name,
		Timestamp: start,
		Details:   make(map[string]string),
	}

	// Basic check - if manager exists, workers are considered operational
	duration := time.Since(start)
	check.Duration = duration

	if c.Manager == nil {
		check.Status = StatusDown
		check.Message = "Worker manager not initialized"
	} else {
		check.Status = StatusUp
		check.Message = "Worker manager is operational"
	}

	check.Details["response_time"] = duration.String()

	return check
}

// HealthChecker orchestrates multiple health checks
type HealthChecker struct {
	checkers  []Checker
	version   string
	startTime time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		checkers:  make([]Checker, 0),
		version:   version,
		startTime: time.Now(),
	}
}

// AddChecker adds a health checker
func (h *HealthChecker) AddChecker(checker Checker) {
	h.checkers = append(h.checkers, checker)
}

// CheckHealth performs all health checks
func (h *HealthChecker) CheckHealth(ctx context.Context) HealthReport {
	checks := make(map[string]Check)
	overallStatus := StatusUp

	// Run all checks
	for _, checker := range h.checkers {
		check := checker.Check(ctx)
		checks[check.Name] = check

		// If any check fails, overall status is down
		if check.Status == StatusDown {
			overallStatus = StatusDown
		}
	}

	return HealthReport{
		Status:    overallStatus,
		Version:   h.version,
		Timestamp: time.Now(),
		Checks:    checks,
		Uptime:    time.Since(h.startTime),
	}
}

// CheckReadiness performs readiness checks (subset of health checks)
func (h *HealthChecker) CheckReadiness(ctx context.Context) HealthReport {
	// For readiness, we only check essential components
	readinessCheckers := make([]Checker, 0)

	// Add only critical checkers for readiness
	for _, checker := range h.checkers {
		switch checker.(type) {
		case *PostgresChecker, *RedisChecker:
			// Only include database checks for readiness
			readinessCheckers = append(readinessCheckers, checker)
		}
	}

	checks := make(map[string]Check)
	overallStatus := StatusUp

	// Run readiness checks
	for _, checker := range readinessCheckers {
		check := checker.Check(ctx)
		checks[check.Name] = check

		if check.Status == StatusDown {
			overallStatus = StatusDown
		}
	}

	return HealthReport{
		Status:    overallStatus,
		Version:   h.version,
		Timestamp: time.Now(),
		Checks:    checks,
		Uptime:    time.Since(h.startTime),
	}
}

// CheckLiveness performs liveness checks (basic application health)
func (h *HealthChecker) CheckLiveness(ctx context.Context) HealthReport {
	// Liveness is just a basic "is the application running" check
	return HealthReport{
		Status:    StatusUp,
		Version:   h.version,
		Timestamp: time.Now(),
		Checks: map[string]Check{
			"application": {
				Name:      "application",
				Status:    StatusUp,
				Message:   "Application is running",
				Timestamp: time.Now(),
				Duration:  0,
			},
		},
		Uptime: time.Since(h.startTime),
	}
}
