package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/pkg/worker_registry"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// sseClient represents a connected SSE client
type sseClient struct {
	id       string
	messages chan sseMessage
}

type sseMessage struct {
	event string
	data  []byte
}

// sseHub manages all connected SSE clients
type sseHub struct {
	clients    map[string]*sseClient
	register   chan *sseClient
	unregister chan *sseClient
	broadcast  chan sseMessage
	mu         sync.RWMutex
}

func newSSEHub() *sseHub {
	return &sseHub{
		clients:    make(map[string]*sseClient),
		register:   make(chan *sseClient),
		unregister: make(chan *sseClient),
		broadcast:  make(chan sseMessage, 256),
	}
}

func (h *sseHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			log.Printf("SSE client connected: %s (total: %d)", client.id, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.messages)
			}
			h.mu.Unlock()
			log.Printf("SSE client disconnected: %s (total: %d)", client.id, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.messages <- message:
				default:
					// Client's channel is full, skip this message
				}
			}
			h.mu.RUnlock()
		}
	}
}

var (
	hub     *sseHub
	hubOnce sync.Once
)

func getHub() *sseHub {
	hubOnce.Do(func() {
		hub = newSSEHub()
		go hub.run()
	})
	return hub
}

func writeSSEMessage(w io.Writer, msg sseMessage) error {
	if msg.event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", msg.event); err != nil {
			return err
		}
	}

	// SSE allows multiple `data:` lines; split to be safe.
	data := strings.TrimRight(string(msg.data), "\n")
	if data == "" {
		if _, err := io.WriteString(w, "data: \n\n"); err != nil {
			return err
		}
		return nil
	}

	for _, line := range strings.Split(data, "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// GetAdminEvents returns a handler for Server-Sent Events
func GetAdminEvents(workerManager WorkerStatusProvider, redisClient *redis.Client, cfg config.WorkerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

		// Create client
		clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
		client := &sseClient{
			id:       clientID,
			messages: make(chan sseMessage, 10),
		}

		// Register client
		h := getHub()
		h.register <- client
		defer func() {
			h.unregister <- client
		}()

		// Start background worker to send updates
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Broadcast worker status
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
							})
						}
					}

					// Remote worker instances (published by worker processes)
					if redisClient != nil {
						namespace := cfg.RegistryNamespace
						if namespace == "" {
							namespace = "flights"
						}

						// Quick timeout for SSE to avoid blocking
						timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
						heartbeats, err := getRemoteWorkers(timeoutCtx, redisClient, namespace, cfg.HeartbeatTTL)
						timeoutCancel()
						if err == nil {
							out = append(out, heartbeats...)
						}
					}

					if data, err := json.Marshal(out); err == nil {
						h.broadcast <- sseMessage{event: "worker-status", data: data}
					}
				}
			}
		}()

		// Stream messages to client
		c.Stream(func(w io.Writer) bool {
			select {
			case <-ctx.Done():
				return false
			case msg, ok := <-client.messages:
				if !ok {
					return false
				}
				if err := writeSSEMessage(w, msg); err != nil {
					return false
				}
				c.Writer.Flush()
				return true
			}
		})
	}
}

// Helper to get remote workers (extracted from GetWorkerStatus)
func getRemoteWorkers(ctx context.Context, redisClient *redis.Client, namespace string, ttl time.Duration) ([]workerStatusResponse, error) {
	reg := worker_registry.New(redisClient, namespace)
	heartbeats, err := reg.ListActive(ctx, ttl, 200)
	if err != nil {
		return nil, err
	}

	out := make([]workerStatusResponse, 0, len(heartbeats))
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
		})
	}

	return out, nil
}
