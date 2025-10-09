package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/queue"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// Helper functions to parse string values into flights enum types
// (Replicated from api/handlers.go as they are not exported)
func parseClass(class string) flights.Class {
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

func parseStops(stops string) flights.Stops {
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

// workerState tracks runtime statistics for a worker goroutine.
type workerState struct {
	ID            int
	Status        string
	CurrentJob    string
	ProcessedJobs int
	StartedAt     time.Time
	LastHeartbeat time.Time
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullInt32(value int) sql.NullInt32 {
	if value == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(value), Valid: true}
}

func durationToNullMinutes(d time.Duration) sql.NullInt32 {
	minutes := int(d.Minutes())
	if minutes <= 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(minutes), Valid: true}
}

func flightsToJSON(flights []flights.Flight) []byte {
	legs := make([]map[string]interface{}, 0, len(flights))
	for _, f := range flights {
		leg := map[string]interface{}{
			"dep_airport_code": f.DepAirportCode,
			"dep_airport_name": f.DepAirportName,
			"dep_city":         f.DepCity,
			"arr_airport_code": f.ArrAirportCode,
			"arr_airport_name": f.ArrAirportName,
			"arr_city":         f.ArrCity,
			"dep_time":         f.DepTime,
			"arr_time":         f.ArrTime,
			"duration_minutes": int(f.Duration.Minutes()),
			"airplane":         f.Airplane,
			"flight_number":    f.FlightNumber,
			"airline_name":     f.AirlineName,
			"legroom":          f.Legroom,
		}
		if f.Unknown != nil {
			leg["unknown"] = f.Unknown
		}
		legs = append(legs, leg)
	}
	data, err := json.Marshal(legs)
	if err != nil {
		return []byte("[]")
	}
	return data
}

func offerToJSON(offer flights.FullOffer) []byte {
	data, err := json.Marshal(offer)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// WorkerStatus is a snapshot of worker metrics exposed via the API.
type WorkerStatus struct {
	ID            int    `json:"id"`
	Status        string `json:"status"`
	CurrentJob    string `json:"current_job,omitempty"`
	ProcessedJobs int    `json:"processed_jobs"`
	Uptime        int64  `json:"uptime"` // seconds
}

// Manager manages a pool of workers
type Manager struct {
	queue        queue.Queue
	postgresDB   db.PostgresDB    // Changed type from pointer to interface
	neo4jDB      db.Neo4jDatabase // Use the interface type
	config       config.WorkerConfig
	workers      []*Worker
	stopChan     chan struct{}
	workerWg     sync.WaitGroup
	flightCache  map[string]*flights.Session
	cacheMutex   sync.RWMutex
	scheduler    *Scheduler
	statsMutex   sync.RWMutex
	workerStates []*workerState
}

// NewManager creates a new worker manager
func NewManager(queue queue.Queue, postgresDB db.PostgresDB, neo4jDB db.Neo4jDatabase, config config.WorkerConfig) *Manager { // Changed neo4jDB parameter type to interface
	// Pass nil for Cronner to use the default cron instance
	scheduler := NewScheduler(queue, postgresDB, nil)
	return &Manager{
		queue:        queue,
		postgresDB:   postgresDB, // Use unexported field name
		neo4jDB:      neo4jDB,    // Use unexported field name
		config:       config,
		stopChan:     make(chan struct{}),
		flightCache:  make(map[string]*flights.Session),
		scheduler:    scheduler,
		workerStates: make([]*workerState, config.Concurrency),
	}
}

// updateWorkerState applies the provided update function while holding the mutex.
func (m *Manager) updateWorkerState(workerIndex int, updateFn func(*workerState)) {
	if updateFn == nil || workerIndex < 0 || workerIndex >= len(m.workerStates) {
		return
	}

	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	state := m.workerStates[workerIndex]
	if state == nil {
		state = &workerState{ID: workerIndex + 1}
		m.workerStates[workerIndex] = state
	}
	updateFn(state)
}

// WorkerStatuses returns a snapshot of current worker metrics.
func (m *Manager) WorkerStatuses() []WorkerStatus {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	statuses := make([]WorkerStatus, 0, len(m.workerStates))
	now := time.Now()
	for _, state := range m.workerStates {
		if state == nil {
			continue
		}

		uptime := int64(0)
		if !state.StartedAt.IsZero() {
			uptime = int64(now.Sub(state.StartedAt).Seconds())
			if uptime < 0 {
				uptime = 0
			}
		}

		statuses = append(statuses, WorkerStatus{
			ID:            state.ID,
			Status:        state.Status,
			CurrentJob:    state.CurrentJob,
			ProcessedJobs: state.ProcessedJobs,
			Uptime:        uptime,
		})
	}

	return statuses
}

// Start starts the worker pool and scheduler
func (m *Manager) Start() {
	log.Printf("Starting worker pool with %d workers", m.config.Concurrency)

	now := time.Now()
	m.statsMutex.Lock()
	for i := 0; i < m.config.Concurrency; i++ {
		m.workerStates[i] = &workerState{
			ID:            i + 1,
			Status:        "starting",
			StartedAt:     now,
			LastHeartbeat: now,
		}
	}
	m.statsMutex.Unlock()

	// Create and start workers
	for i := 0; i < m.config.Concurrency; i++ {
		worker := &Worker{
			postgresDB: m.postgresDB,
			neo4jDB:    m.neo4jDB,
		}
		m.workers = append(m.workers, worker)

		m.workerWg.Add(1)
		go m.runWorker(i, worker)
	}

	// Start the scheduler
	if err := m.scheduler.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler: %v", err)
	} else {
		log.Println("Scheduler started successfully")
	}
}

// Stop stops the worker pool and scheduler
func (m *Manager) Stop() {
	log.Println("Stopping worker pool and scheduler")

	now := time.Now()
	m.statsMutex.Lock()
	for _, state := range m.workerStates {
		if state != nil {
			state.Status = "stopping"
			state.CurrentJob = ""
			state.LastHeartbeat = now
		}
	}
	m.statsMutex.Unlock()

	// Stop the scheduler
	m.scheduler.Stop()

	// Signal all workers to stop
	close(m.stopChan)

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		m.workerWg.Wait()
		close(done)
	}()

	// Wait for workers to finish or timeout
	select {
	case <-done:
		log.Println("All workers stopped gracefully")
	case <-time.After(m.config.ShutdownTimeout):
		log.Println("Worker shutdown timed out")
	}

	m.statsMutex.Lock()
	for _, state := range m.workerStates {
		if state != nil {
			state.Status = "stopped"
			state.LastHeartbeat = time.Now()
			state.CurrentJob = ""
		}
	}
	m.statsMutex.Unlock()

	// Clear flight sessions cache
	m.cacheMutex.Lock()
	m.flightCache = make(map[string]*flights.Session)
	m.cacheMutex.Unlock()
}

