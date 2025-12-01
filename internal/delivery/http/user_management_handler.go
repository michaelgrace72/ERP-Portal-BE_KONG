package http

import (
	"net/http"
	"strconv"

	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserManagementHandler struct {
	userManagementUseCase *usecase.UserManagementUseCase
}

func NewUserManagementHandler(userManagementUseCase *usecase.UserManagementUseCase) *UserManagementHandler {
	return &UserManagementHandler{
		userManagementUseCase: userManagementUseCase,
	}
}

// GetMyProfile handles GET /api/v1/users/me
func (h *UserManagementHandler) GetMyProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.userManagementUseCase.GetUserProfile(c.Request.Context(), userID.(int64))
	if err != nil {
		response.Error(c, "failed to get profile", err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, "profile retrieved successfully", profile, http.StatusOK)
}

// CreateUser handles POST /api/v1/users
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}
	_ = requestorUserID // Will be used for permission checks in the future

	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error(), "", http.StatusBadRequest)
		return
	}

	user, err := h.userManagementUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusInternalServerError)
		return
	}

	response.Success(c, "user created successfully", user, http.StatusCreated)
}

// AssignUserToTenant handles POST /api/v1/memberships
func (h *UserManagementHandler) AssignUserToTenant(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	var req model.AssignUserToTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error(), "", http.StatusBadRequest)
		return
	}

	result, err := h.userManagementUseCase.AssignUserToTenant(c.Request.Context(), &req, requestorUserID.(int64))
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "user assigned to tenant successfully", result, http.StatusCreated)
}

// RemoveUserFromTenant handles DELETE /api/v1/memberships
func (h *UserManagementHandler) RemoveUserFromTenant(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	var req model.RemoveUserFromTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error(), "", http.StatusBadRequest)
		return
	}

	err := h.userManagementUseCase.RemoveUserFromTenant(c.Request.Context(), &req, requestorUserID.(int64))
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "user removed from tenant successfully", nil, http.StatusOK)
}

// UpdateUserRole handles PUT /api/v1/memberships/:id/role
func (h *UserManagementHandler) UpdateUserRole(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	// Get membership ID from URL
	membershipIDStr := c.Param("id")
	membershipID, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		response.Error(c, "invalid membership ID", "", http.StatusBadRequest)
		return
	}

	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error(), "", http.StatusBadRequest)
		return
	}

	// Set membership ID from URL
	req.MembershipID = membershipID

	err = h.userManagementUseCase.UpdateUserRole(c.Request.Context(), &req, requestorUserID.(int64))
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "user role updated successfully", nil, http.StatusOK)
}

// GetTenantMembers handles GET /api/v1/tenants/:id/members
func (h *UserManagementHandler) GetTenantMembers(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	// Get tenant ID from URL
	tenantIDStr := c.Param("id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response.Error(c, "invalid tenant ID", "", http.StatusBadRequest)
		return
	}

	members, err := h.userManagementUseCase.GetTenantMembers(c.Request.Context(), tenantID, requestorUserID.(int64))
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "tenant members retrieved successfully", members, http.StatusOK)
}

// UpdateTenant handles PUT /api/v1/tenants/:id
func (h *UserManagementHandler) UpdateTenant(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	// Get tenant ID from URL
	tenantIDStr := c.Param("id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response.Error(c, "invalid tenant ID", "", http.StatusBadRequest)
		return
	}

	var req model.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err.Error(), "", http.StatusBadRequest)
		return
	}

	err = h.userManagementUseCase.UpdateTenant(c.Request.Context(), tenantID, &req, requestorUserID.(int64))
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "tenant updated successfully", nil, http.StatusOK)
}

// GetTenantRoles handles GET /api/v1/tenants/:id/roles
func (h *UserManagementHandler) GetTenantRoles(c *gin.Context) {
	// Get requestor user ID from context
	requestorUserID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, "user not authenticated", "", http.StatusUnauthorized)
		return
	}

	// Get tenant ID from URL
	tenantIDStr := c.Param("id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		response.Error(c, "invalid tenant ID", "", http.StatusBadRequest)
		return
	}

	// Check if permissions should be included (query param)
	includePermissions := c.Query("include_permissions") == "true"

	roles, err := h.userManagementUseCase.GetTenantRoles(c.Request.Context(), tenantID, requestorUserID.(int64), includePermissions)
	if err != nil {
		response.Error(c, err.Error(), "", http.StatusForbidden)
		return
	}

	response.Success(c, "tenant roles retrieved successfully", roles, http.StatusOK)
}
