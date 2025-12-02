package middleware

import (
	"go-gin-clean/internal/delivery/http/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// KongAuthMiddleware reads context from Kong-injected headers
// This middleware is used for services behind Kong that use the phantom token pattern
// Kong validates the token and injects headers (X-Tenant-ID, X-User-ID, etc.)
type KongAuthMiddleware struct{}

func NewKongAuthMiddleware() *KongAuthMiddleware {
	return &KongAuthMiddleware{}
}

// RequireAuth validates that Kong has injected the required headers
func (m *KongAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is authenticated (Kong sets this header)
		authenticated := c.GetHeader("X-Authenticated")
		if authenticated != "true" {
			response.Error(c, "authentication required", "missing authentication headers from Kong", http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Read tenant ID
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			response.Error(c, "tenant context required", "missing X-Tenant-ID header", http.StatusUnauthorized)
			c.Abort()
			return
		}

		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
		if err != nil {
			response.Error(c, "invalid tenant ID", "X-Tenant-ID must be a valid integer", http.StatusBadRequest)
			c.Abort()
			return
		}

		// Read user ID
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			response.Error(c, "user context required", "missing X-User-ID header", http.StatusUnauthorized)
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			response.Error(c, "invalid user ID", "X-User-ID must be a valid integer", http.StatusBadRequest)
			c.Abort()
			return
		}

		// Read optional role info
		roleIDStr := c.GetHeader("X-Role-ID")
		roleID := int64(0)
		if roleIDStr != "" {
			roleID, _ = strconv.ParseInt(roleIDStr, 10, 64)
		}

		roleName := c.GetHeader("X-Role-Name")

		// Read permissions (comma-separated)
		permissionsStr := c.GetHeader("X-Permissions")
		var permissions []string
		if permissionsStr != "" {
			permissions = strings.Split(permissionsStr, ",")
		}

		// Set context for downstream handlers
		c.Set("tenant_id", tenantID)
		c.Set("user_id", userID)
		c.Set("role_id", roleID)
		c.Set("role_name", roleName)
		c.Set("permissions", permissions)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequirePermission checks if user has a specific permission
func (m *KongAuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			response.Error(c, "permission denied", "no permissions found in context", http.StatusForbidden)
			c.Abort()
			return
		}

		permList, ok := permissions.([]string)
		if !ok {
			response.Error(c, "permission denied", "invalid permissions format", http.StatusForbidden)
			c.Abort()
			return
		}

		// Check if user has the required permission
		hasPermission := false
		for _, p := range permList {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Error(c, "permission denied", "insufficient permissions", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has a specific role
func (m *KongAuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole, exists := c.Get("role_name")
		if !exists {
			response.Error(c, "access denied", "no role found in context", http.StatusForbidden)
			c.Abort()
			return
		}

		if currentRole != roleName {
			response.Error(c, "access denied", "insufficient role", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