// runWorker runs a worker in a goroutine
func (m *Manager) runWorker(id int, worker *Worker) {
	defer m.workerWg.Done()
	displayID := id + 1
	now := time.Now()
	m.updateWorkerState(id, func(state *workerState) {
		if state.StartedAt.IsZero() {
			state.StartedAt = now
		}
		state.Status = "active"
		state.CurrentJob = ""
		state.LastHeartbeat = now
	})

	log.Printf("Worker %d started", displayID)

	for {
		select {
		case <-m.stopChan:
			log.Printf("Worker %d stopping", displayID)
			m.updateWorkerState(id, func(state *workerState) {
				state.Status = "stopped"
				state.CurrentJob = ""
				state.LastHeartbeat = time.Now()
			})
			return
		default:
			// Process jobs from different queues
			if err := m.processQueue(id, worker, "flight_search"); err != nil {
				log.Printf("Worker %d error processing flight_search queue: %v", displayID, err)
			}

			if err := m.processQueue(id, worker, "bulk_search"); err != nil {
				log.Printf("Worker %d error processing bulk_search queue: %v", displayID, err)
			}

			// Sleep briefly to avoid hammering the queue
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processQueue processes a job from the specified queue
func (m *Manager) processQueue(workerIndex int, worker *Worker, queueName string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.config.JobTimeout)
	defer cancel()

	// Log context deadline for debugging
	if deadline, ok := ctx.Deadline(); ok {
		log.Printf("Created context with deadline: %v (timeout: %v)", deadline.Format("15:04:05"), m.config.JobTimeout)
	}

	// Dequeue a job
	job, err := m.queue.Dequeue(ctx, queueName)
	if err != nil {
		// Don't return error if context times out waiting for job
		if ctx.Err() == context.DeadlineExceeded {
			m.updateWorkerState(workerIndex, func(state *workerState) {
				state.LastHeartbeat = time.Now()
			})
			return nil
		}
		m.updateWorkerState(workerIndex, func(state *workerState) {
			state.Status = "error"
			state.CurrentJob = ""
			state.LastHeartbeat = time.Now()
		})
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	// No jobs available
	if job == nil {
		m.updateWorkerState(workerIndex, func(state *workerState) {
			if state.Status != "processing" {
				state.Status = "active"
			}
			state.CurrentJob = ""
			state.LastHeartbeat = time.Now()
		})
		return nil
	}

	m.updateWorkerState(workerIndex, func(state *workerState) {
		state.Status = "processing"
		state.CurrentJob = fmt.Sprintf("%s:%s", queueName, job.ID)
		state.LastHeartbeat = time.Now()
	})

	jobStartTime := time.Now()
	log.Printf("Processing %s job %s (started at %v)", queueName, job.ID, jobStartTime.Format("15:04:05"))

	// Process the job
	err = m.processJob(ctx, worker, queueName, job)
	jobDuration := time.Since(jobStartTime)

	if err != nil {
		log.Printf("Error processing job %s after %v: %v", job.ID, jobDuration, err)

		// Check if context deadline was exceeded
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Job %s timed out after %v (deadline exceeded)", job.ID, jobDuration)
		}

		// Nack the job
		if nackErr := m.queue.Nack(ctx, queueName, job.ID); nackErr != nil {
			log.Printf("Error nacking job %s: %v", job.ID, nackErr)
		}
		m.updateWorkerState(workerIndex, func(state *workerState) {
			state.Status = "active"
			state.CurrentJob = ""
			state.LastHeartbeat = time.Now()
		})
		return fmt.Errorf("failed to process job: %w", err)
	}

	// Ack the job
	if ackErr := m.queue.Ack(ctx, queueName, job.ID); ackErr != nil {
		log.Printf("Error acking job %s: %v", job.ID, ackErr)
		m.updateWorkerState(workerIndex, func(state *workerState) {
			state.Status = "active"
			state.CurrentJob = ""
			state.LastHeartbeat = time.Now()
		})
		return fmt.Errorf("failed to ack job: %w", ackErr)
	}

	m.updateWorkerState(workerIndex, func(state *workerState) {
		state.Status = "active"
		state.CurrentJob = ""
		state.ProcessedJobs++
		state.LastHeartbeat = time.Now()
	})

	log.Printf("Completed %s job %s in %v", queueName, job.ID, jobDuration)
	return nil
}

// processJob processes a job based on its type
func (m *Manager) processJob(ctx context.Context, worker *Worker, queueName string, job *queue.Job) error {
	switch queueName {
	case "flight_search":
		// Get cached session for direct search (works fine)
		session, err := m.getFlightSession("direct_search")
		if err != nil {
			return fmt.Errorf("failed to get flight session: %w", err)
		}

		var payload FlightSearchPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal flight search payload: %w", err)
		}

		// Process the flight search
		return m.processFlightSearch(ctx, worker, session, payload)
	case "bulk_search":
		// Get fresh session for bulk search (avoids stale session issues)
		session, err := m.getFlightSession("bulk_search")
		if err != nil {
			return fmt.Errorf("failed to get flight session: %w", err)
		}

		var payload BulkSearchPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal bulk search payload: %w", err)
		}

		// Process the bulk search
		return m.processBulkSearch(ctx, worker, session, payload)

	default:
		return fmt.Errorf("unknown job type: %s", queueName)
	}
}

