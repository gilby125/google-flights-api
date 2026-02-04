package api

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/hotels"
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
func RegisterRoutes(router *gin.Engine, postgresDB db.PostgresDB, neo4jDB *db.Neo4jDB, queue queue.Queue, workerManager *worker.Manager, cfg *config.Config, hotelSession *hotels.Session) {
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
	router.Use(middleware.RequestID())
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
		apiGroup.POST("/search", DirectFlightSearch(postgresDB, neo4jDB))
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
		v1.GET("/airports/top", GetTopAirports())

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

		// Hotel routes
		hotelsGroup := v1.Group("/hotels")
		{
			hotelsGroup.POST("/search", DirectHotelSearch(hotelSession))
		}

		// Bulk search routes
		v1.POST("/bulk-search", CreateBulkSearch(queue, postgresDB, workerManager))
		v1.GET("/bulk-search/:id", getBulkSearchById(postgresDB))

		// Price history routes
		v1.GET("/price-history/:origin/:destination", getPriceHistory(neo4jDB))

		// Graph traversal routes (Neo4j-powered)
		graph := v1.Group("/graph")
		{
			graph.GET("/path", GetCheapestPath(neo4jDB))
			graph.GET("/connections", GetConnections(neo4jDB))
			graph.GET("/route-stats", GetRouteStats(neo4jDB))
			graph.GET("/route-details", GetRouteDetails(neo4jDB, cfg.FlightConfig.ExcludedAirlines))
			graph.GET("/explore", GetExplore(neo4jDB))
		}

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
			admin.GET("/queue/:name/backlog", GetQueueBacklog(queue))
			admin.GET("/queue/:name/jobs", ListQueueJobs(queue))
			admin.GET("/queue/:name/jobs/:id", GetQueueJob(queue))
			admin.POST("/queue/:name/jobs/:id/cancel", CancelQueueJob(queue))
			admin.GET("/queue/:name/enqueues", GetQueueEnqueueMetrics(queue))
			admin.POST("/queue/:name/cancel-processing", CancelQueueProcessing(queue))
			admin.POST("/queue/:name/drain", DrainQueue(queue))
			admin.POST("/queue/:name/clear", ClearQueue(queue))
			admin.POST("/queue/:name/clear-failed", ClearQueueFailed(queue))
			admin.POST("/queue/:name/clear-processing", ClearQueueProcessing(queue))
			admin.POST("/queue/:name/retry-failed", RetryQueueFailed(queue))

			// Real-time events via Server-Sent Events
			admin.GET("/events", GetAdminEvents(workerManager, redisClient, cfg.WorkerConfig))

			// Continuous sweep endpoints
			admin.GET("/continuous-sweep/status", getContinuousSweepStatus(workerManager, postgresDB))
			admin.POST("/continuous-sweep/start", startContinuousSweep(workerManager, postgresDB, cfg))
			admin.POST("/continuous-sweep/stop", stopContinuousSweep(workerManager, postgresDB))
			admin.POST("/continuous-sweep/pause", pauseContinuousSweep(workerManager, postgresDB))
			admin.POST("/continuous-sweep/resume", resumeContinuousSweep(workerManager, postgresDB))
			admin.PUT("/continuous-sweep/config", updateContinuousSweepConfig(workerManager))
			admin.POST("/continuous-sweep/skip", skipCurrentRoute(workerManager))
			admin.POST("/continuous-sweep/restart", restartCurrentSweep(workerManager))
			admin.GET("/continuous-sweep/stats", getContinuousSweepStats(postgresDB))
			admin.GET("/continuous-sweep/results", getContinuousSweepResults(postgresDB))

			// Deal detection endpoints
			admin.GET("/deals", listDeals(postgresDB))
			admin.GET("/deal-alerts", listDealAlerts(postgresDB))
		}
	}

	// Web UI routes
	router.GET("/search-page/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Type", "text/html")
		c.File("./web/search/index.html")
	})

	// Search page route - serve index.html directly
	router.GET("/search", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Type", "text/html")
		c.File("./web/search/index.html")
	})

	// Bulk search results page
	router.GET("/bulk-search", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/bulk-search/index.html")
	})

	// Explore page (map/globe) + assets.
	// NOTE: Use a custom handler to avoid Gin conflicts between exact and wildcard static routes.
	router.GET("/explore/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")

		p := c.Param("filepath")
		if p == "" || p == "/" {
			c.Header("Content-Type", "text/html")
			c.File("./web/explore/index.html")
			return
		}

		rel := strings.TrimPrefix(p, "/")
		rel = path.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, "..") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.File(filepath.Join("./web/explore", rel))
	})

	// Serve static files for the web UI
	// NOTE: Use a custom handler for admin assets so we can set Cache-Control headers,
	// and avoid Gin route conflicts between exact and wildcard static routes.
	router.GET("/admin/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")

		p := c.Param("filepath")
		if p == "" || p == "/" {
			c.File("./web/admin/index.html")
			return
		}

		rel := strings.TrimPrefix(p, "/")
		rel = path.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, "..") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.File(filepath.Join("./web/admin", rel))
	})

	router.GET("/search.js", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Type", "application/javascript")
		c.File("./web/search/search.js")
	})
	router.StaticFile("/", "./web/index.html")
}
