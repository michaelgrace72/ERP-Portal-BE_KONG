package http

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	registrationUseCase *usecase.RegistrationUseCase
}

func NewRegistrationHandler(registrationUseCase *usecase.RegistrationUseCase) *RegistrationHandler {
	return &RegistrationHandler{
		registrationUseCase: registrationUseCase,
	}
}

// RegisterWithTenant handles user registration with tenant creation
// POST /auth/register
func (h *RegistrationHandler) RegisterWithTenant(c *gin.Context) {
	var req model.RegisterWithTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to parse request", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.registrationUseCase.RegisterWithTenant(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "Registration failed", err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, "Registration successful. User and tenant created.", result, http.StatusCreated)
}
