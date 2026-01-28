package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/iata"
	"github.com/gilby125/google-flights-api/pkg/deals"
	"github.com/gilby125/google-flights-api/pkg/geo"
	"github.com/gilby125/google-flights-api/pkg/worker_registry"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/redis/go-redis/v9"
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

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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

func timeToNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// calculateDistanceAndCostPerMile returns distance in miles and cost per mile for a route.
// Uses IATA airport coordinates for calculation. Returns null values if coordinates are unavailable.
func calculateDistanceAndCostPerMile(origin, destination string, price float64) (sql.NullFloat64, sql.NullFloat64) {
	originLoc := iata.IATATimeZone(origin)
	destLoc := iata.IATATimeZone(destination)

	// Check if we have valid coordinates (non-zero lat/lon)
	if (originLoc.Lat == 0 && originLoc.Lon == 0) || (destLoc.Lat == 0 && destLoc.Lon == 0) {
		return sql.NullFloat64{}, sql.NullFloat64{}
	}

	distance := geo.Haversine(originLoc.Lat, originLoc.Lon, destLoc.Lat, destLoc.Lon)
	costPerMile := geo.CostPerMile(price, distance)

	return sql.NullFloat64{Float64: distance, Valid: true},
		sql.NullFloat64{Float64: costPerMile, Valid: true}
}

