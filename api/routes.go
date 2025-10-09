package api

import (
	"net/http"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/pkg/cache"
	"github.com/gilby125/google-flights-api/pkg/health"
	"github.com/gilby125/google-flights-api/pkg/middleware"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, postgresDB db.PostgresDB, neo4jDB *db.Neo4jDB, queue queue.Queue, workerManager *worker.Manager, cfg *config.Config) {
	// Initialize cache manager
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.Host + ":" + cfg.RedisConfig.Port,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})

	redisCache := cache.NewRedisCache(redisClient, "flights_api")
	cacheManager := cache.NewCacheManager(redisCache)

	// Initialize health checker
	healthChecker := health.NewHealthChecker("1.0.0")
	healthChecker.AddChecker(&health.PostgresChecker{DB: postgresDB, Name: "postgres"})
	healthChecker.AddChecker(&health.Neo4jChecker{DB: neo4jDB, Name: "neo4j"})
	healthChecker.AddChecker(&health.RedisChecker{Client: redisClient, Name: "redis"})
	healthChecker.AddChecker(&health.QueueChecker{Queue: queue, Name: "queue"})
	healthChecker.AddChecker(&health.WorkerChecker{Manager: workerManager, Name: "workers"})

	// Setup middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		report := healthChecker.CheckHealth(c.Request.Context())
		status := http.StatusOK
		if report.Status == health.StatusDown {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, report)
	})

	router.GET("/health/ready", func(c *gin.Context) {
		report := healthChecker.CheckReadiness(c.Request.Context())
		status := http.StatusOK
		if report.Status == health.StatusDown {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, report)
	})

	router.GET("/health/live", func(c *gin.Context) {
		report := healthChecker.CheckLiveness(c.Request.Context())
		c.JSON(http.StatusOK, report)
	})

	// Legacy API routes (for backward compatibility)
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.ResponseCache(cacheManager, middleware.CacheConfig{
		TTL:         cache.MediumTTL,
		KeyPrefix:   "http_cache",
		SkipPaths:   []string{"/api/search"},
		OnlyMethods: []string{"GET"},
	}))
	{
		// Direct flight search (immediate results, bypasses queue)
		apiGroup.POST("/search", DirectFlightSearch())
		apiGroup.GET("/search-test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Search test endpoint working"})
		})

		// Cached endpoints for development
		apiGroup.GET("/airports", CachedAirportsHandler(cacheManager))
		apiGroup.GET("/price-history", MockPriceHistoryHandler())
	}

	// API v1 routes (production endpoints)
	v1 := router.Group("/api/v1")
	v1.Use(middleware.ResponseCache(cacheManager, middleware.CacheConfig{
		TTL:         cache.MediumTTL,
		KeyPrefix:   "http_cache",
		SkipPaths:   []string{"/api/v1/admin"},
		OnlyMethods: []string{"GET"},
	}))
	{
		// Airport routes
		v1.GET("/airports", GetAirports(postgresDB))

		// Airline routes
		v1.GET("/airlines", GetAirlines(postgresDB))

		// Flight search routes (queued searches)
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
			admin.POST("/bulk-jobs", createBulkJob(postgresDB, workerManager))
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

	// Web UI routes
	router.GET("/search-page/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/search/index.html")
	})

	// Search page route - serve index.html directly
	router.GET("/search", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/search/index.html")
	})

	// Bulk search results page
	router.GET("/bulk-search", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/bulk-search/index.html")
	})

	// Serve static files for the web UI
	router.Static("/admin", "./web/admin")
	router.Static("/search", "./web/search")
	router.Static("/bulk-search", "./web/bulk-search")
	router.StaticFile("/", "./web/index.html")
}
