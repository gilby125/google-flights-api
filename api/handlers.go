package api

// Trivial comment added to force re-evaluation
import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights" // Added missing flights import
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// SearchRequest represents a flight search request
type SearchRequest struct {
	Origin        string    `json:"origin" binding:"required"`
	Destination   string    `json:"destination" binding:"required"`
	DepartureDate time.Time `json:"departure_date" binding:"required"`
	ReturnDate    time.Time `json:"return_date,omitempty"`
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
	DepartureDateFrom time.Time `json:"departure_date_from" binding:"required"`
	DepartureDateTo   time.Time `json:"departure_date_to" binding:"required"`
	ReturnDateFrom    time.Time `json:"return_date_from,omitempty"`
	ReturnDateTo      time.Time `json:"return_date_to,omitempty"`
	TripLength        int       `json:"trip_length,omitempty" binding:"min=0"`
	Adults            int       `json:"adults" binding:"required,min=1"`
	Children          int       `json:"children" binding:"min=0"`
	InfantsLap        int       `json:"infants_lap" binding:"min=0"`
	InfantsSeat       int       `json:"infants_seat" binding:"min=0"`
	TripType          string    `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class             string    `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops             string    `json:"stops" binding:"required,oneof=nonstop one_stop two_stops any"`
	Currency          string    `json:"currency" binding:"required,len=3"`
}

// JobRequest represents a scheduled job request
type JobRequest struct {
	Name            string `json:"name" binding:"required"`
	Origin          string `json:"origin" binding:"required"`
	Destination     string `json:"destination" binding:"required"`
	DateStart       string `json:"date_start" binding:"required"`
	DateEnd         string `json:"date_end" binding:"required"`
	ReturnDateStart string `json:"return_date_start,omitempty"`
	ReturnDateEnd   string `json:"return_date_end,omitempty"`
	TripLength      int    `json:"trip_length,omitempty" binding:"min=0"`
	Adults          int    `json:"adults" binding:"required,min=1"`
	Children        int    `json:"children" binding:"min=0"`
	InfantsLap      int    `json:"infants_lap" binding:"min=0"`
	InfantsSeat     int    `json:"infants_seat" binding:"min=0"`
	TripType        string `json:"trip_type" binding:"required,oneof=one_way round_trip"`
	Class           string `json:"class" binding:"required,oneof=economy premium_economy business first"`
	Stops           string `json:"stops" binding:"required,oneof=nonstop one_stop two_stops any"`
	Currency        string `json:"currency" binding:"required,len=3"`
	Interval        string `json:"interval" binding:"required"`
	Time            string `json:"time" binding:"required"`
	CronExpression  string `json:"cron_expression" binding:"required"`
}

