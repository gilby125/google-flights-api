package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// Manager manages a pool of workers
type Manager struct {
	queue       queue.Queue
	postgresDB  db.PostgresDB    // Changed type from pointer to interface
	neo4jDB     db.Neo4jDatabase // Use the interface type
	config      config.WorkerConfig
	workers     []*Worker
	stopChan    chan struct{}
	workerWg    sync.WaitGroup
	flightCache map[string]*flights.Session
	cacheMutex  sync.RWMutex
	scheduler   *Scheduler
}

// NewManager creates a new worker manager
func NewManager(queue queue.Queue, postgresDB db.PostgresDB, neo4jDB db.Neo4jDatabase, config config.WorkerConfig) *Manager { // Changed neo4jDB parameter type to interface
	// Pass nil for Cronner to use the default cron instance
	scheduler := NewScheduler(queue, postgresDB, nil)
	return &Manager{
		queue:       queue,
		postgresDB:  postgresDB, // Use unexported field name
		neo4jDB:     neo4jDB,    // Use unexported field name
		config:      config,
		stopChan:    make(chan struct{}),
		flightCache: make(map[string]*flights.Session),
		scheduler:   scheduler,
	}
}

// Start starts the worker pool and scheduler
func (m *Manager) Start() {
	log.Printf("Starting worker pool with %d workers", m.config.Concurrency)

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

	// Clear flight sessions cache
	m.cacheMutex.Lock()
	m.flightCache = make(map[string]*flights.Session)
	m.cacheMutex.Unlock()
}

// runWorker runs a worker in a goroutine
func (m *Manager) runWorker(id int, worker *Worker) {
	defer m.workerWg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-m.stopChan:
			log.Printf("Worker %d stopping", id)
			return
		default:
			// Process jobs from different queues
			if err := m.processQueue(worker, "flight_search"); err != nil {
				log.Printf("Worker %d error processing flight_search queue: %v", id, err)
			}

			if err := m.processQueue(worker, "bulk_search"); err != nil {
				log.Printf("Worker %d error processing bulk_search queue: %v", id, err)
			}

			// Sleep briefly to avoid hammering the queue
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processQueue processes a job from the specified queue
func (m *Manager) processQueue(worker *Worker, queueName string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.config.JobTimeout)
	defer cancel()

	// Dequeue a job
	job, err := m.queue.Dequeue(ctx, queueName)
	if err != nil {
		// Don't return error if context times out waiting for job
		if ctx.Err() == context.DeadlineExceeded {
			return nil
		}
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	// No jobs available
	if job == nil {
		return nil
	}

	log.Printf("Processing %s job %s", queueName, job.ID)

	// Process the job
	err = m.processJob(ctx, worker, queueName, job)
	if err != nil {
		log.Printf("Error processing job %s: %v", job.ID, err)
		// Nack the job
		if nackErr := m.queue.Nack(ctx, queueName, job.ID); nackErr != nil {
			log.Printf("Error nacking job %s: %v", job.ID, nackErr)
		}
		return fmt.Errorf("failed to process job: %w", err)
	}

	// Ack the job
	if ackErr := m.queue.Ack(ctx, queueName, job.ID); ackErr != nil {
		log.Printf("Error acking job %s: %v", job.ID, ackErr)
		return fmt.Errorf("failed to ack job: %w", ackErr)
	}

	log.Printf("Completed %s job %s", queueName, job.ID)
	return nil
}

// processJob processes a job based on its type
func (m *Manager) processJob(ctx context.Context, worker *Worker, queueName string, job *queue.Job) error {
	// Get or create a flight session
	session, err := m.getFlightSession()
	if err != nil {
		return fmt.Errorf("failed to get flight session: %w", err)
	}

	switch queueName {
	case "flight_search":
		var payload FlightSearchPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal flight search payload: %w", err)
		}

		// Process the flight search
		return m.processFlightSearch(ctx, worker, session, payload) // processFlightSearch calls worker.StoreFlightOffers internally
	case "bulk_search":
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

// getFlightSession gets or creates a flight session
func (m *Manager) getFlightSession() (*flights.Session, error) {
	// Use a simple round-robin approach for now
	sessionKey := "default"

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
func (m *Manager) processBulkSearch(ctx context.Context, worker *Worker, session *flights.Session, payload BulkSearchPayload) error {
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

	// FIX: Use the first origin/destination from the slices
	var origin string
	if len(payload.Origins) > 0 {
		origin = payload.Origins[0]
	} else {
		log.Printf("Warning: Bulk search payload has empty Origins slice.")
		// Optionally return an error or handle as appropriate
		// return fmt.Errorf("bulk search payload requires at least one origin")
	}
	var destination string
	if len(payload.Destinations) > 0 {
		destination = payload.Destinations[0]
	} else {
		log.Printf("Warning: Bulk search payload has empty Destinations slice.")
		// Optionally return an error or handle as appropriate
		// return fmt.Errorf("bulk search payload requires at least one destination")
	}

	// Get price graph data
	offers, err := session.GetPriceGraph(
		ctx,
		flights.PriceGraphArgs{
			RangeStartDate: payload.DepartureDateFrom,
			RangeEndDate:   payload.DepartureDateTo,
			TripLength:     payload.TripLength,
			SrcAirports:    []string{origin},      // Use extracted origin
			DstAirports:    []string{destination}, // Use extracted destination
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
		return fmt.Errorf("failed to get price graph data: %w", err)
	}

	// Convert Offer to FullOffer
	fullOffers := make([]flights.FullOffer, 0, len(offers))
	for _, offer := range offers {
		fullOffer := flights.FullOffer{
			Offer:          offer,
			SrcAirportCode: origin,      // Use extracted origin
			DstAirportCode: destination, // Use extracted destination
			// We don't have flight details from price graph, so we'll leave other fields empty
			Flight:         []flights.Flight{},
			FlightDuration: 0,
		}
		fullOffers = append(fullOffers, fullOffer)
	}

	// Create a FlightSearchPayload from the BulkSearchPayload
	// FIX: Use extracted origin/destination
	flightSearchPayload := FlightSearchPayload{
		Origin:        origin,
		Destination:   destination,
		DepartureDate: payload.DepartureDateFrom, // Use RangeStartDate as the representative departure date
		// ReturnDate needs careful consideration for bulk searches.
		// If TripLength is defined, we might calculate it, otherwise leave it zero.
		// For simplicity here, we'll leave it zero. A more robust solution might be needed.
		ReturnDate:  time.Time{},
		Adults:      payload.Adults,
		Children:    payload.Children,
		InfantsLap:  payload.InfantsLap,
		InfantsSeat: payload.InfantsSeat,
		TripType:    payload.TripType,
		Class:       payload.Class, // Keep as string for StoreFlightOffers
		Stops:       payload.Stops, // Keep as string for StoreFlightOffers
		Currency:    payload.Currency,
	}

	// Process the results
	if err := worker.StoreFlightOffers(ctx, flightSearchPayload, fullOffers, nil); err != nil { // Use exported method
		return fmt.Errorf("failed to store flight offers: %w", err)
	}

	return nil
}

// GetScheduler returns the scheduler instance
func (m *Manager) GetScheduler() *Scheduler {
	return m.scheduler
}