func flightsToJSON(flights []flights.Flight) []byte {
	legs := make([]map[string]any, 0, len(flights))
	for _, f := range flights {
		leg := map[string]any{
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

func airlineCodesFromOffer(offer flights.FullOffer) []string {
	codesSet := make(map[string]struct{})
	for _, leg := range offer.Flight {
		if len(leg.FlightNumber) >= 2 {
			codesSet[leg.FlightNumber[:2]] = struct{}{}
		}
	}
	for _, leg := range offer.ReturnFlight {
		if len(leg.FlightNumber) >= 2 {
			codesSet[leg.FlightNumber[:2]] = struct{}{}
		}
	}
	codes := make([]string, 0, len(codesSet))
	for code := range codesSet {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
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
	queue         queue.Queue
	postgresDB    db.PostgresDB    // Changed type from pointer to interface
	neo4jDB       db.Neo4jDatabase // Use the interface type
	config        config.WorkerConfig
	flightConfig  config.FlightConfig
	dealConfig    config.DealConfig
	topNDeals     int
	excludedSet   map[string]bool
	workers       []*Worker
	stopChan      chan struct{}
	workerWg      sync.WaitGroup
	flightCache   map[string]*flights.Session
	cacheMutex    sync.RWMutex
	scheduler     *Scheduler
	statsMutex    sync.RWMutex
	workerStates  []*workerState
	leaderElector *LeaderElector
	redisClient   *redis.Client
	sweepRunner   *ContinuousSweepRunner
	sweepMutex    sync.RWMutex // Protects sweepRunner access

	bulkBusyMu        sync.Mutex
	bulkBusyCached    bool
	bulkBusyCheckedAt time.Time
}

// NewManager creates a new worker manager.
// If redisClient is provided, leader election is enabled for the scheduler.
// If redisClient is nil, the scheduler runs on every instance (legacy behavior).
func NewManager(queue queue.Queue, redisClient *redis.Client, postgresDB db.PostgresDB, neo4jDB db.Neo4jDatabase, workerConfig config.WorkerConfig, flightConfig config.FlightConfig, dealConfig config.DealConfig) *Manager {
	// Pass nil for Cronner to use the default cron instance
	scheduler := NewScheduler(queue, postgresDB, nil)

	topNDeals := flightConfig.TopNDeals
	if topNDeals <= 0 {
		topNDeals = 3
	}

	excludedSet := make(map[string]bool, len(flightConfig.ExcludedAirlines))
	for _, code := range flightConfig.ExcludedAirlines {
		code = strings.TrimSpace(strings.ToUpper(code))
		if code == "" {
			continue
		}
		excludedSet[code] = true
	}

	m := &Manager{
		queue:        queue,
		postgresDB:   postgresDB,
		neo4jDB:      neo4jDB,
		config:       workerConfig,
		flightConfig: flightConfig,
		dealConfig:   dealConfig,
		topNDeals:    topNDeals,
		excludedSet:  excludedSet,
		stopChan:     make(chan struct{}),
		flightCache:  make(map[string]*flights.Session),
		scheduler:    scheduler,
		workerStates: make([]*workerState, workerConfig.Concurrency),
		redisClient:  redisClient,
	}

	// Create leader elector if Redis client is provided
	if redisClient != nil {
		m.leaderElector = NewLeaderElector(
			redisClient,
			workerConfig.SchedulerLockKey,
			workerConfig.SchedulerLockTTL,
			workerConfig.SchedulerLockRenew,
			m.onBecomeLeader,
			m.onLoseLeader,
		)
	}

	return m
}

func (m *Manager) bulkSearchBusy() bool {
	if m == nil || m.queue == nil {
		return false
	}

	m.bulkBusyMu.Lock()
	defer m.bulkBusyMu.Unlock()

	// Cache for a short time to avoid hammering Redis from every worker loop.
	const cacheTTL = 2 * time.Second
	if !m.bulkBusyCheckedAt.IsZero() && time.Since(m.bulkBusyCheckedAt) < cacheTTL {
		return m.bulkBusyCached
	}

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	stats, err := m.queue.GetQueueStats(ctx, "bulk_search")
	cancel()
	if err != nil {
		// If stats can't be fetched, fall back to the last known value.
		m.bulkBusyCheckedAt = time.Now()
		return m.bulkBusyCached
	}

	pending := stats["pending"]
	processing := stats["processing"]
	m.bulkBusyCached = pending > 0 || processing > 0
	m.bulkBusyCheckedAt = time.Now()
	return m.bulkBusyCached
}

// onBecomeLeader is called when this instance becomes the scheduler leader.
func (m *Manager) onBecomeLeader() {
	log.Println("This instance became the scheduler leader - starting scheduler")
	if err := m.scheduler.Start(); err != nil {
		log.Printf("Failed to start scheduler after becoming leader: %v", err)
	}
}

// onLoseLeader is called when this instance loses scheduler leadership.
func (m *Manager) onLoseLeader() {
	log.Println("This instance lost scheduler leadership - stopping scheduler")
	m.scheduler.Stop()
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

// Start starts the worker pool and scheduler.
// If leader election is enabled, only the leader instance runs the scheduler.
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

	m.startRegistryHeartbeat()

	// Create and start workers (ALL instances run workers)
	for i := 0; i < m.config.Concurrency; i++ {
		worker := &Worker{
			postgresDB: m.postgresDB,
			neo4jDB:    m.neo4jDB,
		}
		m.workers = append(m.workers, worker)

		m.workerWg.Add(1)
		go m.runWorker(i, worker)
	}

	// Start leader election if enabled, otherwise start scheduler directly
	if m.leaderElector != nil {
		m.leaderElector.Start()
		log.Println("Leader election started - scheduler will run on leader instance only")
	} else {
		// Legacy behavior: start scheduler on every instance
		if err := m.scheduler.Start(); err != nil {
			log.Printf("Warning: Failed to start scheduler: %v", err)
		} else {
			log.Println("Scheduler started successfully (no leader election)")
		}
	}
}

func (m *Manager) GetQueue() queue.Queue {
	if m == nil {
		return nil
	}
	return m.queue
}

func (m *Manager) startRegistryHeartbeat() {
	if m == nil || m.redisClient == nil || m.config.WorkerID == "" {
		return
	}
	namespace := m.config.RegistryNamespace
	if namespace == "" {
		namespace = "flights"
	}

	reg := worker_registry.New(m.redisClient, namespace)
	hostname, _ := os.Hostname()
	startedAt := time.Now().UTC()
	interval := m.config.HeartbeatInterval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ttl := m.config.HeartbeatTTL
	if ttl <= 0 {
		ttl = 45 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_ = reg.Publish(ctx, m.buildRegistryHeartbeat(hostname, startedAt, time.Now().UTC(), "stopped"), ttl)
				cancel()
				return
			case <-ticker.C:
				now := time.Now().UTC()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				if err := reg.Publish(ctx, m.buildRegistryHeartbeat(hostname, startedAt, now, ""), ttl); err != nil {
					log.Printf("Failed to publish worker heartbeat: %v", err)
				}
				cancel()
			}
		}
	}()
}

func (m *Manager) buildRegistryHeartbeat(hostname string, startedAt, now time.Time, forceStatus string) worker_registry.WorkerHeartbeat {
	hb := worker_registry.WorkerHeartbeat{
		ID:            m.config.WorkerID,
		Hostname:      hostname,
		Status:        "active",
		CurrentJob:    "",
		ProcessedJobs: 0,
		Concurrency:   m.config.Concurrency,
		StartedAt:     startedAt,
		LastHeartbeat: now,
		Version:       "1.0.0",
	}

	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	status := "active"
	currentJob := ""
	processedTotal := 0

	for _, state := range m.workerStates {
		if state == nil {
			continue
		}
		processedTotal += state.ProcessedJobs
		if currentJob == "" && state.CurrentJob != "" {
			currentJob = state.CurrentJob
		}
		switch state.Status {
		case "error":
			status = "error"
		case "processing":
			if status != "error" {
				status = "processing"
			}
		}
	}

	if forceStatus != "" {
		status = forceStatus
	}

	hb.Status = status
	hb.CurrentJob = currentJob
	hb.ProcessedJobs = processedTotal

	return hb
}

// Stop stops the worker pool and scheduler.
// If leader election is enabled, it releases leadership first.
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

	// Stop leader election first (releases lock and stops scheduler if leader)
	if m.leaderElector != nil {
		m.leaderElector.Stop()
	} else {
		// Legacy behavior: stop scheduler directly
		m.scheduler.Stop()
	}

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

			// If any bulk search is pending/processing, avoid running background sweeps.
			// This prevents background jobs from competing for rate-limited Google endpoints,
			// which can stall user-initiated bulk searches.
			if m.bulkSearchBusy() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := m.processQueue(id, worker, "price_graph_sweep"); err != nil {
				log.Printf("Worker %d error processing price_graph_sweep queue: %v", displayID, err)
			}

			if err := m.processQueue(id, worker, "continuous_price_graph"); err != nil {
				log.Printf("Worker %d error processing continuous_price_graph queue: %v", displayID, err)
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

		// Cheap-first currently requires TripLength for round trips; fall back to the legacy
		// implementation when TripLength isn't provided (return-window mode).
		if payload.TripType == "round_trip" && payload.TripLength == 0 {
			log.Printf("[BulkSearch] TripLength=0 for round_trip; falling back to legacy bulk search")
			return m.processBulkSearch(ctx, worker, session, payload)
		}

		// Process the bulk search using 2-phase cheap-first strategy.
		// This reduces API calls from O(routes × dates) to O(routes + routes × topNDeals)
		return m.processBulkSearchCheapFirst(ctx, worker, session, payload)
	case "price_graph_sweep":
		session, err := m.getFlightSession("price_graph")
		if err != nil {
			return fmt.Errorf("failed to get flight session: %w", err)
		}

		var payload PriceGraphSweepPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal price graph sweep payload: %w", err)
		}

		return m.processPriceGraphSweep(ctx, worker, session, payload)
	case "continuous_price_graph":
		session, err := m.getFlightSession("price_graph")
		if err != nil {
			return fmt.Errorf("failed to get flight session: %w", err)
		}

		var payload ContinuousPriceGraphPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal continuous price graph payload: %w", err)
		}

		return m.processContinuousPriceGraph(ctx, worker, session, payload)

	default:
		return fmt.Errorf("unknown job type: %s", queueName)
	}
}

