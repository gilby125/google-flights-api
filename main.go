package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/pkg/logger"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
)

// connectWithRetry attempts to connect with exponential backoff retry logic
func connectWithRetry[T any](connectFunc func() (T, error), serviceName, address string, maxRetries int, baseDelay time.Duration) (T, error) {
	var zero T
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := connectFunc()
		if err == nil {
			logger.Info(fmt.Sprintf("Successfully connected to %s", serviceName),
				"service", serviceName,
				"address", address,
				"attempt", attempt)
			return result, nil
		}

		if attempt == maxRetries {
			return zero, fmt.Errorf("failed to connect to %s after %d attempts: %w", serviceName, maxRetries, err)
		}

		delay := baseDelay * time.Duration(attempt) // Exponential backoff
		logger.Warn(fmt.Sprintf("Failed to connect to %s, retrying in %v", serviceName, delay),
			"service", serviceName,
			"address", address,
			"attempt", attempt,
			"maxRetries", maxRetries,
			"error", err)

		time.Sleep(delay)
	}
	return zero, fmt.Errorf("unexpected error in connectWithRetry")
}

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		panic(err) // Can't use logger yet
	}

	// Initialize structured logging
	logger.Init(logger.Config{
		Level:  cfg.LoggingConfig.Level,
		Format: cfg.LoggingConfig.Format,
	})

	logger.Info("Starting Google Flights API server",
		"version", "1.0.0",
		"environment", cfg.Environment,
		"port", cfg.Port)

	logger.Info("Configuration loaded successfully")

	// Initialize database connections with retry logic
	logger.Info("Connecting to databases...")
	postgresDB, err := connectWithRetry(func() (db.PostgresDB, error) {
		return db.NewPostgresDB(cfg.PostgresConfig)
	}, "PostgreSQL", cfg.PostgresConfig.Host, 10, 10*time.Second)
	if err != nil {
		logger.Fatal(err, "Failed to connect to PostgreSQL after retries", "host", cfg.PostgresConfig.Host)
	}
	defer postgresDB.Close()

	neo4jDB, err := connectWithRetry(func() (*db.Neo4jDB, error) {
		return db.NewNeo4jDB(cfg.Neo4jConfig)
	}, "Neo4j", cfg.Neo4jConfig.URI, 15, 10*time.Second)
	if err != nil {
		logger.Fatal(err, "Failed to connect to Neo4j after retries", "uri", cfg.Neo4jConfig.URI)
	}
	defer neo4jDB.Close()

	// Initialize database schemas
	logger.Info("Initializing database schemas...")
	if err := postgresDB.InitSchema(); err != nil {
		logger.Fatal(err, "Failed to initialize PostgreSQL schema")
	}

	if err := neo4jDB.InitSchema(); err != nil {
		logger.Fatal(err, "Failed to initialize Neo4j schema")
	}

	// Seed Neo4j with data from PostgreSQL
	logger.Info("Seeding Neo4j database...")
	if err := neo4jDB.SeedNeo4jData(context.Background(), &postgresDB); err != nil {
		logger.Fatal(err, "Failed to seed Neo4j database")
	}

	// Initialize queue
	logger.Info("Connecting to Redis queue...")
	redisQueue, err := queue.NewRedisQueue(cfg.RedisConfig)
	if err != nil {
		logger.Fatal(err, "Failed to connect to Redis", "host", cfg.RedisConfig.Host)
	}

	// Get Redis client for leader election
	redisClient := redisQueue.GetClient()

	// Initialize worker manager with Redis client for distributed leader election
	workerManager := worker.NewManager(redisQueue, redisClient, postgresDB, neo4jDB, cfg.WorkerConfig)

	// Start worker pool if enabled
	if cfg.WorkerEnabled {
		logger.Info("Starting worker pool", "concurrency", cfg.WorkerConfig.Concurrency)
		workerManager.Start()
		defer workerManager.Stop()
	} else {
		logger.Info("Worker pool disabled")
	}

	// Initialize API router
	router := gin.New() // Use gin.New() instead of gin.Default() to have full control over middleware
	router.LoadHTMLGlob("templates/*html")

	// Register all API routes from the api package
	api.RegisterRoutes(router, postgresDB, neo4jDB, redisQueue, workerManager, cfg)

	// Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("HTTP server starting", "port", cfg.Port, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err, "Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(err, "Server forced to shutdown")
	}

	logger.Info("Server exited gracefully")
}
