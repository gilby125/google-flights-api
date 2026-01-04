package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/pkg/notify"
	"github.com/gilby125/google-flights-api/queue"
)

// PacingMode defines how delays between queries are calculated
type PacingMode string

const (
	PacingModeAdaptive PacingMode = "adaptive"
	PacingModeFixed    PacingMode = "fixed"
)

// ContinuousSweepConfig holds configuration for continuous sweeps
type ContinuousSweepConfig struct {
	TripLengths         []int
	DepartureWindowDays int
	Class               string
	Stops               string
	Adults              int
	Currency            string
	PacingMode          PacingMode
	TargetDurationHours int
	MinDelayMs          int
	InternationalOnly   bool
}

// DefaultContinuousSweepConfig returns the default configuration
func DefaultContinuousSweepConfig() ContinuousSweepConfig {
	return ContinuousSweepConfig{
		TripLengths:         []int{7, 14},
		DepartureWindowDays: 30,
		Class:               "economy",
		Stops:               "any",
		Adults:              1,
		Currency:            "USD",
		PacingMode:          PacingModeAdaptive,
		TargetDurationHours: 24,
		MinDelayMs:          3000,
		InternationalOnly:   true,
	}
}

// ContinuousSweepRunner manages continuous price graph sweeps
type ContinuousSweepRunner struct {
	mu sync.RWMutex

	// Dependencies
	postgresDB db.PostgresDB
	queue      queue.Queue
	notifier   *notify.NTFYClient

	// Configuration
	config ContinuousSweepConfig
	routes []db.Route

	// Long-lived context (not tied to HTTP request)
	ctx       context.Context
	cancelCtx context.CancelFunc

	// State
	isRunning   bool
	isPaused    bool
	sweepNumber int
	routeIndex  int
	startTime   time.Time

	// Metrics
	queriesCompleted int
	errorsCount      int
	lastError        string
	lastErrorTime    time.Time

	// Stats tracking for sweep completion
	minPriceFound float64
	maxPriceFound float64
	totalDelayMs  int64

	// Error tracking for spike detection
	recentErrors []time.Time

	// Auto-resume tracking (e.g. pause for on-demand sweeps, then resume).
	autoResumeMu     sync.Mutex
	autoResumeActive bool

	// Control channels
	stopCh   chan struct{}
	resumeCh chan struct{}
}

type ContinuousPriceGraphPayload struct {
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	RangeStartDate time.Time `json:"range_start_date"`
	RangeEndDate   time.Time `json:"range_end_date"`
	TripLength     int       `json:"trip_length"`
	Class          string    `json:"class"`
	Stops          string    `json:"stops"`
	Adults         int       `json:"adults"`
	Currency       string    `json:"currency"`
}

// NewContinuousSweepRunner creates a new continuous sweep runner
func NewContinuousSweepRunner(
	postgresDB db.PostgresDB,
	queue queue.Queue,
	notifier *notify.NTFYClient,
	config ContinuousSweepConfig,
) *ContinuousSweepRunner {
	// Generate routes based on configuration
	var routes []db.Route
	if config.InternationalOnly {
		routes = db.GenerateInternationalRoutes(db.Top100Airports)
	} else {
		routes = db.GenerateAllRoutes(db.Top100Airports)
	}

	return &ContinuousSweepRunner{
		postgresDB:   postgresDB,
		queue:        queue,
		notifier:     notifier,
		config:       config,
		routes:       routes,
		recentErrors: make([]time.Time, 0),
	}
}

// Start begins the continuous sweep process.
// Note: This method creates its own long-lived context independent of HTTP requests.
func (r *ContinuousSweepRunner) Start() error {
	r.mu.Lock()
	if r.isRunning {
		r.mu.Unlock()
		return fmt.Errorf("continuous sweep is already running")
	}

	// Create long-lived context (not tied to HTTP request)
	r.ctx, r.cancelCtx = context.WithCancel(context.Background())

	// Reinitialize channels for restartability
	r.stopCh = make(chan struct{})
	r.resumeCh = make(chan struct{})

	r.isRunning = true
	r.mu.Unlock()

	// Try to restore progress from DB (use the long-lived context)
	if err := r.restoreProgress(r.ctx); err != nil {
		log.Printf("Could not restore sweep progress: %v (starting fresh)", err)
		r.sweepNumber = 1
		r.routeIndex = 0
	}

	// Update total routes in progress
	r.saveProgress(r.ctx)

	go r.run(r.ctx)
	return nil
}

