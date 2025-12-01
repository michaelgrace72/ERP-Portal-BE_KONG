package http

import (
	"net/http"

	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Login handles user authentication and creates a phantom token session
// @Summary User Login
// @Description Authenticates user and returns a reference token (phantom token)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.PhantomLoginRequest true "Login credentials"
// @Success 200 {object} model.PhantomLoginResponse "Login successful"
// @Success 200 {object} model.TenantSelectionResponse "Multiple tenants available - selection required"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.PhantomLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Invalid request", err.Error(), http.StatusBadRequest)
		return
	}

	loginResp, tenantSelectionResp, err := h.authUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "Login failed", err.Error(), http.StatusUnauthorized)
		return
	}

	// Check if tenant selection is required
	if tenantSelectionResp != nil && tenantSelectionResp.RequiresChoice {
		response.Success(c, "Tenant selection required", tenantSelectionResp, http.StatusOK)
		return
	}

	response.Success(c, "Login successful", loginResp, http.StatusOK)
}

// SelectTenant handles tenant selection for multi-tenant users
// @Summary Select Tenant
// @Description Allows multi-tenant users to select their active tenant
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.SelectTenantRequest true "Tenant selection"
// @Success 200 {object} model.PhantomLoginResponse "Tenant selected successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /auth/select-tenant [post]
func (h *AuthHandler) SelectTenant(c *gin.Context) {
	var req model.SelectTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Invalid request", err.Error(), http.StatusBadRequest)
		return
	}

	loginResp, err := h.authUseCase.SelectTenant(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "Tenant selection failed", err.Error(), http.StatusUnauthorized)
		return
	}

	response.Success(c, "Tenant selected successfully", loginResp, http.StatusOK)
}

// Logout handles session invalidation
// @Summary Logout
// @Description Invalidates the user's session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.SuccessResponse "Logout successful"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract reference token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, "Missing authorization header", "", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix
	refToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		refToken = authHeader[7:]
	}

	if err := h.authUseCase.Logout(c.Request.Context(), refToken); err != nil {
		response.Error(c, "Logout failed", err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, "Logout successful", nil, http.StatusOK)
}

// RefreshSession extends the session TTL
// @Summary Refresh Session
// @Description Extends the TTL of the current session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} response.SuccessResponse "Session refreshed"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshSession(c *gin.Context) {
	// Extract reference token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, "Missing authorization header", "", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix
	refToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		refToken = authHeader[7:]
	}

	if err := h.authUseCase.RefreshSession(c.Request.Context(), refToken); err != nil {
		response.Error(c, "Session refresh failed", err.Error(), http.StatusUnauthorized)
		return
	}

	response.Success(c, "Session refreshed successfully", nil, http.StatusOK)
}

// GetSession retrieves the current session context
// @Summary Get Session Context
// @Description Retrieves the full session context for debugging
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} model.SessionValue "Session context"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /auth/session [get]
func (h *AuthHandler) GetSession(c *gin.Context) {
	// Extract reference token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, "Missing authorization header", "", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix
	refToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		refToken = authHeader[7:]
	}

	sessionValue, err := h.authUseCase.GetSessionContext(c.Request.Context(), refToken)
	if err != nil {
		response.Error(c, "Session not found", err.Error(), http.StatusUnauthorized)
		return
	}

	response.Success(c, "Session retrieved", sessionValue, http.StatusOK)
}
