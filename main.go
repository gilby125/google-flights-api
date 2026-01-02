package main

import (
	"context"
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
		"port", cfg.Port,
		"api_enabled", cfg.APIEnabled,
		"http_bind_addr", cfg.HTTPBindAddr,
		"worker_enabled", cfg.WorkerEnabled)

	logger.Info("Configuration loaded successfully")

	// Initialize database connections
	logger.Info("Connecting to databases...")
	postgresDB, err := db.NewPostgresDB(cfg.PostgresConfig)
	if err != nil {
		logger.Fatal(err, "Failed to connect to PostgreSQL", "host", cfg.PostgresConfig.Host)
	}
	defer postgresDB.Close()

	neo4jDB, err := db.NewNeo4jDB(cfg.Neo4jConfig)
	if err != nil {
		logger.Fatal(err, "Failed to connect to Neo4j", "uri", cfg.Neo4jConfig.URI)
	}
	defer neo4jDB.Close()

	// Initialize database schemas
	if cfg.InitSchema {
		logger.Info("Initializing database schemas...")
		if err := postgresDB.InitSchema(); err != nil {
			logger.Fatal(err, "Failed to initialize PostgreSQL schema")
		}

		if err := neo4jDB.InitSchema(); err != nil {
			logger.Fatal(err, "Failed to initialize Neo4j schema")
		}
	} else {
		logger.Info("Skipping schema initialization", "init_schema", cfg.InitSchema)
	}

	// Seed Neo4j with data from PostgreSQL
	if cfg.SeedNeo4j {
		logger.Info("Seeding Neo4j database...")
		if err := neo4jDB.SeedNeo4jData(context.Background(), &postgresDB); err != nil {
			logger.Fatal(err, "Failed to seed Neo4j database")
		}
	} else {
		logger.Info("Skipping Neo4j seeding", "seed_neo4j", cfg.SeedNeo4j)
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
	workerManager := worker.NewManager(redisQueue, redisClient, postgresDB, neo4jDB, cfg.WorkerConfig, cfg.FlightConfig)

	// Start worker pool if enabled
	if cfg.WorkerEnabled {
		logger.Info("Starting worker pool", "concurrency", cfg.WorkerConfig.Concurrency)
		workerManager.Start()
		defer workerManager.Stop()
	} else {
		logger.Info("Worker pool disabled")
	}

	var srv *http.Server
	if cfg.APIEnabled {
		// Initialize API router
		router := gin.New() // Use gin.New() instead of gin.Default() to have full control over middleware
		router.LoadHTMLGlob("templates/*html")

		// Register all API routes from the api package
		api.RegisterRoutes(router, postgresDB, neo4jDB, redisQueue, workerManager, cfg)

		addr := ":" + cfg.Port
		if cfg.HTTPBindAddr != "" {
			addr = cfg.HTTPBindAddr + ":" + cfg.Port
		}

		// Start HTTP server
		srv = &http.Server{
			Addr:    addr,
			Handler: router,
		}

		// Start server in a goroutine
		go func() {
			logger.Info("HTTP server starting", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal(err, "Failed to start HTTP server")
			}
		}()
	} else {
		logger.Info("API server disabled", "api_enabled", cfg.APIEnabled)
	}

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal(err, "Server forced to shutdown")
		}
	}

	logger.Info("Process exited gracefully")
}
