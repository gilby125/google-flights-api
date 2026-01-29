package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/pkg/buildinfo"
	"github.com/gilby125/google-flights-api/pkg/cache"
	"github.com/gilby125/google-flights-api/pkg/logger"
	"github.com/gilby125/google-flights-api/pkg/macros"
	"github.com/gilby125/google-flights-api/pkg/middleware"
	"github.com/gilby125/google-flights-api/pkg/notify"
	"github.com/gilby125/google-flights-api/pkg/worker_registry"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

const dateLayout = "2006-01-02"

// DateOnly parses YYYY-MM-DD or RFC3339 timestamps into a date (time at midnight UTC)
type DateOnly struct {
	time.Time
}

// UnmarshalJSON accepts either YYYY-MM-DD or RFC3339 timestamps
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), "\"")
	if str == "" || str == "null" {
		d.Time = time.Time{}
		return nil
	}

	layouts := []string{dateLayout, time.RFC3339}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, str)
		if err == nil {
			d.Time = parsed
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("invalid date format %q: %w", str, lastErr)
}

// MarshalJSON renders dates as YYYY-MM-DD; zero values emit null
func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", d.Time.Format(dateLayout))), nil
}

// SearchRequest represents a flight search request
type SearchRequest struct {
	Origin        string   `json:"origin" binding:"required"`
	Destination   string   `json:"destination" binding:"required"`
	DepartureDate DateOnly `json:"departure_date" binding:"required"`
	ReturnDate    DateOnly `json:"return_date,omitempty"`
	Adults        int      `json:"adults" binding:"required,min=1"`
	Children      int      `json:"children" binding:"min=0"`
	InfantsLap    int      `json:"infants_lap" binding:"min=0"`
	InfantsSeat   int      `json:"infants_seat" binding:"min=0"`
	TripType      string   `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class         string   `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops         string   `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"` // Added two_stops_plus
	Currency      string   `json:"currency" binding:"required,len=3"`
}

// BulkSearchRequest represents a bulk flight search request
type BulkSearchRequest struct {
	Origins           []string `json:"origins" binding:"required,min=1"`
	Destinations      []string `json:"destinations" binding:"required,min=1"`
	DepartureDateFrom DateOnly `json:"departure_date_from" binding:"required"`
	DepartureDateTo   DateOnly `json:"departure_date_to" binding:"required"`
	ReturnDateFrom    DateOnly `json:"return_date_from,omitempty"`
	ReturnDateTo      DateOnly `json:"return_date_to,omitempty"`
	TripLength        int      `json:"trip_length,omitempty" binding:"min=0"`
	Adults            int      `json:"adults" binding:"required,min=1"`
	Children          int      `json:"children" binding:"min=0"`
	InfantsLap        int      `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int      `json:"infants_seat" binding:"min=0"`
	TripType          string   `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string   `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string   `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Currency          string   `json:"currency" binding:"required,len=3"`
	Carriers          []string `json:"carriers,omitempty"`
}

// JobRequest represents a scheduled job request
type JobRequest struct {
	Name              string `json:"name" binding:"required"`
	Origin            string `json:"origin" binding:"required"`
	Destination       string `json:"destination" binding:"required"`
	DateStart         string `json:"date_start" binding:"required"`
	DateEnd           string `json:"date_end" binding:"required"`
	ReturnDateStart   string `json:"return_date_start,omitempty"`
	ReturnDateEnd     string `json:"return_date_end,omitempty"`
	TripLength        int    `json:"trip_length,omitempty" binding:"min=0"`
	DynamicDates      bool   `json:"dynamic_dates,omitempty"`       // Use dates relative to execution time
	DaysFromExecution int    `json:"days_from_execution,omitempty"` // Start searching X days from now
	SearchWindowDays  int    `json:"search_window_days,omitempty"`  // Search within X days window
	Adults            int    `json:"adults" binding:"required,min=1"`
	Children          int    `json:"children" binding:"min=0"`
	InfantsLap        int    `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int    `json:"infants_seat" binding:"min=0"`
	TripType          string `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Currency          string `json:"currency" binding:"required,len=3"`
	Interval          string `json:"interval" binding:"required"`
	Time              string `json:"time" binding:"required"`
	CronExpression    string `json:"cron_expression" binding:"required"`
}

// BulkJobRequest represents a request to create a scheduled bulk search job
type BulkJobRequest struct {
	Name              string   `json:"name" binding:"required"`
	Origins           []string `json:"origins" binding:"required,min=1"`
	Destinations      []string `json:"destinations" binding:"required,min=1"`
	DateStart         string   `json:"date_start" binding:"required"`
	DateEnd           string   `json:"date_end" binding:"required"`
	ReturnDateStart   string   `json:"return_date_start,omitempty"`
	ReturnDateEnd     string   `json:"return_date_end,omitempty"`
	TripLength        int      `json:"trip_length,omitempty" binding:"min=0"`
	DynamicDates      bool     `json:"dynamic_dates,omitempty"`       // Use dates relative to execution time
	DaysFromExecution int      `json:"days_from_execution,omitempty"` // Start searching X days from now
	SearchWindowDays  int      `json:"search_window_days,omitempty"`  // Search within X days window
	Adults            int      `json:"adults" binding:"required,min=1"`
	Children          int      `json:"children" binding:"min=0"`
	InfantsLap        int      `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int      `json:"infants_seat" binding:"min=0"`
	TripType          string   `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string   `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string   `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Currency          string   `json:"currency" binding:"required,len=3"`
	CronExpression    string   `json:"cron_expression" binding:"required"`
}

// PriceGraphSweepRequest represents a request to enqueue a price graph sweep
type PriceGraphSweepRequest struct {
	Origins           []string `json:"origins" binding:"required,min=1"`
	Destinations      []string `json:"destinations" binding:"required,min=1"`
	DepartureDateFrom DateOnly `json:"departure_date_from" binding:"required"`
	DepartureDateTo   DateOnly `json:"departure_date_to" binding:"required"`
	TripLengths       []int    `json:"trip_lengths,omitempty"`
	TripType          string   `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	// Class is kept for backward compatibility; prefer Classes for multi-cabin sweeps.
	Class           string   `json:"class,omitempty"`
	Classes         []string `json:"classes,omitempty"`
	Stops           string   `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Adults          int      `json:"adults" binding:"required,min=1"`
	Children        int      `json:"children" binding:"min=0"`
	InfantsLap      int      `json:"infants_lap" binding:"min=0"`
	InfantsSeat     int      `json:"infants_seat" binding:"min=0"`
	Currency        string   `json:"currency" binding:"required,len=3"`
	RateLimitMillis int      `json:"rate_limit_millis,omitempty" binding:"min=0"`
}

func maybeNullInt(value sql.NullInt32) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Int32
}

func maybeNullTime(value sql.NullTime) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Time
}

func maybeNullString(value sql.NullString) interface{} {
	if !value.Valid {
		return nil
	}
	return value.String
}

func normalizePriceGraphSweepClasses(legacyClass string, classes []string) ([]string, error) {
	if legacyClass != "" && len(classes) > 0 {
		return nil, fmt.Errorf("provide either 'class' or 'classes', not both")
	}

	var input []string
	if legacyClass != "" {
		input = []string{legacyClass}
	} else {
		input = classes
	}

	if len(input) == 0 {
		return nil, fmt.Errorf("either 'class' or 'classes' is required")
	}

	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, c := range input {
		cabin := strings.ToLower(strings.TrimSpace(c))
		if cabin == "" {
			continue
		}
		switch cabin {
		case "economy", "premium_economy", "business", "first":
		default:
			return nil, fmt.Errorf("invalid class %q (allowed: economy, premium_economy, business, first)", c)
		}
		if _, ok := seen[cabin]; ok {
			continue
		}
		seen[cabin] = struct{}{}
		out = append(out, cabin)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("either 'class' or 'classes' must contain at least one valid value")
	}

	return out, nil
}

func containsToken(inputs []string, token string) bool {
	for _, v := range inputs {
		if strings.EqualFold(strings.TrimSpace(v), token) {
			return true
		}
	}
	return false
}

func listAllAirportCodes(ctx context.Context, pgDB db.PostgresDB) ([]string, error) {
	rows, err := pgDB.Search(ctx, `SELECT code FROM airports WHERE length(code) = 3 ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("failed to query airports: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	codes := make([]string, 0, 4096)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("failed to scan airport code: %w", err)
		}
		code = strings.ToUpper(strings.TrimSpace(code))
		if len(code) != 3 {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate airport codes: %w", err)
	}

	return codes, nil
}

// getAirports returns a handler for getting all airports
func GetAirports(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		const statusClientClosedRequest = 499

		ctx := c.Request.Context()
		requestID := middleware.GetRequestID(c)

		start := time.Now()
		rows, err := pgDB.QueryAirports(ctx)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"duration":   time.Since(start),
				"ctx_err":    ctx.Err(),
			}).Error(err, "QueryAirports failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query airports"})
			return
		}
		defer rows.Close()

		airports := []db.Airport{} // Use the defined struct
		for rows.Next() {
			if err := ctx.Err(); err != nil {
				logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"duration":   time.Since(start),
					"rows_read":  len(airports),
					"ctx_err":    err,
				}).Warn("GetAirports client disconnected (context canceled)")
				c.Status(statusClientClosedRequest)
				c.Abort()
				return
			}

			var airport db.Airport
			if err := rows.Scan(&airport.Code, &airport.Name, &airport.City, &airport.Country, &airport.Latitude, &airport.Longitude); err != nil {
				logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"duration":   time.Since(start),
					"rows_read":  len(airports),
					"ctx_err":    ctx.Err(),
				}).Error(err, "GetAirports scan failed")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan airport row"})
				return
			}
			airports = append(airports, airport)
		}
		if err := rows.Err(); err != nil {
			if ctx.Err() != nil {
				logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"duration":   time.Since(start),
					"rows_read":  len(airports),
					"ctx_err":    ctx.Err(),
				}).Warn("GetAirports rows iteration stopped (context canceled)")
				c.Status(statusClientClosedRequest)
				c.Abort()
				return
			}
			logger.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"client_ip":  c.ClientIP(),
				"duration":   time.Since(start),
				"rows_read":  len(airports),
			}).Error(err, "GetAirports rows iteration failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating airport rows"})
			return
		}

		// Convert to map[string]interface{} for JSON response to handle nulls gracefully
		response := []map[string]interface{}{}
		for _, airport := range airports {
			airportMap := map[string]interface{}{
				"code":    airport.Code,
				"name":    airport.Name,
				"city":    airport.City,
				"country": airport.Country,
			}
			if airport.Latitude.Valid {
				airportMap["latitude"] = airport.Latitude.Float64
			}
			if airport.Longitude.Valid {
				airportMap["longitude"] = airport.Longitude.Float64
			}
			response = append(response, airportMap)
		}

		c.JSON(http.StatusOK, response)
	}
}

func ParseClass(class string) flights.Class {
	switch class {
	case "economy":
		return flights.Economy
	case "premium_economy":
		return flights.PremiumEconomy
	case "business":
		return flights.Business
	case "first":
		return flights.First
	default:
		return flights.Economy // Default to economy
	}
}

func ParseStops(stops string) flights.Stops {
	switch stops {
	case "nonstop":
		return flights.Nonstop
	case "one_stop":
		return flights.Stop1
	case "two_stops":
		return flights.Stop2
	case "any":
		return flights.AnyStops
	default:
		return flights.AnyStops // Default to any stops
	}
}

// getAirlines returns a handler for getting all airlines
func GetAirlines(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pgDB.QueryAirlines(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query airlines"})
			return
		}
		defer rows.Close()

		airlines := []db.Airline{} // Use the defined struct
		for rows.Next() {
			var airline db.Airline
			if err := rows.Scan(&airline.Code, &airline.Name, &airline.Country); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan airline row"})
				return
			}
			airlines = append(airlines, airline)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating airline rows"})
			return
		}

		c.JSON(http.StatusOK, airlines)
	}
}

// CreateSearch returns a handler for creating a new flight search
func CreateSearch(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// --- Custom Validation Logic ---
		iataRegex := regexp.MustCompile(`^[A-Z]{3}$`)     // Simple IATA check
		currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`) // Simple Currency check

		// Validate IATA codes format
		if !iataRegex.MatchString(req.Origin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid origin airport code format"})
			return
		}
		if !iataRegex.MatchString(req.Destination) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination airport code format"})
			return
		}
		// Validate Currency format
		req.Currency = strings.ToUpper(req.Currency) // Standardize before check
		if !currencyRegex.MatchString(req.Currency) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency code format"})
			return
		}
		// Validate dates
		now := time.Now().Truncate(24 * time.Hour) // Compare dates only
		if req.DepartureDate.Before(now) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Departure date must be in the future"})
			return
		}
		if req.TripType == "round_trip" {
			if req.ReturnDate.IsZero() {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Return date is required for round trips"})
				return
			}
			if !req.ReturnDate.Time.After(req.DepartureDate.Time) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Return date must be after departure date"})
				return
			}
		} else { // one_way
			// Ensure return date is zero/empty for one-way trips
			if !req.ReturnDate.IsZero() {
				// Return error if user provided a return date for a one-way trip
				c.JSON(http.StatusBadRequest, gin.H{"error": "Return date should not be provided for one-way trips"})
				return
			}
			// Ensure ReturnDate is zeroed in the payload for one-way
			req.ReturnDate = DateOnly{}
		}
		// --- End Custom Validation ---

		// Create a worker payload (only if validation passes)
		payload := worker.FlightSearchPayload{
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate.Time,
			ReturnDate:    req.ReturnDate.Time, // Correctly zeroed for one-way
			Adults:        req.Adults,
			Children:      req.Children,
			InfantsLap:    req.InfantsLap,
			InfantsSeat:   req.InfantsSeat,
			TripType:      req.TripType,
			Class:         req.Class,
			Stops:         req.Stops,
			Currency:      req.Currency, // Already uppercased
		}

		// Enqueue the job
		jobID, err := q.Enqueue(c.Request.Context(), "flight_search", payload)
		if err != nil {
			// Log the internal error? Consider adding logging here.
			// log.Printf("Error enqueuing job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create search job"}) // Generic error for client
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  jobID,
			"message": "Flight search job created successfully",
		})
	}
}

// GetSearchByID returns a handler for getting a search by ID
func GetSearchByID(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		searchID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
			return
		}

		// Get the search query
		query, err := pgDB.GetSearchQueryByID(c.Request.Context(), searchID)
		if err != nil {
			// TODO: Check for specific not found error type if defined in db package
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Search not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search query"})
			}
			return
		}

		// Get the flight offers
		offerRows, err := pgDB.GetFlightOffersBySearchID(c.Request.Context(), searchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight offers"})
			return
		}
		defer offerRows.Close()

		offers := []map[string]interface{}{}
		for offerRows.Next() {
			var offer db.FlightOffer // Use the defined struct
			if err := offerRows.Scan(&offer.ID, &offer.Price, &offer.Currency, &offer.AirlineCodes, &offer.OutboundDuration, &offer.OutboundStops, &offer.ReturnDuration, &offer.ReturnStops, &offer.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan flight offer: " + err.Error()})
				return
			}

			// Get the flight segments for this offer
			segmentRows, err := pgDB.GetFlightSegmentsByOfferID(c.Request.Context(), offer.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight segments"})
				return
			}

			segments := []db.FlightSegment{} // Use the defined struct
			for segmentRows.Next() {
				var segment db.FlightSegment
				if err := segmentRows.Scan(
					&segment.AirlineCode, &segment.FlightNumber, &segment.DepartureAirport,
					&segment.ArrivalAirport, &segment.DepartureTime, &segment.ArrivalTime,
					&segment.Duration, &segment.Airplane, &segment.Legroom, &segment.IsReturn,
				); err != nil {
					segmentRows.Close()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan flight segment"})
					return
				}
				segments = append(segments, segment)
			}
			segmentErr := segmentRows.Err()
			segmentRows.Close() // Close immediately after use, not via defer
			if segmentErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating flight segments"})
				return
			}

			offerMap := map[string]interface{}{
				"id":         offer.ID,
				"price":      offer.Price,
				"currency":   offer.Currency,
				"created_at": offer.CreatedAt,
				"segments":   segments, // Use the struct slice directly
			}

			if offer.AirlineCodes.Valid {
				offerMap["airline_codes"] = offer.AirlineCodes.String
			}
			if offer.OutboundDuration.Valid {
				offerMap["outbound_duration"] = offer.OutboundDuration.Int64
			}
			if offer.OutboundStops.Valid {
				offerMap["outbound_stops"] = offer.OutboundStops.Int64
			}
			if offer.ReturnDuration.Valid {
				offerMap["return_duration"] = offer.ReturnDuration.Int64
			}
			if offer.ReturnStops.Valid {
				offerMap["return_stops"] = offer.ReturnStops.Int64
			}

			offers = append(offers, offerMap)
		}
		if err := offerRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating flight offers: " + err.Error()})
			return
		}

		result := map[string]interface{}{
			"id":             query.ID,
			"origin":         query.Origin,
			"destination":    query.Destination,
			"departure_date": query.DepartureDate,
			"status":         query.Status,
			"created_at":     query.CreatedAt,
			"offers":         offers,
		}

		if query.ReturnDate.Valid {
			result["return_date"] = query.ReturnDate.Time
		}

		c.JSON(http.StatusOK, result)
	}
}

