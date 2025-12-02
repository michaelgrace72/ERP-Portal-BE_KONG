package http

import (
	"net/http"
	"strings"

	"go-gin-clean/internal/usecase"

	"github.com/gin-gonic/gin"
)

type IntrospectionHandler struct {
	introspectionUseCase *usecase.IntrospectionUseCase
}

func NewIntrospectionHandler(introspectionUseCase *usecase.IntrospectionUseCase) *IntrospectionHandler {
	return &IntrospectionHandler{
		introspectionUseCase: introspectionUseCase,
	}
}

// Introspect handles POST /auth/introspect
// This endpoint is called by Kong's auth-request plugin to validate phantom tokens
// It returns session context as HTTP headers that Kong will inject into upstream requests
func (h *IntrospectionHandler) Introspect(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// Return inactive response
		c.JSON(http.StatusOK, gin.H{
			"active": false,
		})
		return
	}

	// Remove Bearer prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	// Introspect the token
	resp, err := h.introspectionUseCase.IntrospectToken(c.Request.Context(), token)
	if err != nil || !resp.Active {
		c.JSON(http.StatusOK, gin.H{
			"active": false,
		})
		return
	}

	// Get headers to inject
	headers := h.introspectionUseCase.GetHeadersForUpstream(resp)

	// Set headers in response (Kong will copy these to upstream request)
	for key, value := range headers {
		c.Header(key, value)
	}

	// Return introspection response
	c.JSON(http.StatusOK, gin.H{
		"active":      resp.Active,
		"sub":         resp.Sub,
		"tenant_id":   resp.TenantID,
		"user_id":     resp.UserID,
		"role_id":     resp.RoleID,
		"role_name":   resp.RoleName,
		"permissions": resp.Permissions,
		"exp":         resp.Exp,
	})
}
