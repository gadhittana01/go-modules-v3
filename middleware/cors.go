package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins (including localhost for development)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://sharehub.gadhittana.com", // Add your production domain
		}

		// Check if origin is in allowed list
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		// Set origin header if allowed, otherwise use first allowed origin as default
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
