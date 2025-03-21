package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/db"
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
	Stops         string    `json:"stops" binding:"required,oneof=nonstop one_stop two_stops any"`
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
}

// getAirports returns a handler for getting all airports
func getAirports(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.GetDB().Query("SELECT code, name, city, country, latitude, longitude FROM airports")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		airports := []map[string]interface{}{}
		for rows.Next() {
			var code, name, city, country string
			var latitude, longitude sql.NullFloat64
			if err := rows.Scan(&code, &name, &city, &country, &latitude, &longitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			airport := map[string]interface{}{
				"code":    code,
				"name":    name,
				"city":    city,
				"country": country,
			}

			if latitude.Valid {
				airport["latitude"] = latitude.Float64
			}
			if longitude.Valid {
				airport["longitude"] = longitude.Float64
			}

			airports = append(airports, airport)
		}

		c.JSON(http.StatusOK, airports)
	}
}

// getAirlines returns a handler for getting all airlines
func getAirlines(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.GetDB().Query("SELECT code, name, country FROM airlines")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		airlines := []map[string]interface{}{}
		for rows.Next() {
			var code, name, country string
			if err := rows.Scan(&code, &name, &country); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			airline := map[string]interface{}{
				"code":    code,
				"name":    name,
				"country": country,
			}

			airlines = append(airlines, airline)
		}

		c.JSON(http.StatusOK, airlines)
	}
}

// createSearch returns a handler for creating a new flight search
func createSearch(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create a worker payload
		payload := worker.FlightSearchPayload{
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate,
			ReturnDate:    req.ReturnDate,
			Adults:        req.Adults,
			Children:      req.Children,
			InfantsLap:    req.InfantsLap,
			InfantsSeat:   req.InfantsSeat,
			TripType:      req.TripType,
			Class:         req.Class,
			Stops:         req.Stops,
			Currency:      req.Currency,
		}

		// Enqueue the job
		jobID, err := q.Enqueue(c.Request.Context(), "flight_search", payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  jobID,
			"message": "Flight search job created successfully",
		})
	}
}

