package usecase

import (
	"context"
	"fmt"
	"go-gin-clean/internal/gateway/session"
	"go-gin-clean/internal/model"
	"strings"
)

type IntrospectionUseCase struct {
	sessionService *session.SessionService
}

func NewIntrospectionUseCase(sessionService *session.SessionService) *IntrospectionUseCase {
	return &IntrospectionUseCase{
		sessionService: sessionService,
	}
}

// IntrospectToken validates a reference token and returns session context for Kong
// This is called by Kong's auth-request plugin to validate the phantom token
func (uc *IntrospectionUseCase) IntrospectToken(ctx context.Context, token string) (*model.IntrospectionResponse, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Validate token format (ref_<64 hex chars>)
	if !strings.HasPrefix(token, "ref_") || len(token) != 68 {
		return &model.IntrospectionResponse{
			Active: false,
		}, nil
	}

	// Retrieve session from Redis
	sessionValue, err := uc.sessionService.GetSession(ctx, token)
	if err != nil {
		// Session not found or expired
		return &model.IntrospectionResponse{
			Active: false,
		}, nil
	}

	// Session is valid, return context
	// Note: Roles is an array, we'll use the first role if available
	roleName := ""
	if len(sessionValue.Roles) > 0 {
		roleName = sessionValue.Roles[0]
	}

	return &model.IntrospectionResponse{
		Active:      true,
		Sub:         fmt.Sprintf("user_%d", sessionValue.UserID),
		TenantID:    sessionValue.TenantID,
		UserID:      sessionValue.UserID,
		RoleID:      0, // We don't store role ID in session, only names
		RoleName:    roleName,
		Permissions: sessionValue.Permissions,
		Exp:         sessionValue.ExpiresAt,
	}, nil
}

// GetHeadersForUpstream generates the headers that Kong should inject
func (uc *IntrospectionUseCase) GetHeadersForUpstream(resp *model.IntrospectionResponse) map[string]string {
	if !resp.Active {
		return nil
	}

	return map[string]string{
		"X-Tenant-ID":    fmt.Sprintf("%d", resp.TenantID),
		"X-User-ID":      fmt.Sprintf("%d", resp.UserID),
		"X-Role-ID":      fmt.Sprintf("%d", resp.RoleID),
		"X-Role-Name":    resp.RoleName,
		"X-Permissions":  strings.Join(resp.Permissions, ","),
		"X-Authenticated": "true",
	}
}