// getFlightSession gets or creates a flight session based on session type
func (m *Manager) getFlightSession(sessionType string) (*flights.Session, error) {
	var sessionKey string

	// Use consistent caching strategy for both search types since regular search works fine
	switch sessionType {
	case "bulk_search":
		// Use cached session like direct search for better reliability
		sessionKey = "bulk_search"
	case "direct_search":
		// Use cached session for direct search (works fine)
		sessionKey = "direct_search"
	default:
		// Fallback to default behavior
		sessionKey = "default"
	}

	// Check if we have a cached session
	m.cacheMutex.RLock()
	session, exists := m.flightCache[sessionKey]
	m.cacheMutex.RUnlock()

	if exists {
		return session, nil
	}

	// Create a new session
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Check again in case another goroutine created the session
	session, exists = m.flightCache[sessionKey]
	if exists {
		return session, nil
	}

	// Create a new session
	newSession, err := flights.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create flight session: %w", err)
	}

	m.flightCache[sessionKey] = newSession
	return newSession, nil
}

// processFlightSearch processes a flight search job
func (m *Manager) processFlightSearch(ctx context.Context, worker *Worker, session *flights.Session, payload FlightSearchPayload) error {
	// Map the payload to flights API arguments
	var tripType flights.TripType
	switch payload.TripType {
	case "one_way":
		tripType = flights.OneWay
	case "round_trip":
		tripType = flights.RoundTrip
	default:
		return fmt.Errorf("invalid trip type: %s", payload.TripType)
	}

	// Parse currency
	cur, err := currency.ParseISO(payload.Currency)
	if err != nil {
		// Default to USD if currency parsing fails
		log.Printf("Warning: Invalid currency '%s', defaulting to USD. Error: %v", payload.Currency, err)
		cur = currency.USD
	}

	// Parse class and stops after unmarshaling
	flightClass := parseClass(payload.Class)
	flightStops := parseStops(payload.Stops)

	// Get flight offers
	offers, priceRange, err := session.GetOffers(
		ctx,
		flights.Args{
			Date:        payload.DepartureDate,
			ReturnDate:  payload.ReturnDate,
			SrcAirports: []string{payload.Origin},
			DstAirports: []string{payload.Destination},
			Options: flights.Options{
				Travelers: flights.Travelers{
					Adults:       payload.Adults,
					Children:     payload.Children,
					InfantOnLap:  payload.InfantsLap,
					InfantInSeat: payload.InfantsSeat,
				},
				Currency: cur,
				Stops:    flightStops, // Use parsed value
				Class:    flightClass, // Use parsed value
				TripType: tripType,
				Lang:     language.English,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get flight offers: %w", err)
	}

	// Store the results
	// Pass the original payload (with string Class/Stops) to StoreFlightOffers
	// StoreFlightOffers itself uses the string values when inserting into search_queries
	return worker.StoreFlightOffers(ctx, payload, offers, priceRange)
}

// processBulkSearch processes a bulk search job
func (m *Manager) processBulkSearch(ctx context.Context, worker *Worker, session *flights.Session, payload BulkSearchPayload) (err error) {
	// Validate origins and destinations
	if len(payload.Origins) == 0 {
		return fmt.Errorf("bulk search payload requires at least one origin")
	}
	if len(payload.Destinations) == 0 {
		return fmt.Errorf("bulk search payload requires at least one destination")
	}

	// Map the payload to flights API arguments
	var tripType flights.TripType
	switch payload.TripType {
	case "one_way":
		tripType = flights.OneWay
	case "round_trip":
		tripType = flights.RoundTrip
	default:
		return fmt.Errorf("invalid trip type: %s", payload.TripType)
	}

	// Parse currency
	cur, err := currency.ParseISO(payload.Currency)
	if err != nil {
		// Default to USD if currency parsing fails
		log.Printf("Warning: Invalid currency '%s', defaulting to USD. Error: %v", payload.Currency, err)
		cur = currency.USD
	}

	// Parse class and stops after unmarshaling
	flightClass := parseClass(payload.Class)
	flightStops := parseStops(payload.Stops)

	totalRoutes := len(payload.Origins) * len(payload.Destinations)
	if totalRoutes == 0 {
		return fmt.Errorf("bulk search payload requires at least one origin and destination combination")
	}

	var bulkSearchID int
	if payload.BulkSearchID > 0 {
		bulkSearchID = payload.BulkSearchID
		if updateErr := m.postgresDB.UpdateBulkSearchStatus(ctx, bulkSearchID, "running"); updateErr != nil {
			log.Printf("Failed to update bulk search %d status to running: %v", bulkSearchID, updateErr)
		}
	} else {
		var jobRef sql.NullInt32
		if payload.JobID > 0 {
			jobRef = sql.NullInt32{Int32: int32(payload.JobID), Valid: true}
		}
		newID, createErr := m.postgresDB.CreateBulkSearchRecord(ctx, jobRef, totalRoutes, payload.Currency, "running")
		if createErr != nil {
			return fmt.Errorf("failed to create bulk search record: %w", createErr)
		}
		bulkSearchID = newID
	}

	defer func() {
		if bulkSearchID > 0 && err != nil {
			if updateErr := m.postgresDB.UpdateBulkSearchStatus(ctx, bulkSearchID, "failed"); updateErr != nil {
				log.Printf("Failed to mark bulk search %d as failed: %v", bulkSearchID, updateErr)
			}
		}
	}()

	// Process all origin/destination combinations to find lowest fares
	log.Printf("Starting bulk search: %d origins × %d destinations = %d route combinations",
		len(payload.Origins), len(payload.Destinations), totalRoutes)

	// Calculate date range for searches
	dateRange := m.generateDateRange(payload.DepartureDateFrom, payload.DepartureDateTo, payload.TripLength)
	log.Printf("Searching across %d dates: %s to %s", len(dateRange),
		payload.DepartureDateFrom.Format("2006-01-02"), payload.DepartureDateTo.Format("2006-01-02"))

	// Track lowest fares for each route
	type RouteLowestFare struct {
		Route         string
		Origin        string
		Destination   string
		BestOffer     flights.FullOffer
		BestDate      time.Time
		BestReturn    time.Time
		SearchedDates int
		TotalOffers   int
	}

	routeResults := make(map[string]*RouteLowestFare)
	var searchErrors []error

	for _, origin := range payload.Origins {
		for _, destination := range payload.Destinations {
			routeKey := fmt.Sprintf("%s-%s", origin, destination)
			log.Printf("Processing route: %s", routeKey)

			routeResult := &RouteLowestFare{
				Route:       routeKey,
				Origin:      origin,
				Destination: destination,
				BestOffer:   flights.FullOffer{Offer: flights.Offer{Price: math.MaxFloat64}}, // Initialize with max price
			}

			// Search across all dates in range to find lowest fare
			for _, searchDate := range dateRange {
				var returnDate time.Time
				if tripType == flights.RoundTrip {
					if payload.TripLength > 0 {
						returnDate = searchDate.AddDate(0, 0, payload.TripLength)
					} else {
						// Use return date range midpoint if no trip length specified
						returnDateRange := payload.ReturnDateTo.Sub(payload.ReturnDateFrom)
						returnDate = payload.ReturnDateFrom.Add(returnDateRange / 2)
					}
				}

				// Search for flights on this specific date
				offers, priceRange, err := session.GetOffers(
					ctx,
					flights.Args{
						Date:        searchDate,
						ReturnDate:  returnDate,
						SrcAirports: []string{origin},
						DstAirports: []string{destination},
						Options: flights.Options{
							Travelers: flights.Travelers{
								Adults:       payload.Adults,
								Children:     payload.Children,
								InfantOnLap:  payload.InfantsLap,
								InfantInSeat: payload.InfantsSeat,
							},
							Currency: cur,
							Stops:    flightStops,
							Class:    flightClass,
							TripType: tripType,
							Lang:     language.English,
						},
					},
				)

				if err != nil {
					log.Printf("Error searching %s -> %s on %s: %v", origin, destination, searchDate.Format("2006-01-02"), err)
					searchErrors = append(searchErrors, fmt.Errorf("search %s->%s on %s failed: %w", origin, destination, searchDate.Format("2006-01-02"), err))
					continue
				}

				routeResult.SearchedDates++
				routeResult.TotalOffers += len(offers)

				// Find the lowest price offer from this date
				for _, offer := range offers {
					if offer.Price < routeResult.BestOffer.Price {
						routeResult.BestOffer = offer
						routeResult.BestDate = searchDate
						routeResult.BestReturn = returnDate
						log.Printf("New lowest fare for %s: $%.2f on %s", routeKey, offer.Price, searchDate.Format("2006-01-02"))
					}
				}

				// Store all results for this specific search (for historical tracking)
				if len(offers) > 0 {
					searchPayload := FlightSearchPayload{
						Origin:        origin,
						Destination:   destination,
						DepartureDate: searchDate,
						ReturnDate:    returnDate,
						Adults:        payload.Adults,
						Children:      payload.Children,
						InfantsLap:    payload.InfantsLap,
						InfantsSeat:   payload.InfantsSeat,
						TripType:      payload.TripType,
						Class:         payload.Class,
						Stops:         payload.Stops,
						Currency:      payload.Currency,
					}

					if err := worker.StoreFlightOffers(ctx, searchPayload, offers, priceRange); err != nil {
						log.Printf("Error storing offers for %s -> %s on %s: %v", origin, destination, searchDate.Format("2006-01-02"), err)
						searchErrors = append(searchErrors, fmt.Errorf("failed to store offers for %s->%s on %s: %w", origin, destination, searchDate.Format("2006-01-02"), err))
					}
				}
			}

			// Only keep routes that found valid offers
			if routeResult.BestOffer.Price < math.MaxFloat64 {
				routeResults[routeKey] = routeResult
				log.Printf("Route %s completed: lowest fare $%.2f on %s (searched %d dates, found %d total offers)",
					routeKey, routeResult.BestOffer.Price, routeResult.BestDate.Format("2006-01-02"),
					routeResult.SearchedDates, routeResult.TotalOffers)
			} else {
				log.Printf("Route %s: no valid offers found across %d dates", routeKey, routeResult.SearchedDates)
			}
		}
	}

	// Store the lowest fare results summary
	log.Printf("Bulk search summary: found lowest fares for %d out of %d routes", len(routeResults), totalRoutes)
	for routeKey, result := range routeResults {
		// Store the best offer for each route
		bestSearchPayload := FlightSearchPayload{
			Origin:        result.Origin,
			Destination:   result.Destination,
			DepartureDate: result.BestDate,
			ReturnDate:    result.BestReturn,
			Adults:        payload.Adults,
			Children:      payload.Children,
			InfantsLap:    payload.InfantsLap,
			InfantsSeat:   payload.InfantsSeat,
			TripType:      payload.TripType,
			Class:         payload.Class,
			Stops:         payload.Stops,
			Currency:      payload.Currency,
		}

		// Store just the best offer with a special marker
		if err := worker.StoreFlightOffers(ctx, bestSearchPayload, []flights.FullOffer{result.BestOffer}, nil); err != nil {
			log.Printf("Error storing best offer for route %s: %v", routeKey, err)
			searchErrors = append(searchErrors, fmt.Errorf("failed to store best offer for route %s: %w", routeKey, err))
		}
	}

	// If we had some errors but also some successes, log the errors but don't fail the job
	if len(searchErrors) > 0 {
		if len(routeResults) > 0 {
			log.Printf("Bulk search completed with %d errors, but found lowest fares for %d routes", len(searchErrors), len(routeResults))
			for _, err := range searchErrors {
				log.Printf("Search error: %v", err)
			}
		} else {
			// All searches failed
			status := "failed"
			summary := db.BulkSearchSummary{
				ID:          bulkSearchID,
				Status:      status,
				Completed:   0,
				TotalOffers: 0,
				ErrorCount:  len(searchErrors),
			}
			if completeErr := m.postgresDB.CompleteBulkSearch(ctx, summary); completeErr != nil {
				log.Printf("Failed to finalize failed bulk search %d: %v", bulkSearchID, completeErr)
			}
			return fmt.Errorf("all bulk searches failed: %d errors occurred", len(searchErrors))
		}
	}

	// Persist best offers per route
	completedRoutes := 0
	totalOffers := 0
	var minPrice float64 = math.MaxFloat64
	var maxPrice float64
	var sumPrice float64

	for _, result := range routeResults {
		if result.BestOffer.Price == math.MaxFloat64 {
			continue
		}

		departureDate := result.BestDate
		returnDate := sql.NullTime{}
		if !result.BestReturn.IsZero() {
			returnDate = sql.NullTime{Time: result.BestReturn, Valid: true}
		}

		airlineCode := ""
		if len(result.BestOffer.Flight) > 0 {
			flightNumber := result.BestOffer.Flight[0].FlightNumber
			if len(flightNumber) >= 2 {
				airlineCode = flightNumber[:2]
			}
		}

		flightDuration := int(result.BestOffer.FlightDuration.Minutes())
		returnFlightDuration := int(result.BestOffer.ReturnFlightDuration.Minutes())
		totalDuration := flightDuration + returnFlightDuration

		record := db.BulkSearchResultRecord{
			BulkSearchID:         bulkSearchID,
			Origin:               result.Origin,
			Destination:          result.Destination,
			DepartureDate:        departureDate,
			ReturnDate:           returnDate,
			Price:                result.BestOffer.Price,
			Currency:             strings.ToUpper(payload.Currency),
			AirlineCode:          nullString(airlineCode),
			Duration:             nullInt32(totalDuration),
			SrcAirportCode:       nullString(result.BestOffer.SrcAirportCode),
			DstAirportCode:       nullString(result.BestOffer.DstAirportCode),
			SrcCity:              nullString(result.BestOffer.SrcCity),
			DstCity:              nullString(result.BestOffer.DstCity),
			FlightDuration:       durationToNullMinutes(result.BestOffer.FlightDuration),
			ReturnFlightDuration: durationToNullMinutes(result.BestOffer.ReturnFlightDuration),
			OutboundFlightsJSON:  flightsToJSON(result.BestOffer.Flight),
			ReturnFlightsJSON:    flightsToJSON(result.BestOffer.ReturnFlight),
			OfferJSON:            offerToJSON(result.BestOffer),
		}

		if insertErr := m.postgresDB.InsertBulkSearchResult(ctx, record); insertErr != nil {
			log.Printf("Failed to insert bulk search result for %s -> %s: %v", result.Origin, result.Destination, insertErr)
		}

		completedRoutes++
		totalOffers += result.TotalOffers
		sumPrice += result.BestOffer.Price
		if result.BestOffer.Price < minPrice {
			minPrice = result.BestOffer.Price
		}
		if result.BestOffer.Price > maxPrice {
			maxPrice = result.BestOffer.Price
		}
	}

	var minPriceNull, maxPriceNull, avgPriceNull sql.NullFloat64
	if completedRoutes > 0 {
		minPriceNull = sql.NullFloat64{Float64: minPrice, Valid: true}
		maxPriceNull = sql.NullFloat64{Float64: maxPrice, Valid: true}
		avgPriceNull = sql.NullFloat64{Float64: sumPrice / float64(completedRoutes), Valid: true}
	} else {
		minPriceNull = sql.NullFloat64{Valid: false}
		maxPriceNull = sql.NullFloat64{Valid: false}
		avgPriceNull = sql.NullFloat64{Valid: false}
	}

	status := "completed"
	if len(searchErrors) > 0 && completedRoutes > 0 {
		status = "completed_with_errors"
	}
	if completedRoutes == 0 {
		status = "failed"
	}

	log.Printf("Bulk search completed: found lowest fares for %d routes, processed %d total offers", completedRoutes, totalOffers)

	summary := db.BulkSearchSummary{
		ID:           bulkSearchID,
		Status:       status,
		Completed:    completedRoutes,
		TotalOffers:  totalOffers,
		ErrorCount:   len(searchErrors),
		MinPrice:     minPriceNull,
		MaxPrice:     maxPriceNull,
		AveragePrice: avgPriceNull,
	}

	if completeErr := m.postgresDB.CompleteBulkSearch(ctx, summary); completeErr != nil {
		log.Printf("Failed to update bulk search summary for %d: %v", bulkSearchID, completeErr)
	}

	return nil
}

// generateDateRange generates a slice of dates within the given range for searching
func (m *Manager) generateDateRange(startDate, endDate time.Time, tripLength int) []time.Time {
	var dates []time.Time

	// If start and end are the same, just return that date
	if startDate.Equal(endDate) {
		return []time.Time{startDate}
	}

	// Generate dates within the range (limit to reasonable number to avoid too many API calls)
	current := startDate
	maxDates := 14 // Limit to 2 weeks max to avoid overwhelming the API
	count := 0

	for !current.After(endDate) && count < maxDates {
		dates = append(dates, current)
		current = current.AddDate(0, 0, 1) // Add one day
		count++
	}

	// If no dates were generated, at least return the start date
	if len(dates) == 0 {
		dates = append(dates, startDate)
	}

	return dates
}

// GetScheduler returns the scheduler instance
func (m *Manager) GetScheduler() *Scheduler {
	return m.scheduler
}
