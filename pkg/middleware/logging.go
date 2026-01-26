package middleware

import (
	"time"

	"github.com/gilby125/google-flights-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RequestLogger creates a structured logging middleware for Gin
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Create fields map
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency,
			"client_ip":  clientIP,
			"user_agent": c.Request.UserAgent(),
		}

		if requestID := GetRequestID(c); requestID != "" {
			fields["request_id"] = requestID
		}

		if raw != "" {
			fields["query"] = raw
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		logMsg := "HTTP Request"

		switch {
		case statusCode >= 500:
			logger.WithFields(fields).Error(nil, logMsg)
		case statusCode >= 400:
			logger.WithFields(fields).Warn(logMsg)
		default:
			logger.WithFields(fields).Info(logMsg)
		}
	}
}

// Recovery creates a recovery middleware with structured logging
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		fields := map[string]interface{}{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"client_ip": c.ClientIP(),
			"panic":     recovered,
		}

		logger.WithFields(fields).Error(nil, "Panic recovered")
		c.AbortWithStatus(500)
	})
}