// listSearches returns a handler for listing recent searches
func ListSearches(pgDB db.PostgresDB) gin.HandlerFunc { // Changed parameter type, EXPORTED
	return func(c *gin.Context) {
		// Get pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 10
		}
		offset := (page - 1) * perPage

		// Get the total count
		total, err := pgDB.CountSearches(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count searches: " + err.Error()})
			return
		}

		// Get the search queries
		rows, err := pgDB.QuerySearchesPaginated(c.Request.Context(), perPage, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query searches: " + err.Error()})
			return
		}
		defer rows.Close()

		queries := []map[string]interface{}{}
		for rows.Next() {
			var query db.SearchQuery // Use the defined struct
			if err := rows.Scan(&query.ID, &query.Origin, &query.Destination, &query.DepartureDate, &query.ReturnDate, &query.Status, &query.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan search query: " + err.Error()})
				return
			}

			queryMap := map[string]interface{}{
				"id":             query.ID,
				"origin":         query.Origin,
				"destination":    query.Destination,
				"departure_date": query.DepartureDate,
				"status":         query.Status,
				"created_at":     query.CreatedAt,
			}

			if query.ReturnDate.Valid {
				queryMap["return_date"] = query.ReturnDate.Time
			}

			queries = append(queries, queryMap)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating search queries: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total":       total,
			"page":        page,
			"per_page":    perPage,
			"total_pages": (total + perPage - 1) / perPage,
			"data":        queries,
		})
	}
}

// deleteJob returns a handler for deleting a job
func DeleteJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc { // Changed parameter type, EXPORTED
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Begin a transaction
		tx, err := pgDB.BeginTx(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
			return
		}
		defer tx.Rollback() // Ensure rollback on error

		// Delete the job details
		ctx := c.Request.Context()
		err = pgDB.DeleteJobDetailsByJobID(ctx, tx, jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job details"})
			return
		}

		// Delete the job
		rowsAffected, err := pgDB.DeleteScheduledJobByID(ctx, tx, jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduled job"})
			return
		}

		// Check if the job was found
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
			return
		}

		// Remove the job from the scheduler
		// scheduler := workerManager.GetScheduler()
		// scheduler.RemoveJob(jobID)

		c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
	}
}

// runJob returns a handler for manually triggering a job
func runJob(q queue.Queue, pgDB db.PostgresDB) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Get the job details
		details, err := pgDB.GetJobDetailsByID(c.Request.Context(), jobID)
		if err != nil {
			// TODO: Check for specific not found error type
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job details: " + err.Error()})
			}
			return
		}

		// Create a bulk search payload
		payload := worker.BulkSearchPayload{
			Origins:           []string{details.Origin},
			Destinations:      []string{details.Destination},
			DepartureDateFrom: details.DepartureDateStart,
			DepartureDateTo:   details.DepartureDateEnd,
			Adults:            details.Adults,
			Children:          details.Children,
			InfantsLap:        details.InfantsLap,
			InfantsSeat:       details.InfantsSeat,
			TripType:          details.TripType,
			Class:             details.Class,    // Pass the original string
			Stops:             details.Stops,    // Pass the original string
			Currency:          details.Currency, // Use currency from details
		}

		// Add optional fields if present
		if details.ReturnDateStart.Valid {
			payload.ReturnDateFrom = details.ReturnDateStart.Time
		}
		if details.ReturnDateEnd.Valid {
			payload.ReturnDateTo = details.ReturnDateEnd.Time
		}
		if details.TripLength.Valid {
			payload.TripLength = int(details.TripLength.Int32)
		}

		// Enqueue the job
		jobQueueID, err := q.Enqueue(c.Request.Context(), "bulk_search", payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue job: " + err.Error()})
			return
		}

		// Update the last run time
		err = pgDB.UpdateJobLastRun(c.Request.Context(), jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job last run time: " + err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  jobQueueID,
			"message": "Job triggered successfully",
		})
	}
}

// enableJob returns a handler for enabling a job
func enableJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		rowsAffected, err := pgDB.UpdateJobEnabled(c.Request.Context(), jobID, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable job: " + err.Error()})
			return
		}

		// Check if the job was found
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// Get the job's cron expression - Removed as cronExpr is not used below
		// cronExpr, err := pgDB.GetJobCronExpression(c.Request.Context(), jobID)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job cron expression: " + err.Error()})
		// 	// Consider rolling back the enable status or logging a warning
		// 	return
		// }

		// TODO: Add logic here if the scheduler needs to be explicitly notified
		//       about the newly enabled job (e.g., reload jobs from DB).
		//       The current AddJob call was incorrect.
		// scheduler := workerManager.GetScheduler()
		// if err := scheduler.AddJob(...); err != nil { ... }

		c.JSON(http.StatusOK, gin.H{"message": "Job enabled successfully"})
	}
}

// disableJob returns a handler for disabling a job
func disableJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		rowsAffected, err := pgDB.UpdateJobEnabled(c.Request.Context(), jobID, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable job: " + err.Error()})
			return
		}

		// Check if the job was found
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// TODO: Consider removing the job from the scheduler here as well
		// scheduler := workerManager.GetScheduler()
		// scheduler.RemoveJob(jobID)

		c.JSON(http.StatusOK, gin.H{"message": "Job disabled successfully"})
	}
}

// getWorkerStatus returns a handler for getting worker status
type workerStatusResponse struct {
	ID                  any    `json:"id"`
	Status              string `json:"status"`
	CurrentJob          string `json:"current_job,omitempty"`
	ProcessedJobs       int    `json:"processed_jobs"`
	Uptime              int64  `json:"uptime"`
	Source              string `json:"source,omitempty"`
	Hostname            string `json:"hostname,omitempty"`
	Concurrency         int    `json:"concurrency,omitempty"`
	HeartbeatAgeSeconds int64  `json:"heartbeat_age_seconds,omitempty"`
	Version             string `json:"version,omitempty"`
}

// GetWorkerStatus returns a handler for getting worker status.
// It combines local in-process worker goroutines with remote worker instances (via Redis heartbeats).
func GetWorkerStatus(workerManager WorkerStatusProvider, redisClient *redis.Client, cfg config.WorkerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		out := make([]workerStatusResponse, 0)

		// Local worker goroutines (in this process)
		if workerManager != nil {
			statuses := workerManager.WorkerStatuses()
			for _, s := range statuses {
				out = append(out, workerStatusResponse{
					ID:            s.ID,
					Status:        s.Status,
					CurrentJob:    s.CurrentJob,
					ProcessedJobs: s.ProcessedJobs,
					Uptime:        s.Uptime,
					Source:        "local",
					Version:       buildinfo.VersionString(),
				})
			}
		}

		// Remote worker instances (published by worker processes)
		if redisClient != nil {
			namespace := cfg.RegistryNamespace
			if namespace == "" {
				namespace = "flights"
			}

			reg := worker_registry.New(redisClient, namespace)

			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			heartbeats, err := reg.ListActive(ctx, cfg.HeartbeatTTL, 200)
			cancel()
			if err != nil {
				logger.Warn("Failed to list remote workers", "error", err)
			} else {
				now := time.Now().UTC()
				for _, hb := range heartbeats {
					uptime := int64(0)
					if !hb.StartedAt.IsZero() {
						uptime = int64(now.Sub(hb.StartedAt).Seconds())
						if uptime < 0 {
							uptime = 0
						}
					}
					age := int64(0)
					if !hb.LastHeartbeat.IsZero() {
						age = int64(now.Sub(hb.LastHeartbeat).Seconds())
						if age < 0 {
							age = 0
						}
					}

					out = append(out, workerStatusResponse{
						ID:                  hb.ID,
						Status:              hb.Status,
						CurrentJob:          hb.CurrentJob,
						ProcessedJobs:       hb.ProcessedJobs,
						Uptime:              uptime,
						Source:              "remote",
						Hostname:            hb.Hostname,
						Concurrency:         hb.Concurrency,
						HeartbeatAgeSeconds: age,
						Version:             hb.Version,
					})
				}
			}
		}

		c.JSON(http.StatusOK, out)
	}
}

// getBulkSearchById returns a handler for getting a bulk search by ID
func getBulkSearchById(pgDB db.PostgresDB) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		searchID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk search ID"})
			return
		}

		// Get bulk search metadata
		search, err := pgDB.GetBulkSearchByID(c.Request.Context(), searchID)
		if err != nil {
			// TODO: Check for specific not found error type
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Bulk search not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bulk search metadata"})
			}
			return
		}

		// Get paginated results
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 25
		}
		offset := (page - 1) * perPage

		rows, err := pgDB.QueryBulkSearchResultsPaginated(c.Request.Context(), searchID, perPage, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query bulk search results"})
			return
		}
		defer rows.Close()

		results := []db.BulkSearchResult{} // Use the defined struct
		for rows.Next() {
			var res db.BulkSearchResult
			err := rows.Scan(
				&res.Origin,
				&res.Destination,
				&res.DepartureDate,
				&res.ReturnDate,
				&res.Price,
				&res.Currency,
				&res.AirlineCode,
				&res.Duration,
				&res.SrcAirportCode,
				&res.DstAirportCode,
				&res.SrcCity,
				&res.DstCity,
				&res.FlightDuration,
				&res.ReturnFlightDuration,
				&res.OutboundFlights,
				&res.ReturnFlights,
				&res.OfferJSON,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan bulk search result: " + err.Error()})
				return
			}
			results = append(results, res)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating bulk search results: " + err.Error()})
			return
		}

		// Convert results to map[string]interface{} for JSON response to handle nulls
		resultsMap := []map[string]interface{}{}
		for _, res := range results {
			resultMap := map[string]interface{}{
				"origin":         res.Origin,
				"destination":    res.Destination,
				"departure_date": res.DepartureDate,
				"price":          res.Price,
				"currency":       res.Currency,
				"airline_code":   res.AirlineCode,
				"duration":       res.Duration,
			}
			if res.ReturnDate.Valid {
				resultMap["return_date"] = res.ReturnDate.Time
			}
			resultsMap = append(resultsMap, resultMap)
		}

		var jobID interface{}
		if search.JobID.Valid {
			jobID = search.JobID.Int32
		}

		response := map[string]interface{}{
			"id":             search.ID,
			"job_id":         jobID,
			"status":         search.Status,
			"total_searches": search.TotalSearches,
			"completed":      search.Completed,
			"total_offers":   search.TotalOffers,
			"error_count":    search.ErrorCount,
			"created_at":     search.CreatedAt,
			"updated_at":     search.UpdatedAt,
			"results":        resultsMap, // Use the converted map
			"pagination": map[string]interface{}{
				"page":     page,
				"per_page": perPage,
				// Ensure TotalSearches is used for total pages calculation
				"total_pages": (search.TotalSearches + perPage - 1) / perPage,
			},
		}

		if search.CompletedAt.Valid {
			response["completed_at"] = search.CompletedAt.Time
		}
		if search.MinPrice.Valid {
			response["min_price"] = search.MinPrice.Float64
		}
		if search.MaxPrice.Valid {
			response["max_price"] = search.MaxPrice.Float64
		}
		if search.AveragePrice.Valid {
			response["average_price"] = search.AveragePrice.Float64
		}

		c.JSON(http.StatusOK, response)
	}
}

// getPriceHistory returns a handler for getting price history for a route
func getPriceHistory(neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context() // Use request context
		origin := c.Param("origin")
		destination := c.Param("destination")

		includeGroups := c.QueryArray("include_airline_groups")
		excludeGroups := c.QueryArray("exclude_airline_groups")

		includeSet := make(map[string]struct{})
		if len(includeGroups) > 0 {
			codes, _, err := macros.ExpandAirlineTokens(includeGroups)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid include_airline_groups: " + err.Error()})
				return
			}
			for _, code := range codes {
				includeSet[code] = struct{}{}
			}
		}

		excludeSet := make(map[string]struct{})
		if len(excludeGroups) > 0 {
			codes, _, err := macros.ExpandAirlineTokens(excludeGroups)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exclude_airline_groups: " + err.Error()})
				return
			}
			for _, code := range codes {
				excludeSet[code] = struct{}{}
			}
		}

		query := "MATCH (origin:Airport {code: $originCode})-[r:PRICE_POINT]->(dest:Airport {code: $destCode}) " +
			"RETURN r.date AS date, r.price AS price, r.airline AS airline ORDER BY r.date"
		params := map[string]interface{}{
			"originCode": origin,
			"destCode":   destination,
		}

		result, err := neo4jDB.ExecuteReadQuery(ctx, query, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query price history"})
			return
		}
		defer result.Close() // Ensure session is closed when done

		priceHistory := []map[string]interface{}{}
		for result.Next() {
			record := result.Record()
			date, _ := record.Get("date")
			price, _ := record.Get("price")
			airline, _ := record.Get("airline")

			airlineCode := strings.ToUpper(strings.TrimSpace(fmt.Sprint(airline)))
			if airlineCode != "" {
				if _, excluded := excludeSet[airlineCode]; excluded {
					continue
				}
				if len(includeSet) > 0 {
					if _, ok := includeSet[airlineCode]; !ok {
						continue
					}
				}
			} else if len(includeSet) > 0 {
				continue
			}

			var dateVal interface{}
			if dt, ok := date.(neo4j.Date); ok {
				dateVal = dt.Time() // Convert neo4j.Date to time.Time
			} else {
				dateVal = date
			}

			priceHistory = append(priceHistory, map[string]interface{}{
				"date":    dateVal,
				"price":   price,
				"airline": airline,
			})
		}
		if err = result.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing price history results"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"origin":      origin,
			"destination": destination,
			"history":     priceHistory,
			"filter": map[string]interface{}{
				"include_airline_groups": includeGroups,
				"exclude_airline_groups": excludeGroups,
			},
		})
	}
}

// listJobs returns a handler for listing all scheduled jobs
func listJobs(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("ListJobs called") // Debug log
		rows, err := pgDB.ListJobs(c.Request.Context())
		if err != nil {
			log.Printf("ListJobs query error: %v", err) // Debug log
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list jobs: " + err.Error()})
			return
		}
		defer rows.Close()
		log.Printf("ListJobs query successful, starting to scan rows") // Debug log

		jobs := []map[string]interface{}{}
		for rows.Next() {
			var job db.ScheduledJob // Use the defined struct

			// Variables for the job details from LEFT JOIN
			var origin, destination sql.NullString
			var dynamicDates sql.NullBool
			var daysFromExecution, searchWindowDays, tripLength sql.NullInt32

			if err := rows.Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt,
				&origin, &destination, &dynamicDates, &daysFromExecution, &searchWindowDays, &tripLength); err != nil {
				log.Printf("Scan error: %v", err) // Debug log
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan job: " + err.Error()})
				return
			}

			log.Printf("Scanned job %d: %s (origin=%v, dest=%v, dynamic=%v)", job.ID, job.Name, origin.String, destination.String, dynamicDates.Bool)

			jobMap := map[string]interface{}{
				"id":              job.ID,
				"name":            job.Name,
				"cron_expression": job.CronExpression,
				"enabled":         job.Enabled,
				"created_at":      job.CreatedAt,
				"updated_at":      job.UpdatedAt,
			}

			if job.LastRun.Valid {
				jobMap["last_run"] = job.LastRun.Time
			} else {
				jobMap["last_run"] = nil // Explicitly set null if not valid
			}

			// Add job details from database
			details := map[string]interface{}{}

			if origin.Valid && destination.Valid {
				// Use real data from database
				details["origin"] = origin.String
				details["destination"] = destination.String
				details["dynamic_dates"] = dynamicDates.Bool

				if daysFromExecution.Valid {
					details["days_from_execution"] = daysFromExecution.Int32
				}
				if searchWindowDays.Valid {
					details["search_window_days"] = searchWindowDays.Int32
				}
				if tripLength.Valid {
					details["trip_length"] = tripLength.Int32
				}
			} else {
				// Fallback - this should not happen for properly created jobs
				log.Printf("Job %d has no details in database", job.ID)
				details["origin"] = "N/A"
				details["destination"] = "N/A"
				details["dynamic_dates"] = false
			}

			jobMap["details"] = details

			jobs = append(jobs, jobMap)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating jobs: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, jobs)
	}
}

// listBulkSearches returns a handler for listing bulk search runs
func listBulkSearches(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		rows, err := pgDB.ListBulkSearches(c.Request.Context(), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list bulk searches: " + err.Error()})
			return
		}
		defer rows.Close()

		searches := []map[string]interface{}{}
		for rows.Next() {
			var search db.BulkSearch
			if err := rows.Scan(&search.ID, &search.JobID, &search.Status, &search.TotalSearches, &search.Completed,
				&search.TotalOffers, &search.ErrorCount, &search.Currency, &search.CreatedAt, &search.UpdatedAt,
				&search.CompletedAt, &search.MinPrice, &search.MaxPrice, &search.AveragePrice); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan bulk search: " + err.Error()})
				return
			}

			record := map[string]interface{}{
				"id":           search.ID,
				"status":       search.Status,
				"total_routes": search.TotalSearches,
				"completed":    search.Completed,
				"total_offers": search.TotalOffers,
				"error_count":  search.ErrorCount,
				"currency":     search.Currency,
				"created_at":   search.CreatedAt,
				"updated_at":   search.UpdatedAt,
			}
			if search.JobID.Valid {
				record["job_id"] = search.JobID.Int32
			}
			if search.CompletedAt.Valid {
				record["completed_at"] = search.CompletedAt.Time
			}
			if search.MinPrice.Valid {
				record["min_price"] = search.MinPrice.Float64
			}
			if search.MaxPrice.Valid {
				record["max_price"] = search.MaxPrice.Float64
			}
			if search.AveragePrice.Valid {
				record["average_price"] = search.AveragePrice.Float64
			}

			searches = append(searches, record)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating bulk searches: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": searches,
			"count": len(searches),
		})
	}
}