// Stop gracefully stops the continuous sweep
func (r *ContinuousSweepRunner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRunning {
		return
	}

	// Cancel context and close stop channel
	if r.cancelCtx != nil {
		r.cancelCtx()
	}
	if r.stopCh != nil {
		close(r.stopCh)
		r.stopCh = nil // Prevent double-close
	}
	r.isRunning = false
}

// Pause pauses the sweep (can be resumed)
func (r *ContinuousSweepRunner) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isRunning && !r.isPaused {
		r.isPaused = true
	}
}

// Resume resumes a paused sweep
func (r *ContinuousSweepRunner) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isRunning && r.isPaused {
		r.isPaused = false
		select {
		case r.resumeCh <- struct{}{}:
		default:
		}
	}
}

func (r *ContinuousSweepRunner) PauseAndAutoResumeAfterQueueDrain(queueName string) {
	r.mu.RLock()
	running := r.isRunning
	paused := r.isPaused
	r.mu.RUnlock()

	if !running || paused {
		return
	}

	r.Pause()

	r.autoResumeMu.Lock()
	if r.autoResumeActive {
		r.autoResumeMu.Unlock()
		return
	}
	r.autoResumeActive = true
	r.autoResumeMu.Unlock()

	go func() {
		defer func() {
			r.autoResumeMu.Lock()
			r.autoResumeActive = false
			r.autoResumeMu.Unlock()
		}()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			r.mu.RLock()
			running := r.isRunning
			paused := r.isPaused
			r.mu.RUnlock()

			if !running {
				return
			}
			if !paused {
				// User manually resumed; don't fight it.
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			stats, err := r.queue.GetQueueStats(ctx, queueName)
			cancel()
			if err == nil {
				pending := stats["pending"]
				processing := stats["processing"]
				if pending == 0 && processing == 0 {
					r.Resume()
					return
				}
			}

			select {
			case <-ticker.C:
			case <-r.stopCh:
				return
			case <-r.ctx.Done():
				return
			}
		}
	}()
}

// SetConfig updates the configuration (safe to call while running)
func (r *ContinuousSweepRunner) SetConfig(config ContinuousSweepConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = config
}

// GetConfig returns a copy of the current configuration.
func (r *ContinuousSweepRunner) GetConfig() ContinuousSweepConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// GetStatus returns the current status
func (r *ContinuousSweepRunner) GetStatus() db.SweepStatusResponse {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := db.SweepStatusResponse{
		IsRunning:           r.isRunning,
		IsPaused:            r.isPaused,
		SweepNumber:         r.sweepNumber,
		RouteIndex:          r.routeIndex,
		TotalRoutes:         len(r.routes) * len(r.config.TripLengths),
		QueriesCompleted:    r.queriesCompleted,
		ErrorsCount:         r.errorsCount,
		LastError:           r.lastError,
		PacingMode:          string(r.config.PacingMode),
		TargetDurationHours: r.config.TargetDurationHours,
		Class:               r.config.Class,
		Stops:               r.config.Stops,
		TripLengths:         append([]int(nil), r.config.TripLengths...),
	}

	if status.TotalRoutes > 0 {
		status.ProgressPercent = float64(r.routeIndex*len(r.config.TripLengths)) / float64(status.TotalRoutes) * 100
	}

	if len(r.routes) > 0 && r.routeIndex < len(r.routes) {
		status.CurrentOrigin = r.routes[r.routeIndex].Origin
		status.CurrentDestination = r.routes[r.routeIndex].Destination
	}

	if !r.startTime.IsZero() {
		status.SweepStartedAt = r.startTime

		// Calculate estimated completion
		if r.queriesCompleted > 0 {
			elapsed := time.Since(r.startTime)
			remaining := status.TotalRoutes - (r.routeIndex * len(r.config.TripLengths))
			avgPerQuery := elapsed / time.Duration(r.queriesCompleted)
			status.EstimatedCompletion = time.Now().Add(avgPerQuery * time.Duration(remaining))
			status.QueriesPerHour = float64(r.queriesCompleted) / elapsed.Hours()
		}
	}

	status.CurrentDelayMs = r.calculateDelay()

	return status
}

