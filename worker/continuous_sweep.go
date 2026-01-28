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

type continuousSweepControlStore interface {
	GetContinuousSweepControlFlags(ctx context.Context) (*queue.ContinuousSweepControl, error)
	SetContinuousSweepControlFlags(ctx context.Context, isRunning, isPaused *bool) (*queue.ContinuousSweepControl, error)
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
	r.mu.Unlock()

	// Best-effort global guard: if another instance is actively running and recently updated its heartbeat,
	// refuse to start a second runner.
	if r.postgresDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		progress, err := r.postgresDB.GetContinuousSweepProgress(ctx)
		cancel()
		if err == nil && progress != nil && progress.IsRunning {
			if !progress.LastUpdated.IsZero() && time.Since(progress.LastUpdated) < 10*time.Minute {
				return fmt.Errorf("continuous sweep appears to be running elsewhere (last_updated=%s)", progress.LastUpdated.Format(time.RFC3339))
			}
		}
	} else if r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			ctrl, err := store.GetContinuousSweepControlFlags(ctx)
			cancel()
			if err == nil && ctrl != nil && ctrl.IsRunning {
				if !ctrl.LastUpdated.IsZero() && time.Since(ctrl.LastUpdated) < 10*time.Minute {
					return fmt.Errorf("continuous sweep appears to be running elsewhere (redis last_updated=%s)", ctrl.LastUpdated.Format(time.RFC3339))
				}
			}
		}
	}

	r.mu.Lock()

	// Create long-lived context (not tied to HTTP request)
	baseCtx := queue.WithEnqueueMeta(context.Background(), queue.EnqueueMeta{Actor: "continuous_sweep"})
	r.ctx, r.cancelCtx = context.WithCancel(baseCtx)

	// Reinitialize channels for restartability
	r.stopCh = make(chan struct{})
	r.resumeCh = make(chan struct{})

	r.isRunning = true
	r.isPaused = false
	r.mu.Unlock()

	// Persist control flags immediately so other processes/UI can see the sweep is running and so
	// the runner doesn't stop itself on the first DB sync.
	if r.postgresDB != nil {
		running := true
		paused := false
		ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = r.postgresDB.SetContinuousSweepControlFlags(ctrlCtx, &running, &paused)
		cancel()
	}
	if r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			running := true
			paused := false
			ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = store.SetContinuousSweepControlFlags(ctrlCtx, &running, &paused)
			cancel()
		}
	}

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

func (r *ContinuousSweepRunner) syncControlFromDB(ctx context.Context) (stop bool) {
	var (
		dbRunningSet bool
		dbRunning    bool
		dbPausedSet  bool
		dbPaused     bool
		dbLast       time.Time
	)
	if r.postgresDB != nil {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		progress, err := r.postgresDB.GetContinuousSweepProgress(checkCtx)
		cancel()
		if err == nil && progress != nil {
			dbRunningSet = true
			dbRunning = progress.IsRunning
			dbPausedSet = true
			dbPaused = progress.IsPaused
			dbLast = progress.LastUpdated
		}
	}

	var (
		redisRunningSet bool
		redisRunning    bool
		redisPausedSet  bool
		redisPaused     bool
		redisLast       time.Time
	)
	if r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			ctrl, err := store.GetContinuousSweepControlFlags(checkCtx)
			cancel()
			if err == nil && ctrl != nil {
				redisRunningSet = true
				redisRunning = ctrl.IsRunning
				redisPausedSet = true
				redisPaused = ctrl.IsPaused
				redisLast = ctrl.LastUpdated
			}
		}
	}

	// STOP is treated as a kill-switch: if either control plane explicitly says "not running", stop.
	if (dbRunningSet && !dbRunning) || (redisRunningSet && !redisRunning) {
		return true
	}

	// Determine paused state (prefer Postgres if available, otherwise Redis).
	var nextPaused *bool
	switch {
	case dbPausedSet:
		nextPaused = &dbPaused
	case redisPausedSet:
		nextPaused = &redisPaused
	}

	if nextPaused == nil {
		return false
	}

	r.mu.Lock()
	wasPaused := r.isPaused
	r.isPaused = *nextPaused
	resumeCh := r.resumeCh
	r.mu.Unlock()

	// If we were paused and control says resume, poke the channel so the pause select unblocks quickly.
	if wasPaused && !*nextPaused && resumeCh != nil {
		select {
		case resumeCh <- struct{}{}:
		default:
		}
	}

	_ = dbLast
	_ = redisLast
	return false
}