// enqueuePriceGraphSweep enqueues a price graph sweep job for execution.
// To avoid competing for rate-limited Google endpoints, it refuses to start when the continuous sweep is running.
func enqueuePriceGraphSweep(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if workerManager == nil || workerManager.GetScheduler() == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Price graph scheduler unavailable"})
			return
		}

		if runner := workerManager.GetSweepRunner(); runner != nil {
			status := runner.GetStatus()
			if status.IsRunning && !status.IsPaused {
				// Continuous sweeps are often 24/7; pause it temporarily so on-demand sweeps can run,
				// then auto-resume when the sweep queue drains.
				runner.PauseAndAutoResumeAfterQueueDrain("price_graph_sweep")
			}
		}

		var req PriceGraphSweepRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.DepartureDateFrom.Time.After(req.DepartureDateTo.Time) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "departure_date_from must be before departure_date_to"})
			return
		}

		stops := req.Stops
		if stops == "two_stops_plus" {
			stops = "any"
		}

		classes, err := normalizePriceGraphSweepClasses(req.Class, req.Classes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
		var overrides map[string][]string
		var worldAllCount int
		if containsToken(req.Origins, macros.RegionWorldAll) || containsToken(req.Destinations, macros.RegionWorldAll) {
			worldAll, err := listAllAirportCodes(ctx, pgDB)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			overrides = map[string][]string{macros.RegionWorldAll: worldAll}
			worldAllCount = len(worldAll)
		}

		// Expand region tokens in origins and destinations
		expandedOrigins, originWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Origins, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid origin: " + err.Error()})
			return
		}
		expandedDestinations, destinationWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Destinations, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination: " + err.Error()})
			return
		}

		warnings := append([]string{}, originWarnings...)
		warnings = append(warnings, destinationWarnings...)
		if worldAllCount > 0 {
			warnings = append(warnings, fmt.Sprintf("%s expanded to %d airports", macros.RegionWorldAll, worldAllCount))
		}

		// Guard against accidental workload explosion from region tokens
		totalRoutes := len(expandedOrigins) * len(expandedDestinations)
		if totalRoutes == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "At least one origin and one destination are required (after REGION:* expansion)",
				"warnings": warnings,
			})
			return
		}
		const maxSweepRoutes = 10000
		if totalRoutes > maxSweepRoutes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        fmt.Sprintf("Too many routes: %d (max %d). Region tokens can expand to many airports; consider narrowing your search.", totalRoutes, maxSweepRoutes),
				"total_routes": totalRoutes,
				"max_routes":   maxSweepRoutes,
				"origins":      len(expandedOrigins),
				"destinations": len(expandedDestinations),
				"warnings":     warnings,
			})
			return
		}

		payload := worker.PriceGraphSweepPayload{
			Origins:           expandedOrigins,
			Destinations:      expandedDestinations,
			DepartureDateFrom: req.DepartureDateFrom.Time,
			DepartureDateTo:   req.DepartureDateTo.Time,
			TripLengths:       req.TripLengths,
			TripType:          req.TripType,
			Class:             classes[0],
			Classes:           classes,
			Stops:             stops,
			Adults:            req.Adults,
			Children:          req.Children,
			InfantsLap:        req.InfantsLap,
			InfantsSeat:       req.InfantsSeat,
			Currency:          strings.ToUpper(req.Currency),
			RateLimitMillis:   req.RateLimitMillis,
		}

		sweepID, err := workerManager.GetScheduler().EnqueuePriceGraphSweep(ctx, payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue price graph sweep: " + err.Error()})
			return
		}

		sweep, err := pgDB.GetPriceGraphSweepByID(ctx, sweepID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sweep metadata: " + err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message":  "Price graph sweep enqueued",
			"sweep_id": sweepID,
			"classes":  classes,
			"warnings": warnings,
			"sweep": map[string]interface{}{
				"id":                sweep.ID,
				"status":            sweep.Status,
				"origin_count":      sweep.OriginCount,
				"destination_count": sweep.DestinationCount,
				"trip_length_min":   maybeNullInt(sweep.TripLengthMin),
				"trip_length_max":   maybeNullInt(sweep.TripLengthMax),
				"currency":          sweep.Currency,
				"created_at":        sweep.CreatedAt,
				"updated_at":        sweep.UpdatedAt,
				"started_at":        maybeNullTime(sweep.StartedAt),
				"completed_at":      maybeNullTime(sweep.CompletedAt),
				"error_count":       sweep.ErrorCount,
			},
		})
	}
}

// listPriceGraphSweeps lists price graph sweep runs
func listPriceGraphSweeps(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		rows, err := pgDB.ListPriceGraphSweeps(c.Request.Context(), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list price graph sweeps: " + err.Error()})
			return
		}
		defer rows.Close()

		results := make([]map[string]interface{}, 0)
		for rows.Next() {
			var sweep db.PriceGraphSweep
			if err := rows.Scan(&sweep.ID, &sweep.JobID, &sweep.Status, &sweep.OriginCount, &sweep.DestinationCount,
				&sweep.TripLengthMin, &sweep.TripLengthMax, &sweep.Currency, &sweep.ErrorCount,
				&sweep.CreatedAt, &sweep.UpdatedAt, &sweep.StartedAt, &sweep.CompletedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan price graph sweep: " + err.Error()})
				return
			}

			entry := map[string]interface{}{
				"id":                sweep.ID,
				"status":            sweep.Status,
				"origin_count":      sweep.OriginCount,
				"destination_count": sweep.DestinationCount,
				"trip_length_min":   maybeNullInt(sweep.TripLengthMin),
				"trip_length_max":   maybeNullInt(sweep.TripLengthMax),
				"currency":          sweep.Currency,
				"error_count":       sweep.ErrorCount,
				"created_at":        sweep.CreatedAt,
				"updated_at":        sweep.UpdatedAt,
				"started_at":        maybeNullTime(sweep.StartedAt),
				"completed_at":      maybeNullTime(sweep.CompletedAt),
			}
			if sweep.JobID.Valid {
				entry["job_id"] = sweep.JobID.Int32
			}
			results = append(results, entry)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating sweeps: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": results,
			"count": len(results),
		})
	}
}

// getPriceGraphSweepResults returns stored price graph results for a sweep
func getPriceGraphSweepResults(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sweepID, err := strconv.Atoi(c.Param("id"))
		if err != nil || sweepID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sweep ID"})
			return
		}

		ctx := c.Request.Context()
		summary, err := pgDB.GetPriceGraphSweepByID(ctx, sweepID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Price graph sweep not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sweep metadata: " + err.Error()})
			}
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		rows, err := pgDB.ListPriceGraphResults(ctx, sweepID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sweep results: " + err.Error()})
			return
		}
		defer rows.Close()

		results := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id          int
				sid         int
				origin      string
				destination string
				departure   time.Time
				returnDate  sql.NullTime
				tripLength  sql.NullInt32
				price       float64
				currency    string
				adults      int
				children    int
				infantsLap  int
				infantsSeat int
				tripType    string
				class       string
				stops       string
				searchURL   sql.NullString
				queriedAt   time.Time
				createdAt   time.Time
			)

			if err := rows.Scan(
				&id,
				&sid,
				&origin,
				&destination,
				&departure,
				&returnDate,
				&tripLength,
				&price,
				&currency,
				&adults,
				&children,
				&infantsLap,
				&infantsSeat,
				&tripType,
				&class,
				&stops,
				&searchURL,
				&queriedAt,
				&createdAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan result: " + err.Error()})
				return
			}

			entry := map[string]interface{}{
				"id":             id,
				"sweep_id":       sid,
				"origin":         origin,
				"destination":    destination,
				"departure_date": departure,
				"return_date":    maybeNullTime(returnDate),
				"trip_length":    maybeNullInt(tripLength),
				"price":          price,
				"currency":       currency,
				"adults":         adults,
				"children":       children,
				"infants_lap":    infantsLap,
				"infants_seat":   infantsSeat,
				"trip_type":      tripType,
				"class":          class,
				"stops":          stops,
				"search_url":     maybeNullString(searchURL),
				"queried_at":     queriedAt,
				"created_at":     createdAt,
			}
			results = append(results, entry)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating results: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sweep": map[string]interface{}{
				"id":                summary.ID,
				"status":            summary.Status,
				"origin_count":      summary.OriginCount,
				"destination_count": summary.DestinationCount,
				"trip_length_min":   maybeNullInt(summary.TripLengthMin),
				"trip_length_max":   maybeNullInt(summary.TripLengthMax),
				"currency":          summary.Currency,
				"error_count":       summary.ErrorCount,
				"created_at":        summary.CreatedAt,
				"updated_at":        summary.UpdatedAt,
				"started_at":        maybeNullTime(summary.StartedAt),
				"completed_at":      maybeNullTime(summary.CompletedAt),
			},
			"results": results,
			"count":   len(results),
		})
	}
}

// getBulkSearchResults returns a handler that fetches summary and results for a bulk search
func getBulkSearchResults(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		bulkID, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk search ID"})
			return
		}

		summary, err := pgDB.GetBulkSearchByID(c.Request.Context(), bulkID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Bulk search not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load bulk search: " + err.Error()})
			}
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		rows, err := pgDB.QueryBulkSearchResultsPaginated(c.Request.Context(), bulkID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query bulk search results"})
			return
		}
		defer rows.Close()

		results := []map[string]interface{}{}
		for rows.Next() {
			var result db.BulkSearchResult
			if err := rows.Scan(&result.Origin, &result.Destination, &result.DepartureDate, &result.ReturnDate,
				&result.Price, &result.Currency, &result.AirlineCode, &result.Duration,
				&result.SrcAirportCode, &result.DstAirportCode, &result.SrcCity, &result.DstCity,
				&result.FlightDuration, &result.ReturnFlightDuration, &result.OutboundFlights, &result.ReturnFlights, &result.OfferJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan bulk search result: " + err.Error()})
				return
			}

			entry := map[string]interface{}{
				"origin":         result.Origin,
				"destination":    result.Destination,
				"departure_date": result.DepartureDate,
				"price":          result.Price,
				"currency":       result.Currency,
			}
			if result.ReturnDate.Valid {
				entry["return_date"] = result.ReturnDate.Time
			}
			if result.AirlineCode.Valid {
				entry["airline_code"] = result.AirlineCode.String
			}
			if result.Duration.Valid {
				entry["duration"] = result.Duration.Int32
			}
			if result.SrcAirportCode.Valid {
				entry["src_airport_code"] = result.SrcAirportCode.String
			}
			if result.DstAirportCode.Valid {
				entry["dst_airport_code"] = result.DstAirportCode.String
			}
			if result.SrcCity.Valid {
				entry["src_city"] = result.SrcCity.String
			}
			if result.DstCity.Valid {
				entry["dst_city"] = result.DstCity.String
			}
			if result.FlightDuration.Valid {
				entry["flight_duration"] = result.FlightDuration.Int32
			}
			if result.ReturnFlightDuration.Valid {
				entry["return_flight_duration"] = result.ReturnFlightDuration.Int32
			}
			if len(result.OutboundFlights) > 0 {
				entry["outbound_flights"] = json.RawMessage(result.OutboundFlights)
			}
			if len(result.ReturnFlights) > 0 {
				entry["return_flights"] = json.RawMessage(result.ReturnFlights)
			}
			if len(result.OfferJSON) > 0 {
				entry["offer_json"] = json.RawMessage(result.OfferJSON)
			}

			results = append(results, entry)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating bulk search results: " + err.Error()})
			return
		}

		summaryMap := map[string]interface{}{
			"id":           summary.ID,
			"status":       summary.Status,
			"total_routes": summary.TotalSearches,
			"completed":    summary.Completed,
			"total_offers": summary.TotalOffers,
			"error_count":  summary.ErrorCount,
			"currency":     summary.Currency,
			"created_at":   summary.CreatedAt,
			"updated_at":   summary.UpdatedAt,
		}
		if summary.JobID.Valid {
			summaryMap["job_id"] = summary.JobID.Int32
		}
		if summary.CompletedAt.Valid {
			summaryMap["completed_at"] = summary.CompletedAt.Time
		}
		if summary.MinPrice.Valid {
			summaryMap["min_price"] = summary.MinPrice.Float64
		}
		if summary.MaxPrice.Valid {
			summaryMap["max_price"] = summary.MaxPrice.Float64
		}
		if summary.AveragePrice.Valid {
			summaryMap["average_price"] = summary.AveragePrice.Float64
		}

		c.JSON(http.StatusOK, gin.H{
			"summary": summaryMap,
			"results": results,
			"count":   len(results),
		})
	}
}

// getBulkSearchOffers returns the full set of offers captured during a bulk search
func getBulkSearchOffers(pgDB db.PostgresDB) gin.HandlerFunc {
	type gridCell struct {
		DepartureDate time.Time  `json:"departure_date"`
		ReturnDate    *time.Time `json:"return_date,omitempty"`
		Price         float64    `json:"price"`
		Currency      string     `json:"currency"`
		AirlineCodes  []string   `json:"airline_codes,omitempty"`
		CreatedAt     time.Time  `json:"created_at"`
	}

	type routeGrid struct {
		Origin      string     `json:"origin"`
		Destination string     `json:"destination"`
		Cells       []gridCell `json:"cells"`
	}

	type gridAccumulator struct {
		origin      string
		destination string
		cells       map[string]*struct {
			departureDate time.Time
			returnDate    sql.NullTime
			price         float64
			currency      string
			airlineCodes  []string
			createdAt     time.Time
		}
	}

	return func(c *gin.Context) {
		idParam := c.Param("id")
		bulkID, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk search ID"})
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "200"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}
		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
			return
		}
		if offset < 0 {
			offset = 0
		}
		sortBy := c.DefaultQuery("sort_by", "price")

		offers, err := pgDB.ListBulkSearchOffers(c.Request.Context(), bulkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load bulk search offers: " + err.Error()})
			return
		}

		// Sort offers based on sort_by parameter
		switch sortBy {
		case "cost_per_mile":
			sort.Slice(offers, func(i, j int) bool {
				// Offers without cost_per_mile go to the end
				if !offers[i].CostPerMile.Valid && !offers[j].CostPerMile.Valid {
					return offers[i].Price < offers[j].Price // fallback to price
				}
				if !offers[i].CostPerMile.Valid {
					return false
				}
				if !offers[j].CostPerMile.Valid {
					return true
				}
				return offers[i].CostPerMile.Float64 < offers[j].CostPerMile.Float64
			})
		default: // "price" or any other value defaults to price sorting
			sort.Slice(offers, func(i, j int) bool {
				return offers[i].Price < offers[j].Price
			})
		}

		total := len(offers)
		if limit <= 0 || limit > total {
			limit = total
		}
		if offset > total {
			offset = total
		}

		end := offset + limit
		if end > total {
			end = total
		}

		offerItems := make([]map[string]interface{}, 0, end-offset)
		for i := offset; i < end; i++ {
			offer := offers[i]
			entry := map[string]interface{}{
				"origin":         offer.Origin,
				"destination":    offer.Destination,
				"departure_date": offer.DepartureDate,
				"price":          offer.Price,
				"currency":       offer.Currency,
				"created_at":     offer.CreatedAt,
			}
			if offer.ReturnDate.Valid {
				entry["return_date"] = offer.ReturnDate.Time
			}
			if len(offer.AirlineCodes) > 0 {
				entry["airline_codes"] = append([]string(nil), offer.AirlineCodes...)
			}
			if offer.SrcAirportCode.Valid {
				entry["src_airport_code"] = offer.SrcAirportCode.String
			}
			if offer.DstAirportCode.Valid {
				entry["dst_airport_code"] = offer.DstAirportCode.String
			}
			if offer.SrcCity.Valid {
				entry["src_city"] = offer.SrcCity.String
			}
			if offer.DstCity.Valid {
				entry["dst_city"] = offer.DstCity.String
			}
			if offer.FlightDuration.Valid {
				entry["flight_duration"] = offer.FlightDuration.Int32
			}
			if offer.ReturnFlightDuration.Valid {
				entry["return_flight_duration"] = offer.ReturnFlightDuration.Int32
			}
			if offer.DistanceMiles.Valid {
				entry["distance_miles"] = offer.DistanceMiles.Float64
			}
			if offer.CostPerMile.Valid {
				entry["cost_per_mile"] = offer.CostPerMile.Float64
			}
			if len(offer.OutboundFlights) > 0 {
				entry["outbound_flights"] = json.RawMessage(offer.OutboundFlights)
			}
			if len(offer.ReturnFlights) > 0 {
				entry["return_flights"] = json.RawMessage(offer.ReturnFlights)
			}
			if len(offer.OfferJSON) > 0 {
				entry["offer_json"] = json.RawMessage(offer.OfferJSON)
			}
			offerItems = append(offerItems, entry)
		}

		routeAccumulators := make(map[string]*gridAccumulator)
		for _, offer := range offers {
			routeKey := offer.Origin + "|" + offer.Destination
			acc, ok := routeAccumulators[routeKey]
			if !ok {
				acc = &gridAccumulator{
					origin:      offer.Origin,
					destination: offer.Destination,
					cells: make(map[string]*struct {
						departureDate time.Time
						returnDate    sql.NullTime
						price         float64
						currency      string
						airlineCodes  []string
						createdAt     time.Time
					}),
				}
				routeAccumulators[routeKey] = acc
			}

			cellKey := offer.DepartureDate.Format("2006-01-02")
			if offer.ReturnDate.Valid {
				cellKey += "|" + offer.ReturnDate.Time.Format("2006-01-02")
			} else {
				cellKey += "|one_way"
			}

			current := acc.cells[cellKey]
			if current == nil || offer.Price < current.price || (offer.Price == current.price && offer.CreatedAt.Before(current.createdAt)) {
				acc.cells[cellKey] = &struct {
					departureDate time.Time
					returnDate    sql.NullTime
					price         float64
					currency      string
					airlineCodes  []string
					createdAt     time.Time
				}{
					departureDate: offer.DepartureDate,
					returnDate:    offer.ReturnDate,
					price:         offer.Price,
					currency:      offer.Currency,
					airlineCodes:  append([]string(nil), offer.AirlineCodes...),
					createdAt:     offer.CreatedAt,
				}
			}
		}

		gridRoutes := make([]routeGrid, 0, len(routeAccumulators))
		routeKeys := make([]string, 0, len(routeAccumulators))
		for key := range routeAccumulators {
			routeKeys = append(routeKeys, key)
		}
		sort.Strings(routeKeys)

		for _, key := range routeKeys {
			acc := routeAccumulators[key]
			cells := make([]gridCell, 0, len(acc.cells))
			for _, cell := range acc.cells {
				var returnDatePtr *time.Time
				if cell.returnDate.Valid {
					rt := cell.returnDate.Time
					returnDatePtr = &rt
				}
				cells = append(cells, gridCell{
					DepartureDate: cell.departureDate,
					ReturnDate:    returnDatePtr,
					Price:         cell.price,
					Currency:      cell.currency,
					AirlineCodes:  append([]string(nil), cell.airlineCodes...),
					CreatedAt:     cell.createdAt,
				})
			}

			sort.Slice(cells, func(i, j int) bool {
				if !cells[i].DepartureDate.Equal(cells[j].DepartureDate) {
					return cells[i].DepartureDate.Before(cells[j].DepartureDate)
				}

				if cells[i].ReturnDate == nil && cells[j].ReturnDate == nil {
					return false
				}
				if cells[i].ReturnDate == nil {
					return true
				}
				if cells[j].ReturnDate == nil {
					return false
				}
				return cells[i].ReturnDate.Before(*cells[j].ReturnDate)
			})

			gridRoutes = append(gridRoutes, routeGrid{
				Origin:      acc.origin,
				Destination: acc.destination,
				Cells:       cells,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"items": offerItems,
			"count": total,
			"grid":  gridRoutes,
		})
	}
}

