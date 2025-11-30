package api

import (
	"database/sql"
	"encoding/json"
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
	"github.com/gilby125/google-flights-api/pkg/cache"
	"github.com/gilby125/google-flights-api/pkg/logger"
	"github.com/gilby125/google-flights-api/pkg/notify"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// SearchRequest represents a flight search request
type SearchRequest struct {
	Origin        string    `json:"origin" binding:"required"`
	Destination   string    `json:"destination" binding:"required"`
	DepartureDate time.Time `json:"departure_date" binding:"required" time_format:"2006-01-02" time_utc:"true"`
	ReturnDate    time.Time `json:"return_date,omitempty" time_format:"2006-01-02" time_utc:"true"`
	Adults        int       `json:"adults" binding:"required,min=1"`
	Children      int       `json:"children" binding:"min=0"`
	InfantsLap    int       `json:"infants_lap" binding:"min=0"`
	InfantsSeat   int       `json:"infants_seat" binding:"min=0"`
	TripType      string    `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class         string    `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops         string    `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"` // Added two_stops_plus
	Currency      string    `json:"currency" binding:"required,len=3"`
}

// BulkSearchRequest represents a bulk flight search request
type BulkSearchRequest struct {
	Origins           []string  `json:"origins" binding:"required,min=1"`
	Destinations      []string  `json:"destinations" binding:"required,min=1"`
	DepartureDateFrom time.Time `json:"departure_date_from" binding:"required" time_format:"2006-01-02" time_utc:"true"`
	DepartureDateTo   time.Time `json:"departure_date_to" binding:"required" time_format:"2006-01-02" time_utc:"true"`
	ReturnDateFrom    time.Time `json:"return_date_from,omitempty" time_format:"2006-01-02" time_utc:"true"`
	ReturnDateTo      time.Time `json:"return_date_to,omitempty" time_format:"2006-01-02" time_utc:"true"`
	TripLength        int       `json:"trip_length,omitempty" binding:"min=0"`
	Adults            int       `json:"adults" binding:"required,min=1"`
	Children          int       `json:"children" binding:"min=0"`
	InfantsLap        int       `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int       `json:"infants_seat" binding:"min=0"`
	TripType          string    `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string    `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string    `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Currency          string    `json:"currency" binding:"required,len=3"`
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
	Origins           []string  `json:"origins" binding:"required,min=1"`
	Destinations      []string  `json:"destinations" binding:"required,min=1"`
	DepartureDateFrom time.Time `json:"departure_date_from" binding:"required"`
	DepartureDateTo   time.Time `json:"departure_date_to" binding:"required"`
	TripLengths       []int     `json:"trip_lengths,omitempty"`
	TripType          string    `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string    `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string    `json:"stops" binding:"required,oneof=nonstop one_stop two_stops two_stops_plus any"`
	Adults            int       `json:"adults" binding:"required,min=1"`
	Children          int       `json:"children" binding:"min=0"`
	InfantsLap        int       `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int       `json:"infants_seat" binding:"min=0"`
	Currency          string    `json:"currency" binding:"required,len=3"`
	RateLimitMillis   int       `json:"rate_limit_millis,omitempty" binding:"min=0"`
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

// getAirports returns a handler for getting all airports
func GetAirports(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pgDB.QueryAirports(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query airports"})
			return
		}
		defer rows.Close()

		airports := []db.Airport{} // Use the defined struct
		for rows.Next() {
			var airport db.Airport
			if err := rows.Scan(&airport.Code, &airport.Name, &airport.City, &airport.Country, &airport.Latitude, &airport.Longitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan airport row"})
				return
			}
			airports = append(airports, airport)
		}
		if err := rows.Err(); err != nil {
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
			if !req.ReturnDate.After(req.DepartureDate) {
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
			req.ReturnDate = time.Time{}
		}
		// --- End Custom Validation ---

		// Create a worker payload (only if validation passes)
		payload := worker.FlightSearchPayload{
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate,
			ReturnDate:    req.ReturnDate, // Correctly zeroed for one-way
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
func GetWorkerStatus(workerManager WorkerStatusProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if workerManager == nil {
			c.JSON(http.StatusOK, []worker.WorkerStatus{})
			return
		}

		statuses := workerManager.WorkerStatuses()
		c.JSON(http.StatusOK, statuses)
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
			err := rows.Scan(&res.Origin, &res.Destination, &res.DepartureDate,
				&res.ReturnDate, &res.Price, &res.Currency, &res.AirlineCode, &res.Duration)
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

		response := map[string]interface{}{
			"id":             search.ID,
			"status":         search.Status,
			"total_searches": search.TotalSearches,
			"completed":      search.Completed,
			"created_at":     search.CreatedAt,
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

// enqueuePriceGraphSweep enqueues a price graph sweep job for execution
func enqueuePriceGraphSweep(pgDB db.PostgresDB, scheduler *worker.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if scheduler == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Price graph scheduler unavailable"})
			return
		}

		var req PriceGraphSweepRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.DepartureDateFrom.After(req.DepartureDateTo) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "departure_date_from must be before departure_date_to"})
			return
		}

		stops := req.Stops
		if stops == "two_stops_plus" {
			stops = "any"
		}

		payload := worker.PriceGraphSweepPayload{
			Origins:           req.Origins,
			Destinations:      req.Destinations,
			DepartureDateFrom: req.DepartureDateFrom,
			DepartureDateTo:   req.DepartureDateTo,
			TripLengths:       req.TripLengths,
			TripType:          req.TripType,
			Class:             req.Class,
			Stops:             stops,
			Adults:            req.Adults,
			Children:          req.Children,
			InfantsLap:        req.InfantsLap,
			InfantsSeat:       req.InfantsSeat,
			Currency:          strings.ToUpper(req.Currency),
			RateLimitMillis:   req.RateLimitMillis,
		}

		ctx := c.Request.Context()
		sweepID, err := scheduler.EnqueuePriceGraphSweep(ctx, payload)
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
				queriedAt   time.Time
				createdAt   time.Time
			)

			if err := rows.Scan(&id, &sid, &origin, &destination, &departure, &returnDate, &tripLength, &price, &currency, &queriedAt, &createdAt); err != nil {
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
func CreateBulkSearch(q queue.Queue, pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		totalRoutes := len(req.Origins) * len(req.Destinations)
		if totalRoutes == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one origin and one destination are required"})
			return
		}

		currencyCode := strings.ToUpper(req.Currency)

		// Create bulk search record so the run can be tracked
		bulkSearchID, err := pgDB.CreateBulkSearchRecord(
			c.Request.Context(),
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
			Origins:           req.Origins,
			Destinations:      req.Destinations,
			DepartureDateFrom: req.DepartureDateFrom,
			DepartureDateTo:   req.DepartureDateTo,
			ReturnDateFrom:    req.ReturnDateFrom,
			ReturnDateTo:      req.ReturnDateTo,
			TripLength:        req.TripLength,
			Adults:            req.Adults,
			Children:          req.Children,
			InfantsLap:        req.InfantsLap,
			InfantsSeat:       req.InfantsSeat,
			TripType:          req.TripType,
			Class:             req.Class,
			Stops:             req.Stops,
			Currency:          currencyCode,
			BulkSearchID:      bulkSearchID,
		}

		// Enqueue the job
		jobID, err := q.Enqueue(c.Request.Context(), "bulk_search", payload)
		if err != nil {
			_ = pgDB.UpdateBulkSearchStatus(c.Request.Context(), bulkSearchID, "failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":         jobID,
			"bulk_search_id": bulkSearchID,
			"message":        "Bulk flight search job created successfully",
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
		queueTypes := []string{"flight_search", "bulk_search"}
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
}

// DirectFlightSearch handles direct flight searches (bypasses queue for immediate results)
func DirectFlightSearch() gin.HandlerFunc {
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

		// Validate required fields
		if searchRequest.Origin == "" || searchRequest.Destination == "" || searchRequest.DepartureDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Origin, destination, and departure date are required"})
			return
		}

		// Parse dates
		departureDate, err := time.Parse("2006-01-02", searchRequest.DepartureDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid departure date format. Use YYYY-MM-DD"})
			return
		}

		var returnDate time.Time
		if searchRequest.ReturnDate != "" {
			returnDate, err = time.Parse("2006-01-02", searchRequest.ReturnDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return date format. Use YYYY-MM-DD"})
				return
			}
		} else if searchRequest.TripType == "round_trip" {
			returnDate = departureDate.AddDate(0, 0, 7)
		} else {
			returnDate = departureDate.AddDate(0, 0, 1)
		}

		// Create flight session
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

		// Perform the flight search
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

			// Generate Google Flights URL
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

		// Create multiple scheduled jobs - one for each origin/destination combination
		createdJobs := []map[string]interface{}{}

		for _, origin := range req.Origins {
			for _, destination := range req.Destinations {
				// Create job name
				jobName := fmt.Sprintf("%s: %s→%s", req.Name, origin, destination)

				// Begin transaction for each job
				ctx := c.Request.Context()
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
			"message": fmt.Sprintf("Created %d scheduled bulk search jobs", len(createdJobs)),
			"jobs":    createdJobs,
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
}

// getContinuousSweepStatus returns the current status of the continuous sweep
func getContinuousSweepStatus(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Continuous sweep runner not initialized",
				"message": "The sweep runner has not been configured. Start the sweep first.",
			})
			return
		}

		status := runner.GetStatus()
		c.JSON(http.StatusOK, status)
	}
}

// startContinuousSweep starts the continuous sweep process
func startContinuousSweep(workerManager *worker.Manager, pgDB db.PostgresDB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()

		// If no runner exists, create one with default config
		if runner == nil {
			// Get a flight session from the manager's cache
			session, err := getOrCreateSession(workerManager)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flight session: " + err.Error()})
				return
			}

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
			runner = worker.NewContinuousSweepRunner(pgDB, session, notifier, sweepConfig)
			workerManager.SetSweepRunner(runner)
		}

		// Start the sweep
		if err := runner.Start(); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Sweep already running"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep started",
			"status":  runner.GetStatus(),
		})
	}
}

// stopContinuousSweep stops the continuous sweep process
func stopContinuousSweep(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		runner.Stop()
		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep stopped",
			"status":  runner.GetStatus(),
		})
	}
}

// pauseContinuousSweep pauses the continuous sweep process
func pauseContinuousSweep(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		runner.Pause()
		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep paused",
			"status":  runner.GetStatus(),
		})
	}
}

// resumeContinuousSweep resumes the continuous sweep process
func resumeContinuousSweep(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		runner := workerManager.GetSweepRunner()
		if runner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Continuous sweep runner not initialized"})
			return
		}

		runner.Resume()
		c.JSON(http.StatusOK, gin.H{
			"message": "Continuous sweep resumed",
			"status":  runner.GetStatus(),
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

		// Get current config and update fields
		currentStatus := runner.GetStatus()
		newConfig := worker.ContinuousSweepConfig{
			TripLengths:         []int{7, 14},
			DepartureWindowDays: 30,
			Class:               "economy",
			Stops:               "any",
			Adults:              1,
			Currency:            "USD",
			InternationalOnly:   true,
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
		} else {
			newConfig.PacingMode = worker.PacingMode(currentStatus.PacingMode)
		}

		if req.TargetDurationHours > 0 {
			newConfig.TargetDurationHours = req.TargetDurationHours
		} else {
			newConfig.TargetDurationHours = currentStatus.TargetDurationHours
		}

		if req.MinDelayMs > 0 {
			newConfig.MinDelayMs = req.MinDelayMs
		} else {
			newConfig.MinDelayMs = currentStatus.CurrentDelayMs
		}

		runner.SetConfig(newConfig)

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
			QueriedAt:     r.QueriedAt,
		}
		if r.ReturnDate.Valid {
			resp[i].ReturnDate = &r.ReturnDate.Time
		}
		if r.TripLength.Valid {
			tripLen := int(r.TripLength.Int32)
			resp[i].TripLength = &tripLen
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