func (m *Manager) processContinuousPriceGraph(ctx context.Context, worker *Worker, session *flights.Session, payload ContinuousPriceGraphPayload) error {
	if payload.Origin == "" || payload.Destination == "" {
		return fmt.Errorf("continuous price graph payload requires origin and destination")
	}
	if payload.RangeStartDate.IsZero() || payload.RangeEndDate.IsZero() {
		return fmt.Errorf("continuous price graph payload requires a date range")
	}
	if payload.RangeEndDate.Before(payload.RangeStartDate) {
		return fmt.Errorf("continuous price graph payload has invalid date range")
	}

	// Continuous jobs can sit in the queue long enough that their date range becomes stale.
	// Shift the whole range forward to start at (or after) today's date in the range's timezone.
	payload.RangeStartDate, payload.RangeEndDate = shiftRangeToToday(payload.RangeStartDate, payload.RangeEndDate)
	if payload.RangeEndDate.Equal(payload.RangeStartDate) || payload.RangeEndDate.Before(payload.RangeStartDate) {
		return fmt.Errorf("continuous price graph payload has invalid date range after normalization")
	}

	args := flights.PriceGraphArgs{
		RangeStartDate: payload.RangeStartDate,
		RangeEndDate:   payload.RangeEndDate,
		TripLength:     payload.TripLength,
		SrcAirports:    []string{payload.Origin},
		DstAirports:    []string{payload.Destination},
	}

	cur, err := currency.ParseISO(strings.ToUpper(payload.Currency))
	if err != nil {
		cur = currency.USD
	}

	adults := payload.Adults
	if adults <= 0 {
		adults = 1
	}

	args.Options = flights.Options{
		Travelers: flights.Travelers{Adults: adults},
		Currency:  cur,
		Stops:     parseStops(payload.Stops),
		Class:     parseClass(payload.Class),
		TripType:  flights.RoundTrip,
		Lang:      language.English,
	}

	offers, _, err := session.GetPriceGraph(ctx, args)
	if err != nil {
		return fmt.Errorf("price graph query failed for %s->%s (trip length %d): %w", payload.Origin, payload.Destination, payload.TripLength, err)
	}

	if len(offers) == 0 {
		return nil
	}

	var cheapest *flights.Offer
	for i := range offers {
		price := offers[i].Price
		if price <= 0 {
			continue
		}
		if cheapest == nil || price < cheapest.Price {
			cheapest = &offers[i]
		}
	}
	if cheapest == nil {
		return nil
	}

	// Calculate distance and cost per mile
	distanceMiles, costPerMile := calculateDistanceAndCostPerMile(payload.Origin, payload.Destination, cheapest.Price)

	record := db.PriceGraphResultRecord{
		SweepID:       0,
		Origin:        payload.Origin,
		Destination:   payload.Destination,
		DepartureDate: cheapest.StartDate,
		ReturnDate:    sql.NullTime{Time: cheapest.ReturnDate, Valid: payload.TripLength > 0},
		TripLength:    sql.NullInt32{Int32: int32(payload.TripLength), Valid: true},
		Price:         cheapest.Price,
		Currency:      strings.ToUpper(payload.Currency),
		DistanceMiles: distanceMiles,
		CostPerMile:   costPerMile,
		Adults:        adults,
		Children:      0,
		InfantsLap:    0,
		InfantsSeat:   0,
		TripType:      "round_trip",
		Class:         defaultString(payload.Class, "economy"),
		Stops:         defaultString(payload.Stops, "any"),
		QueriedAt:     time.Now(),
	}

	searchURL, urlErr := session.SerializeURL(ctx, flights.Args{
		Date:        cheapest.StartDate,
		ReturnDate:  cheapest.ReturnDate,
		SrcAirports: []string{payload.Origin},
		DstAirports: []string{payload.Destination},
		Options:     args.Options,
	})
	if urlErr == nil && searchURL != "" {
		record.SearchURL = sql.NullString{String: searchURL, Valid: true}
	} else if urlErr != nil {
		log.Printf("Failed to serialize Google Flights URL for %s->%s: %v", payload.Origin, payload.Destination, urlErr)
	}

	if err := worker.postgresDB.InsertPriceGraphResult(ctx, record); err != nil {
		return err
	}

	// Sync price point to Neo4j for graph analytics (idempotent via MERGE)
	if m.neo4jDB != nil {
		dateStr := cheapest.StartDate.Format("2006-01-02")
		if syncErr := m.neo4jDB.AddPricePoint(
			payload.Origin,
			payload.Destination,
			dateStr,
			cheapest.Price,
			"", // No specific airline for price graph results
		); syncErr != nil {
			// Log but don't fail - Postgres is the source of truth
			log.Printf("Warning: failed to sync price point to Neo4j for %s->%s: %v", payload.Origin, payload.Destination, syncErr)
		}
	}

	// Detect deals from the price result
	detector := deals.NewDealDetector(m.postgresDB, m.dealConfig)
	deal, detectErr := detector.DetectDeal(ctx, record)
	if detectErr != nil {
		log.Printf("Warning: deal detection failed for %s->%s: %v", payload.Origin, payload.Destination, detectErr)
	} else if deal != nil {
		// Store or upsert the detected deal
		if upsertErr := m.postgresDB.UpsertDetectedDeal(ctx, *deal); upsertErr != nil {
			log.Printf("Warning: failed to upsert detected deal for %s->%s: %v", payload.Origin, payload.Destination, upsertErr)
		} else {
			log.Printf("Detected %s deal for %s->%s: $%.2f (%.0f%% off, score %d)",
				deal.DealClassification.String, payload.Origin, payload.Destination,
				deal.Price, deal.DiscountPercent.Float64, deal.DealScore.Int32)
		}
	}

	return nil
}