// SkipRoute skips the current route and moves to the next one
func (r *ContinuousSweepRunner) SkipRoute() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.routeIndex < len(r.routes)-1 {
		r.routeIndex++
	}
}

// RestartSweep restarts the sweep from the beginning
func (r *ContinuousSweepRunner) RestartSweep() {
	r.mu.Lock()
	r.routeIndex = 0
	r.sweepNumber++
	r.queriesCompleted = 0
	r.errorsCount = 0
	r.recentErrors = make([]time.Time, 0)
	r.lastError = ""
	r.lastErrorTime = time.Time{}
	r.totalDelayMs = 0
	r.startTime = time.Now()
	r.mu.Unlock()

	// Immediately save progress
	if r.ctx != nil {
		r.saveProgress(r.ctx)
	}
}

func (r *ContinuousSweepRunner) run(ctx context.Context) {
	defer func() {
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
		r.saveProgress(ctx)
	}()

	log.Printf("Starting continuous sweep runner with %d routes, pacing mode: %s",
		len(r.routes), r.config.PacingMode)

	// Notify sweep start
	if r.notifier != nil && r.notifier.IsEnabled() {
		estimatedDuration := time.Duration(len(r.routes)*len(r.config.TripLengths)) * time.Duration(r.calculateDelay()) * time.Millisecond
		r.notifier.AlertSweepStarted(r.sweepNumber, len(r.routes)*len(r.config.TripLengths), estimatedDuration)
	}

	r.startTime = time.Now()
	lastProgressSave := time.Now()
	lastActivityTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			log.Println("Continuous sweep stopped: context cancelled")
			return
		case <-r.stopCh:
			log.Println("Continuous sweep stopped: stop requested")
			return
		default:
		}

		// Check if paused
		r.mu.RLock()
		paused := r.isPaused
		r.mu.RUnlock()

		if paused {
			select {
			case <-r.resumeCh:
				log.Println("Continuous sweep resumed")
			case <-r.stopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				continue
			}
		}

		// Check for stall
		r.mu.RLock()
		routeIdx := r.routeIndex
		sweepNum := r.sweepNumber
		r.mu.RUnlock()

		if r.notifier != nil && r.notifier.IsEnabled() {
			stallThreshold := r.notifier.GetConfig().StallThreshold
			if time.Since(lastActivityTime) > stallThreshold {
				route := ""
				if routeIdx < len(r.routes) {
					route = fmt.Sprintf("%s->%s", r.routes[routeIdx].Origin, r.routes[routeIdx].Destination)
				}
				r.notifier.AlertStall(sweepNum, route, time.Since(lastActivityTime))
			}
		}

		// Process current route
		if routeIdx >= len(r.routes) {
			// Sweep complete, start next
			r.completeSweep(ctx)
			continue
		}

		route := r.routes[routeIdx]
		err := r.processRoute(ctx, route)

		lastActivityTime = time.Now()

		if err != nil {
			r.recordError(ctx, err)
		}

		// Move to next route
		r.mu.Lock()
		r.routeIndex++
		r.mu.Unlock()

		// Save progress periodically (every 5 minutes or every 100 queries)
		if time.Since(lastProgressSave) > 5*time.Minute || r.queriesCompleted%100 == 0 {
			r.saveProgress(ctx)
			lastProgressSave = time.Now()
		}

		// Wait based on pacing mode
		delay := time.Duration(r.calculateDelay()) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (r *ContinuousSweepRunner) processRoute(ctx context.Context, route db.Route) error {
	r.mu.RLock()
	tripLengths := r.config.TripLengths
	config := r.config
	r.mu.RUnlock()

	// Calculate date range
	startDate := time.Now().AddDate(0, 0, 7) // Start 1 week from now
	endDate := startDate.AddDate(0, 0, config.DepartureWindowDays)

	for _, tripLength := range tripLengths {
		if r.queue == nil {
			return fmt.Errorf("queue is not configured")
		}

		payload := ContinuousPriceGraphPayload{
			Origin:         route.Origin,
			Destination:    route.Destination,
			RangeStartDate: startDate,
			RangeEndDate:   endDate,
			TripLength:     tripLength,
			Class:          config.Class,
			Stops:          config.Stops,
			Adults:         config.Adults,
			Currency:       config.Currency,
		}

		if _, err := r.queue.Enqueue(ctx, "continuous_price_graph", payload); err != nil {
			return fmt.Errorf("failed to enqueue continuous price graph job for %s->%s: %w", route.Origin, route.Destination, err)
		}

		// Track delay for avg calculation
		delay := r.calculateDelay()
		r.mu.Lock()
		r.totalDelayMs += int64(delay)
		r.queriesCompleted++
		r.mu.Unlock()
	}

	return nil
}

