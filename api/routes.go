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
func RegisterRoutes(router *gin.Engine, postgresDB db.PostgresDB, neo4jDB *db.Neo4jDB, queue queue.Queue, workerManager *worker.Manager, cfg *config.Config) { // Changed postgresDB type
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
		v1.GET("/airports", GetAirports(postgresDB))

		// Airline routes
		v1.GET("/airlines", GetAirlines(postgresDB))

		// Flight search routes
		v1.POST("/search", CreateSearch(queue))
		v1.GET("/search/:id", GetSearchByID(postgresDB))
		v1.GET("/search", ListSearches(postgresDB))

		// Bulk search routes
		v1.POST("/bulk-search", CreateBulkSearch(queue))
		v1.GET("/bulk-search/:id", getBulkSearchById(postgresDB))

		// Price history routes
		v1.GET("/price-history/:origin/:destination", getPriceHistory(neo4jDB))

		// Admin routes
		admin := v1.Group("/admin")
		{
			// Job routes
			admin.GET("/jobs", listJobs(postgresDB))
			admin.POST("/jobs", createJob(postgresDB, workerManager))
			admin.GET("/jobs/:id", getJobById(postgresDB))
			admin.PUT("/jobs/:id", updateJob(postgresDB, workerManager))
			admin.DELETE("/jobs/:id", DeleteJob(postgresDB, workerManager))

			// Job actions
			admin.POST("/jobs/:id/run", runJob(queue, postgresDB))
			admin.POST("/jobs/:id/enable", enableJob(postgresDB, workerManager))
			admin.POST("/jobs/:id/disable", disableJob(postgresDB, workerManager))

			// Worker and queue status
			admin.GET("/workers", GetWorkerStatus(workerManager))
			admin.GET("/queue", GetQueueStatus(queue))
		}
	}

	// Serve static files for the web UI
	router.Static("/admin", "./web/admin")
	router.Static("/search", "./web/search")
	router.StaticFile("/", "./web/index.html")
}
