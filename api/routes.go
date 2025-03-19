package api

import (
	"net/http"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, postgresDB *db.PostgresDB, neo4jDB *db.Neo4jDB, queue queue.Queue, workerManager *worker.Manager, cfg *config.Config) {
	// Setup middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Airport routes
		v1.GET("/airports", getAirports(postgresDB))

		// Airline routes
		v1.GET("/airlines", getAirlines(postgresDB))

		// Flight search routes
		v1.POST("/search", createSearch(queue))
		v1.GET("/search/:id", getSearchById(postgresDB))
		v1.GET("/search", listSearches(postgresDB))

		// Bulk search routes
		v1.POST("/bulk-search", createBulkSearch(queue))
		v1.GET("/bulk-search/:id", getBulkSearchById(postgresDB))

		// Price history routes
		v1.GET("/price-history/:origin/:destination", getPriceHistory(neo4jDB))

		// Admin routes
		admin := v1.Group("/admin")
		{
			// Job routes
			admin.GET("/jobs", listJobs(postgresDB))
			admin.POST("/jobs", createJob(postgresDB))
			admin.GET("/jobs/:id", getJobById(postgresDB))
			admin.PUT("/jobs/:id", updateJob(postgresDB))
			admin.DELETE("/jobs/:id", deleteJob(postgresDB))

			// Job actions
			admin.POST("/jobs/:id/run", runJob(queue, postgresDB))
			admin.POST("/jobs/:id/enable", enableJob(postgresDB))
			admin.POST("/jobs/:id/disable", disableJob(postgresDB))

			// Worker and queue status
			admin.GET("/workers", getWorkerStatus(workerManager))
			admin.GET("/queue", getQueueStatus(queue))
		}
	}

	// Serve static files for the web UI
	router.Static("/admin", "./web/admin")
	router.Static("/search", "./web/search")
	router.StaticFile("/", "./web/index.html")
}
