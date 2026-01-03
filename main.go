package main

import (
	"context"
	"database/sql"
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
	"github.com/redis/go-redis/v9"
)

func main() {
	// Handle health check flag before loading full config/logger
	for _, arg := range os.Args[1:] {
		if arg == "-health-check" {
			// Container orchestrators should check readiness (core deps) rather than full health.
			// Full /health includes optional dependencies like Neo4j and can flap the container.
			resp, err := http.Get("http://localhost:8080/health/ready")
			if err != nil || resp.StatusCode != http.StatusOK {
				os.Exit(1)
			}
			os.Exit(0)
		}

		if arg == "-health-check-worker" {
			cfg, err := config.Load()
			if err != nil {
				os.Exit(1)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Postgres auth + connectivity check
			connStr := fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
				cfg.PostgresConfig.Host, cfg.PostgresConfig.Port, cfg.PostgresConfig.User, cfg.PostgresConfig.Password,
				cfg.PostgresConfig.DBName, cfg.PostgresConfig.SSLMode, cfg.PostgresConfig.SSLCert, cfg.PostgresConfig.SSLKey, cfg.PostgresConfig.SSLRootCert)

			postgresDB, err := sql.Open("postgres", connStr)
			if err != nil {
				os.Exit(1)
			}
			defer postgresDB.Close()

			if err := postgresDB.PingContext(ctx); err != nil {
				os.Exit(1)
			}

			// Redis auth + connectivity check
			redisClient := redis.NewClient(&redis.Options{
				Addr:     cfg.RedisConfig.Host + ":" + cfg.RedisConfig.Port,
				Password: cfg.RedisConfig.Password,
				DB:       cfg.RedisConfig.DB,
			})
			defer redisClient.Close()

			if _, err := redisClient.Ping(ctx).Result(); err != nil {
				os.Exit(1)
			}

			// Optional Neo4j connectivity check
			if cfg.Neo4jEnabled {
				neo4jDB, err := db.NewNeo4jDB(cfg.Neo4jConfig)
				if err != nil {
					os.Exit(1)
				}
				_ = neo4jDB.Close()
			}

			os.Exit(0)
		}
	}

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
		"worker_enabled", cfg.WorkerEnabled,
		"neo4j_enabled", cfg.Neo4jEnabled)

	logger.Info("Configuration loaded successfully")

	if cfg.Environment == "production" && cfg.InitSchema {
		logger.Info("INIT_SCHEMA enabled in production (safe/idempotent)", "init_schema", cfg.InitSchema)
	}

	// Initialize database connections with retries
	var postgresDB db.PostgresDB
	var neo4jDB *db.Neo4jDB
	var redisQueue *queue.RedisQueue

	maxRetries := 10
	retryDelay := 5 * time.Second

	logger.Info("Connecting to databases (with retries)...")

	for i := 0; i < maxRetries; i++ {
		var pErr, nErr, rErr error

		if postgresDB == nil {
			postgresDB, pErr = db.NewPostgresDB(cfg.PostgresConfig)
			if pErr != nil {
				logger.Warn("Failed to connect to PostgreSQL, retrying...", "error", pErr, "attempt", i+1)
			}
		}

		if cfg.Neo4jEnabled && neo4jDB == nil {
			neo4jDB, nErr = db.NewNeo4jDB(cfg.Neo4jConfig)
			if nErr != nil {
				logger.Warn("Failed to connect to Neo4j, retrying...", "error", nErr, "attempt", i+1)
			}
		}

		if redisQueue == nil {
			redisQueue, rErr = queue.NewRedisQueue(cfg.RedisConfig)
			if rErr != nil {
				logger.Warn("Failed to connect to Redis, retrying...", "error", rErr, "attempt", i+1)
			}
		}

		if pErr == nil && rErr == nil && (!cfg.Neo4jEnabled || nErr == nil) {
			logger.Info("All database connections established successfully")
			break
		}

		if i == maxRetries-1 {
			logger.Fatal(fmt.Errorf("db connection timeout"), "All database connection attempts failed")
		}

		time.Sleep(retryDelay)
	}
	defer postgresDB.Close()
	if neo4jDB != nil {
		defer neo4jDB.Close()
	}

	// Initialize database schemas (migrations) and Neo4j constraints/indexes
	if cfg.InitSchema {
		logger.Info("Running database migrations...")
		if err := db.RunMigrations(db.BuildPostgresConnString(cfg.PostgresConfig)); err != nil {
			logger.Fatal(err, "Failed to run PostgreSQL migrations")
		}

		if neo4jDB != nil {
			if err := neo4jDB.InitSchema(); err != nil {
				logger.Fatal(err, "Failed to initialize Neo4j schema")
			}
		}
	} else {
		logger.Info("Skipping schema initialization", "init_schema", cfg.InitSchema)
	}

	// Seed Neo4j with data from PostgreSQL
	if cfg.SeedNeo4j {
		if neo4jDB == nil {
			logger.Fatal(fmt.Errorf("neo4j is disabled"), "SEED_NEO4J=true requires Neo4jEnabled=true")
		}
		logger.Info("Seeding Neo4j database...")
		if err := neo4jDB.SeedNeo4jData(context.Background(), postgresDB); err != nil {
			logger.Fatal(err, "Failed to seed Neo4j database")
		}
	} else {
		logger.Info("Skipping Neo4j seeding", "seed_neo4j", cfg.SeedNeo4j)
	}

	// Initialize queue
	// redisQueue already initialized in retry loop

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