// createJob returns a handler for creating a new scheduled job
func createJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		var req JobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.CronExpression == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CronExpression is required"})
			return
		}

		// Parse date strings to time.Time objects
		dateStart, err := time.Parse("2006-01-02", req.DateStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_start format. Use YYYY-MM-DD: " + err.Error()})
			return
		}

		dateEnd, err := time.Parse("2006-01-02", req.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_end format. Use YYYY-MM-DD: " + err.Error()})
			return
		}

		// Parse optional return dates if provided
		var returnDateStart, returnDateEnd time.Time
		if req.ReturnDateStart != "" {
			returnDateStart, err = time.Parse("2006-01-02", req.ReturnDateStart)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_start format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
		}

		if req.ReturnDateEnd != "" {
			returnDateEnd, err = time.Parse("2006-01-02", req.ReturnDateEnd)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_end format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
		}

		// Begin a transaction
		tx, err := pgDB.BeginTx(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
			return
		}
		defer tx.Rollback() // Ensure rollback on error

		// Validate cron expression format (basic validation)
		parts := strings.Fields(req.CronExpression)
		if len(parts) != 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression format. Must have 5 space-separated fields: minute hour day month weekday"})
			return
		}
		// TODO: Add more robust cron expression validation if needed (e.g., using a library)

		// Insert the job
		ctx := c.Request.Context()
		jobID, err := pgDB.CreateScheduledJob(ctx, tx, req.Name, req.CronExpression, true) // Assume enabled by default
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled job"})
			return
		}

		// Prepare job details struct
		details := db.JobDetails{
			JobID:              jobID,
			Origin:             req.Origin,
			Destination:        req.Destination,
			DepartureDateStart: dateStart,
			DepartureDateEnd:   dateEnd,
			DynamicDates:       req.DynamicDates,
			Adults:             req.Adults,
			Children:           req.Children,
			InfantsLap:         req.InfantsLap,
			InfantsSeat:        req.InfantsSeat,
			TripType:           req.TripType,
			Class:              req.Class,
			Stops:              req.Stops,
			Currency:           req.Currency, // Assuming Currency is part of JobDetails now
		}
		if req.ReturnDateStart != "" {
			details.ReturnDateStart = sql.NullTime{Time: returnDateStart, Valid: true}
		}
		if req.ReturnDateEnd != "" {
			details.ReturnDateEnd = sql.NullTime{Time: returnDateEnd, Valid: true}
		}
		if req.TripLength > 0 { // Assuming 0 means not set
			details.TripLength = sql.NullInt32{Int32: int32(req.TripLength), Valid: true}
		}
		if req.DaysFromExecution > 0 {
			details.DaysFromExecution = sql.NullInt32{Int32: int32(req.DaysFromExecution), Valid: true}
		}
		if req.SearchWindowDays > 0 {
			details.SearchWindowDays = sql.NullInt32{Int32: int32(req.SearchWindowDays), Valid: true}
		}

		// Insert the job details
		log.Printf("Creating job details for job ID %d: origin=%s, destination=%s, dynamic_dates=%v",
			jobID, req.Origin, req.Destination, req.DynamicDates)
		err = pgDB.CreateJobDetails(ctx, tx, details)
		if err != nil {
			log.Printf("Failed to create job details: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job details"})
			return
		}
		log.Printf("Successfully created job details for job ID %d", jobID)

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
			return
		}

		// The job is saved to the DB. The scheduler should pick it up based on its loading logic.
		// The previous call to scheduler.AddJob was incorrect based on its signature.
		// TODO: Verify scheduler loading logic handles new jobs correctly.

		c.JSON(http.StatusCreated, gin.H{
			"id":      jobID,
			"message": "Job created and scheduled successfully",
		})
	}
}

// createBulkSearch returns a handler for creating a bulk flight search
func CreateBulkSearch(q queue.Queue, pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// If the continuous sweep is running, pause it while bulk searches are active to avoid
		// competing for rate-limited Google endpoints. Auto-resume when the bulk queue drains.
		// (Best-effort; bulk search still runs even if a runner is not configured.)
		if workerManager != nil {
			if runner := workerManager.GetSweepRunner(); runner != nil {
				status := runner.GetStatus()
				if status.IsRunning && !status.IsPaused {
					runner.PauseAndAutoResumeAfterQueueDrain("bulk_search")
				}
			}
		}

		ctx := c.Request.Context()
		var overrides map[string][]string
		var worldAllCount int
		if containsToken(req.Origins, macros.RegionWorldAll) || containsToken(req.Destinations, macros.RegionWorldAll) {
			worldAll, err := listAllAirportCodes(ctx, pgDB)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			overrides = map[string][]string{macros.RegionWorldAll: worldAll}
			worldAllCount = len(worldAll)
		}

		// Expand region tokens in origins and destinations
		expandedOrigins, originWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Origins, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid origin: " + err.Error()})
			return
		}
		expandedDestinations, destinationWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Destinations, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination: " + err.Error()})
			return
		}

		warnings := append([]string{}, originWarnings...)
		warnings = append(warnings, destinationWarnings...)
		if worldAllCount > 0 {
			warnings = append(warnings, fmt.Sprintf("%s expanded to %d airports", macros.RegionWorldAll, worldAllCount))
		}

		// Never schedule origin==destination routes.
		// For region<>region searches (e.g. CARIBBEAN<>CARIBBEAN) this prevents N self-routes (A->A),
		// which are never useful and can add significant work.
		destSet := make(map[string]struct{}, len(expandedDestinations))
		for _, d := range expandedDestinations {
			destSet[d] = struct{}{}
		}
		selfRoutes := 0
		for _, o := range expandedOrigins {
			if _, ok := destSet[o]; ok {
				selfRoutes++
			}
		}

		totalRoutes := len(expandedOrigins)*len(expandedDestinations) - selfRoutes
		if totalRoutes == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "At least one non-self route is required (origin != destination after REGION:* expansion)",
				"warnings": warnings,
			})
			return
		}
		if selfRoutes > 0 {
			warnings = append(warnings, fmt.Sprintf("skipped %d origin==destination route(s)", selfRoutes))
		}

		// Guard against accidental workload explosion from region tokens
		const maxBulkSearchRoutes = 10000
		if totalRoutes > maxBulkSearchRoutes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        fmt.Sprintf("Too many routes: %d (max %d). Region tokens can expand to many airports; consider narrowing your search.", totalRoutes, maxBulkSearchRoutes),
				"total_routes": totalRoutes,
				"max_routes":   maxBulkSearchRoutes,
				"origins":      len(expandedOrigins),
				"destinations": len(expandedDestinations),
				"warnings":     warnings,
			})
			return
		}

		currencyCode := strings.ToUpper(req.Currency)

		// Create bulk search record so the run can be tracked
		bulkSearchID, err := pgDB.CreateBulkSearchRecord(
			ctx,
			sql.NullInt32{}, // No associated scheduled job for on-demand requests
			totalRoutes,
			currencyCode,
			"queued",
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bulk search record: " + err.Error()})
			return
		}

		// Create a worker payload
		payload := worker.BulkSearchPayload{
			Origins:           expandedOrigins,
			Destinations:      expandedDestinations,
			DepartureDateFrom: req.DepartureDateFrom.Time,
			DepartureDateTo:   req.DepartureDateTo.Time,
			ReturnDateFrom:    req.ReturnDateFrom.Time,
			ReturnDateTo:      req.ReturnDateTo.Time,
			TripLength:        req.TripLength,
			Adults:            req.Adults,
			Children:          req.Children,
			InfantsLap:        req.InfantsLap,
			InfantsSeat:       req.InfantsSeat,
			TripType:          req.TripType,
			Class:             req.Class,
			Stops:             req.Stops,
			Currency:          currencyCode,
			Carriers:          req.Carriers,
			BulkSearchID:      bulkSearchID,
		}

		// Enqueue the job
		jobID, err := q.Enqueue(ctx, "bulk_search", payload)
		if err != nil {
			_ = pgDB.UpdateBulkSearchStatus(ctx, bulkSearchID, "failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":         jobID,
			"bulk_search_id": bulkSearchID,
			"message":        "Bulk flight search job created successfully",
			"warnings":       warnings,
		})
	}
}

// updateJob returns a handler for updating a job
func updateJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		var req JobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse date strings (assuming YYYY-MM-DD format from JobRequest)
		dateStart, err := time.Parse("2006-01-02", req.DateStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_start format. Use YYYY-MM-DD: " + err.Error()})
			return
		}
		dateEnd, err := time.Parse("2006-01-02", req.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_end format. Use YYYY-MM-DD: " + err.Error()})
			return
		}
		var returnDateStart, returnDateEnd time.Time
		var returnStartValid, returnEndValid bool
		if req.ReturnDateStart != "" {
			returnDateStart, err = time.Parse("2006-01-02", req.ReturnDateStart)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_start format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
			returnStartValid = true
		}
		if req.ReturnDateEnd != "" {
			returnDateEnd, err = time.Parse("2006-01-02", req.ReturnDateEnd)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_end format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
			returnEndValid = true
		}

		// Begin a transaction
		ctx := c.Request.Context()
		tx, err := pgDB.BeginTx(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback() // Ensure rollback on error

		// Update the job
		err = pgDB.UpdateScheduledJob(ctx, tx, jobID, req.Name, req.CronExpression)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled job"})
			return
		}

		// Prepare job details struct for update
		details := db.JobDetails{
			// JobID is used in the WHERE clause, not SET
			Origin:             req.Origin,
			Destination:        req.Destination,
			DepartureDateStart: dateStart,
			DepartureDateEnd:   dateEnd,
			Adults:             req.Adults,
			Children:           req.Children,
			InfantsLap:         req.InfantsLap,
			InfantsSeat:        req.InfantsSeat,
			TripType:           req.TripType,
			Class:              req.Class,
			Stops:              req.Stops,
			Currency:           req.Currency,
		}
		if returnStartValid {
			details.ReturnDateStart = sql.NullTime{Time: returnDateStart, Valid: true}
		}
		if returnEndValid {
			details.ReturnDateEnd = sql.NullTime{Time: returnDateEnd, Valid: true}
		}
		if req.TripLength > 0 {
			details.TripLength = sql.NullInt32{Int32: int32(req.TripLength), Valid: true}
		}

		// Update the job details
		err = pgDB.UpdateJobDetails(ctx, tx, jobID, details)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job details"})
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
			return
		}

		// Update the job schedule using the scheduler
		// scheduler := workerManager.GetScheduler()
		// if err := scheduler.UpdateJob(jobID, req.CronExpression); err != nil { // TODO: Fix UpdateJob signature/usage
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Job updated in database but rescheduling failed: " + err.Error()})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"id":      jobID,
			"message": "Job updated and rescheduled successfully",
		})
	}
}

// getQueueStatus returns a handler for getting the status of the queue
func GetQueueStatus(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get stats for all queue types
		queueTypes := []string{"flight_search", "bulk_search", "bulk_search_route", "price_graph_sweep", "continuous_price_graph"}
		allStats := make(map[string]map[string]int64)

		for _, queueType := range queueTypes {
			stats, err := q.GetQueueStats(c.Request.Context(), queueType)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			allStats[queueType] = stats
		}

		c.JSON(http.StatusOK, allStats)
	}
}

func isAllowedQueueName(queueName string) bool {
	switch queueName {
	case "flight_search", "bulk_search", "bulk_search_route", "price_graph_sweep", "continuous_price_graph":
		return true
	default:
		return false
	}
}

// GetQueueBacklog returns recent unacked stream entries for a queue (admin/debug endpoint).
func GetQueueBacklog(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		limit := 200
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		jobs, err := q.GetBacklog(c.Request.Context(), queueName, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue": queueName,
			"limit": limit,
			"jobs":  jobs,
		})
	}
}

// ListQueueJobs lists jobs from a queue status set (admin/debug endpoint).
func ListQueueJobs(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		state := c.DefaultQuery("state", "failed")
		limit := 100
		offset := 0
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}
		if v := c.Query("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		jobs, err := q.ListJobs(c.Request.Context(), queueName, state, limit, offset)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":  queueName,
			"state":  state,
			"limit":  limit,
			"offset": offset,
			"count":  len(jobs),
			"jobs":   jobs,
		})
	}
}

// GetQueueJob fetches a persisted job by ID (admin/debug endpoint).
func GetQueueJob(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job id"})
			return
		}

		job, err := q.GetJob(c.Request.Context(), jobID)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue": queueName,
			"job":   job,
		})
	}
}

// GetQueueEnqueueMetrics shows who/what is enqueueing jobs (admin/debug endpoint).
func GetQueueEnqueueMetrics(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		minutes := 60
		if v := c.Query("minutes"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				minutes = n
			}
		}

		stats, err := q.GetEnqueueMetrics(c.Request.Context(), queueName, minutes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"minutes": minutes,
			"sources": stats,
		})
	}
}

// ClearQueue clears pending jobs from a queue (admin/debug endpoint).
func ClearQueue(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		cleared, err := q.ClearQueue(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"cleared": cleared,
			"stats":   stats,
		})
	}
}

// ClearQueueFailed clears failed jobs from a queue (admin/debug endpoint).
func ClearQueueFailed(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		cleared, err := q.ClearFailed(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"cleared": cleared,
			"stats":   stats,
		})
	}
}

// ClearQueueProcessing clears "processing" jobs for a queue (admin/debug endpoint).
func ClearQueueProcessing(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		cleared, err := q.ClearProcessing(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"cleared": cleared,
			"stats":   stats,
		})
	}
}

// RetryQueueFailed retries failed jobs for a queue (admin/debug endpoint).
func RetryQueueFailed(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		limit := 200
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		retried, err := q.RetryFailed(c.Request.Context(), queueName, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"retried": retried,
			"limit":   limit,
			"stats":   stats,
		})
	}
}

// CancelQueueJob requests cancellation for a specific queue job (admin/debug endpoint).
func CancelQueueJob(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job id"})
			return
		}

		if err := q.CancelJob(c.Request.Context(), queueName, jobID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":   queueName,
			"job_id":  jobID,
			"message": "Cancel requested",
		})
	}
}

// CancelQueueProcessing requests cancellation for all currently processing jobs in a queue (admin/debug endpoint).
func CancelQueueProcessing(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		canceled, err := q.CancelProcessing(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":    queueName,
			"canceled": canceled,
			"stats":    stats,
		})
	}
}

