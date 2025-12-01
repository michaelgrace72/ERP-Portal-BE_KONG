package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/kong"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/errors"

	"gorm.io/gorm"
)

type RegistrationUseCase struct {
	db                   *gorm.DB
	userRepo             *repository.UserRepository
	tenantRepo           *repository.TenantRepository
	tenantRoleRepo       *repository.TenantRoleRepository
	membershipRepo       *repository.MembershipRepository
	bcryptService        *security.BcryptService
	kongClient           *kong.KongAdminClient
}

func NewRegistrationUseCase(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	tenantRoleRepo *repository.TenantRoleRepository,
	membershipRepo *repository.MembershipRepository,
	bcryptService *security.BcryptService,
	kongClient *kong.KongAdminClient,
) *RegistrationUseCase {
	return &RegistrationUseCase{
		db:             db,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		tenantRoleRepo: tenantRoleRepo,
		membershipRepo: membershipRepo,
		bcryptService:  bcryptService,
		kongClient:     kongClient,
	}
}

// RegisterWithTenant creates a new user, tenant, and sets the user as the owner
func (uc *RegistrationUseCase) RegisterWithTenant(ctx context.Context, req *model.RegisterWithTenantRequest) (*model.RegisterWithTenantResponse, error) {
	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" || req.CompanyName == "" {
		return nil, errors.ErrValidationFailed
	}

	// Check if user already exists
	existingUser, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.ErrEmailAlreadyExists
	}

	// Generate tenant slug from company name
	tenantSlug := generateSlug(req.CompanyName)
	
	// Check if tenant slug already exists
	if uc.tenantRepo.ExistsBySlug(ctx, tenantSlug) {
		// Append timestamp to make it unique
		tenantSlug = fmt.Sprintf("%s-%d", tenantSlug, time.Now().Unix())
	}

	// Hash password
	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user *entity.User
	var tenant *entity.Tenant
	var ownerRole *entity.TenantRole
	var membership *entity.Membership

	// 1. Create User
	user = &entity.User{
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		IsActive:   true,
		IsVerified: false,
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Fetch the user to get the generated UUID
	if err := tx.First(user, user.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch created user: %w", err)
	}

	// 2. Create Tenant
	tenant = &entity.Tenant{
		Name:     req.CompanyName,
		Slug:     tenantSlug,
		Config:   "{}",
		IsActive: true,
	}

	if err := tx.Create(tenant).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 3. Copy system roles to the new tenant
	if err := uc.copySystemRolesToTenant(ctx, tx, tenant.ID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to copy system roles: %w", err)
	}

	// 4. Find the Tenant Owner role for this tenant
	if err := tx.Where("tenant_id = ? AND name = ?", tenant.ID, "Tenant Owner").First(&ownerRole).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to find Tenant Owner role: %w", err)
	}

	// 5. Create Membership (link user, tenant, and role)
	membership = &entity.Membership{
		UserID:   user.ID,
		TenantID: tenant.ID,
		RoleID:   ownerRole.ID,
	}

	if err := tx.Create(membership).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// 6. Create Kong Consumer using User UUID
	kongConsumerReq := &kong.ConsumerRequest{
		Username: user.UUID,
		CustomID: fmt.Sprintf("%d", user.ID),
		Tags:     []string{"portal-user", fmt.Sprintf("tenant:%s", tenantSlug)},
	}

	_, err = uc.kongClient.CreateConsumer(ctx, kongConsumerReq)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create Kong consumer: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		// If commit fails, try to cleanup Kong consumer
		_ = uc.kongClient.DeleteConsumer(ctx, user.UUID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return response
	return &model.RegisterWithTenantResponse{
		UserID:     user.ID,
		UserUUID:   user.UUID,
		Email:      user.Email,
		Name:       user.Name,
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		TenantSlug: tenant.Slug,
		Role:       ownerRole.Name,
	}, nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)
	
	// Replace spaces and special characters with hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1 // Remove other characters
	}, slug)
	
	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	
	return slug
}

// copySystemRolesToTenant copies the system role templates to a specific tenant
func (uc *RegistrationUseCase) copySystemRolesToTenant(ctx context.Context, tx *gorm.DB, tenantID int64) error {
	// Get system tenant
	var systemTenant entity.Tenant
	if err := tx.Where("slug = ?", "system").First(&systemTenant).Error; err != nil {
		return fmt.Errorf("system tenant not found. Please run the seeder first: %w", err)
	}

	// Get all system roles (templates)
	var systemRoles []entity.TenantRole
	if err := tx.Where("tenant_id = ?", systemTenant.ID).Find(&systemRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch system roles: %w", err)
	}

	if len(systemRoles) == 0 {
		return fmt.Errorf("no system roles found. Please run the seeder first")
	}

	// Copy each role and its permissions to the new tenant
	for _, systemRole := range systemRoles {
		// Create new role for the tenant
		newRole := entity.TenantRole{
			TenantID:    tenantID,
			Name:        systemRole.Name,
			Description: systemRole.Description,
		}

		if err := tx.Create(&newRole).Error; err != nil {
			return fmt.Errorf("failed to create role %s: %w", systemRole.Name, err)
		}

		// Get all permissions from the system role
		var systemPermissions []entity.Permission
		if err := tx.Where("role_id = ?", systemRole.ID).Find(&systemPermissions).Error; err != nil {
			return fmt.Errorf("failed to fetch permissions for role %s: %w", systemRole.Name, err)
		}

		// Copy permissions to the new role
		for _, systemPerm := range systemPermissions {
			newPerm := entity.Permission{
				RoleID:   newRole.ID,
				Resource: systemPerm.Resource,
				Action:   systemPerm.Action,
			}

			if err := tx.Create(&newPerm).Error; err != nil {
				return fmt.Errorf("failed to create permission: %w", err)
			}
		}
	}

	return nil
}
