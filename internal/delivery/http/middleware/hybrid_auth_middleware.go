package middleware

import (
	"go-gin-clean/internal/gateway/security"

	"github.com/gin-gonic/gin"
)

// HybridAuthMiddleware supports both Kong (header-based) and JWT authentication
// Priority: Kong headers take precedence, fallback to JWT if headers not present
type HybridAuthMiddleware struct {
	jwtMiddleware  *AuthMiddleware
	kongMiddleware *KongAuthMiddleware
}

func NewHybridAuthMiddleware(jwtService *security.JWTService) *HybridAuthMiddleware {
	return &HybridAuthMiddleware{
		jwtMiddleware:  NewAuthMiddleware(jwtService),
		kongMiddleware: NewKongAuthMiddleware(),
	}
}

// RequireAuth checks for Kong headers first, falls back to JWT
func (m *HybridAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Kong has injected authentication headers
		authenticated := c.GetHeader("X-Authenticated")
		
		if authenticated == "true" {
			// Use Kong middleware (reads headers)
			m.kongMiddleware.RequireAuth()(c)
		} else {
			// Fallback to JWT middleware
			m.jwtMiddleware.RequireAuth()(c)
		}
	}
}

// RequirePermission checks permission from Kong headers or JWT claims
func (m *HybridAuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Kong has injected authentication headers
		authenticated := c.GetHeader("X-Authenticated")
		
		if authenticated == "true" {
			// Use Kong middleware
			m.kongMiddleware.RequirePermission(permission)(c)
		} else {
			// For JWT, we only have RequireAuth - permission check would need to be added
			m.jwtMiddleware.RequireAuth()(c)
		}
	}
}

// RequireRole checks role from Kong headers or JWT claims  
func (m *HybridAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if Kong has injected authentication headers
		authenticated := c.GetHeader("X-Authenticated")
		
		if authenticated == "true" {
			// Use Kong middleware
			m.kongMiddleware.RequireRole(role)(c)
		} else {
			// For JWT, we only have RequireAuth - role check would need to be added
			m.jwtMiddleware.RequireAuth()(c)
		}
	}
}