func (r *ContinuousSweepRunner) completeSweep(ctx context.Context) {
	r.mu.Lock()
	sweepNum := r.sweepNumber
	duration := time.Since(r.startTime)
	queries := r.queriesCompleted
	errors := r.errorsCount
	minPrice := r.minPriceFound
	maxPrice := r.maxPriceFound
	totalDelay := r.totalDelayMs

	// Record sweep stats
	stats := db.ContinuousSweepStats{
		SweepNumber:       sweepNum,
		StartedAt:         r.startTime,
		CompletedAt:       sql.NullTime{Time: time.Now(), Valid: true},
		TotalRoutes:       len(r.routes),
		SuccessfulQueries: queries - errors,
		FailedQueries:     errors,
	}
	if duration.Seconds() > 0 {
		stats.TotalDurationSeconds = sql.NullInt32{Int32: int32(duration.Seconds()), Valid: true}
	}
	if queries > 0 {
		avgDelay := int32(totalDelay / int64(queries))
		stats.AvgDelayMs = sql.NullInt32{Int32: avgDelay, Valid: true}
	}
	if minPrice > 0 {
		stats.MinPriceFound = sql.NullFloat64{Float64: minPrice, Valid: true}
	}
	if maxPrice > 0 {
		stats.MaxPriceFound = sql.NullFloat64{Float64: maxPrice, Valid: true}
	}

	// Reset for next sweep
	r.sweepNumber++
	r.routeIndex = 0
	r.queriesCompleted = 0
	r.errorsCount = 0
	r.minPriceFound = 0
	r.maxPriceFound = 0
	r.totalDelayMs = 0
	r.startTime = time.Now()
	r.mu.Unlock()

	// Save stats
	if err := r.postgresDB.InsertContinuousSweepStats(ctx, stats); err != nil {
		log.Printf("Failed to save sweep stats: %v", err)
	}

	// Notify completion
	if r.notifier != nil && r.notifier.IsEnabled() {
		r.notifier.AlertSweepComplete(sweepNum, duration, queries, errors)
	}

	log.Printf("Sweep #%d complete: %d queries in %v with %d errors", sweepNum, queries, duration, errors)
}

func (r *ContinuousSweepRunner) recordError(ctx context.Context, err error) {
	r.mu.Lock()
	r.errorsCount++
	r.lastError = err.Error()
	r.lastErrorTime = time.Now()

	// Track recent errors for spike detection
	r.recentErrors = append(r.recentErrors, time.Now())

	// Clean old errors (older than error window)
	if r.notifier != nil {
		errorWindow := r.notifier.GetConfig().ErrorWindow
		cutoff := time.Now().Add(-errorWindow)
		filtered := r.recentErrors[:0]
		for _, t := range r.recentErrors {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		r.recentErrors = filtered

		// Check for error spike
		if len(r.recentErrors) >= r.notifier.GetConfig().ErrorThreshold {
			sweepNum := r.sweepNumber
			errorCount := len(r.recentErrors)
			lastErr := r.lastError
			r.mu.Unlock()
			r.notifier.AlertErrorSpike(sweepNum, errorCount, errorWindow, lastErr)
			return
		}
	}
	r.mu.Unlock()

	log.Printf("Sweep error: %v", err)
}

func (r *ContinuousSweepRunner) calculateDelay() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.config.PacingMode == PacingModeFixed {
		return r.config.MinDelayMs
	}

	// Adaptive mode: calculate delay to complete in target duration
	totalQueries := len(r.routes) * len(r.config.TripLengths)
	if totalQueries == 0 {
		return r.config.MinDelayMs
	}

	targetDuration := time.Duration(r.config.TargetDurationHours) * time.Hour
	delayMs := int(targetDuration.Milliseconds() / int64(totalQueries))

	// Enforce minimum delay
	if delayMs < r.config.MinDelayMs {
		delayMs = r.config.MinDelayMs
	}

	return delayMs
}

