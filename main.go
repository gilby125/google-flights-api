package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

func main() {
	// Log that the main function has been entered
	log.Println("Entering main function")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log that the configuration has been loaded successfully
	log.Println("Configuration loaded successfully")

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
	log.Println("Entering InitSchema function")
	if err := postgresDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize PostgreSQL schema: %v", err)
	}

	if err := neo4jDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize Neo4j schema: %v", err)
	}

	// Seed database with initial data
	log.Println("Seeding database with initial data...")
	// if err := postgresDB.SeedData(); err != nil { // SeedData is on *PostgresDBImpl, not the interface. Commenting out for build.
	// 	log.Fatalf("Failed to seed database: %v", err)
	// }

	// Seed Neo4j with data from PostgreSQL
	log.Println("Seeding Neo4j database...")
	if err := neo4jDB.SeedNeo4jData(context.Background(), &postgresDB); err != nil { // Pass address of interface
		log.Fatalf("Failed to seed Neo4j database: %v", err)
	}

	// Initialize queue
	queue, err := queue.NewRedisQueue(cfg.RedisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize worker manager
	workerManager := worker.NewManager(queue, postgresDB, neo4jDB, cfg.WorkerConfig)

	// Log the value of cfg.WorkerEnabled
	log.Printf("cfg.WorkerEnabled: %v", cfg.WorkerEnabled)

	// Start worker pool if enabled
	if cfg.WorkerEnabled {
		workerManager.Start()
		defer workerManager.Stop()
	}

	// In the main function, update how routes are registered
	// Initialize API router
	router := gin.Default()
	router.LoadHTMLGlob("templates/*html")

	// Register API routes from the api package
	api.RegisterRoutes(router, postgresDB, neo4jDB, queue, workerManager, cfg)

	// Also register our custom routes to ensure the search functionality works
	setupRoutes(router)

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

// Make sure the search API endpoint is properly defined
// Update the setupRoutes function to properly handle the search API
func setupRoutes(r *gin.Engine) {
	// We'll keep this function but make sure it doesn't conflict with api.RegisterRoutes

	// Add CORS middleware to allow requests from the frontend
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Log all requests for debugging
	r.Use(func(c *gin.Context) {
		log.Printf("Request received: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Serve the search page with proper content type
	r.GET("/search-page/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/search/index.html")
	})

	// Add a specific handler for the search API
	apiGroup := r.Group("/api")
	{
		// Register our handlers
		apiGroup.POST("/search", searchFlightsHandler)
		apiGroup.GET("/airports", airportsHandler)
		apiGroup.GET("/price-history", priceHistoryHandler)

		// Add a GET version of the search endpoint for testing
		apiGroup.GET("/search-test", func(c *gin.Context) {
			log.Println("GET /api/search-test endpoint hit")
			c.JSON(http.StatusOK, gin.H{"message": "Search test endpoint working"})
		})
	}
}

// Update the searchFlightsHandler to ensure it properly processes the request
func searchFlightsHandler(c *gin.Context) {
	log.Println("Search flights handler called")

	// Try to bind JSON first
	var searchRequest struct {
		Origin        string `json:"origin" form:"origin"`
		Destination   string `json:"destination" form:"destination"`
		DepartureDate string `json:"departure_date" form:"departure_date"`
		ReturnDate    string `json:"return_date" form:"return_date"`
		TripType      string `json:"trip_type" form:"trip_type"`
		Class         string `json:"class" form:"class"`
		Stops         string `json:"stops" form:"stops"`
		Adults        int    `json:"adults" form:"adults"`
		Children      int    `json:"children" form:"children"`
		InfantsLap    int    `json:"infants_lap" form:"infants_lap"`
		InfantsSeat   int    `json:"infants_seat" form:"infants_seat"`
		Currency      string `json:"currency" form:"currency"`
	}

	// Try to bind JSON first, then form data if that fails
	if err := c.ShouldBindJSON(&searchRequest); err != nil {
		log.Printf("JSON binding failed, trying form binding: %v", err)
		if err := c.ShouldBind(&searchRequest); err != nil {
			log.Printf("Form binding also failed: %v", err)
			// Just log the error but continue processing with whatever data we got
		}
	}

	// Log the search request for debugging
	log.Printf("Search request received: %+v", searchRequest)

	// Set defaults for optional fields
	if searchRequest.Adults <= 0 {
		searchRequest.Adults = 1
	}

	if searchRequest.Currency == "" {
		searchRequest.Currency = "USD"
	}

	if searchRequest.Class == "" {
		searchRequest.Class = "economy"
	}

	if searchRequest.Stops == "" {
		searchRequest.Stops = "any"
	}

	if searchRequest.TripType == "" {
		searchRequest.TripType = "one_way"
	}

	// Validate required fields
	if searchRequest.Origin == "" || searchRequest.Destination == "" || searchRequest.DepartureDate == "" {
		log.Printf("Missing required fields: origin=%s, destination=%s, departure_date=%s",
			searchRequest.Origin, searchRequest.Destination, searchRequest.DepartureDate)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Origin, destination, and departure date are required"})
		return
	}

	// Parse dates
	departureDate, err := time.Parse("2006-01-02", searchRequest.DepartureDate)
	if err != nil {
		log.Printf("Invalid departure date format: %s, error: %v", searchRequest.DepartureDate, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid departure date format. Use YYYY-MM-DD"})
		return
	}

	var returnDate time.Time
	if searchRequest.ReturnDate != "" {
		returnDate, err = time.Parse("2006-01-02", searchRequest.ReturnDate)
		if err != nil {
			log.Printf("Invalid return date format: %s, error: %v", searchRequest.ReturnDate, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return date format. Use YYYY-MM-DD"})
			return
		}
	} else if searchRequest.TripType == "round_trip" {
		// Default return date for round trips (7 days after departure)
		returnDate = departureDate.AddDate(0, 0, 7)
	} else {
		// For one-way trips, set returnDate to a future date to pass validation
		// This won't be used in the actual API request for one-way trips
		returnDate = departureDate.AddDate(0, 0, 1)
	}

	// Create a new flight session
	session, err := flights.New()
	if err != nil {
		log.Printf("Error creating flight session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize flight search"})
		return
	}

	// Map trip type
	var tripType flights.TripType
	switch searchRequest.TripType {
	case "one_way":
		tripType = flights.OneWay
	case "round_trip", "":
		tripType = flights.RoundTrip
	default:
		tripType = flights.RoundTrip
	}

	// Map class
	var class flights.Class
	switch searchRequest.Class {
	case "economy", "":
		class = flights.Economy
	case "premium_economy":
		class = flights.PremiumEconomy
	case "business":
		class = flights.Business
	case "first":
		class = flights.First
	default:
		class = flights.Economy
	}

	// Map stops
	var stops flights.Stops
	switch searchRequest.Stops {
	case "nonstop":
		stops = flights.Nonstop
	case "one_stop":
		stops = flights.Stop1
	case "two_stops":
		stops = flights.Stop2
	case "any", "":
		stops = flights.AnyStops
	default:
		stops = flights.AnyStops
	}

	// Parse currency
	cur, err := currency.ParseISO(searchRequest.Currency)
	if err != nil {
		log.Printf("Invalid currency %s, using USD", searchRequest.Currency)
		cur = currency.USD
	}

	log.Printf("Performing flight search with: origin=%s, destination=%s, departure=%s, return=%s, class=%s, stops=%s",
		searchRequest.Origin, searchRequest.Destination, departureDate, returnDate, searchRequest.Class, searchRequest.Stops)

	// Perform the actual flight search
	offers, priceRange, err := session.GetOffers(
		c.Request.Context(),
		flights.Args{
			Date:        departureDate,
			ReturnDate:  returnDate,
			SrcAirports: []string{searchRequest.Origin},
			DstAirports: []string{searchRequest.Destination},
			Options: flights.Options{
				Travelers: flights.Travelers{
					Adults:       searchRequest.Adults,
					Children:     searchRequest.Children,
					InfantOnLap:  searchRequest.InfantsLap,
					InfantInSeat: searchRequest.InfantsSeat,
				},
				Currency: cur,
				Stops:    stops,
				Class:    class,
				TripType: tripType,
				Lang:     language.English,
			},
		},
	)

	if err != nil {
		log.Printf("Error searching flights: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search flights: " + err.Error()})
		return
	}

	log.Printf("Search successful. Found %d offers", len(offers))

	// Convert offers to response format
	responseOffers := make([]map[string]interface{}, 0, len(offers))
	for i, offer := range offers {
		// Create segments from flights
		segments := make([]map[string]interface{}, 0, len(offer.Flight))
		for _, flight := range offer.Flight {
			segment := map[string]interface{}{
				"departure_airport": flight.DepAirportCode,
				"arrival_airport":   flight.ArrAirportCode,
				"departure_time":    flight.DepTime.Format(time.RFC3339),
				"arrival_time":      flight.ArrTime.Format(time.RFC3339),
				"airline":           flight.AirlineName,
				"flight_number":     flight.FlightNumber,
				"duration":          int(flight.Duration.Minutes()),
				"airplane":          flight.Airplane,
				"legroom":           flight.Legroom,
			}
			segments = append(segments, segment)
		}

		// Generate Google Flights URL using the backend SerializeURL function
		googleFlightsUrl, err := session.SerializeURL(
			c.Request.Context(),
			flights.Args{
				Date:        offer.StartDate,
				ReturnDate:  offer.ReturnDate,
				SrcAirports: []string{searchRequest.Origin},
				DstAirports: []string{searchRequest.Destination},
				Options: flights.Options{
					Travelers: flights.Travelers{
						Adults:       searchRequest.Adults,
						Children:     searchRequest.Children,
						InfantOnLap:  searchRequest.InfantsLap,
						InfantInSeat: searchRequest.InfantsSeat,
					},
					Currency: cur,
					Stops:    stops,
					Class:    class,
					TripType: tripType,
					Lang:     language.English,
				},
			},
		)
		if err != nil {
			log.Printf("Error generating Google Flights URL: %v", err)
			// If there's an error, we'll continue without the URL
			googleFlightsUrl = ""
		}

		responseOffer := map[string]interface{}{
			"id":                 fmt.Sprintf("offer%d", i+1),
			"price":              offer.Price,
			"currency":           searchRequest.Currency,
			"total_duration":     int(offer.FlightDuration.Minutes()),
			"segments":           segments,
			"departure_date":     offer.StartDate.Format("2006-01-02"),
			"return_date":        offer.ReturnDate.Format("2006-01-02"),
			"google_flights_url": googleFlightsUrl,
		}

		responseOffers = append(responseOffers, responseOffer)
	}

	// Build the response
	response := map[string]interface{}{
		"offers":        responseOffers,
		"search_params": searchRequest,
	}

	// Add price range if available
	if priceRange != nil {
		response["price_range"] = map[string]interface{}{
			"low":  priceRange.Low,
			"high": priceRange.High,
		}
	}

	c.JSON(http.StatusOK, response)
}

func airportsHandler(c *gin.Context) {
	// Mock airport data for now
	airports := []map[string]string{
		{"code": "JFK", "name": "John F. Kennedy International Airport", "city": "New York"},
		{"code": "LAX", "name": "Los Angeles International Airport", "city": "Los Angeles"},
		{"code": "ORD", "name": "O'Hare International Airport", "city": "Chicago"},
		{"code": "LHR", "name": "Heathrow Airport", "city": "London"},
		{"code": "CDG", "name": "Charles de Gaulle Airport", "city": "Paris"},
	}

	c.JSON(http.StatusOK, airports)
}

func priceHistoryHandler(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")

	// Mock price history data
	priceHistory := []map[string]interface{}{
		{"date": "2023-06-01", "price": 299, "origin": origin, "destination": destination},
		{"date": "2023-06-02", "price": 310, "origin": origin, "destination": destination},
		{"date": "2023-06-03", "price": 305, "origin": origin, "destination": destination},
		{"date": "2023-06-04", "price": 295, "origin": origin, "destination": destination},
		{"date": "2023-06-05", "price": 320, "origin": origin, "destination": destination},
		{"date": "2023-06-06", "price": 315, "origin": origin, "destination": destination},
		{"date": "2023-06-07", "price": 300, "origin": origin, "destination": destination},
	}

	c.JSON(http.StatusOK, priceHistory)
}