// DrainQueue cancels all processing jobs and clears all pending jobs for a queue (admin/debug endpoint).
func DrainQueue(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		queueName := c.Param("name")
		if !isAllowedQueueName(queueName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue name"})
			return
		}

		canceled, err := q.CancelProcessing(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cleared, err := q.ClearQueue(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats, err := q.GetQueueStats(c.Request.Context(), queueName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"queue":    queueName,
			"canceled": canceled,
			"cleared":  cleared,
			"stats":    stats,
		})
	}
}

// getJobById returns a handler for getting a job by ID
func getJobById(pgDB db.PostgresDB) gin.HandlerFunc { // Changed parameter type
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Get the job
		job, err := pgDB.GetJobByID(c.Request.Context(), jobID)
		if err != nil {
			// TODO: Check for specific not found error type
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job: " + err.Error()})
			}
			return
		}

		// Get the job details
		details, err := pgDB.GetJobDetailsByID(c.Request.Context(), jobID)
		if err != nil {
			// If job exists but details don't, it's an inconsistency, but handle gracefully
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job details: " + err.Error()})
			return
		}

		// Build the response
		jobMap := map[string]interface{}{
			"id":              job.ID,
			"name":            job.Name,
			"cron_expression": job.CronExpression,
			"enabled":         job.Enabled,
			"created_at":      job.CreatedAt,
			"updated_at":      job.UpdatedAt,
			"details": map[string]interface{}{
				"origin":               details.Origin,
				"destination":          details.Destination,
				"departure_date_start": details.DepartureDateStart.Format("2006-01-02"), // Format dates
				"departure_date_end":   details.DepartureDateEnd.Format("2006-01-02"),
				"adults":               details.Adults,
				"children":             details.Children,
				"infants_lap":          details.InfantsLap,
				"infants_seat":         details.InfantsSeat,
				"trip_type":            details.TripType,
				"class":                details.Class,
				"stops":                details.Stops,
				"currency":             details.Currency, // Include currency
			},
		}

		if job.LastRun.Valid {
			jobMap["last_run"] = job.LastRun.Time
		} else {
			jobMap["last_run"] = nil
		}

		detailsMap := jobMap["details"].(map[string]interface{})
		if details.ReturnDateStart.Valid {
			detailsMap["return_date_start"] = details.ReturnDateStart.Time.Format("2006-01-02")
		} else {
			detailsMap["return_date_start"] = nil
		}

		if details.ReturnDateEnd.Valid {
			detailsMap["return_date_end"] = details.ReturnDateEnd.Time.Format("2006-01-02")
		} else {
			detailsMap["return_date_end"] = nil
		}

		if details.TripLength.Valid {
			detailsMap["trip_length"] = details.TripLength.Int32
		} else {
			detailsMap["trip_length"] = nil
		}

		c.JSON(http.StatusOK, jobMap)
	}
}

// DirectSearchRequest represents a direct flight search request (non-queued)
type DirectSearchRequest struct {
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

	IncludePriceGraph           bool   `json:"include_price_graph" form:"include_price_graph"`
	PriceGraphWindowDays        int    `json:"price_graph_window_days" form:"price_graph_window_days"`
	PriceGraphDepartureDateFrom string `json:"price_graph_departure_date_from" form:"price_graph_departure_date_from"`
	PriceGraphDepartureDateTo   string `json:"price_graph_departure_date_to" form:"price_graph_departure_date_to"`
	PriceGraphTripLengthDays    int    `json:"price_graph_trip_length_days" form:"price_graph_trip_length_days"`
	PriceGraphTopN              int    `json:"price_graph_top_n" form:"price_graph_top_n"`

	// DebugBatches enables per-batch diagnostics in multi-route searches.
	// Intended for troubleshooting missing origin→destination pairs.
	DebugBatches bool `json:"debug_batches" form:"debug_batches"`

	// Airline group filtering (best-effort).
	// - If IncludeAirlineGroups is non-empty, offers must match at least one included group.
	// - If ExcludeAirlineGroups is non-empty, offers matching any excluded group are removed.
	//
	// Group tokens are like: GROUP:LOW_COST, GROUP:STAR_ALLIANCE, GROUP:ONEWORLD, GROUP:SKYTEAM.
	IncludeAirlineGroups []string `json:"include_airline_groups" form:"include_airline_groups"`
	ExcludeAirlineGroups []string `json:"exclude_airline_groups" form:"exclude_airline_groups"`

	// Carriers optionally restricts results at the Google query level (affects offers and calendar graph).
	// Values: airline IATA 2-letter codes (e.g., "UA", "DL") and/or alliance strings
	// ("STAR_ALLIANCE", "ONEWORLD", "SKYTEAM").
	//
	// NOTE: This is an inclusive filter. Exclusion semantics are not supported here; use group filters
	// for post-fetch filtering.
	Carriers []string `json:"carriers" form:"carriers"`
}