func (r *ContinuousSweepRunner) saveProgress(ctx context.Context) {
	r.mu.RLock()
	progress := db.ContinuousSweepProgress{
		ID:                  1, // Single row
		SweepNumber:         r.sweepNumber,
		RouteIndex:          r.routeIndex,
		TotalRoutes:         len(r.routes) * len(r.config.TripLengths),
		QueriesCompleted:    r.queriesCompleted,
		ErrorsCount:         r.errorsCount,
		LastError:           sql.NullString{String: r.lastError, Valid: r.lastError != ""},
		SweepStartedAt:      sql.NullTime{Time: r.startTime, Valid: !r.startTime.IsZero()},
		LastUpdated:         time.Now(),
		TripLengths:         append([]int(nil), r.config.TripLengths...),
		PacingMode:          string(r.config.PacingMode),
		TargetDurationHours: r.config.TargetDurationHours,
		MinDelayMs:          r.config.MinDelayMs,
		IsRunning:           r.isRunning,
		IsPaused:            r.isPaused,
		InternationalOnly:   r.config.InternationalOnly,
	}
	if r.routeIndex < len(r.routes) {
		progress.CurrentOrigin = sql.NullString{String: r.routes[r.routeIndex].Origin, Valid: true}
		progress.CurrentDestination = sql.NullString{String: r.routes[r.routeIndex].Destination, Valid: true}
	}
	r.mu.RUnlock()

	if err := r.postgresDB.SaveContinuousSweepProgress(ctx, progress); err != nil {
		log.Printf("Failed to save sweep progress: %v", err)
	}
}

func (r *ContinuousSweepRunner) restoreProgress(ctx context.Context) error {
	progress, err := r.postgresDB.GetContinuousSweepProgress(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if InternationalOnly config has changed - if so, reset the sweep
	// since the route set would be different
	if progress.InternationalOnly != r.config.InternationalOnly {
		log.Printf("InternationalOnly config changed (was %v, now %v), resetting sweep progress",
			progress.InternationalOnly, r.config.InternationalOnly)
		r.sweepNumber = progress.SweepNumber + 1
		r.routeIndex = 0
		r.queriesCompleted = 0
		r.errorsCount = 0
		r.startTime = time.Now()
		return nil
	}

	r.sweepNumber = progress.SweepNumber
	r.routeIndex = progress.RouteIndex
	r.queriesCompleted = progress.QueriesCompleted
	r.errorsCount = progress.ErrorsCount
	if progress.LastError.Valid {
		r.lastError = progress.LastError.String
	}
	if progress.SweepStartedAt.Valid {
		r.startTime = progress.SweepStartedAt.Time
	}
	if len(progress.TripLengths) > 0 {
		r.config.TripLengths = append([]int(nil), progress.TripLengths...)
	}
	r.config.PacingMode = PacingMode(progress.PacingMode)
	if progress.TargetDurationHours > 0 {
		r.config.TargetDurationHours = progress.TargetDurationHours
	}
	if progress.MinDelayMs > 0 {
		r.config.MinDelayMs = progress.MinDelayMs
	}

	log.Printf("Restored sweep progress: sweep #%d, route %d/%d",
		r.sweepNumber, r.routeIndex, len(r.routes))

	return nil
}