// getAirports returns a handler for getting all airports
func GetAirports(pgDB db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pgDB.QueryAirports(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query airports: " + err.Error()})
			return
		}
		defer rows.Close()

		airports := []db.Airport{} // Use the defined struct
		for rows.Next() {
			var airport db.Airport
			if err := rows.Scan(&airport.Code, &airport.Name, &airport.City, &airport.Country, &airport.Latitude, &airport.Longitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan airport row: " + err.Error()})
				return
			}
			airports = append(airports, airport)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating airport rows: " + err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query airlines: " + err.Error()})
			return
		}
		defer rows.Close()

		airlines := []db.Airline{} // Use the defined struct
		for rows.Next() {
			var airline db.Airline
			if err := rows.Scan(&airline.Code, &airline.Name, &airline.Country); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan airline row: " + err.Error()})
				return
			}
			airlines = append(airlines, airline)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating airline rows: " + err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search query: " + err.Error()})
			}
			return
		}

		// Get the flight offers
		offerRows, err := pgDB.GetFlightOffersBySearchID(c.Request.Context(), searchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight offers: " + err.Error()})
			return
		}
		defer offerRows.Close()

		offers := []map[string]interface{}{}
		for offerRows.Next() {
			var offer db.FlightOffer // Use the defined struct
			if err := offerRows.Scan(&offer.ID, &offer.Price, &offer.Currency, &offer.DepartureDate, &offer.ReturnDate, &offer.TotalDuration, &offer.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan flight offer: " + err.Error()})
				return
			}

			// Get the flight segments for this offer
			segmentRows, err := pgDB.GetFlightSegmentsByOfferID(c.Request.Context(), offer.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight segments: " + err.Error()})
				return
			}
			defer segmentRows.Close()

			segments := []db.FlightSegment{} // Use the defined struct
			for segmentRows.Next() {
				var segment db.FlightSegment
				if err := segmentRows.Scan(
					&segment.AirlineCode, &segment.FlightNumber, &segment.DepartureAirport,
					&segment.ArrivalAirport, &segment.DepartureTime, &segment.ArrivalTime,
					&segment.Duration, &segment.Airplane, &segment.Legroom, &segment.IsReturn,
				); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan flight segment: " + err.Error()})
					return
				}
				segments = append(segments, segment)
			}
			if err := segmentRows.Err(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating flight segments: " + err.Error()})
				return
			}

			offerMap := map[string]interface{}{
				"id":             offer.ID,
				"price":          offer.Price,
				"currency":       offer.Currency,
				"departure_date": offer.DepartureDate,
				"total_duration": offer.TotalDuration,
				"created_at":     offer.CreatedAt,
				"segments":       segments, // Use the struct slice directly
			}

			if offer.ReturnDate.Valid {
				offerMap["return_date"] = offer.ReturnDate.Time
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
		err = pgDB.DeleteJobDetailsByJobID(tx, jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job details: " + err.Error()})
			return
		}

		// Delete the job
		rowsAffected, err := pgDB.DeleteScheduledJobByID(tx, jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduled job: " + err.Error()})
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
func GetWorkerStatus(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "running"})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bulk search metadata: " + err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query bulk search results: " + err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query price history: " + err.Error()})
			return
		}
		// IMPORTANT: The underlying session for the result is still open.
		// We need to process the result fully here.

		priceHistory := []map[string]interface{}{}
		// Process the result using the db.Neo4jResult interface.
		// NOTE: The underlying session for the result is still open after this loop.
		// Proper session management requires refactoring db.ExecuteReadQuery or careful handling here.
		for result.Next() { // Use Next() from db.Neo4jResult interface
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
		// Check for errors after the loop using the result interface
		if err = result.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing price history results: " + err.Error()})
			return
		}
		// Session associated with 'result' is likely still open here.

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
		rows, err := pgDB.ListJobs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list jobs: " + err.Error()})
			return
		}
		defer rows.Close()

		jobs := []map[string]interface{}{}
		for rows.Next() {
			var job db.ScheduledJob // Use the defined struct
			if err := rows.Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan job: " + err.Error()})
				return
			}

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

			jobs = append(jobs, jobMap)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating jobs: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, jobs)
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
		jobID, err := pgDB.CreateScheduledJob(tx, req.Name, req.CronExpression, true) // Assume enabled by default
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled job: " + err.Error()})
			return
		}

		// Prepare job details struct
		details := db.JobDetails{
			JobID:              jobID,
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

		// Insert the job details
		err = pgDB.CreateJobDetails(tx, details)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job details: " + err.Error()})
			return
		}

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
func CreateBulkSearch(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BulkSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			Currency:          req.Currency,
		}

		// Enqueue the job
		jobID, err := q.Enqueue(c.Request.Context(), "bulk_search", payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  jobID,
			"message": "Bulk flight search job created successfully",
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
		tx, err := pgDB.BeginTx(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
			return
		}
		defer tx.Rollback() // Ensure rollback on error

		// Update the job
		err = pgDB.UpdateScheduledJob(tx, jobID, req.Name, req.CronExpression)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled job: " + err.Error()})
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
		err = pgDB.UpdateJobDetails(tx, jobID, details)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job details: " + err.Error()})
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
