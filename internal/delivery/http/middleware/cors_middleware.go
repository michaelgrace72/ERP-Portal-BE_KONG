package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowOrigin := ""

		// Check if origin is allowed
		for _, o := range allowedOrigins {
			if o == "*" {
				allowOrigin = "*"
				break
			}
			if o == origin {
				allowOrigin = origin
				break
			}

			// Check for wildcard patterns (e.g. *.erplabiim.com or https://*.erplabiim.com)
			if strings.Contains(o, "*") {
				parts := strings.SplitN(o, "*", 2)
				if len(parts) == 2 {
					prefix := parts[0]
					suffix := parts[1]
					if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
						allowOrigin = origin
						break
					}
				}
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Refresh-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