func shiftRangeToToday(rangeStartDate, rangeEndDate time.Time) (time.Time, time.Time) {
	startDay := truncateToDay(rangeStartDate)
	nowDay := truncateToDay(time.Now().In(startDay.Location()))

	if !startDay.Before(nowDay) {
		return rangeStartDate, rangeEndDate
	}

	shiftDays := 0
	for d := startDay; d.Before(nowDay); d = d.AddDate(0, 0, 1) {
		shiftDays++
	}

	return rangeStartDate.AddDate(0, 0, shiftDays), rangeEndDate.AddDate(0, 0, shiftDays)
}

func truncateToDay(d time.Time) time.Time {
	loc := d.Location()
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
}

// getFlightSession gets or creates a flight session based on session type
func (m *Manager) getFlightSession(sessionType string) (*flights.Session, error) {
	var sessionKey string

	// Use consistent caching strategy for both search types since regular search works fine
	switch sessionType {
	case "bulk_search":
		// Use cached session like direct search for better reliability
		sessionKey = "bulk_search"
	case "price_graph":
		sessionKey = "price_graph"
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
			routeResults[routeKey] = routeResult

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
							Carriers: payload.Carriers,
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

					if bulkSearchID > 0 {
						for _, offer := range offers {
							distanceMiles, costPerMile := calculateDistanceAndCostPerMile(origin, destination, offer.Price)
							offerRecord := db.BulkSearchOfferRecord{
								BulkSearchID:         bulkSearchID,
								Origin:               origin,
								Destination:          destination,
								DepartureDate:        offer.StartDate,
								ReturnDate:           timeToNullTime(offer.ReturnDate),
								Price:                offer.Price,
								Currency:             strings.ToUpper(payload.Currency),
								AirlineCodes:         airlineCodesFromOffer(offer),
								SrcAirportCode:       nullString(offer.SrcAirportCode),
								DstAirportCode:       nullString(offer.DstAirportCode),
								SrcCity:              nullString(offer.SrcCity),
								DstCity:              nullString(offer.DstCity),
								FlightDuration:       durationToNullMinutes(offer.FlightDuration),
								ReturnFlightDuration: durationToNullMinutes(offer.ReturnFlightDuration),
								DistanceMiles:        distanceMiles,
								CostPerMile:          costPerMile,
								OutboundFlightsJSON:  flightsToJSON(offer.Flight),
								ReturnFlightsJSON:    flightsToJSON(offer.ReturnFlight),
								OfferJSON:            offerToJSON(offer),
							}

							if insertErr := m.postgresDB.InsertBulkSearchOffer(ctx, offerRecord); insertErr != nil {
								log.Printf("Failed to insert bulk offer for %s -> %s: %v", origin, destination, insertErr)
							}
						}
					}
				}
			}
			if routeResult.BestOffer.Price < math.MaxFloat64 {
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
		// Skip routes with no valid offers (Price == math.MaxFloat64 is the sentinel value)
		if result.BestOffer.Price >= math.MaxFloat64 {
			log.Printf("Skipping storage for route %s: no valid offers found", routeKey)
			continue
		}

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

// isExcludedAirline checks if any flight in the offer uses an excluded airline.
func (m *Manager) isExcludedAirline(offer flights.FullOffer) bool {
	if len(m.excludedSet) == 0 {
		return false
	}
	// Check outbound flights
	for _, flight := range offer.Flight {
		if len(flight.FlightNumber) >= 2 {
			airlineCode := flight.FlightNumber[:2]
			if m.excludedSet[airlineCode] {
				return true
			}
		}
	}
	// Check return flights
	for _, flight := range offer.ReturnFlight {
		if len(flight.FlightNumber) >= 2 {
			airlineCode := flight.FlightNumber[:2]
			if m.excludedSet[airlineCode] {
				return true
			}
		}
	}
	return false
}

// scoreDeal calculates a composite deal score where lower is better.
// It considers price, duration overhead, stops, and departure time.
func scoreDeal(offer flights.FullOffer, distanceMiles float64) float64 {
	score := offer.Price

	// Penalize duration: each hour over baseline (500mph) adds $10 equivalent
	if distanceMiles > 0 {
		baselineHours := distanceMiles / 500.0
		if offer.ReturnFlightDuration > 0 || len(offer.ReturnFlight) > 0 {
			baselineHours *= 2
		}
		actualHours := offer.FlightDuration.Hours() + offer.ReturnFlightDuration.Hours()
		extraHours := actualHours - baselineHours
		if extraHours > 0 {
			score += extraHours * 10
		}
	}

	// Penalize stops: $30 per stop on outbound, $20 per stop on return
	outboundStops := 0
	if len(offer.Flight) > 1 {
		outboundStops = len(offer.Flight) - 1
	}
	returnStops := 0
	if len(offer.ReturnFlight) > 1 {
		returnStops = len(offer.ReturnFlight) - 1
	}
	score += float64(outboundStops)*30 + float64(returnStops)*20

	// Penalize red-eye departures (10pm-6am)
	if len(offer.Flight) > 0 {
		hour := offer.Flight[0].DepTime.Hour()
		if hour >= 22 || hour < 6 {
			score += 20
		}
	}

	return score
}

// processBulkSearchCheapFirst implements a 2-phase "cheap-first" bulk search strategy:
// Phase 1: Use GetPriceGraph to find prices for ALL dates in one API call per route
// Phase 2: Call GetOffers only for the top N cheapest dates to get full itineraries
//
// This reduces API calls from O(routes × dates) to O(routes + routes × N)
// Example: 100 routes × 30 dates → 3000 calls becomes 100 + 300 = 400 calls
func (m *Manager) processBulkSearchCheapFirst(ctx context.Context, worker *Worker, session *flights.Session, payload BulkSearchPayload) (err error) {
	topN := m.topNDeals

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
		log.Printf("Warning: Invalid currency '%s', defaulting to USD. Error: %v", payload.Currency, err)
		cur = currency.USD
	}

	// Parse class and stops
	flightClass := parseClass(payload.Class)
	flightStops := parseStops(payload.Stops)

	totalRoutes := len(payload.Origins) * len(payload.Destinations)
	log.Printf("[CheapFirst] Starting 2-phase bulk search: %d origins × %d destinations = %d routes",
		len(payload.Origins), len(payload.Destinations), totalRoutes)

	// Create or get bulk search record
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

	var searchErrors []error

	// Track stats
	var (
		phase1Calls int
		phase2Calls int
		totalOffers int
		minPrice    float64 = math.MaxFloat64
		maxPrice    float64
		sumPrice    float64
		processed   int
		withResults int
	)

	for _, origin := range payload.Origins {
		for _, destination := range payload.Destinations {
			routeKey := fmt.Sprintf("%s-%s", origin, destination)
			log.Printf("[CheapFirst] Phase 1: Getting price graph for %s", routeKey)

			// Calculate distance for deal scoring
			distanceMiles, _ := calculateDistanceAndCostPerMile(origin, destination, 1.0)
			dist := 0.0
			if distanceMiles.Valid {
				dist = distanceMiles.Float64
			}

			// PHASE 1: Get prices for all dates in ONE call
			priceGraphArgs := flights.PriceGraphArgs{
				RangeStartDate: payload.DepartureDateFrom,
				RangeEndDate:   payload.DepartureDateTo,
				TripLength:     payload.TripLength,
				SrcAirports:    []string{origin},
				DstAirports:    []string{destination},
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
					Carriers: payload.Carriers,
				},
			}

			priceOffers, _, err := session.GetPriceGraph(ctx, priceGraphArgs)
			phase1Calls++

			if err != nil {
				log.Printf("[CheapFirst] Error getting price graph for %s: %v", routeKey, err)
				searchErrors = append(searchErrors, fmt.Errorf("price graph for %s failed: %w", routeKey, err))
				// Count the route as processed for progress reporting even if it failed.
				processed++
				if processed%5 == 0 {
					_ = m.postgresDB.UpdateBulkSearchProgress(ctx, bulkSearchID, processed, totalOffers, len(searchErrors))
				}
				continue
			}

			if len(priceOffers) == 0 {
				log.Printf("[CheapFirst] No prices found for %s", routeKey)
				processed++
				if processed%5 == 0 {
					_ = m.postgresDB.UpdateBulkSearchProgress(ctx, bulkSearchID, processed, totalOffers, len(searchErrors))
				}
				continue
			}

			// Price=0 means price unavailable; also skip values that cannot be stored in DB.
			filtered := priceOffers[:0]
			for _, offer := range priceOffers {
				if !isDBSafePrice(offer.Price) {
					continue
				}
				filtered = append(filtered, offer)
			}
			priceOffers = filtered
			if len(priceOffers) == 0 {
				log.Printf("[CheapFirst] No priced offers found for %s (all offers were unpriced or invalid)", routeKey)
				processed++
				if processed%5 == 0 {
					_ = m.postgresDB.UpdateBulkSearchProgress(ctx, bulkSearchID, processed, totalOffers, len(searchErrors))
				}
				continue
			}

			// Sort offers by price to find top N cheapest dates
			sort.Slice(priceOffers, func(i, j int) bool {
				return priceOffers[i].Price < priceOffers[j].Price
			})

			// Take top N (or fewer if not enough offers)
			topOffers := priceOffers
			if len(topOffers) > topN {
				topOffers = topOffers[:topN]
			}

			log.Printf("[CheapFirst] Phase 2: Getting full itineraries for top %d dates for %s", len(topOffers), routeKey)

			// PHASE 2: Get full itineraries only for cheapest dates
			var bestOffer *flights.FullOffer
			var bestScore float64 = math.MaxFloat64
			var bestDate, bestReturn time.Time

			for _, priceOffer := range topOffers {
				args := flights.Args{
					Date:        priceOffer.StartDate,
					ReturnDate:  priceOffer.ReturnDate,
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
						Carriers: payload.Carriers,
					},
				}

				fullOffers, _, err := session.GetOffers(ctx, args)
				phase2Calls++

				if err != nil {
					log.Printf("[CheapFirst] Error getting offers for %s on %s: %v",
						routeKey, priceOffer.StartDate.Format("2006-01-02"), err)
					continue
				}

				totalOffers += len(fullOffers)

				// Find best offer by composite deal score (skip excluded airlines)
				for i := range fullOffers {
					// Skip offers with non-storable prices (prevents DB overflows and bogus min/max).
					if !isDBSafePrice(fullOffers[i].Price) {
						continue
					}
					// Skip offers from excluded airlines (Spirit, Allegiant, Frontier)
					if m.isExcludedAirline(fullOffers[i]) {
						continue
					}
					score := scoreDeal(fullOffers[i], dist)
					if score < bestScore {
						bestScore = score
						bestOffer = &fullOffers[i]
						bestDate = priceOffer.StartDate
						bestReturn = priceOffer.ReturnDate
					}
				}
			}

			if bestOffer != nil {
				log.Printf("[CheapFirst] Route %s: best deal $%.2f (score: %.2f) on %s",
					routeKey, bestOffer.Price, bestScore, bestDate.Format("2006-01-02"))

				// Update price stats
				if bestOffer.Price < minPrice {
					minPrice = bestOffer.Price
				}
				if bestOffer.Price > maxPrice {
					maxPrice = bestOffer.Price
				}
				sumPrice += bestOffer.Price
				withResults++

				returnDate := sql.NullTime{}
				if !bestReturn.IsZero() {
					returnDate = sql.NullTime{Time: bestReturn, Valid: true}
				}

				airlineCode := ""
				if len(bestOffer.Flight) > 0 {
					flightNumber := bestOffer.Flight[0].FlightNumber
					if len(flightNumber) >= 2 {
						airlineCode = flightNumber[:2]
					}
				}

				flightDuration := int(bestOffer.FlightDuration.Minutes())
				returnFlightDuration := int(bestOffer.ReturnFlightDuration.Minutes())
				totalDuration := flightDuration + returnFlightDuration

				record := db.BulkSearchResultRecord{
					BulkSearchID:         bulkSearchID,
					Origin:               origin,
					Destination:          destination,
					DepartureDate:        bestDate,
					ReturnDate:           returnDate,
					Price:                bestOffer.Price,
					Currency:             strings.ToUpper(payload.Currency),
					AirlineCode:          nullString(airlineCode),
					Duration:             nullInt32(totalDuration),
					SrcAirportCode:       nullString(bestOffer.SrcAirportCode),
					DstAirportCode:       nullString(bestOffer.DstAirportCode),
					SrcCity:              nullString(bestOffer.SrcCity),
					DstCity:              nullString(bestOffer.DstCity),
					FlightDuration:       durationToNullMinutes(bestOffer.FlightDuration),
					ReturnFlightDuration: durationToNullMinutes(bestOffer.ReturnFlightDuration),
					OutboundFlightsJSON:  flightsToJSON(bestOffer.Flight),
					ReturnFlightsJSON:    flightsToJSON(bestOffer.ReturnFlight),
					OfferJSON:            offerToJSON(*bestOffer),
				}

				if insertErr := m.postgresDB.InsertBulkSearchResult(ctx, record); insertErr != nil {
					log.Printf("[CheapFirst] Failed to insert result for %s -> %s: %v", origin, destination, insertErr)
				}
			} else {
				log.Printf("[CheapFirst] Route %s: no valid offers found", routeKey)
			}

			processed++
			if processed%5 == 0 || processed == totalRoutes {
				_ = m.postgresDB.UpdateBulkSearchProgress(ctx, bulkSearchID, processed, totalOffers, len(searchErrors))
			}
		}
	}

	log.Printf("[CheapFirst] API call summary: Phase1=%d, Phase2=%d, Total=%d (vs old method: %d)",
		phase1Calls, phase2Calls, phase1Calls+phase2Calls,
		totalRoutes*len(m.generateDateRange(payload.DepartureDateFrom, payload.DepartureDateTo, payload.TripLength)))

	// Finalize bulk search
	minPriceNull, maxPriceNull, avgPriceNull := cheapFirstFinalizePriceStats(withResults, minPrice, maxPrice, sumPrice)
	status := cheapFirstFinalizeStatus(withResults, len(searchErrors))

	log.Printf("[CheapFirst] Bulk search completed: processed=%d, with_results=%d, offers=%d, status=%s",
		processed, withResults, totalOffers, status)

	summary := db.BulkSearchSummary{
		ID:           bulkSearchID,
		Status:       status,
		Completed:    processed,
		TotalOffers:  totalOffers,
		ErrorCount:   len(searchErrors),
		MinPrice:     minPriceNull,
		MaxPrice:     maxPriceNull,
		AveragePrice: avgPriceNull,
	}

	if completeErr := m.postgresDB.CompleteBulkSearch(ctx, summary); completeErr != nil {
		log.Printf("[CheapFirst] Failed to update bulk search summary for %d: %v", bulkSearchID, completeErr)
	}

	return nil
}

