package api

import (
	"net/http"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/pkg/cache"
	"github.com/gilby125/google-flights-api/pkg/health"
	"github.com/gilby125/google-flights-api/pkg/macros"
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

		// Macro/metadata endpoints (region and airline group tokens)
		v1.GET("/regions", func(c *gin.Context) {
			c.JSON(http.StatusOK, macros.GetAllRegionInfo())
		})
		v1.GET("/airline-groups", func(c *gin.Context) {
			c.JSON(http.StatusOK, macros.GetAllAirlineGroupInfo())
		})

		// Flight search routes (queued searches)
		v1.POST("/search", CreateSearch(queue))
		v1.GET("/search/:id", GetSearchByID(postgresDB))
		v1.GET("/search", ListSearches(postgresDB))

		// Bulk search routes
		v1.POST("/bulk-search", CreateBulkSearch(queue, postgresDB))
		v1.GET("/bulk-search/:id", getBulkSearchById(postgresDB))

		// Price history routes
		v1.GET("/price-history/:origin/:destination", getPriceHistory(neo4jDB))

		// Admin routes (with optional authentication)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuth(cfg.AdminAuthConfig))
		{
			// Job routes
			admin.GET("/jobs", listJobs(postgresDB))
			admin.POST("/jobs", createJob(postgresDB, workerManager))
			admin.GET("/bulk-jobs", listBulkSearches(postgresDB))
			admin.GET("/bulk-jobs/:id", getBulkSearchResults(postgresDB))
			admin.POST("/bulk-jobs", createBulkJob(postgresDB, workerManager))
			admin.GET("/bulk-jobs/:id/offers", getBulkSearchOffers(postgresDB))
			admin.POST("/price-graph-sweeps", enqueuePriceGraphSweep(postgresDB, workerManager))
			admin.GET("/price-graph-sweeps", listPriceGraphSweeps(postgresDB))
			admin.GET("/price-graph-sweeps/:id", getPriceGraphSweepResults(postgresDB))
			admin.GET("/jobs/:id", getJobById(postgresDB))
			admin.PUT("/jobs/:id", updateJob(postgresDB, workerManager))
			admin.DELETE("/jobs/:id", DeleteJob(postgresDB, workerManager))

			// Job actions
			admin.POST("/jobs/:id/run", runJob(queue, postgresDB))
			admin.POST("/jobs/:id/enable", enableJob(postgresDB, workerManager))
			admin.POST("/jobs/:id/disable", disableJob(postgresDB, workerManager))

			// Worker and queue status
			admin.GET("/workers", GetWorkerStatus(workerManager, redisClient, cfg.WorkerConfig))
			admin.GET("/queue", GetQueueStatus(queue))

			// Real-time events via Server-Sent Events
			admin.GET("/events", GetAdminEvents(workerManager, redisClient, cfg.WorkerConfig))

			// Continuous sweep endpoints
			admin.GET("/continuous-sweep/status", getContinuousSweepStatus(workerManager))
			admin.POST("/continuous-sweep/start", startContinuousSweep(workerManager, postgresDB, cfg))
			admin.POST("/continuous-sweep/stop", stopContinuousSweep(workerManager))
			admin.POST("/continuous-sweep/pause", pauseContinuousSweep(workerManager))
			admin.POST("/continuous-sweep/resume", resumeContinuousSweep(workerManager))
			admin.PUT("/continuous-sweep/config", updateContinuousSweepConfig(workerManager))
			admin.POST("/continuous-sweep/skip", skipCurrentRoute(workerManager))
			admin.POST("/continuous-sweep/restart", restartCurrentSweep(workerManager))
			admin.GET("/continuous-sweep/stats", getContinuousSweepStats(postgresDB))
			admin.GET("/continuous-sweep/results", getContinuousSweepResults(postgresDB))
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
	router.StaticFile("/search.js", "./web/search/search.js")
	router.StaticFile("/", "./web/index.html")
}