// DirectFlightSearch handles direct flight searches (bypasses queue for immediate results)
func DirectFlightSearch(pgDB db.PostgresDB, neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Direct search flights handler called")

		var searchRequest DirectSearchRequest

		// Try to bind JSON first, then form data if that fails
		if err := c.ShouldBindJSON(&searchRequest); err != nil {
			log.Printf("JSON binding failed, trying form binding: %v", err)
			if err := c.ShouldBind(&searchRequest); err != nil {
				log.Printf("Form binding also failed: %v", err)
			}
		}

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
		if searchRequest.IncludePriceGraph && searchRequest.PriceGraphTopN <= 0 {
			searchRequest.PriceGraphTopN = 10
		}
		if searchRequest.PriceGraphTopN > 50 {
			searchRequest.PriceGraphTopN = 50
		}

		originTokens, destinationTokens, err := ParseRouteInputs(searchRequest.Origin, searchRequest.Destination)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plan, err := PlanDirectSearchDates(time.Now().UTC(), searchRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		departureDate := plan.DepartureDate
		returnDate := plan.ReturnDate
		priceGraphOnly := plan.PriceGraphOnly
		if priceGraphOnly {
			if strings.TrimSpace(searchRequest.PriceGraphDepartureDateFrom) == "" {
				searchRequest.PriceGraphDepartureDateFrom = plan.PriceGraph.DepartureDateFrom
			}
			if strings.TrimSpace(searchRequest.PriceGraphDepartureDateTo) == "" {
				searchRequest.PriceGraphDepartureDateTo = plan.PriceGraph.DepartureDateTo
			}
			if searchRequest.PriceGraphTripLengthDays <= 0 {
				searchRequest.PriceGraphTripLengthDays = plan.PriceGraph.TripLengthDays
			}
		}

		// Create flight session
		session, err := flights.New()
		if err != nil {
			log.Printf("Error creating flight session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize flight search"})
			return
		}

		searchWorker := worker.NewWorker(pgDB, neo4jDB)
		persistOffers := func(routeOrigin, routeDestination string, offers []flights.FullOffer, priceRange *flights.PriceRange) {
			ctx := c.Request.Context()
			payload := worker.FlightSearchPayload{
				Origin:        routeOrigin,
				Destination:   routeDestination,
				DepartureDate: departureDate,
				ReturnDate:    returnDate,
				Adults:        searchRequest.Adults,
				Children:      searchRequest.Children,
				InfantsLap:    searchRequest.InfantsLap,
				InfantsSeat:   searchRequest.InfantsSeat,
				TripType:      searchRequest.TripType,
				Class:         searchRequest.Class,
				Stops:         searchRequest.Stops,
				Currency:      searchRequest.Currency,
			}

			switch {
			case pgDB != nil:
				if err := searchWorker.StoreFlightOffers(ctx, payload, offers, priceRange); err != nil {
					log.Printf("Failed to persist direct search %s->%s in Postgres: %v", routeOrigin, routeDestination, err)
				}
			case neo4jDB != nil:
				for _, offer := range offers {
					if err := searchWorker.StoreFlightInNeo4j(ctx, offer, searchRequest.Class); err != nil {
						log.Printf("Failed to persist direct search %s->%s in Neo4j: %v", routeOrigin, routeDestination, err)
						break
					}
				}
			default:
				return
			}
		}

		// Map trip type
		var tripType flights.TripType
		switch searchRequest.TripType {
		case "one_way":
			tripType = flights.OneWay
		case "round_trip":
			tripType = flights.RoundTrip
		default:
			tripType = flights.RoundTrip
		}

		// Map class
		var class flights.Class
		switch searchRequest.Class {
		case "economy":
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
		case "any":
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

		baseOptions := flights.Options{
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
		}

		priceGraphParams := plan.PriceGraph

		normalizeGroupTokens := func(tokens []string) map[string]struct{} {
			out := make(map[string]struct{}, len(tokens))
			for _, raw := range tokens {
				token := strings.ToUpper(strings.TrimSpace(raw))
				if token == "" {
					continue
				}
				if !macros.IsAirlineGroupToken(token) {
					continue
				}
				out[token] = struct{}{}
			}
			return out
		}

		includeGroups := normalizeGroupTokens(searchRequest.IncludeAirlineGroups)
		excludeGroups := normalizeGroupTokens(searchRequest.ExcludeAirlineGroups)

		googleCarrierFilterFromGroups := func(include, exclude map[string]struct{}) []string {
			// Only support a single alliance include filter for now.
			// If you need AND/NOT semantics, keep using post-fetch filtering.
			if len(exclude) > 0 {
				return nil
			}
			if _, ok := include[macros.GroupLowCost]; ok {
				return nil
			}

			selected := ""
			for _, token := range []string{macros.GroupStarAlliance, macros.GroupOneworld, macros.GroupSkyTeam} {
				if _, ok := include[token]; !ok {
					continue
				}
				if selected != "" {
					return nil
				}
				selected = token
			}
			switch selected {
			case macros.GroupStarAlliance:
				return []string{"STAR_ALLIANCE"}
			case macros.GroupOneworld:
				return []string{"ONEWORLD"}
			case macros.GroupSkyTeam:
				return []string{"SKYTEAM"}
			default:
				return nil
			}
		}

		baseOptions.Carriers = googleCarrierFilterFromGroups(includeGroups, excludeGroups)
		if len(searchRequest.Carriers) > 0 {
			baseOptions.Carriers = append([]string{}, searchRequest.Carriers...)
		}

		maybeGetPriceGraph := func(routeOrigin, routeDestination string) map[string]interface{} {
			if !searchRequest.IncludePriceGraph {
				return nil
			}

			args, err := BuildPriceGraphArgs(time.Now().UTC(), routeOrigin, routeDestination, departureDate, returnDate, baseOptions, priceGraphParams)
			if err != nil {
				return SerializePriceGraphResponse(routeOrigin, routeDestination, searchRequest.Currency, flights.PriceGraphArgs{}, nil, nil, err)
			}

			offers, parseErrors, pgErr := session.GetPriceGraph(c.Request.Context(), args)
			response := SerializePriceGraphResponse(routeOrigin, routeDestination, searchRequest.Currency, args, offers, parseErrors, pgErr)
			if len(searchRequest.IncludeAirlineGroups) > 0 || len(searchRequest.ExcludeAirlineGroups) > 0 {
				if len(baseOptions.Carriers) == 0 {
					response["filter_warning"] = "Google price graph could not apply the selected airline-group filters; it reflects all carriers returned by Google for the route/date range."
				} else {
					response["filter_note"] = map[string]interface{}{
						"google_carriers": baseOptions.Carriers,
					}
				}
				response["filter"] = map[string]interface{}{
					"include_airline_groups": searchRequest.IncludeAirlineGroups,
					"exclude_airline_groups": searchRequest.ExcludeAirlineGroups,
				}
			}
			if pgErr != nil {
				return response
			}

			points, ok := response["points"].([]map[string]interface{})
			if !ok || len(points) == 0 {
				return response
			}

			for i := range points {
				if i >= len(offers) {
					break
				}

				var returnDate sql.NullTime
				if !offers[i].ReturnDate.IsZero() {
					returnDate = sql.NullTime{Time: offers[i].ReturnDate, Valid: true}
				}

				url, urlErr := session.SerializeURL(
					c.Request.Context(),
					flights.Args{
						Date:        offers[i].StartDate,
						ReturnDate:  offers[i].ReturnDate,
						SrcAirports: []string{routeOrigin},
						DstAirports: []string{routeDestination},
						Options:     baseOptions,
					},
				)
				if urlErr == nil && url != "" {
					points[i]["google_flights_url"] = url
				}

				// Best-effort persistence of price graph points for history/tracking.
				// We store these in Postgres (price_graph_results) under sweep_id=0.
				if pgDB != nil && offers[i].Price > 0 {
					record := db.PriceGraphResultRecord{
						SweepID:       0,
						Origin:        routeOrigin,
						Destination:   routeDestination,
						DepartureDate: offers[i].StartDate,
						ReturnDate:    returnDate,
						TripLength:    sql.NullInt32{Int32: int32(args.TripLength), Valid: true},
						Price:         offers[i].Price,
						Currency:      searchRequest.Currency,
						Adults:        searchRequest.Adults,
						Children:      searchRequest.Children,
						InfantsLap:    searchRequest.InfantsLap,
						InfantsSeat:   searchRequest.InfantsSeat,
						TripType:      searchRequest.TripType,
						Class:         searchRequest.Class,
						Stops:         searchRequest.Stops,
						SearchURL:     sql.NullString{String: url, Valid: url != ""},
						QueriedAt:     time.Now().UTC(),
					}
					if insertErr := pgDB.InsertPriceGraphResult(c.Request.Context(), record); insertErr != nil {
						log.Printf("Failed to persist price graph point %s->%s %s: %v",
							routeOrigin, routeDestination, offers[i].StartDate.Format("2006-01-02"), insertErr)
					}
				}
			}
			response["points"] = points
			if searchRequest.PriceGraphTopN > 0 {
				if top := TopPriceGraphPoints(points, searchRequest.PriceGraphTopN); len(top) > 0 {
					response["top_points"] = top
				}
			}
			return response
		}

		offerAirlineCodes := func(offer flights.FullOffer) []string {
			out := make([]string, 0, len(offer.Flight)+len(offer.ReturnFlight))
			for _, flight := range offer.Flight {
				code := macros.ExtractAirlineCodeFromFlightNumber(flight.FlightNumber)
				if code != "" {
					out = append(out, code)
				}
			}
			for _, flight := range offer.ReturnFlight {
				code := macros.ExtractAirlineCodeFromFlightNumber(flight.FlightNumber)
				if code != "" {
					out = append(out, code)
				}
			}
			return out
		}

		shouldIncludeOffer := func(offer flights.FullOffer) bool {
			if len(includeGroups) == 0 && len(excludeGroups) == 0 {
				return true
			}

			codes := offerAirlineCodes(offer)
			groups := macros.AirlineGroupsForCodes(codes)

			for _, g := range groups {
				if _, ok := excludeGroups[g]; ok {
					return false
				}
			}

			if len(includeGroups) == 0 {
				return true
			}
			for _, g := range groups {
				if _, ok := includeGroups[g]; ok {
					return true
				}
			}
			return false
		}

		filterOffers := func(offers []flights.FullOffer) (kept []flights.FullOffer, filteredOut int) {
			if len(includeGroups) == 0 && len(excludeGroups) == 0 {
				return offers, 0
			}
			out := make([]flights.FullOffer, 0, len(offers))
			for _, offer := range offers {
				if shouldIncludeOffer(offer) {
					out = append(out, offer)
				}
			}
			return out, len(offers) - len(out)
		}

		expandedOrigins, originWarnings, err := macros.ExpandAirportTokens(originTokens)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid origin: " + err.Error()})
			return
		}
		expandedDestinations, destinationWarnings, err := macros.ExpandAirportTokens(destinationTokens)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination: " + err.Error()})
			return
		}

		warnings := append([]string{}, originWarnings...)
		warnings = append(warnings, destinationWarnings...)

		totalRoutes := len(expandedOrigins) * len(expandedDestinations)
		const maxDirectAirportsTotal = 60
		// NOTE: We enforce a *total* expansion cap to keep the direct search endpoint safe/fast.
		// Per-side caps are intentionally kept equal to the total cap so a single large region token
		// (e.g., REGION:CARIBBEAN) can still be used as either origins or destinations.
		const maxDirectAirportsPerSide = maxDirectAirportsTotal
		if len(expandedOrigins) > maxDirectAirportsPerSide || len(expandedDestinations) > maxDirectAirportsPerSide || (len(expandedOrigins)+len(expandedDestinations)) > maxDirectAirportsTotal {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"too many airports after expansion (origins=%d, destinations=%d; max %d total across origins+destinations). Narrow your inputs or use /api/v1/bulk-search for large grids.",
					len(expandedOrigins),
					len(expandedDestinations),
					maxDirectAirportsTotal,
				),
				"warnings": warnings,
			})
			return
		}

		convertOffers := func(routeOrigin, routeDestination string, offers []flights.FullOffer) ([]map[string]interface{}, float64, map[string]interface{}, bool) {
			responseOffers := make([]map[string]interface{}, 0, len(offers))
			var cheapest map[string]interface{}
			var cheapestPrice float64
			cheapestSet := false

			for i, offer := range offers {
				segments := make([]map[string]interface{}, 0, len(offer.Flight))
				airlineCodes := make([]string, 0, len(offer.Flight))
				for _, flight := range offer.Flight {
					airlineCode := macros.ExtractAirlineCodeFromFlightNumber(flight.FlightNumber)
					segment := map[string]interface{}{
						"departure_airport": flight.DepAirportCode,
						"arrival_airport":   flight.ArrAirportCode,
						"departure_time":    flight.DepTime.Format(time.RFC3339),
						"arrival_time":      flight.ArrTime.Format(time.RFC3339),
						"airline":           flight.AirlineName,
						"airline_code":      airlineCode,
						"flight_number":     flight.FlightNumber,
						"duration":          int(flight.Duration.Minutes()),
						"airplane":          flight.Airplane,
						"legroom":           flight.Legroom,
					}
					segments = append(segments, segment)
					if airlineCode != "" {
						airlineCodes = append(airlineCodes, airlineCode)
					}
				}
				for _, flight := range offer.ReturnFlight {
					airlineCode := macros.ExtractAirlineCodeFromFlightNumber(flight.FlightNumber)
					if airlineCode != "" {
						airlineCodes = append(airlineCodes, airlineCode)
					}
				}

				airlineGroups := macros.AirlineGroupsForCodes(airlineCodes)

				googleFlightsUrl, err := session.SerializeURL(
					c.Request.Context(),
					flights.Args{
						Date:        offer.StartDate,
						ReturnDate:  offer.ReturnDate,
						SrcAirports: []string{routeOrigin},
						DstAirports: []string{routeDestination},
						Options:     baseOptions,
					},
				)
				if err != nil {
					log.Printf("Error generating Google Flights URL: %v", err)
					googleFlightsUrl = ""
				}

				responseOffer := map[string]interface{}{
					"id":                 fmt.Sprintf("offer%d", i+1),
					"price":              offer.Price,
					"price_available":    offer.Price > 0,
					"currency":           searchRequest.Currency,
					"total_duration":     int(offer.FlightDuration.Minutes()),
					"segments":           segments,
					"departure_date":     offer.StartDate.Format("2006-01-02"),
					"return_date":        offer.ReturnDate.Format("2006-01-02"),
					"google_flights_url": googleFlightsUrl,
					"airline_groups":     airlineGroups,
				}

				if offer.Price > 0 && (!cheapestSet || offer.Price < cheapestPrice) {
					cheapestSet = true
					cheapestPrice = offer.Price
					cheapest = responseOffer
				}

				responseOffers = append(responseOffers, responseOffer)
			}

			return responseOffers, cheapestPrice, cheapest, cheapestSet
		}

		// Single route: preserve the legacy response shape.
		if len(expandedOrigins) == 1 && len(expandedDestinations) == 1 {
			searchRequest.Origin = expandedOrigins[0]
			searchRequest.Destination = expandedDestinations[0]

			if priceGraphOnly {
				response := map[string]interface{}{
					"offers":           []map[string]interface{}{},
					"search_params":    searchRequest,
					"price_graph_only": true,
				}

				if priceGraph := maybeGetPriceGraph(searchRequest.Origin, searchRequest.Destination); priceGraph != nil {
					response["price_graph"] = priceGraph
				}

				c.JSON(http.StatusOK, response)
				return
			}

			offers, priceRange, err := session.GetOffers(
				c.Request.Context(),
				flights.Args{
					Date:        departureDate,
					ReturnDate:  returnDate,
					SrcAirports: []string{searchRequest.Origin},
					DstAirports: []string{searchRequest.Destination},
					Options:     baseOptions,
				},
			)

			if err != nil {
				log.Printf("Error searching flights: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search flights: " + err.Error()})
				return
			}

			persistOffers(searchRequest.Origin, searchRequest.Destination, offers, priceRange)

			filteredOffers, filteredOut := filterOffers(offers)
			responseOffers, _, _, _ := convertOffers(searchRequest.Origin, searchRequest.Destination, filteredOffers)
			response := map[string]interface{}{
				"offers":        responseOffers,
				"search_params": searchRequest,
			}
			if filteredOut > 0 {
				response["filter"] = map[string]interface{}{
					"include_airline_groups": searchRequest.IncludeAirlineGroups,
					"exclude_airline_groups": searchRequest.ExcludeAirlineGroups,
					"filtered_out":           filteredOut,
				}
			}

			if priceRange != nil {
				response["price_range"] = map[string]interface{}{
					"low":  priceRange.Low,
					"high": priceRange.High,
				}
			}

			if priceGraph := maybeGetPriceGraph(searchRequest.Origin, searchRequest.Destination); priceGraph != nil {
				response["price_graph"] = priceGraph
			}

			c.JSON(http.StatusOK, response)
			return
		}

		if priceGraphOnly {
			const maxDirectRouteFallback = 16
			if totalRoutes > maxDirectRouteFallback {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"price-graph-only mode supports up to %d routes per request (requested %d). Narrow your route grid or use the admin price-graph sweep pipeline.",
						maxDirectRouteFallback,
						totalRoutes,
					),
					"warnings": warnings,
				})
				return
			}

			routes := make([]map[string]interface{}, 0, totalRoutes)
			for _, origin := range expandedOrigins {
				for _, destination := range expandedDestinations {
					route := map[string]interface{}{
						"origin":      origin,
						"destination": destination,
					}

					routeParams := searchRequest
					routeParams.Origin = origin
					routeParams.Destination = destination
					route["search_params"] = routeParams

					if priceGraph := maybeGetPriceGraph(origin, destination); priceGraph != nil {
						route["google_price_graph"] = priceGraph

						if top, ok := priceGraph["top_points"].([]map[string]interface{}); ok && len(top) > 0 {
							first := top[0]
							summary := map[string]interface{}{
								"currency":       priceGraph["currency"],
								"price":          first["price"],
								"departure_date": first["departure_date"],
								"return_date":    first["return_date"],
							}
							route["summary"] = summary
						}
					}

					routes = append(routes, route)
				}
			}

			response := map[string]interface{}{
				"routes": routes,
				"search_params": map[string]interface{}{
					"origins":               expandedOrigins,
					"destinations":          expandedDestinations,
					"departure_date":        searchRequest.DepartureDate,
					"return_date":           searchRequest.ReturnDate,
					"trip_type":             searchRequest.TripType,
					"class":                 searchRequest.Class,
					"stops":                 searchRequest.Stops,
					"adults":                searchRequest.Adults,
					"children":              searchRequest.Children,
					"infants_lap":           searchRequest.InfantsLap,
					"infants_seat":          searchRequest.InfantsSeat,
					"currency":              searchRequest.Currency,
					"requested_route_count": totalRoutes,
					"returned_route_count":  len(routes),
					"price_graph_only":      true,
				},
				"price_graph_only": true,
			}
			if len(searchRequest.IncludeAirlineGroups) > 0 || len(searchRequest.ExcludeAirlineGroups) > 0 {
				response["filter"] = map[string]interface{}{
					"include_airline_groups": searchRequest.IncludeAirlineGroups,
					"exclude_airline_groups": searchRequest.ExcludeAirlineGroups,
				}
			}
			if len(warnings) > 0 {
				response["warnings"] = warnings
			}

			c.JSON(http.StatusOK, response)
			return
		}

		// Multi-route: single query with multiple Src/Dst airports, then group offers by route.
		// Fallback to per-route queries for small grids, since Google responses can vary with multi-airport queries.
		const maxDirectRouteFallback = 16
		if totalRoutes <= maxDirectRouteFallback {
			routes := make([]map[string]interface{}, 0, totalRoutes)
			var overallCheapestOffer map[string]interface{}
			var overallCheapestOrigin string
			var overallCheapestDestination string
			var overallCheapestPrice float64
			overallCheapestSet := false

			for _, origin := range expandedOrigins {
				for _, destination := range expandedDestinations {
					offers, priceRange, err := session.GetOffers(
						c.Request.Context(),
						flights.Args{
							Date:        departureDate,
							ReturnDate:  returnDate,
							SrcAirports: []string{origin},
							DstAirports: []string{destination},
							Options:     baseOptions,
						},
					)

					route := map[string]interface{}{
						"origin":      origin,
						"destination": destination,
					}

					if err != nil {
						log.Printf("Error searching flights for %s->%s: %v", origin, destination, err)
						route["error"] = "Failed to search flights: " + err.Error()
						routes = append(routes, route)
						continue
					}

					persistOffers(origin, destination, offers, priceRange)

					filteredOffers, filteredOut := filterOffers(offers)
					routeOffers, routeCheapestPrice, routeCheapestOffer, routeCheapestSet := convertOffers(origin, destination, filteredOffers)
					route["offers"] = routeOffers
					if filteredOut > 0 {
						route["filter"] = map[string]interface{}{
							"filtered_out": filteredOut,
						}
						if len(routeOffers) == 0 {
							route["warning"] = "All offers were filtered out by airline group filters."
						}
					}

					routeParams := searchRequest
					routeParams.Origin = origin
					routeParams.Destination = destination
					route["search_params"] = routeParams

					if priceRange != nil {
						route["price_range"] = map[string]interface{}{
							"low":  priceRange.Low,
							"high": priceRange.High,
						}
					}

					if routeCheapestSet && routeCheapestOffer != nil && (!overallCheapestSet || routeCheapestPrice < overallCheapestPrice) {
						overallCheapestSet = true
						overallCheapestPrice = routeCheapestPrice
						overallCheapestOffer = routeCheapestOffer
						overallCheapestOrigin = origin
						overallCheapestDestination = destination
					}

					routes = append(routes, route)
				}
			}

			response := map[string]interface{}{
				"routes": routes,
				"search_params": map[string]interface{}{
					"origins":               expandedOrigins,
					"destinations":          expandedDestinations,
					"departure_date":        searchRequest.DepartureDate,
					"return_date":           searchRequest.ReturnDate,
					"trip_type":             searchRequest.TripType,
					"class":                 searchRequest.Class,
					"stops":                 searchRequest.Stops,
					"adults":                searchRequest.Adults,
					"children":              searchRequest.Children,
					"infants_lap":           searchRequest.InfantsLap,
					"infants_seat":          searchRequest.InfantsSeat,
					"currency":              searchRequest.Currency,
					"requested_route_count": totalRoutes,
					"returned_route_count":  len(routes),
				},
			}
			if len(searchRequest.IncludeAirlineGroups) > 0 || len(searchRequest.ExcludeAirlineGroups) > 0 {
				response["filter"] = map[string]interface{}{
					"include_airline_groups": searchRequest.IncludeAirlineGroups,
					"exclude_airline_groups": searchRequest.ExcludeAirlineGroups,
				}
			}
			if len(warnings) > 0 {
				response["warnings"] = warnings
			}
			if overallCheapestSet && overallCheapestOffer != nil {
				response["cheapest"] = map[string]interface{}{
					"origin":      overallCheapestOrigin,
					"destination": overallCheapestDestination,
					"offer":       overallCheapestOffer,
				}
			}

			if overallCheapestSet {
				if priceGraph := maybeGetPriceGraph(overallCheapestOrigin, overallCheapestDestination); priceGraph != nil {
					response["price_graph"] = priceGraph
				}
			}

			c.JSON(http.StatusOK, response)
			return
		}

		var overallCheapestOffer map[string]interface{}
		var overallCheapestOrigin string
		var overallCheapestDestination string
		var overallCheapestPrice float64
		overallCheapestSet := false

		// For large grids, do batched queries by the smaller side (e.g., per-origin with many destinations).
		// This is more reliable than sending many Src and many Dst airports in a single Google Flights request.
		offers := make([]flights.FullOffer, 0, 128)
		chunkStrings := func(values []string, size int) [][]string {
			if size <= 0 || len(values) == 0 {
				return nil
			}
			chunks := make([][]string, 0, (len(values)+size-1)/size)
			for i := 0; i < len(values); i += size {
				end := i + size
				if end > len(values) {
					end = len(values)
				}
				chunks = append(chunks, values[i:end])
			}
			return chunks
		}

		const maxAirportsPerRequestSide = 10
		const maxBatchedRequests = 64

		type batchDiagnostics struct {
			Mode         string `json:"mode"`
			ChunkSize    int    `json:"chunk_size"`
			TotalBatches int    `json:"total_batches"`
			Failed       int    `json:"failed_batches"`
			Empty        int    `json:"empty_batches"`
			TotalOffers  int    `json:"total_offers"`
			DurationMs   int64  `json:"duration_ms"`
		}

		type batchDebugEntry struct {
			Mode                string         `json:"mode"`
			Origin              string         `json:"origin,omitempty"`
			Destination         string         `json:"destination,omitempty"`
			Origins             []string       `json:"origins,omitempty"`
			Destinations        []string       `json:"destinations,omitempty"`
			Offers              int            `json:"offers"`
			DurationMs          int64          `json:"duration_ms"`
			Error               string         `json:"error,omitempty"`
			DestinationOfferMap map[string]int `json:"destination_offer_map,omitempty"`
			OriginOfferMap      map[string]int `json:"origin_offer_map,omitempty"`
			MissingDestinations []string       `json:"missing_destinations,omitempty"`
			MissingOrigins      []string       `json:"missing_origins,omitempty"`
		}

		diag := batchDiagnostics{ChunkSize: maxAirportsPerRequestSide}
		debugEntries := make([]batchDebugEntry, 0, 16)
		batchStartOverall := time.Now()

		queryArgs := func(srcAirports, dstAirports []string) flights.Args {
			return flights.Args{
				Date:        departureDate,
				ReturnDate:  returnDate,
				SrcAirports: srcAirports,
				DstAirports: dstAirports,
				Options:     baseOptions,
			}
		}

		appendDebug := func(entry batchDebugEntry) {
			if !searchRequest.DebugBatches {
				return
			}
			// Avoid leaking giant payloads: cap stored codes.
			const maxCodes = 12
			if len(entry.Origins) > maxCodes {
				entry.Origins = append(entry.Origins[:maxCodes], fmt.Sprintf("…(+%d)", len(entry.Origins)-maxCodes))
			}
			if len(entry.Destinations) > maxCodes {
				entry.Destinations = append(entry.Destinations[:maxCodes], fmt.Sprintf("…(+%d)", len(entry.Destinations)-maxCodes))
			}
			debugEntries = append(debugEntries, entry)
		}

		if len(expandedOrigins) <= len(expandedDestinations) {
			diag.Mode = "origin_x_dest_chunks"
			destChunks := chunkStrings(expandedDestinations, maxAirportsPerRequestSide)
			totalBatches := len(expandedOrigins) * len(destChunks)
			diag.TotalBatches = totalBatches
			if totalBatches > maxBatchedRequests {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"too many batched requests to execute safely (%d; max %d). Narrow your inputs or use /api/v1/bulk-search.",
						totalBatches,
						maxBatchedRequests,
					),
					"warnings": warnings,
				})
				return
			}

			for _, origin := range expandedOrigins {
				for _, destChunk := range destChunks {
					batchStart := time.Now()
					batchOffers, _, err := session.GetOffers(c.Request.Context(), queryArgs([]string{origin}, destChunk))
					batchMs := time.Since(batchStart).Milliseconds()
					if err != nil {
						log.Printf("Error searching flights for %s->* (chunk): %v", origin, err)
						warnings = append(warnings, fmt.Sprintf("search failed for origin %s: %v", origin, err))
						diag.Failed++
						appendDebug(batchDebugEntry{
							Mode:         diag.Mode,
							Origin:       origin,
							Destinations: destChunk,
							Offers:       0,
							DurationMs:   batchMs,
							Error:        err.Error(),
						})
						continue
					}
					diag.TotalOffers += len(batchOffers)
					if len(batchOffers) == 0 {
						diag.Empty++
					}
					destCounts := make(map[string]int, len(destChunk))
					for _, code := range destChunk {
						destCounts[code] = 0
					}
					for _, offer := range batchOffers {
						if offer.DstAirportCode == "" {
							continue
						}
						destCounts[offer.DstAirportCode]++
					}
					missingDests := make([]string, 0)
					for _, code := range destChunk {
						if destCounts[code] == 0 {
							missingDests = append(missingDests, code)
						}
					}
					appendDebug(batchDebugEntry{
						Mode:                diag.Mode,
						Origin:              origin,
						Destinations:        destChunk,
						Offers:              len(batchOffers),
						DurationMs:          batchMs,
						DestinationOfferMap: destCounts,
						MissingDestinations: missingDests,
					})
					offers = append(offers, batchOffers...)
				}
			}
		} else {
			diag.Mode = "dest_x_origin_chunks"
			originChunks := chunkStrings(expandedOrigins, maxAirportsPerRequestSide)
			totalBatches := len(expandedDestinations) * len(originChunks)
			diag.TotalBatches = totalBatches
			if totalBatches > maxBatchedRequests {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"too many batched requests to execute safely (%d; max %d). Narrow your inputs or use /api/v1/bulk-search.",
						totalBatches,
						maxBatchedRequests,
					),
					"warnings": warnings,
				})
				return
			}

			for _, destination := range expandedDestinations {
				for _, originChunk := range originChunks {
					batchStart := time.Now()
					batchOffers, _, err := session.GetOffers(c.Request.Context(), queryArgs(originChunk, []string{destination}))
					batchMs := time.Since(batchStart).Milliseconds()
					if err != nil {
						log.Printf("Error searching flights for *->%s (chunk): %v", destination, err)
						warnings = append(warnings, fmt.Sprintf("search failed for destination %s: %v", destination, err))
						diag.Failed++
						appendDebug(batchDebugEntry{
							Mode:        diag.Mode,
							Destination: destination,
							Origins:     originChunk,
							Offers:      0,
							DurationMs:  batchMs,
							Error:       err.Error(),
						})
						continue
					}
					diag.TotalOffers += len(batchOffers)
					if len(batchOffers) == 0 {
						diag.Empty++
					}
					originCounts := make(map[string]int, len(originChunk))
					for _, code := range originChunk {
						originCounts[code] = 0
					}
					for _, offer := range batchOffers {
						if offer.SrcAirportCode == "" {
							continue
						}
						originCounts[offer.SrcAirportCode]++
					}
					missingOrigins := make([]string, 0)
					for _, code := range originChunk {
						if originCounts[code] == 0 {
							missingOrigins = append(missingOrigins, code)
						}
					}
					appendDebug(batchDebugEntry{
						Mode:           diag.Mode,
						Destination:    destination,
						Origins:        originChunk,
						Offers:         len(batchOffers),
						DurationMs:     batchMs,
						OriginOfferMap: originCounts,
						MissingOrigins: missingOrigins,
					})
					offers = append(offers, batchOffers...)
				}
			}
		}

		diag.DurationMs = time.Since(batchStartOverall).Milliseconds()

		offersByRoute := make(map[string][]flights.FullOffer, len(offers))
		for _, offer := range offers {
			if offer.SrcAirportCode == "" || offer.DstAirportCode == "" {
				continue
			}
			key := offer.SrcAirportCode + "->" + offer.DstAirportCode
			offersByRoute[key] = append(offersByRoute[key], offer)
		}

		type routeKey struct {
			origin      string
			destination string
		}

		routeKeys := make([]routeKey, 0, len(offersByRoute))
		for key := range offersByRoute {
			parts := strings.SplitN(key, "->", 2)
			if len(parts) != 2 {
				continue
			}
			routeKeys = append(routeKeys, routeKey{origin: parts[0], destination: parts[1]})
		}

		sort.Slice(routeKeys, func(i, j int) bool {
			if routeKeys[i].origin == routeKeys[j].origin {
				return routeKeys[i].destination < routeKeys[j].destination
			}
			return routeKeys[i].origin < routeKeys[j].origin
		})

		routes := make([]map[string]interface{}, 0, len(routeKeys))
		for _, rk := range routeKeys {
			key := rk.origin + "->" + rk.destination
			fullOffers := offersByRoute[key]
			if len(fullOffers) == 0 {
				continue
			}

			persistOffers(rk.origin, rk.destination, fullOffers, nil)

			filteredOffers, filteredOut := filterOffers(fullOffers)
			routeOffers, routeCheapestPrice, routeCheapestOffer, routeCheapestSet := convertOffers(rk.origin, rk.destination, filteredOffers)
			route := map[string]interface{}{
				"origin":      rk.origin,
				"destination": rk.destination,
				"offers":      routeOffers,
			}
			if filteredOut > 0 {
				route["filter"] = map[string]interface{}{
					"filtered_out": filteredOut,
				}
				if len(routeOffers) == 0 {
					route["warning"] = "All offers were filtered out by airline group filters."
				}
			}

			routeParams := searchRequest
			routeParams.Origin = rk.origin
			routeParams.Destination = rk.destination
			route["search_params"] = routeParams

			low := 0.0
			high := 0.0
			haveRange := false
			for _, offer := range routeOffers {
				price, ok := offer["price"].(float64)
				if !ok || price <= 0 {
					continue
				}
				if !haveRange {
					haveRange = true
					low = price
					high = price
					continue
				}
				if price < low {
					low = price
				}
				if price > high {
					high = price
				}
			}
			if haveRange {
				route["price_range"] = map[string]interface{}{"low": low, "high": high}
			}

			if routeCheapestSet && routeCheapestOffer != nil && (!overallCheapestSet || routeCheapestPrice < overallCheapestPrice) {
				overallCheapestSet = true
				overallCheapestPrice = routeCheapestPrice
				overallCheapestOffer = routeCheapestOffer
				overallCheapestOrigin = rk.origin
				overallCheapestDestination = rk.destination
			}

			routes = append(routes, route)
		}

		if len(offers) > 0 && len(routes) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"search returned %d offers, but none could be grouped into routes; try narrowing the search or switching to /api/v1/bulk-search",
					len(offers),
				),
				"warnings":      warnings,
				"batch_summary": diag,
				"batch_debug":   debugEntries,
			})
			return
		}

		response := map[string]interface{}{
			"routes":        routes,
			"batch_summary": diag,
			"search_params": map[string]interface{}{
				"origins":               expandedOrigins,
				"destinations":          expandedDestinations,
				"departure_date":        searchRequest.DepartureDate,
				"return_date":           searchRequest.ReturnDate,
				"trip_type":             searchRequest.TripType,
				"class":                 searchRequest.Class,
				"stops":                 searchRequest.Stops,
				"adults":                searchRequest.Adults,
				"children":              searchRequest.Children,
				"infants_lap":           searchRequest.InfantsLap,
				"infants_seat":          searchRequest.InfantsSeat,
				"currency":              searchRequest.Currency,
				"offers_count":          len(offers),
				"requested_route_count": totalRoutes,
				"returned_route_count":  len(routes),
			},
		}
		if len(searchRequest.IncludeAirlineGroups) > 0 || len(searchRequest.ExcludeAirlineGroups) > 0 {
			response["filter"] = map[string]interface{}{
				"include_airline_groups": searchRequest.IncludeAirlineGroups,
				"exclude_airline_groups": searchRequest.ExcludeAirlineGroups,
			}
		}
		if searchRequest.DebugBatches {
			response["batch_debug"] = debugEntries
		}
		if len(warnings) > 0 {
			response["warnings"] = warnings
		}

		if overallCheapestSet && overallCheapestOffer != nil {
			response["cheapest"] = map[string]interface{}{
				"origin":      overallCheapestOrigin,
				"destination": overallCheapestDestination,
				"offer":       overallCheapestOffer,
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// CachedAirportsHandler provides cached airport data
func CachedAirportsHandler(cacheManager *cache.CacheManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey := cache.AirportsKey()

		// Try to get from cache first
		var airports []map[string]string
		err := cacheManager.GetJSON(c.Request.Context(), cacheKey, &airports)
		if err == nil {
			logger.WithField("cache_key", cacheKey).Debug("Airport data served from cache")
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, airports)
			return
		}

		if err != cache.ErrCacheMiss {
			logger.WithField("cache_key", cacheKey).Error(err, "Cache get error for airports")
		}

		// Cache miss - generate data and cache it
		airports = []map[string]string{
			{"code": "JFK", "name": "John F. Kennedy International Airport", "city": "New York"},
			{"code": "LAX", "name": "Los Angeles International Airport", "city": "Los Angeles"},
			{"code": "ORD", "name": "O'Hare International Airport", "city": "Chicago"},
			{"code": "LHR", "name": "Heathrow Airport", "city": "London"},
			{"code": "CDG", "name": "Charles de Gaulle Airport", "city": "Paris"},
			{"code": "DXB", "name": "Dubai International Airport", "city": "Dubai"},
			{"code": "NRT", "name": "Narita International Airport", "city": "Tokyo"},
			{"code": "SIN", "name": "Singapore Changi Airport", "city": "Singapore"},
		}

		// Cache the data for 24 hours
		if err := cacheManager.SetJSON(c.Request.Context(), cacheKey, airports, cache.LongTTL); err != nil {
			logger.WithField("cache_key", cacheKey).Error(err, "Cache set error for airports")
		} else {
			logger.WithField("cache_key", cacheKey).Debug("Airport data cached")
		}

		c.Header("X-Cache", "MISS")
		c.JSON(http.StatusOK, airports)
	}
}

