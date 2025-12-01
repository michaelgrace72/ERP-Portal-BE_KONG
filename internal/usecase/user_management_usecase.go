package usecase

import (
	"context"
	"fmt"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/errors"

	"gorm.io/gorm"
)

type UserManagementUseCase struct {
	db             *gorm.DB
	userRepo       *repository.UserRepository
	tenantRepo     *repository.TenantRepository
	tenantRoleRepo *repository.TenantRoleRepository
	membershipRepo *repository.MembershipRepository
	permissionRepo *repository.PermissionRepository
	bcryptService  *security.BcryptService
}

func NewUserManagementUseCase(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	tenantRoleRepo *repository.TenantRoleRepository,
	membershipRepo *repository.MembershipRepository,
	permissionRepo *repository.PermissionRepository,
	bcryptService *security.BcryptService,
) *UserManagementUseCase {
	return &UserManagementUseCase{
		db:             db,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		tenantRoleRepo: tenantRoleRepo,
		membershipRepo: membershipRepo,
		permissionRepo: permissionRepo,
		bcryptService:  bcryptService,
	}
}

// GetUserProfile returns the authenticated user's complete profile with all memberships
func (uc *UserManagementUseCase) GetUserProfile(ctx context.Context, userID int64) (*model.GetUserProfileResponse, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Get all memberships for this user
	memberships, err := uc.membershipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memberships: %w", err)
	}

	// Build membership responses with tenant and role details
	membershipResponses := make([]model.MembershipResponse, 0, len(memberships))
	for _, m := range memberships {
		tenant, err := uc.tenantRepo.FindByID(ctx, m.TenantID)
		if err != nil {
			continue // Skip if tenant not found
		}

		role, err := uc.tenantRoleRepo.FindByID(ctx, m.RoleID)
		if err != nil {
			continue // Skip if role not found
		}

		// Get permissions for this role
		permissions, err := uc.permissionRepo.FindByRoleID(ctx, role.ID)
		if err != nil {
			continue
		}

		permissionStrings := make([]string, 0, len(permissions))
		for _, p := range permissions {
			permissionStrings = append(permissionStrings, fmt.Sprintf("%s:%s", p.Resource, p.Action))
		}

		membershipResponses = append(membershipResponses, model.MembershipResponse{
			MembershipID: m.ID,
			TenantID:     tenant.ID,
			TenantName:   tenant.Name,
			TenantSlug:   tenant.Slug,
			RoleID:       role.ID,
			RoleName:     role.Name,
			Permissions:  permissionStrings,
		})
	}

	return &model.GetUserProfileResponse{
		UserID:      user.ID,
		UUID:        user.UUID,
		Code:        user.Code,
		Name:        user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
		IsActive:    user.IsActive,
		IsVerified:  user.IsVerified,
		Memberships: membershipResponses,
	}, nil
}

// CreateUser creates a new user without tenant (for invitation to existing tenant)
func (uc *UserManagementUseCase) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.CreateUserResponse, error) {
	// Check if user already exists
	existingUser, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entity.User{
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		IsActive:   true,
		IsVerified: false,
	}

	if err := uc.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &model.CreateUserResponse{
		UserID:   user.ID,
		UUID:     user.UUID,
		Code:     user.Code,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: user.IsActive,
	}, nil
}

// AssignUserToTenant adds a user to a tenant with a specific role
func (uc *UserManagementUseCase) AssignUserToTenant(ctx context.Context, req *model.AssignUserToTenantRequest, requestorUserID int64) (*model.AssignUserToTenantResponse, error) {
	// Verify requestor has permission (must be owner/admin of the tenant)
	if err := uc.verifyTenantAdmin(ctx, requestorUserID, req.TenantID); err != nil {
		return nil, err
	}

	// Verify user exists
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Verify tenant exists
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Verify role exists and belongs to this tenant
	role, err := uc.tenantRoleRepo.FindByID(ctx, req.RoleID)
	if err != nil || role.TenantID != req.TenantID {
		return nil, fmt.Errorf("role not found or does not belong to this tenant")
	}

	// Check if user is already a member of this tenant
	existingMemberships, _ := uc.membershipRepo.FindByUserID(ctx, req.UserID)
	for _, m := range existingMemberships {
		if m.TenantID == req.TenantID {
			return nil, fmt.Errorf("user is already a member of this tenant")
		}
	}

	// Create membership
	membership := &entity.Membership{
		UserID:   req.UserID,
		TenantID: req.TenantID,
		RoleID:   req.RoleID,
	}

	if err := uc.db.Create(membership).Error; err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	return &model.AssignUserToTenantResponse{
		MembershipID: membership.ID,
		UserID:       user.ID,
		UserName:     user.Name,
		TenantID:     tenant.ID,
		TenantName:   tenant.Name,
		RoleID:       role.ID,
		RoleName:     role.Name,
	}, nil
}

// RemoveUserFromTenant removes a user's membership from a tenant
func (uc *UserManagementUseCase) RemoveUserFromTenant(ctx context.Context, req *model.RemoveUserFromTenantRequest, requestorUserID int64) error {
	// Verify requestor has permission
	if err := uc.verifyTenantAdmin(ctx, requestorUserID, req.TenantID); err != nil {
		return err
	}

	// Cannot remove yourself if you're the only owner
	if requestorUserID == req.UserID {
		return fmt.Errorf("cannot remove yourself from tenant")
	}

	// Find and delete membership
	var membership entity.Membership
	if err := uc.db.Where("user_id = ? AND tenant_id = ?", req.UserID, req.TenantID).First(&membership).Error; err != nil {
		return fmt.Errorf("membership not found")
	}

	if err := uc.db.Delete(&membership).Error; err != nil {
		return fmt.Errorf("failed to remove membership: %w", err)
	}

	return nil
}

