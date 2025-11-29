package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeStall         AlertType = "stall"
	AlertTypeErrorSpike    AlertType = "error_spike"
	AlertTypeSweepComplete AlertType = "sweep_complete"
	AlertTypeRateLimited   AlertType = "rate_limited"
	AlertTypeSweepStarted  AlertType = "sweep_started"
	AlertTypeInfo          AlertType = "info"
)

// Priority levels for NTFY
type Priority int

const (
	PriorityMin     Priority = 1
	PriorityLow     Priority = 2
	PriorityDefault Priority = 3
	PriorityHigh    Priority = 4
	PriorityUrgent  Priority = 5
)

// NTFYConfig holds configuration for NTFY notifications
type NTFYConfig struct {
	ServerURL       string
	Topic           string
	Username        string // Optional basic auth
	Password        string // Optional basic auth
	Enabled         bool
	StallThreshold  time.Duration // Alert if no progress for this duration
	ErrorThreshold  int           // Alert if this many errors in error window
	ErrorWindow     time.Duration // Time window for error counting
	DefaultPriority Priority
}

// NTFYClient handles sending notifications via NTFY
type NTFYClient struct {
	config     NTFYConfig
	httpClient *http.Client
	mu         sync.Mutex

	// Rate limiting to prevent notification spam
	lastAlerts map[AlertType]time.Time
	minGap     time.Duration
}

// NTFYMessage represents a message to send
type NTFYMessage struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title,omitempty"`
	Message  string   `json:"message"`
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Click    string   `json:"click,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
}

// Action represents an action button in the notification
type Action struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url,omitempty"`
	Clear  bool   `json:"clear,omitempty"`
}

// NewNTFYClient creates a new NTFY client
func NewNTFYClient(config NTFYConfig) *NTFYClient {
	if config.ServerURL == "" {
		config.ServerURL = "https://ntfy.sh"
	}
	if config.DefaultPriority == 0 {
		config.DefaultPriority = PriorityDefault
	}
	if config.StallThreshold == 0 {
		config.StallThreshold = 15 * time.Minute
	}
	if config.ErrorThreshold == 0 {
		config.ErrorThreshold = 10
	}
	if config.ErrorWindow == 0 {
		config.ErrorWindow = 5 * time.Minute
	}

	return &NTFYClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		lastAlerts: make(map[AlertType]time.Time),
		minGap:     5 * time.Minute, // Minimum 5 minutes between same alert type
	}
}

// SendAlert sends a notification with rate limiting
func (c *NTFYClient) SendAlert(alertType AlertType, title, message string, priority Priority) error {
	if !c.config.Enabled || c.config.Topic == "" {
		return nil
	}

	c.mu.Lock()
	// Check rate limiting
	if lastTime, ok := c.lastAlerts[alertType]; ok {
		if time.Since(lastTime) < c.minGap {
			c.mu.Unlock()
			return nil // Skip, too soon
		}
	}
	c.lastAlerts[alertType] = time.Now()
	c.mu.Unlock()

	return c.send(title, message, priority, c.tagsForAlertType(alertType))
}

// SendImmediate sends a notification immediately without rate limiting
func (c *NTFYClient) SendImmediate(title, message string, priority Priority, tags []string) error {
	if !c.config.Enabled || c.config.Topic == "" {
		return nil
	}
	return c.send(title, message, priority, tags)
}

func (c *NTFYClient) send(title, message string, priority Priority, tags []string) error {
	msg := NTFYMessage{
		Topic:    c.config.Topic,
		Title:    title,
		Message:  message,
		Priority: int(priority),
		Tags:     tags,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal NTFY message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create NTFY request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add basic auth if configured
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send NTFY notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("NTFY returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (c *NTFYClient) tagsForAlertType(alertType AlertType) []string {
	switch alertType {
	case AlertTypeStall:
		return []string{"warning", "hourglass"}
	case AlertTypeErrorSpike:
		return []string{"rotating_light", "x"}
	case AlertTypeSweepComplete:
		return []string{"white_check_mark", "airplane"}
	case AlertTypeRateLimited:
		return []string{"stop_sign", "snail"}
	case AlertTypeSweepStarted:
		return []string{"rocket", "airplane"}
	default:
		return []string{"information_source"}
	}
}

// AlertStall sends a stall alert
func (c *NTFYClient) AlertStall(sweepNumber int, lastRoute string, duration time.Duration) error {
	title := fmt.Sprintf("Sweep #%d Stalled", sweepNumber)
	message := fmt.Sprintf("No progress for %v. Last route: %s", duration.Round(time.Minute), lastRoute)
	return c.SendAlert(AlertTypeStall, title, message, PriorityHigh)
}

// AlertErrorSpike sends an error spike alert
func (c *NTFYClient) AlertErrorSpike(sweepNumber int, errorCount int, window time.Duration, lastError string) error {
	title := fmt.Sprintf("Sweep #%d Error Spike", sweepNumber)
	message := fmt.Sprintf("%d errors in %v. Last: %s", errorCount, window.Round(time.Minute), lastError)
	return c.SendAlert(AlertTypeErrorSpike, title, message, PriorityHigh)
}

// AlertSweepComplete sends a sweep completion notification
func (c *NTFYClient) AlertSweepComplete(sweepNumber int, duration time.Duration, routesProcessed int, errors int) error {
	title := fmt.Sprintf("Sweep #%d Complete", sweepNumber)
	message := fmt.Sprintf("Processed %d routes in %v with %d errors", routesProcessed, duration.Round(time.Minute), errors)
	return c.SendAlert(AlertTypeSweepComplete, title, message, PriorityDefault)
}

// AlertRateLimited sends a rate limit alert
func (c *NTFYClient) AlertRateLimited(sweepNumber int, route string) error {
	title := fmt.Sprintf("Sweep #%d Rate Limited", sweepNumber)
	message := fmt.Sprintf("Got 429 response on route: %s", route)
	return c.SendAlert(AlertTypeRateLimited, title, message, PriorityUrgent)
}

// AlertSweepStarted sends a sweep start notification
func (c *NTFYClient) AlertSweepStarted(sweepNumber int, totalRoutes int, estimatedDuration time.Duration) error {
	title := fmt.Sprintf("Sweep #%d Started", sweepNumber)
	message := fmt.Sprintf("Processing %d routes. Est. completion: %v", totalRoutes, estimatedDuration.Round(time.Minute))
	return c.SendAlert(AlertTypeSweepStarted, title, message, PriorityLow)
}

// IsEnabled returns whether notifications are enabled
func (c *NTFYClient) IsEnabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.Enabled && c.config.Topic != ""
}

// GetConfig returns the current configuration
func (c *NTFYClient) GetConfig() NTFYConfig {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config
}

// UpdateConfig updates the configuration
func (c *NTFYClient) UpdateConfig(config NTFYConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
}