// MockAirportsHandler provides mock airport data (deprecated - use CachedAirportsHandler)
func MockAirportsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		airports := []map[string]string{
			{"code": "JFK", "name": "John F. Kennedy International Airport", "city": "New York"},
			{"code": "LAX", "name": "Los Angeles International Airport", "city": "Los Angeles"},
			{"code": "ORD", "name": "O'Hare International Airport", "city": "Chicago"},
			{"code": "LHR", "name": "Heathrow Airport", "city": "London"},
			{"code": "CDG", "name": "Charles de Gaulle Airport", "city": "Paris"},
		}
		c.JSON(http.StatusOK, airports)
	}
}

// MockPriceHistoryHandler provides mock price history data
func MockPriceHistoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Query("origin")
		destination := c.Query("destination")

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
}

// createBulkJob returns a handler for creating a new scheduled bulk search job
func createBulkJob(pgDB db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.CronExpression == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CronExpression is required"})
			return
		}

		// Parse date strings
		dateStart, err := time.Parse("2006-01-02", req.DateStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_start format. Use YYYY-MM-DD: " + err.Error()})
			return
		}
		dateEnd, err := time.Parse("2006-01-02", req.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_end format. Use YYYY-MM-DD: " + err.Error()})
			return
		}

		var returnDateStart, returnDateEnd sql.NullTime
		if req.ReturnDateStart != "" {
			parsedDate, err := time.Parse("2006-01-02", req.ReturnDateStart)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_start format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
			returnDateStart = sql.NullTime{Time: parsedDate, Valid: true}
		}
		if req.ReturnDateEnd != "" {
			parsedDate, err := time.Parse("2006-01-02", req.ReturnDateEnd)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return_date_end format. Use YYYY-MM-DD: " + err.Error()})
				return
			}
			returnDateEnd = sql.NullTime{Time: parsedDate, Valid: true}
		}

		ctx := c.Request.Context()
		var overrides map[string][]string
		var worldAllCount int
		if containsToken(req.Origins, macros.RegionWorldAll) || containsToken(req.Destinations, macros.RegionWorldAll) {
			worldAll, err := listAllAirportCodes(ctx, pgDB)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			overrides = map[string][]string{macros.RegionWorldAll: worldAll}
			worldAllCount = len(worldAll)
		}

		// Expand region tokens in origins and destinations
		expandedOrigins, originWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Origins, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid origin: " + err.Error()})
			return
		}
		expandedDestinations, destinationWarnings, err := macros.ExpandAirportTokensWithOverrides(req.Destinations, overrides)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid destination: " + err.Error()})
			return
		}

		warnings := append([]string{}, originWarnings...)
		warnings = append(warnings, destinationWarnings...)
		if worldAllCount > 0 {
			warnings = append(warnings, fmt.Sprintf("%s expanded to %d airports", macros.RegionWorldAll, worldAllCount))
		}

		// Guard against accidental explosion of scheduled jobs from region tokens
		// This is stricter than bulk search since each job persists to DB and runs repeatedly
		totalJobs := len(expandedOrigins) * len(expandedDestinations)
		if totalJobs == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "At least one origin and one destination are required (after REGION:* expansion)",
				"warnings": warnings,
			})
			return
		}
		const maxScheduledJobs = 500
		if totalJobs > maxScheduledJobs {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":        fmt.Sprintf("Too many scheduled jobs to create: %d (max %d). Region tokens can expand to many airports; consider narrowing your search or using bulk-search for one-off runs.", totalJobs, maxScheduledJobs),
				"total_jobs":   totalJobs,
				"max_jobs":     maxScheduledJobs,
				"origins":      len(expandedOrigins),
				"destinations": len(expandedDestinations),
				"warnings":     warnings,
			})
			return
		}

		// Create multiple scheduled jobs - one for each origin/destination combination
		createdJobs := []map[string]interface{}{}

		for _, origin := range expandedOrigins {
			for _, destination := range expandedDestinations {
				// Create job name
				jobName := fmt.Sprintf("%s: %s→%s", req.Name, origin, destination)

				// Begin transaction for each job
				tx, err := pgDB.BeginTx(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
					return
				}

				// Create the scheduled job
				jobID, err := pgDB.CreateScheduledJob(ctx, tx, jobName, req.CronExpression, true)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled job"})
					return
				}

				// Create job details with nullable fields
				var tripLength sql.NullInt32
				if req.TripLength > 0 {
					tripLength = sql.NullInt32{Int32: int32(req.TripLength), Valid: true}
				}

				var daysFromExecution sql.NullInt32
				if req.DaysFromExecution > 0 {
					daysFromExecution = sql.NullInt32{Int32: int32(req.DaysFromExecution), Valid: true}
				}

				var searchWindowDays sql.NullInt32
				if req.SearchWindowDays > 0 {
					searchWindowDays = sql.NullInt32{Int32: int32(req.SearchWindowDays), Valid: true}
				}

				details := db.JobDetails{
					JobID:              jobID,
					Origin:             origin,
					Destination:        destination,
					DepartureDateStart: dateStart,
					DepartureDateEnd:   dateEnd,
					ReturnDateStart:    returnDateStart,
					ReturnDateEnd:      returnDateEnd,
					TripLength:         tripLength,
					DynamicDates:       req.DynamicDates,
					DaysFromExecution:  daysFromExecution,
					SearchWindowDays:   searchWindowDays,
					Adults:             req.Adults,
					Children:           req.Children,
					InfantsLap:         req.InfantsLap,
					InfantsSeat:        req.InfantsSeat,
					TripType:           req.TripType,
					Class:              req.Class,
					Stops:              req.Stops,
					Currency:           req.Currency,
				}

				if err := pgDB.CreateJobDetails(ctx, tx, details); err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job details"})
					return
				}

				// Commit the transaction
				if err := tx.Commit(); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
					return
				}

				createdJobs = append(createdJobs, map[string]interface{}{
					"id":              jobID,
					"name":            jobName,
					"origin":          origin,
					"destination":     destination,
					"cron_expression": req.CronExpression,
				})

				log.Printf("Created scheduled bulk search job: %s (ID: %d)", jobName, jobID)
			}
		}

		// Restart the scheduler to pick up the new jobs
		if scheduler := workerManager.GetScheduler(); scheduler != nil {
			log.Printf("Restarting scheduler to load new bulk search jobs")
			scheduler.Stop()
			if err := scheduler.Start(); err != nil {
				log.Printf("Warning: Failed to restart scheduler: %v", err)
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  fmt.Sprintf("Created %d scheduled bulk search jobs", len(createdJobs)),
			"jobs":     createdJobs,
			"warnings": warnings,
		})
	}
}

// WorkerStatusProvider exposes worker status metrics for the admin API.
type WorkerStatusProvider interface {
	WorkerStatuses() []worker.WorkerStatus
}

// --- Continuous Sweep Handlers ---

// SweepConfigRequest represents a request to update sweep configuration
type SweepConfigRequest struct {
	PacingMode          string `json:"pacing_mode,omitempty"`
	TargetDurationHours int    `json:"target_duration_hours,omitempty"`
	MinDelayMs          int    `json:"min_delay_ms,omitempty"`
	Class               string `json:"class,omitempty"`
	TripLengths         *[]int `json:"trip_lengths,omitempty"`
}

func normalizeContinuousSweepTripLengths(input []int) ([]int, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("trip_lengths cannot be empty")
	}

	seen := make(map[int]struct{}, len(input))
	out := make([]int, 0, len(input))
	for _, v := range input {
		if v < 1 || v > 30 {
			return nil, fmt.Errorf("trip_lengths values must be between 1 and 30 (got %d)", v)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	sort.Ints(out)
	if len(out) > 30 {
		return nil, fmt.Errorf("trip_lengths has too many values (max 30)")
	}

	return out, nil
}

func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func updateContinuousSweepDBFlags(ctx context.Context, pgDB db.PostgresDB, isRunning, isPaused *bool) (*db.ContinuousSweepProgress, error) {
	if pgDB == nil {
		return nil, fmt.Errorf("postgres is not configured")
	}

	if err := pgDB.SetContinuousSweepControlFlags(ctx, isRunning, isPaused); err != nil {
		return nil, err
	}

	progress, err := pgDB.GetContinuousSweepProgress(ctx)
	if err != nil {
		return nil, err
	}
	return progress, nil
}

type continuousSweepControlStore interface {
	SetContinuousSweepControlFlags(ctx context.Context, isRunning, isPaused *bool) (*queue.ContinuousSweepControl, error)
	GetContinuousSweepControlFlags(ctx context.Context) (*queue.ContinuousSweepControl, error)
}

// getContinuousSweepStatus returns the current status of the continuous sweep
func getContinuousSweepStatus(workerManager *worker.Manager, pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		out := gin.H{}
		if pgDB != nil {
			dbProgress, err := pgDB.GetContinuousSweepProgress(c.Request.Context())
			if err != nil {
				out["db_error"] = err.Error()
			} else if dbProgress != nil {
				out["db"] = dbProgress
			}
		}

		// Include a best-effort control state sourced from Redis so STOP/PAUSE can work even when Postgres is
		// disabled/misconfigured on the API process.
		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				if store, ok := q.(continuousSweepControlStore); ok {
					ctrl, err := store.GetContinuousSweepControlFlags(c.Request.Context())
					if err != nil {
						out["control_error"] = err.Error()
					} else if ctrl != nil {
						out["control"] = ctrl
					}
				}
			}
		}

		if pgDB == nil {
			out["db_error"] = "postgres is not configured"
		}

		out["server_time"] = time.Now().Format(time.RFC3339)

		if workerManager == nil {
			out["initialized"] = false
			out["message"] = "Worker manager not available"
			c.JSON(http.StatusOK, out)
			return
		}

		// Include queue stats so the UI can show whether work is still in-flight/backlogged even if the runner
		// isn't initialized in this process.
		if q := workerManager.GetQueue(); q != nil {
			queues := gin.H{}
			for _, name := range []string{"continuous_price_graph"} {
				stats, err := q.GetQueueStats(c.Request.Context(), name)
				if err != nil {
					queues[name] = gin.H{"error": err.Error()}
					continue
				}
				queues[name] = gin.H{
					"pending":    stats["pending"],
					"processing": stats["processing"],
					"completed":  stats["completed"],
					"failed":     stats["failed"],
				}
			}
			out["queues"] = queues
		}

		runner := workerManager.GetSweepRunner()
		if runner == nil {
			out["initialized"] = false
			out["message"] = "Continuous sweep runner not initialized in this process."
			c.JSON(http.StatusOK, out)
			return
		}

		status := runner.GetStatus()
		out["initialized"] = true
		out["status"] = status
		c.JSON(http.StatusOK, out)
	}
}