// Stop gracefully stops the continuous sweep
func (r *ContinuousSweepRunner) Stop() {
	r.mu.Lock()
	wasRunning := r.isRunning

	// Cancel context and close stop channel
	cancelFn := r.cancelCtx
	stopCh := r.stopCh
	r.stopCh = nil // Prevent double-close

	r.isRunning = false
	r.isPaused = false
	r.mu.Unlock()

	if !wasRunning {
		return
	}

	if cancelFn != nil {
		cancelFn()
	}
	if stopCh != nil {
		close(stopCh)
	}

	if r.postgresDB != nil {
		running := false
		paused := false
		ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = r.postgresDB.SetContinuousSweepControlFlags(ctrlCtx, &running, &paused)
		cancel()
	}
	if r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			running := false
			paused := false
			ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = store.SetContinuousSweepControlFlags(ctrlCtx, &running, &paused)
			cancel()
		}
	}
}

// Pause pauses the sweep (can be resumed)
func (r *ContinuousSweepRunner) Pause() {
	r.mu.Lock()
	running := r.isRunning
	paused := r.isPaused
	if running && !paused {
		r.isPaused = true
	}
	r.mu.Unlock()

	if running && !paused && r.postgresDB != nil {
		nextPaused := true
		ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = r.postgresDB.SetContinuousSweepControlFlags(ctrlCtx, nil, &nextPaused)
		cancel()
	}
	if running && !paused && r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			nextPaused := true
			ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = store.SetContinuousSweepControlFlags(ctrlCtx, nil, &nextPaused)
			cancel()
		}
	}
}

// Resume resumes a paused sweep
func (r *ContinuousSweepRunner) Resume() {
	r.mu.Lock()
	running := r.isRunning
	paused := r.isPaused
	resumeCh := r.resumeCh
	if running && paused {
		r.isPaused = false
		select {
		case resumeCh <- struct{}{}:
		default:
		}
	}
	r.mu.Unlock()

	if running && paused && r.postgresDB != nil {
		nextPaused := false
		ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = r.postgresDB.SetContinuousSweepControlFlags(ctrlCtx, nil, &nextPaused)
		cancel()
	}
	if running && paused && r.queue != nil {
		if store, ok := r.queue.(continuousSweepControlStore); ok {
			nextPaused := false
			ctrlCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = store.SetContinuousSweepControlFlags(ctrlCtx, nil, &nextPaused)
			cancel()
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
		MinDelayMs:          r.config.MinDelayMs,
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
	// Capture channels at goroutine start - these won't change during execution.
	// Stop() will close stopCh to signal us; reading from the struct field directly
	// is racy because Stop() sets r.stopCh = nil before closing.
	r.mu.RLock()
	stopCh := r.stopCh
	resumeCh := r.resumeCh
	r.mu.RUnlock()

	defer func() {
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
		r.saveProgress(ctx)
	}()

	// Watch DB control flags in the background so STOP/PAUSE/RESUME apply quickly even if
	// the main loop is sleeping on pacing delays.
	go r.watchControlFromDB(ctx)

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
	lastControlSync := time.Time{}

	for {
		if lastControlSync.IsZero() || time.Since(lastControlSync) > 5*time.Second {
			lastControlSync = time.Now()
			if r.syncControlFromDB(ctx) {
				log.Println("Continuous sweep stopped: external stop requested")
				r.Stop()
				return
			}
		}

		select {
		case <-ctx.Done():
			log.Println("Continuous sweep stopped: context cancelled")
			return
		case <-stopCh:
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
			case <-resumeCh:
				log.Println("Continuous sweep resumed")
			case <-stopCh:
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
		case <-stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (r *ContinuousSweepRunner) watchControlFromDB(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if r.syncControlFromDB(ctx) {
			r.mu.RLock()
			cancelFn := r.cancelCtx
			r.mu.RUnlock()
			if cancelFn != nil {
				cancelFn()
			}
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
		// Enforce STOP quickly and prevent enqueuing new jobs after the sweep is stopped in DB.
		if r.syncControlFromDB(ctx) {
			r.Stop()
			return nil
		}

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