// getSearchById returns a handler for getting a search by ID
func getSearchById(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		searchID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
			return
		}

		// Get the search query
		var query struct {
			ID            int
			Origin        string
			Destination   string
			DepartureDate time.Time
			ReturnDate    sql.NullTime
			Status        string
			CreatedAt     time.Time
		}

		err = db.GetDB().QueryRow(
			`SELECT id, origin, destination, departure_date, return_date, status, created_at 
			FROM search_queries WHERE id = $1`,
			searchID,
		).Scan(&query.ID, &query.Origin, &query.Destination, &query.DepartureDate, &query.ReturnDate, &query.Status, &query.CreatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Search not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the flight offers
		rows, err := db.GetDB().Query(
			`SELECT id, price, currency, departure_date, return_date, total_duration, created_at 
			FROM flight_offers WHERE search_query_id = $1`,
			searchID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		offers := []map[string]interface{}{}
		for rows.Next() {
			var offer struct {
				ID            int
				Price         float64
				Currency      string
				DepartureDate time.Time
				ReturnDate    sql.NullTime
				TotalDuration int
				CreatedAt     time.Time
			}

			if err := rows.Scan(&offer.ID, &offer.Price, &offer.Currency, &offer.DepartureDate, &offer.ReturnDate, &offer.TotalDuration, &offer.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Get the flight segments for this offer
			segmentRows, err := db.GetDB().Query(
				`SELECT airline_code, flight_number, departure_airport, arrival_airport, 
				departure_time, arrival_time, duration, airplane, legroom, is_return 
				FROM flight_segments WHERE flight_offer_id = $1`,
				offer.ID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer segmentRows.Close()

			segments := []map[string]interface{}{}
			for segmentRows.Next() {
				var segment struct {
					AirlineCode      string
					FlightNumber     string
					DepartureAirport string
					ArrivalAirport   string
					DepartureTime    time.Time
					ArrivalTime      time.Time
					Duration         int
					Airplane         string
					Legroom          string
					IsReturn         bool
				}

				if err := segmentRows.Scan(
					&segment.AirlineCode, &segment.FlightNumber, &segment.DepartureAirport,
					&segment.ArrivalAirport, &segment.DepartureTime, &segment.ArrivalTime,
					&segment.Duration, &segment.Airplane, &segment.Legroom, &segment.IsReturn,
				); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				segments = append(segments, map[string]interface{}{
					"airline_code":      segment.AirlineCode,
					"flight_number":     segment.FlightNumber,
					"departure_airport": segment.DepartureAirport,
					"arrival_airport":   segment.ArrivalAirport,
					"departure_time":    segment.DepartureTime,
					"arrival_time":      segment.ArrivalTime,
					"duration":          segment.Duration,
					"airplane":          segment.Airplane,
					"legroom":           segment.Legroom,
					"is_return":         segment.IsReturn,
				})
			}

			offerMap := map[string]interface{}{
				"id":             offer.ID,
				"price":          offer.Price,
				"currency":       offer.Currency,
				"departure_date": offer.DepartureDate,
				"total_duration": offer.TotalDuration,
				"created_at":     offer.CreatedAt,
				"segments":       segments,
			}

			if offer.ReturnDate.Valid {
				offerMap["return_date"] = offer.ReturnDate.Time
			}

			offers = append(offers, offerMap)
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
func listSearches(db *db.PostgresDB) gin.HandlerFunc {
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
		var total int
		err := db.GetDB().QueryRow("SELECT COUNT(*) FROM search_queries").Scan(&total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the search queries
		rows, err := db.GetDB().Query(
			`SELECT id, origin, destination, departure_date, return_date, status, created_at 
			FROM search_queries ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			perPage, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		queries := []map[string]interface{}{}
		for rows.Next() {
			var query struct {
				ID            int
				Origin        string
				Destination   string
				DepartureDate time.Time
				ReturnDate    sql.NullTime
				Status        string
				CreatedAt     time.Time
			}

			if err := rows.Scan(&query.ID, &query.Origin, &query.Destination, &query.DepartureDate, &query.ReturnDate, &query.Status, &query.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
func deleteJob(db *db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Begin a transaction
		tx, err := db.GetDB().Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Delete the job details
		_, err = tx.Exec("DELETE FROM job_details WHERE job_id = $1", jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the job
		result, err := tx.Exec("DELETE FROM scheduled_jobs WHERE id = $1", jobID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if the job was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Remove the job from the scheduler
		scheduler := workerManager.GetScheduler()
		scheduler.RemoveJob(jobID)

		c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
	}
}

// runJob returns a handler for manually triggering a job
func runJob(q queue.Queue, db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Get the job details
		var details struct {
			Origin             string
			Destination        string
			DepartureDateStart time.Time
			DepartureDateEnd   time.Time
			ReturnDateStart    sql.NullTime
			ReturnDateEnd      sql.NullTime
			TripLength         sql.NullInt32
			Adults             int
			Children           int
			InfantsLap         int
			InfantsSeat        int
			TripType           string
			Class              string
			Stops              string
		}

		err = db.GetDB().QueryRow(
			`SELECT origin, destination, departure_date_start, departure_date_end, 
			return_date_start, return_date_end, trip_length, adults, children, 
			infants_lap, infants_seat, trip_type, class, stops 
			FROM job_details WHERE job_id = $1`,
			jobID,
		).Scan(
			&details.Origin, &details.Destination, &details.DepartureDateStart, &details.DepartureDateEnd,
			&details.ReturnDateStart, &details.ReturnDateEnd, &details.TripLength, &details.Adults, &details.Children,
			&details.InfantsLap, &details.InfantsSeat, &details.TripType, &details.Class, &details.Stops,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			Class:             details.Class,
			Stops:             details.Stops,
			Currency:          "USD", // Default currency
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update the last run time
		_, err = db.GetDB().Exec(
			"UPDATE scheduled_jobs SET last_run = NOW() WHERE id = $1",
			jobID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  jobQueueID,
			"message": "Job triggered successfully",
		})
	}
}

// enableJob returns a handler for enabling a job
func enableJob(db *db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		result, err := db.GetDB().Exec(
			"UPDATE scheduled_jobs SET enabled = true, updated_at = NOW() WHERE id = $1",
			jobID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if the job was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// Get the job's cron expression
		var cronExpr string
		err = db.GetDB().QueryRow("SELECT cron_expression FROM scheduled_jobs WHERE id = $1", jobID).Scan(&cronExpr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job cron expression: " + err.Error()})
			return
		}

		// Add the job to the scheduler
		scheduler := workerManager.GetScheduler()
		if err := scheduler.AddJob(jobID, cronExpr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Job enabled in database but scheduling failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Job enabled successfully"})
	}
}

// disableJob returns a handler for disabling a job
func disableJob(db *db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		result, err := db.GetDB().Exec(
			"UPDATE scheduled_jobs SET enabled = false, updated_at = NOW() WHERE id = $1",
			jobID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if the job was found
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Job disabled successfully"})
	}
}

// getWorkerStatus returns a handler for getting worker status
func getWorkerStatus(workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "running"})
	}
}

// getBulkSearchById returns a handler for getting a bulk search by ID
func getBulkSearchById(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would be similar to getSearchById but would aggregate results from multiple searches
		// For now, we'll return a placeholder
		c.JSON(http.StatusOK, gin.H{"message": "Bulk search results will be implemented here"})
	}
}

// getPriceHistory returns a handler for getting price history for a route
func getPriceHistory(neo4jDB *db.Neo4jDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Param("origin")
		destination := c.Param("destination")

		// Get the Neo4j session
		session := neo4jDB.GetDriver().NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
		defer session.Close()

		// Query price history
		result, err := session.Run(
			"MATCH (origin:Airport {code: $originCode})-[r:PRICE_POINT]->(dest:Airport {code: $destCode}) "+
				"RETURN r.date AS date, r.price AS price, r.airline AS airline ORDER BY r.date",
			map[string]interface{}{
				"originCode": origin,
				"destCode":   destination,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Process the results
		priceHistory := []map[string]interface{}{}
		for result.Next() {
			record := result.Record()
			date, _ := record.Get("date")
			price, _ := record.Get("price")
			airline, _ := record.Get("airline")

			priceHistory = append(priceHistory, map[string]interface{}{
				"date":    date,
				"price":   price,
				"airline": airline,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"origin":      origin,
			"destination": destination,
			"history":     priceHistory,
		})
	}
}

// listJobs returns a handler for listing all scheduled jobs
func listJobs(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.GetDB().Query(
			`SELECT id, name, cron_expression, enabled, last_run, created_at, updated_at 
			FROM scheduled_jobs ORDER BY created_at DESC`,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		jobs := []map[string]interface{}{}
		for rows.Next() {
			var job struct {
				ID             int
				Name           string
				CronExpression string
				Enabled        bool
				LastRun        sql.NullTime
				CreatedAt      time.Time
				UpdatedAt      time.Time
			}

			if err := rows.Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			}

			jobs = append(jobs, jobMap)
		}

		c.JSON(http.StatusOK, jobs)
	}
}

// createJob returns a handler for creating a new scheduled job
func createJob(db *db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		tx, err := db.GetDB().Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Validate cron expression
		parts := strings.Fields(req.CronExpression)
		if len(parts) != 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cron expression format. Must have 5 space-separated fields: minute hour day month weekday"})
			return
		}

		// Insert the job
		var jobID int
		err = tx.QueryRow(
			`INSERT INTO scheduled_jobs (name, cron_expression, enabled) 
			VALUES ($1, $2, $3) RETURNING id`,
			req.Name, strings.Join(parts, " "), true,
		).Scan(&jobID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert the job details
		_, err = tx.Exec(
			`INSERT INTO job_details 
			(job_id, origin, destination, departure_date_start, departure_date_end, 
			return_date_start, return_date_end, trip_length, adults, children, 
			infants_lap, infants_seat, trip_type, class, stops) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
			jobID, req.Origin, req.Destination, dateStart, dateEnd,
			returnDateStart, returnDateEnd, req.TripLength, req.Adults, req.Children,
			req.InfantsLap, req.InfantsSeat, req.TripType, req.Class, req.Stops,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Schedule the job using the scheduler
		scheduler := workerManager.GetScheduler()
		if err := scheduler.AddJob(jobID, req.CronExpression); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Job created in database but scheduling failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":      jobID,
			"message": "Job created and scheduled successfully",
		})
	}
}

// createBulkSearch returns a handler for creating a bulk flight search
func createBulkSearch(q queue.Queue) gin.HandlerFunc {
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
func updateJob(db *db.PostgresDB, workerManager *worker.Manager) gin.HandlerFunc {
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

		// Begin a transaction
		tx, err := db.GetDB().Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Update the job
		_, err = tx.Exec(
			`UPDATE scheduled_jobs SET name = $1, cron_expression = $2, updated_at = NOW() WHERE id = $3`,
			req.Name, req.CronExpression, jobID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update the job details
		_, err = tx.Exec(
			`UPDATE job_details SET origin = $1, destination = $2, departure_date_start = $3, departure_date_end = $4, 
			return_date_start = $5, return_date_end = $6, trip_length = $7, adults = $8, children = $9, 
			infants_lap = $10, infants_seat = $11, trip_type = $12, class = $13, stops = $14 WHERE job_id = $15`,
			req.Origin, req.Destination, req.DateStart, req.DateEnd,
			req.ReturnDateStart, req.ReturnDateEnd, req.TripLength, req.Adults, req.Children,
			req.InfantsLap, req.InfantsSeat, req.TripType, req.Class, req.Stops,
			jobID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update the job schedule using the scheduler
		scheduler := workerManager.GetScheduler()
		if err := scheduler.UpdateJob(jobID, req.CronExpression); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Job updated in database but rescheduling failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      jobID,
			"message": "Job updated and rescheduled successfully",
		})
	}
}

// getQueueStatus returns a handler for getting the status of the queue
func getQueueStatus(q queue.Queue) gin.HandlerFunc {
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
func getJobById(db *db.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		jobID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}

		// Get the job
		var job struct {
			ID             int
			Name           string
			CronExpression string
			Enabled        bool
			LastRun        sql.NullTime
			CreatedAt      time.Time
			UpdatedAt      time.Time
		}

		err = db.GetDB().QueryRow(
			`SELECT id, name, cron_expression, enabled, last_run, created_at, updated_at 
			FROM scheduled_jobs WHERE id = $1`,
			jobID,
		).Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the job details
		var details struct {
			Origin             string
			Destination        string
			DepartureDateStart time.Time
			DepartureDateEnd   time.Time
			ReturnDateStart    sql.NullTime
			ReturnDateEnd      sql.NullTime
			TripLength         sql.NullInt32
			Adults             int
			Children           int
			InfantsLap         int
			InfantsSeat        int
			TripType           string
			Class              string
			Stops              string
		}

		err = db.GetDB().QueryRow(
			`SELECT origin, destination, departure_date_start, departure_date_end, 
			return_date_start, return_date_end, trip_length, adults, children, 
			infants_lap, infants_seat, trip_type, class, stops 
			FROM job_details WHERE job_id = $1`,
			jobID,
		).Scan(
			&details.Origin, &details.Destination, &details.DepartureDateStart, &details.DepartureDateEnd,
			&details.ReturnDateStart, &details.ReturnDateEnd, &details.TripLength, &details.Adults, &details.Children,
			&details.InfantsLap, &details.InfantsSeat, &details.TripType, &details.Class, &details.Stops,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
				"departure_date_start": details.DepartureDateStart,
				"departure_date_end":   details.DepartureDateEnd,
				"adults":               details.Adults,
				"children":             details.Children,
				"infants_lap":          details.InfantsLap,
				"infants_seat":         details.InfantsSeat,
				"trip_type":            details.TripType,
				"class":                details.Class,
				"stops":                details.Stops,
			},
		}

		if job.LastRun.Valid {
			jobMap["last_run"] = job.LastRun.Time
		}

		if details.ReturnDateStart.Valid {
			jobMap["details"].(map[string]interface{})["return_date_start"] = details.ReturnDateStart.Time
		}

		if details.ReturnDateEnd.Valid {
			jobMap["details"].(map[string]interface{})["return_date_end"] = details.ReturnDateEnd.Time
		}

		if details.TripLength.Valid {
			jobMap["details"].(map[string]interface{})["trip_length"] = details.TripLength.Int32
		}

		c.JSON(http.StatusOK, jobMap)
	}
}