// UpdateUserRole changes a user's role within a tenant
func (uc *UserManagementUseCase) UpdateUserRole(ctx context.Context, req *model.UpdateUserRoleRequest, requestorUserID int64) error {
	// Find membership
	membership, err := uc.membershipRepo.FindByID(ctx, req.MembershipID)
	if err != nil {
		return fmt.Errorf("membership not found")
	}

	// Verify requestor has permission
	if err := uc.verifyTenantAdmin(ctx, requestorUserID, membership.TenantID); err != nil {
		return err
	}

	// Verify new role exists and belongs to this tenant
	role, err := uc.tenantRoleRepo.FindByID(ctx, req.RoleID)
	if err != nil || role.TenantID != membership.TenantID {
		return fmt.Errorf("role not found or does not belong to this tenant")
	}

	// Update role
	membership.RoleID = req.RoleID
	if err := uc.db.Save(membership).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// GetTenantMembers lists all members of a tenant
func (uc *UserManagementUseCase) GetTenantMembers(ctx context.Context, tenantID int64, requestorUserID int64) (*model.GetTenantMembersResponse, error) {
	// Verify requestor is a member of this tenant
	if err := uc.verifyTenantMember(ctx, requestorUserID, tenantID); err != nil {
		return nil, err
	}

	// Get tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Get all memberships for this tenant
	var memberships []entity.Membership
	if err := uc.db.Where("tenant_id = ?", tenantID).Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch memberships: %w", err)
	}

	members := make([]model.TenantMemberResponse, 0, len(memberships))
	for _, m := range memberships {
		user, err := uc.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			continue
		}

		role, err := uc.tenantRoleRepo.FindByID(ctx, m.RoleID)
		if err != nil {
			continue
		}

		members = append(members, model.TenantMemberResponse{
			MembershipID: m.ID,
			UserID:       user.ID,
			UserUUID:     user.UUID,
			UserCode:     user.Code,
			UserName:     user.Name,
			UserEmail:    user.Email,
			RoleID:       role.ID,
			RoleName:     role.Name,
			JoinedAt:     m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &model.GetTenantMembersResponse{
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		Members:    members,
	}, nil
}

// GetTenantRoles lists all available roles in a tenant
func (uc *UserManagementUseCase) GetTenantRoles(ctx context.Context, tenantID int64, requestorUserID int64, includePermissions bool) (*model.GetTenantRolesResponse, error) {
	// Verify requestor is a member of this tenant
	if err := uc.verifyTenantMember(ctx, requestorUserID, tenantID); err != nil {
		return nil, err
	}

	// Get all roles for this tenant
	var roles []entity.TenantRole
	if err := uc.db.Where("tenant_id = ?", tenantID).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	roleDetails := make([]model.TenantRoleDetail, 0, len(roles))
	for _, r := range roles {
		// Get permissions for this role
		permissions, err := uc.permissionRepo.FindByRoleID(ctx, r.ID)
		if err != nil {
			continue
		}

		var permissionStrings []string
		if includePermissions {
			permissionStrings = make([]string, 0, len(permissions))
			for _, p := range permissions {
				permissionStrings = append(permissionStrings, fmt.Sprintf("%s:%s", p.Resource, p.Action))
			}
		}

		roleDetails = append(roleDetails, model.TenantRoleDetail{
			RoleID:          r.ID,
			RoleName:        r.Name,
			Description:     r.Description,
			PermissionCount: len(permissions),
			Permissions:     permissionStrings,
		})
	}

	return &model.GetTenantRolesResponse{
		TenantID: tenantID,
		Roles:    roleDetails,
	}, nil
}

// UpdateTenant allows tenant owner to update tenant details
func (uc *UserManagementUseCase) UpdateTenant(ctx context.Context, tenantID int64, req *model.UpdateTenantRequest, requestorUserID int64) error {
	// Verify requestor is owner of this tenant
	if err := uc.verifyTenantOwner(ctx, requestorUserID, tenantID); err != nil {
		return err
	}

	// Get tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found")
	}

	// Update tenant
	tenant.Name = req.Name
	if err := uc.db.Save(tenant).Error; err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// Helper functions for authorization checks

func (uc *UserManagementUseCase) verifyTenantMember(ctx context.Context, userID int64, tenantID int64) error {
	memberships, err := uc.membershipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify membership")
	}

	for _, m := range memberships {
		if m.TenantID == tenantID {
			return nil
		}
	}

	return fmt.Errorf("user is not a member of this tenant")
}

func (uc *UserManagementUseCase) verifyTenantAdmin(ctx context.Context, userID int64, tenantID int64) error {
	memberships, err := uc.membershipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify membership")
	}

	for _, m := range memberships {
		if m.TenantID == tenantID {
			role, err := uc.tenantRoleRepo.FindByID(ctx, m.RoleID)
			if err != nil {
				continue
			}
			// Admin roles: Tenant Owner, Administrator, Super Administrator
			if role.Name == "Tenant Owner" || role.Name == "Administrator" || role.Name == "Super Administrator" {
				return nil
			}
		}
	}

	return fmt.Errorf("user does not have admin permission for this tenant")
}

func (uc *UserManagementUseCase) verifyTenantOwner(ctx context.Context, userID int64, tenantID int64) error {
	memberships, err := uc.membershipRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify membership")
	}

	for _, m := range memberships {
		if m.TenantID == tenantID {
			role, err := uc.tenantRoleRepo.FindByID(ctx, m.RoleID)
			if err != nil {
				continue
			}
			if role.Name == "Tenant Owner" {
				return nil
			}
		}
	}

	return fmt.Errorf("user is not the owner of this tenant")
}
