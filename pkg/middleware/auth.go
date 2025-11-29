package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gin-gonic/gin"
)

// AdminAuth returns a middleware that authenticates admin requests.
// If auth is disabled in config, it passes all requests through.
// Supports both Basic Auth and Bearer Token authentication.
func AdminAuth(cfg config.AdminAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If auth is disabled, allow all requests
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Check for Bearer token first
		authHeader := c.GetHeader("Authorization")
		if cfg.Token != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.Token)) == 1 {
				c.Next()
				return
			}
		}

		// Check Basic Auth
		if cfg.Username != "" && cfg.Password != "" {
			username, password, hasAuth := c.Request.BasicAuth()
			if hasAuth {
				usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Username)) == 1
				passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Password)) == 1
				if usernameMatch && passwordMatch {
					c.Next()
					return
				}
			}
		}

		// No valid auth found
		c.Header("WWW-Authenticate", `Basic realm="Admin API"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: valid credentials required for admin API access",
		})
	}
}
