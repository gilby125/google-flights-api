package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/pkg/cache"
	"github.com/gilby125/google-flights-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CacheConfig holds cache middleware configuration
type CacheConfig struct {
	TTL         time.Duration
	KeyPrefix   string
	SkipPaths   []string
	OnlyMethods []string
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// ResponseCache creates a middleware that caches HTTP responses
func ResponseCache(cacheManager *cache.CacheManager, config CacheConfig) gin.HandlerFunc {
	if config.OnlyMethods == nil {
		config.OnlyMethods = []string{"GET"}
	}

	return func(c *gin.Context) {
		// Skip if method not in allowed list
		if !contains(config.OnlyMethods, c.Request.Method) {
			c.Next()
			return
		}

		// Skip if path is in skip list
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, skipPath) {
				c.Next()
				return
			}
		}

		// Skip caching for HTML responses (mostly browser requests)
		if strings.Contains(c.GetHeader("Accept"), "text/html") {
			c.Next()
			return
		}

		// Generate cache key
		cacheKey := generateCacheKey(config.KeyPrefix, c.Request)

		// Try to get from cache
		var cachedResponse CachedResponse
		err := cacheManager.GetJSON(c.Request.Context(), cacheKey, &cachedResponse)
		if err == nil {
			// Cache hit - return cached response
			logger.WithField("cache_key", cacheKey).Debug("Cache hit")

			// Set headers
			for key, value := range cachedResponse.Headers {
				c.Header(key, value)
			}
			c.Header("X-Cache", "HIT")

			c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
			c.Abort() // stop remaining handlers from executing
			return
		}

		if err != cache.ErrCacheMiss {
			logger.WithField("cache_key", cacheKey).Error(err, "Cache get error")
		}

		// Cache miss - execute request and cache response
		logger.WithField("cache_key", cacheKey).Debug("Cache miss")

		// Wrap response writer to capture response
		body := &bytes.Buffer{}
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           body,
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Only cache JSON responses to avoid duplicating HTML/static assets
		if !strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			return
		}

		// Only cache successful responses (2xx status codes)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Prepare cached response
			cachedResp := CachedResponse{
				StatusCode:  c.Writer.Status(),
				Headers:     make(map[string]string),
				Body:        body.Bytes(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				CachedAt:    time.Now(),
			}

			// Copy relevant headers
			for key, values := range c.Writer.Header() {
				if len(values) > 0 && shouldCacheHeader(key) {
					cachedResp.Headers[key] = values[0]
				}
			}

			// Cache the response
			if err := cacheManager.SetJSON(c.Request.Context(), cacheKey, cachedResp, config.TTL); err != nil {
				logger.WithField("cache_key", cacheKey).Error(err, "Cache set error")
			} else {
				logger.WithField("cache_key", cacheKey).Debug("Response cached")
			}
		}

		// Add cache miss header
		c.Header("X-Cache", "MISS")
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	ContentType string            `json:"content_type"`
	CachedAt    time.Time         `json:"cached_at"`
}

// generateCacheKey creates a cache key from the HTTP request
func generateCacheKey(prefix string, req *http.Request) string {
	// Include method, path, and query parameters
	keyData := fmt.Sprintf("%s:%s:%s", req.Method, req.URL.Path, req.URL.RawQuery)

	// Include relevant headers (e.g., Accept, Accept-Language)
	if accept := req.Header.Get("Accept"); accept != "" {
		keyData += ":" + accept
	}
	if acceptLang := req.Header.Get("Accept-Language"); acceptLang != "" {
		keyData += ":" + acceptLang
	}

	// Hash the key data to create a consistent, shorter key
	hash := md5.Sum([]byte(keyData))
	hashStr := fmt.Sprintf("%x", hash)

	if prefix != "" {
		return fmt.Sprintf("%s:response:%s", prefix, hashStr)
	}
	return fmt.Sprintf("response:%s", hashStr)
}

// shouldCacheHeader determines if a header should be cached
func shouldCacheHeader(header string) bool {
	header = strings.ToLower(header)
	cacheable := []string{
		"content-type",
		"content-encoding",
		"cache-control",
		"etag",
		"last-modified",
	}

	for _, h := range cacheable {
		if header == h {
			return true
		}
	}
	return false
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