// startContinuousSweep starts the continuous sweep process
func startContinuousSweep(workerManager *worker.Manager, pgDB db.PostgresDB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()

		// If no runner exists, create one with default config
		if runner == nil {
			// Create NTFY client for notifications
			var notifier *notify.NTFYClient
			if cfg.NTFYConfig.Enabled && cfg.NTFYConfig.Topic != "" {
				notifier = notify.NewNTFYClient(notify.NTFYConfig{
					ServerURL:       cfg.NTFYConfig.ServerURL,
					Topic:           cfg.NTFYConfig.Topic,
					Username:        cfg.NTFYConfig.Username,
					Password:        cfg.NTFYConfig.Password,
					Enabled:         cfg.NTFYConfig.Enabled,
					StallThreshold:  cfg.NTFYConfig.StallThreshold,
					ErrorThreshold:  cfg.NTFYConfig.ErrorThreshold,
					ErrorWindow:     cfg.NTFYConfig.ErrorWindow,
					DefaultPriority: notify.PriorityDefault,
				})
			}

			sweepConfig := worker.DefaultContinuousSweepConfig()
			runner = worker.NewContinuousSweepRunner(pgDB, workerManager.GetQueue(), notifier, sweepConfig)
			workerManager.SetSweepRunner(runner)
		}

		// Start the sweep
		if err := runner.Start(); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Sweep already running"})
			return
		}

		// Best-effort: persist control flags in Redis too.
		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				if store, ok := q.(continuousSweepControlStore); ok {
					running := true
					paused := false
					_, _ = store.SetContinuousSweepControlFlags(c.Request.Context(), &running, &paused)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep started",
			"status":  runner.GetStatus(),
		})
	}
}

// stopContinuousSweep stops the continuous sweep process
func stopContinuousSweep(workerManager *worker.Manager, pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		running := false
		paused := false

		persistErrors := make([]string, 0, 2)

		if _, err := updateContinuousSweepDBFlags(c.Request.Context(), pgDB, &running, &paused); err != nil {
			persistErrors = append(persistErrors, "postgres: "+err.Error())
		}

		// Best-effort: also persist STOP into Redis so worker-only deployments can be controlled.
		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				if store, ok := q.(continuousSweepControlStore); ok {
					if _, err := store.SetContinuousSweepControlFlags(c.Request.Context(), &running, &paused); err != nil {
						persistErrors = append(persistErrors, "redis: "+err.Error())
					}
				}
			}
		}

		if workerManager != nil {
			if runner := workerManager.GetSweepRunner(); runner != nil {
				runner.Stop()
			}
		}

		drainResult := gin.H{}
		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				// When the sweep is stopped, we don't want the continuous queue to keep running in the background.
				// Best-effort: cancel in-flight jobs and clear pending.
				canceled, cancelErr := q.CancelProcessing(c.Request.Context(), "continuous_price_graph")
				if cancelErr != nil {
					drainResult["error"] = cancelErr.Error()
				} else {
					drainResult["canceled_processing"] = canceled
				}
				cleared, clearErr := q.ClearQueue(c.Request.Context(), "continuous_price_graph")
				if clearErr != nil {
					drainResult["error"] = clearErr.Error()
				} else {
					drainResult["cleared_pending"] = cleared
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep stopped",
			"persist": func() any {
				if len(persistErrors) == 0 {
					return nil
				}
				return gin.H{"errors": persistErrors}
			}(),
			"drain": drainResult,
			"status": func() any {
				if workerManager == nil {
					return nil
				}
				if runner := workerManager.GetSweepRunner(); runner != nil {
					return runner.GetStatus()
				}
				return nil
			}(),
		})
	}
}

// pauseContinuousSweep pauses the continuous sweep process
func pauseContinuousSweep(workerManager *worker.Manager, pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paused := true

		persistErrors := make([]string, 0, 2)
		if _, err := updateContinuousSweepDBFlags(c.Request.Context(), pgDB, nil, &paused); err != nil {
			persistErrors = append(persistErrors, "postgres: "+err.Error())
		}

		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				if store, ok := q.(continuousSweepControlStore); ok {
					if _, err := store.SetContinuousSweepControlFlags(c.Request.Context(), nil, &paused); err != nil {
						persistErrors = append(persistErrors, "redis: "+err.Error())
					}
				}
			}
		}

		if workerManager != nil {
			if runner := workerManager.GetSweepRunner(); runner != nil {
				runner.Pause()
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep paused",
			"persist": func() any {
				if len(persistErrors) == 0 {
					return nil
				}
				return gin.H{"errors": persistErrors}
			}(),
			"status": func() any {
				if workerManager == nil {
					return nil
				}
				if runner := workerManager.GetSweepRunner(); runner != nil {
					return runner.GetStatus()
				}
				return nil
			}(),
		})
	}
}

// resumeContinuousSweep resumes the continuous sweep process
func resumeContinuousSweep(workerManager *worker.Manager, pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		paused := false

		persistErrors := make([]string, 0, 2)
		if _, err := updateContinuousSweepDBFlags(c.Request.Context(), pgDB, nil, &paused); err != nil {
			persistErrors = append(persistErrors, "postgres: "+err.Error())
		}

		if workerManager != nil {
			if q := workerManager.GetQueue(); q != nil {
				if store, ok := q.(continuousSweepControlStore); ok {
					if _, err := store.SetContinuousSweepControlFlags(c.Request.Context(), nil, &paused); err != nil {
						persistErrors = append(persistErrors, "redis: "+err.Error())
					}
				}
			}
		}

		if workerManager != nil {
			if runner := workerManager.GetSweepRunner(); runner != nil {
				runner.Resume()
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep resumed",
			"persist": func() any {
				if len(persistErrors) == 0 {
					return nil
				}
				return gin.H{"errors": persistErrors}
			}(),
			"status": func() any {
				if workerManager == nil {
					return nil
				}
				if runner := workerManager.GetSweepRunner(); runner != nil {
					return runner.GetStatus()
				}
				return nil
			}(),
		})
	}
}

// updateContinuousSweepConfig updates the sweep configuration
func updateContinuousSweepConfig(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		var req SweepConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Start from the current config to avoid resetting unrelated fields.
		prevConfig := runner.GetConfig()
		newConfig := prevConfig
		tripLengthsChanged := false

		if req.TripLengths != nil {
			tripLengths, err := normalizeContinuousSweepTripLengths(*req.TripLengths)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			tripLengthsChanged = !equalIntSlice(tripLengths, prevConfig.TripLengths)
			newConfig.TripLengths = tripLengths
		}

		// Apply updates
		if req.PacingMode != "" {
			if req.PacingMode == "adaptive" {
				newConfig.PacingMode = worker.PacingModeAdaptive
			} else if req.PacingMode == "fixed" {
				newConfig.PacingMode = worker.PacingModeFixed
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pacing_mode. Use 'adaptive' or 'fixed'"})
				return
			}
		}

		if req.TargetDurationHours > 0 {
			newConfig.TargetDurationHours = req.TargetDurationHours
		}

		if req.MinDelayMs > 0 {
			newConfig.MinDelayMs = req.MinDelayMs
		}

		if req.Class != "" {
			switch req.Class {
			case "economy", "premium_economy", "business", "first":
				newConfig.Class = req.Class
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class. Use 'economy', 'premium_economy', 'business', or 'first'"})
				return
			}
		}

		runner.SetConfig(newConfig)
		if tripLengthsChanged {
			status := runner.GetStatus()
			if status.IsRunning {
				runner.RestartSweep()
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Sweep configuration updated",
			"status":  runner.GetStatus(),
		})
	}
}

// skipCurrentRoute skips the current route in the sweep
func skipCurrentRoute(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		runner.SkipRoute()
		c.JSON(http.StatusOK, gin.H{
			"message": "Route skipped",
			"status":  runner.GetStatus(),
		})
	}
}

// restartCurrentSweep restarts the current sweep from the beginning
func restartCurrentSweep(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		runner.RestartSweep()
		c.JSON(http.StatusOK, gin.H{
			"message": "Sweep restarted from beginning",
			"status":  runner.GetStatus(),
		})
	}
}

// ContinuousSweepStatsResponse is the JSON-friendly response for sweep stats
type ContinuousSweepStatsResponse struct {
	ID                   int        `json:"id"`
	SweepNumber          int        `json:"sweep_number"`
	StartedAt            time.Time  `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	TotalRoutes          int        `json:"total_routes"`
	SuccessfulQueries    int        `json:"successful_queries"`
	FailedQueries        int        `json:"failed_queries"`
	TotalDurationSeconds *int       `json:"total_duration_seconds,omitempty"`
	AvgDelayMs           *int       `json:"avg_delay_ms,omitempty"`
	MinPriceFound        *float64   `json:"min_price_found,omitempty"`
	MaxPriceFound        *float64   `json:"max_price_found,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// convertSweepStats converts db.ContinuousSweepStats to JSON-friendly response
func convertSweepStats(stats []db.ContinuousSweepStats) []ContinuousSweepStatsResponse {
	result := make([]ContinuousSweepStatsResponse, len(stats))
	for i, s := range stats {
		result[i] = ContinuousSweepStatsResponse{
			ID:                s.ID,
			SweepNumber:       s.SweepNumber,
			StartedAt:         s.StartedAt,
			TotalRoutes:       s.TotalRoutes,
			SuccessfulQueries: s.SuccessfulQueries,
			FailedQueries:     s.FailedQueries,
			CreatedAt:         s.CreatedAt,
		}
		if s.CompletedAt.Valid {
			result[i].CompletedAt = &s.CompletedAt.Time
		}
		if s.TotalDurationSeconds.Valid {
			dur := int(s.TotalDurationSeconds.Int32)
			result[i].TotalDurationSeconds = &dur
		}
		if s.AvgDelayMs.Valid {
			delay := int(s.AvgDelayMs.Int32)
			result[i].AvgDelayMs = &delay
		}
		if s.MinPriceFound.Valid {
			result[i].MinPriceFound = &s.MinPriceFound.Float64
		}
		if s.MaxPriceFound.Valid {
			result[i].MaxPriceFound = &s.MaxPriceFound.Float64
		}
	}
	return result
}

// getContinuousSweepStats returns historical sweep statistics
func getContinuousSweepStats(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		stats, err := pgDB.ListContinuousSweepStats(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sweep stats"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stats": convertSweepStats(stats),
			"count": len(stats),
		})
	}
}

// ContinuousSweepResultResponse is the JSON-friendly response for sweep results
type ContinuousSweepResultResponse struct {
	Origin        string     `json:"origin"`
	Destination   string     `json:"destination"`
	DepartureDate time.Time  `json:"departure_date"`
	ReturnDate    *time.Time `json:"return_date,omitempty"`
	TripLength    *int       `json:"trip_length,omitempty"`
	Price         float64    `json:"price"`
	Currency      string     `json:"currency"`
	Adults        int        `json:"adults"`
	Children      int        `json:"children"`
	InfantsLap    int        `json:"infants_lap"`
	InfantsSeat   int        `json:"infants_seat"`
	TripType      string     `json:"trip_type"`
	Class         string     `json:"class"`
	Stops         string     `json:"stops"`
	SearchURL     *string    `json:"search_url,omitempty"`
	QueriedAt     time.Time  `json:"queried_at"`
}

// convertSweepResults converts db.PriceGraphResultRecord to JSON-friendly response
func convertSweepResults(results []db.PriceGraphResultRecord) []ContinuousSweepResultResponse {
	resp := make([]ContinuousSweepResultResponse, len(results))
	for i, r := range results {
		resp[i] = ContinuousSweepResultResponse{
			Origin:        r.Origin,
			Destination:   r.Destination,
			DepartureDate: r.DepartureDate,
			Price:         r.Price,
			Currency:      r.Currency,
			Adults:        r.Adults,
			Children:      r.Children,
			InfantsLap:    r.InfantsLap,
			InfantsSeat:   r.InfantsSeat,
			TripType:      r.TripType,
			Class:         r.Class,
			Stops:         r.Stops,
			QueriedAt:     r.QueriedAt,
		}
		if r.ReturnDate.Valid {
			resp[i].ReturnDate = &r.ReturnDate.Time
		}
		if r.TripLength.Valid {
			tripLen := int(r.TripLength.Int32)
			resp[i].TripLength = &tripLen
		}
		if r.SearchURL.Valid {
			resp[i].SearchURL = &r.SearchURL.String
		}
	}
	return resp
}

// getContinuousSweepResults returns price graph results from continuous sweeps
func getContinuousSweepResults(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		filters := db.ContinuousSweepResultsFilter{
			Origin:      c.Query("origin"),
			Destination: c.Query("destination"),
			Limit:       limit,
			Offset:      offset,
		}

		// Parse date filters
		if from := c.Query("from"); from != "" {
			if t, err := time.Parse("2006-01-02", from); err == nil {
				filters.FromDate = t
			}
		}
		if to := c.Query("to"); to != "" {
			if t, err := time.Parse("2006-01-02", to); err == nil {
				filters.ToDate = t
			}
		}

		results, err := pgDB.ListContinuousSweepResults(c.Request.Context(), filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sweep results"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": convertSweepResults(results),
			"count":   len(results),
			"filters": gin.H{
				"origin":      filters.Origin,
				"destination": filters.Destination,
				"from":        filters.FromDate,
				"to":          filters.ToDate,
			},
		})
	}
}

// getOrCreateSession gets an existing flight session from the manager or creates a new one
func getOrCreateSession(m *worker.Manager) (*flights.Session, error) {
	// Create a new session
	session, err := flights.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create flights session: %w", err)
	}
	return session, nil
}

// listDeals returns a handler for listing detected deals
func listDeals(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse filter parameters
		filter := db.DealFilter{
			Origin:         c.Query("origin"),
			Destination:    c.Query("destination"),
			Classification: c.Query("classification"),
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 500 {
			limit = 50
		}
		filter.Limit = limit

		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		filter.Offset = offset

		// Filter by status (default to active only)
		status := c.DefaultQuery("status", db.DealStatusActive)
		if status != "" {
			filter.Status = status
		}

		deals, err := pgDB.ListActiveDeals(c.Request.Context(), filter)
		if err != nil {
			log.Printf("Error fetching deals: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deals"})
			return
		}

		// Convert to JSON-friendly format
		response := make([]map[string]interface{}, len(deals))
		for i, deal := range deals {
			response[i] = map[string]interface{}{
				"id":                  deal.ID,
				"origin":              deal.Origin,
				"destination":         deal.Destination,
				"departure_date":      deal.DepartureDate.Format("2006-01-02"),
				"price":               deal.Price,
				"currency":            deal.Currency,
				"discount_percent":    maybeNullFloat(deal.DiscountPercent),
				"deal_score":          maybeNullInt(deal.DealScore),
				"deal_classification": maybeNullString(deal.DealClassification),
				"cost_per_mile":       maybeNullFloat(deal.CostPerMile),
				"cabin_class":         deal.CabinClass,
				"status":              deal.Status,
				"first_seen_at":       deal.FirstSeenAt,
				"times_seen":          deal.TimesSeen,
			}
			if deal.ReturnDate.Valid {
				response[i]["return_date"] = deal.ReturnDate.Time.Format("2006-01-02")
			}
			if deal.SearchURL.Valid {
				response[i]["search_url"] = deal.SearchURL.String
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"deals": response,
			"count": len(deals),
			"filter": gin.H{
				"origin":         filter.Origin,
				"destination":    filter.Destination,
				"classification": filter.Classification,
				"status":         filter.Status,
			},
		})
	}
}

// maybeNullFloat returns the float value or nil if not valid
func maybeNullFloat(value sql.NullFloat64) interface{} {
	if value.Valid {
		return value.Float64
	}
	return nil
}

// listDealAlerts returns a handler for listing published deal alerts
func listDealAlerts(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 500 {
			limit = 50
		}

		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		alerts, err := pgDB.ListDealAlerts(c.Request.Context(), limit, offset)
		if err != nil {
			log.Printf("Error fetching deal alerts: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deal alerts"})
			return
		}

		// Convert to JSON-friendly format
		response := make([]map[string]interface{}, len(alerts))
		for i, alert := range alerts {
			response[i] = map[string]interface{}{
				"id":                  alert.ID,
				"detected_deal_id":    alert.DetectedDealID,
				"origin":              alert.Origin,
				"destination":         alert.Destination,
				"price":               alert.Price,
				"currency":            alert.Currency,
				"discount_percent":    maybeNullFloat(alert.DiscountPercent),
				"deal_classification": maybeNullString(alert.DealClassification),
				"deal_score":          maybeNullInt(alert.DealScore),
				"published_at":        alert.PublishedAt,
				"publish_method":      alert.PublishMethod,
				"notification_sent":   alert.NotificationSent,
			}
			if alert.NotificationSentAt.Valid {
				response[i]["notification_sent_at"] = alert.NotificationSentAt.Time
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"alerts": response,
			"count":  len(alerts),
		})
	}
}
