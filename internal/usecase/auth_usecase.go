package usecase

import (
	"context"
	"fmt"
	"time"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/gateway/session"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/errors"
)

type AuthUseCase struct {
	userRepo       *repository.UserRepository
	membershipRepo *repository.MembershipRepository
	tenantRepo     *repository.TenantRepository
	tenantRoleRepo *repository.TenantRoleRepository
	permissionRepo *repository.PermissionRepository
	bcryptService  *security.BcryptService
	sessionService *session.SessionService
	sessionTTL     time.Duration
}

func NewAuthUseCase(
	userRepo *repository.UserRepository,
	membershipRepo *repository.MembershipRepository,
	tenantRepo *repository.TenantRepository,
	tenantRoleRepo *repository.TenantRoleRepository,
	permissionRepo *repository.PermissionRepository,
	bcryptService *security.BcryptService,
	sessionService *session.SessionService,
	sessionTTL time.Duration,
) *AuthUseCase {
	if sessionTTL == 0 {
		sessionTTL = 30 * time.Minute // Default 30 minutes
	}
	return &AuthUseCase{
		userRepo:       userRepo,
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		tenantRoleRepo: tenantRoleRepo,
		permissionRepo: permissionRepo,
		bcryptService:  bcryptService,
		sessionService: sessionService,
		sessionTTL:     sessionTTL,
	}
}

// Login authenticates a user and creates a phantom token session
func (uc *AuthUseCase) Login(ctx context.Context, req *model.PhantomLoginRequest) (*model.PhantomLoginResponse, *model.TenantSelectionResponse, error) {
	// 1. Validate credentials
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, errors.ErrUserInactive
	}

	// Verify password
	if err := uc.bcryptService.ComparePassword(user.Password, req.Password); err != nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	// 2. Fetch user's memberships
	memberships, err := uc.membershipRepo.FindByUserID(ctx, user.ID)
	if err != nil || len(memberships) == 0 {
		return nil, nil, fmt.Errorf("user has no tenant memberships")
	}

	// 3. Determine tenant selection
	var selectedMembership *entity.Membership
	
	if req.TenantID != nil {
		// User explicitly selected a tenant
		for _, m := range memberships {
			if m.TenantID == *req.TenantID {
				selectedMembership = &m
				break
			}
		}
		if selectedMembership == nil {
			return nil, nil, fmt.Errorf("user does not have access to the specified tenant")
		}
	} else if len(memberships) == 1 {
		// User has only one tenant - auto-select
		selectedMembership = &memberships[0]
	} else {
		// User has multiple tenants - return tenant list for selection
		tenantSelectionResp, err := uc.buildTenantSelectionResponse(ctx, memberships)
		return nil, tenantSelectionResp, err
	}

	// 4. Build session and create phantom token
	loginResp, err := uc.createLoginSession(ctx, user, selectedMembership)
	return loginResp, nil, err
}

// SelectTenant allows a multi-tenant user to select their active tenant
func (uc *AuthUseCase) SelectTenant(ctx context.Context, req *model.SelectTenantRequest) (*model.PhantomLoginResponse, error) {
	// Re-authenticate user
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Verify password
	if err := uc.bcryptService.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Find membership for selected tenant
	membership, err := uc.membershipRepo.FindByUserAndTenant(ctx, user.ID, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("user does not have access to the specified tenant")
	}

	// Create session
	response, err := uc.createLoginSession(ctx, user, membership)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// createLoginSession creates a session and returns login response
func (uc *AuthUseCase) createLoginSession(ctx context.Context, user *entity.User, membership *entity.Membership) (*model.PhantomLoginResponse, error) {
	// Fetch tenant details
	tenant, err := uc.tenantRepo.FindByID(ctx, membership.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenant: %w", err)
	}

	// Fetch role details
	role, err := uc.tenantRoleRepo.FindByID(ctx, membership.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role: %w", err)
	}

	// Fetch permissions for this role
	permissionEntities, err := uc.permissionRepo.FindByRoleID(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %w", err)
	}

	// Convert permissions to string format (resource:action)
	permissions := make([]string, 0, len(permissionEntities))
	for _, perm := range permissionEntities {
		permissions = append(permissions, fmt.Sprintf("%s:%s", perm.Resource, perm.Action))
	}

	scope := uc.buildScope(permissions)

	// Create session value object
	sessionValue := &model.SessionValue{
		UserID:      user.ID,
		UserUUID:    user.UUID,
		TenantID:    tenant.ID,
		TenantSlug:  tenant.Slug,
		Roles:       []string{role.Name},
		Permissions: permissions,
		Scope:       scope,
		Email:       user.Email,
		Name:        user.Name,
	}

	// Generate reference token and store session in Redis
	refToken, err := uc.sessionService.CreateSession(ctx, sessionValue)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Build response
	response := &model.PhantomLoginResponse{
		AccessToken: refToken,
		ExpiresIn:   int(uc.sessionTTL.Seconds()),
		TokenType:   "Bearer",
		User: model.UserSessionInfo{
			ID:    user.ID,
			UUID:  user.UUID,
			Email: user.Email,
			Name:  user.Name,
		},
		Tenant: model.TenantInfo{
			ID:     tenant.ID,
			Name:   tenant.Name,
			Slug:   tenant.Slug,
		},
	}

	return response, nil
}

// buildTenantSelectionResponse creates a response with available tenants
func (uc *AuthUseCase) buildTenantSelectionResponse(ctx context.Context, memberships []entity.Membership) (*model.TenantSelectionResponse, error) {
	tenants := make([]model.TenantMembership, 0, len(memberships))

	for _, m := range memberships {
		tenant, err := uc.tenantRepo.FindByID(ctx, m.TenantID)
		if err != nil {
			continue // Skip if tenant not found
		}

		role, err := uc.tenantRoleRepo.FindByID(ctx, m.RoleID)
		if err != nil {
			continue // Skip if role not found
		}

		tenants = append(tenants, model.TenantMembership{
			ID:   tenant.ID,
			Name: tenant.Name,
			Slug: tenant.Slug,
			Role: role.Name,
		})
	}

	return &model.TenantSelectionResponse{
		Message:        "User has multiple tenants. Please select one.",
		Tenants:        tenants,
		RequiresChoice: true,
	}, nil
}

// buildScope constructs OAuth-style scope string from permissions
func (uc *AuthUseCase) buildScope(permissions []string) string {
	if len(permissions) == 0 {
		return "read:basic"
	}
	
	scope := ""
	for i, perm := range permissions {
		if i > 0 {
			scope += " "
		}
		scope += perm
	}
	return scope
}

// Logout invalidates a user's session
func (uc *AuthUseCase) Logout(ctx context.Context, refToken string) error {
	return uc.sessionService.DeleteSession(ctx, refToken)
}

// RefreshSession extends the session TTL
func (uc *AuthUseCase) RefreshSession(ctx context.Context, refToken string) error {
	return uc.sessionService.RefreshSession(ctx, refToken)
}

// GetSessionContext retrieves the full session context
func (uc *AuthUseCase) GetSessionContext(ctx context.Context, refToken string) (*model.SessionValue, error) {
	return uc.sessionService.GetSession(ctx, refToken)
}
