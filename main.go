package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connections
	postgresDB, err := db.NewPostgresDB(cfg.PostgresConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	neo4jDB, err := db.NewNeo4jDB(cfg.Neo4jConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4jDB.Close()

	// Initialize database schemas
	log.Println("Initializing database schemas...")
	if err := postgresDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL schema: %v", err)
	}

	if err := neo4jDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize Neo4j schema: %v", err)
	}

	// Initialize queue
	queue, err := queue.NewRedisQueue(cfg.RedisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize worker manager
	workerManager := worker.NewManager(queue, postgresDB, neo4jDB, cfg.WorkerConfig)

	// Start worker pool if enabled
	if cfg.WorkerEnabled {
		workerManager.Start()
		defer workerManager.Stop()
	}

	// Initialize API router
	router := gin.Default()
	api.RegisterRoutes(router, postgresDB, neo4jDB, queue, workerManager, cfg)

	// Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