// processPriceGraphSweep executes a price graph sweep job and stores the cheapest fares for each date
func (m *Manager) processPriceGraphSweep(ctx context.Context, worker *Worker, session *flights.Session, payload PriceGraphSweepPayload) (err error) {
	if len(payload.Origins) == 0 {
		return fmt.Errorf("price graph sweep requires at least one origin")
	}
	if len(payload.Destinations) == 0 {
		return fmt.Errorf("price graph sweep requires at least one destination")
	}

	tripLengths := payload.TripLengths
	if len(tripLengths) == 0 {
		tripLengths = []int{0}
	}

	classes := payload.Classes
	if len(classes) == 0 {
		if payload.Class != "" {
			classes = []string{payload.Class}
		} else {
			classes = []string{"economy"}
			payload.Class = "economy"
		}
	}
	seenClass := make(map[string]struct{}, len(classes))
	normalizedClasses := make([]string, 0, len(classes))
	for _, c := range classes {
		cabin := strings.ToLower(strings.TrimSpace(c))
		if cabin == "" {
			continue
		}
		switch cabin {
		case "economy", "premium_economy", "business", "first":
		default:
			return fmt.Errorf("invalid class %q (allowed: economy, premium_economy, business, first)", c)
		}
		if _, ok := seenClass[cabin]; ok {
			continue
		}
		seenClass[cabin] = struct{}{}
		normalizedClasses = append(normalizedClasses, cabin)
	}
	if len(normalizedClasses) == 0 {
		return fmt.Errorf("price graph sweep requires at least one class")
	}

	var tripType flights.TripType
	switch payload.TripType {
	case "one_way":
		tripType = flights.OneWay
	default:
		tripType = flights.RoundTrip
	}

	flightStops := parseStops(payload.Stops)

	cur, err := currency.ParseISO(payload.Currency)
	if err != nil {
		log.Printf("Warning: invalid currency '%s' for price graph sweep, defaulting to USD: %v", payload.Currency, err)
		cur = currency.USD
		payload.Currency = "USD"
	}

	var jobRef sql.NullInt32
	if payload.JobID > 0 {
		jobRef = sql.NullInt32{Int32: int32(payload.JobID), Valid: true}
	}

	var minLen, maxLen sql.NullInt32
	if len(tripLengths) > 0 {
		minVal := tripLengths[0]
		maxVal := tripLengths[0]
		for _, l := range tripLengths {
			if l < minVal {
				minVal = l
			}
			if l > maxVal {
				maxVal = l
			}
		}
		minLen = sql.NullInt32{Int32: int32(minVal), Valid: true}
		maxLen = sql.NullInt32{Int32: int32(maxVal), Valid: true}
	}

	sweepID := payload.SweepID
	if sweepID == 0 {
		newID, createErr := m.postgresDB.CreatePriceGraphSweep(ctx, jobRef, len(payload.Origins), len(payload.Destinations), minLen, maxLen, strings.ToUpper(payload.Currency))
		if createErr != nil {
			return createErr
		}
		sweepID = newID
	}

	resultsInserted := 0
	errorCount := 0

	startedAt := sql.NullTime{Time: time.Now(), Valid: true}
	if updateErr := m.postgresDB.UpdatePriceGraphSweepStatus(ctx, sweepID, "running", startedAt, sql.NullTime{}, 0); updateErr != nil {
		log.Printf("Failed to mark price graph sweep %d as running: %v", sweepID, updateErr)
	}

	defer func() {
		if err != nil {
			if updateErr := m.postgresDB.UpdatePriceGraphSweepStatus(ctx, sweepID, "failed", sql.NullTime{}, sql.NullTime{}, errorCount); updateErr != nil {
				log.Printf("Failed to mark price graph sweep %d as failed: %v", sweepID, updateErr)
			}
		}
	}()

	rateDelay := time.Duration(payload.RateLimitMillis) * time.Millisecond
	if rateDelay <= 0 {
		rateDelay = 750 * time.Millisecond
	}

	options := flights.Options{
		Travelers: flights.Travelers{
			Adults:       payload.Adults,
			Children:     payload.Children,
			InfantOnLap:  payload.InfantsLap,
			InfantInSeat: payload.InfantsSeat,
		},
		Currency: cur,
		Stops:    flightStops,
		TripType: tripType,
		Lang:     language.English,
	}

	for _, origin := range payload.Origins {
		for _, destination := range payload.Destinations {
			for _, length := range tripLengths {
				for _, class := range normalizedClasses {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					options.Class = parseClass(class)
					args := flights.PriceGraphArgs{
						RangeStartDate: payload.DepartureDateFrom,
						RangeEndDate:   payload.DepartureDateTo,
						TripLength:     length,
						SrcAirports:    []string{origin},
						DstAirports:    []string{destination},
						Options:        options,
					}

					offers, _, sweepErr := session.GetPriceGraph(ctx, args)
					if sweepErr != nil {
						errorCount++
						log.Printf("Price graph sweep error for %s -> %s (class %s, length %d): %v", origin, destination, class, length, sweepErr)
						continue
					}

					for _, offer := range offers {
						// Price=0 means price unavailable; skip these rows.
						if offer.Price <= 0 {
							continue
						}
						distanceMiles, costPerMile := calculateDistanceAndCostPerMile(origin, destination, offer.Price)
						record := db.PriceGraphResultRecord{
							SweepID:       sweepID,
							Origin:        origin,
							Destination:   destination,
							DepartureDate: offer.StartDate,
							ReturnDate:    timeToNullTime(offer.ReturnDate),
							TripLength:    nullInt32(length),
							Price:         offer.Price,
							Currency:      strings.ToUpper(payload.Currency),
							DistanceMiles: distanceMiles,
							CostPerMile:   costPerMile,
							Adults:        payload.Adults,
							Children:      payload.Children,
							InfantsLap:    payload.InfantsLap,
							InfantsSeat:   payload.InfantsSeat,
							TripType:      payload.TripType,
							Class:         class,
							Stops:         payload.Stops,
							QueriedAt:     time.Now(),
						}

						searchURL, urlErr := session.SerializeURL(ctx, flights.Args{
							Date:        offer.StartDate,
							ReturnDate:  offer.ReturnDate,
							SrcAirports: []string{origin},
							DstAirports: []string{destination},
							Options:     options,
						})
						if urlErr == nil && searchURL != "" {
							record.SearchURL = sql.NullString{String: searchURL, Valid: true}
						} else if urlErr != nil {
							log.Printf("Failed to serialize Google Flights URL for %s->%s: %v", origin, destination, urlErr)
						}

						if insertErr := m.postgresDB.InsertPriceGraphResult(ctx, record); insertErr != nil {
							errorCount++
							log.Printf("Failed to store price graph result for %s -> %s on %s: %v", origin, destination, offer.StartDate.Format("2006-01-02"), insertErr)
							continue
						}
						resultsInserted++

						// Sync price point to Neo4j for graph analytics (idempotent via MERGE)
						if m.neo4jDB != nil {
							dateStr := offer.StartDate.Format("2006-01-02")
							if syncErr := m.neo4jDB.AddPricePoint(origin, destination, dateStr, offer.Price, ""); syncErr != nil {
								log.Printf("Warning: failed to sync price point to Neo4j for %s->%s: %v", origin, destination, syncErr)
							}
						}
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(rateDelay):
					}
				}
			}
		}
	}

	status := "completed"
	if resultsInserted == 0 && errorCount > 0 {
		status = "failed"
	} else if errorCount > 0 {
		status = "completed_with_errors"
	}

	completedAt := sql.NullTime{Time: time.Now(), Valid: true}
	if updateErr := m.postgresDB.UpdatePriceGraphSweepStatus(ctx, sweepID, status, sql.NullTime{}, completedAt, errorCount); updateErr != nil {
		log.Printf("Failed to finalize price graph sweep %d: %v", sweepID, updateErr)
	}

	log.Printf("Price graph sweep %d finished with %d results (%d errors)", sweepID, resultsInserted, errorCount)
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

// GetSweepRunner returns the continuous sweep runner instance
func (m *Manager) GetSweepRunner() *ContinuousSweepRunner {
	m.sweepMutex.RLock()
	defer m.sweepMutex.RUnlock()
	return m.sweepRunner
}

// SetSweepRunner sets the continuous sweep runner instance
func (m *Manager) SetSweepRunner(runner *ContinuousSweepRunner) {
	m.sweepMutex.Lock()
	defer m.sweepMutex.Unlock()
	m.sweepRunner = runner
}
