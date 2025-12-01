package main

import (
	"context"
	"fmt"
	"log"

	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

// CopySystemRolesToTenant copies the system role templates from the "system" tenant to a specific tenant
// This should be called when a new tenant is created
func CopySystemRolesToTenant(ctx context.Context, db *gorm.DB, tenantID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
			// Check if role already exists for this tenant
			var existingRole entity.TenantRole
			err := tx.Where("tenant_id = ? AND name = ?", tenantID, systemRole.Name).First(&existingRole).Error
			if err == nil {
				fmt.Printf("⊘ Role '%s' already exists for tenant %d, skipping\n", systemRole.Name, tenantID)
				continue
			}

			// Create new role for the tenant
			newRole := entity.TenantRole{
				TenantID:    tenantID,
				Name:        systemRole.Name,
				Description: systemRole.Description,
			}

			if err := tx.Create(&newRole).Error; err != nil {
				return fmt.Errorf("failed to create role %s for tenant %d: %w", systemRole.Name, tenantID, err)
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
					return fmt.Errorf("failed to create permission for role %s: %w", newRole.Name, err)
				}
			}

			fmt.Printf("✓ Copied role '%s' with %d permissions to tenant %d\n", newRole.Name, len(systemPermissions), tenantID)
		}

		return nil
	})
}

// Example usage for testing
func exampleCopyToTenant(db *gorm.DB) {
	ctx := context.Background()
	
	// Example: Copy system roles to tenant with ID 1
	tenantID := int64(1)
	
	if err := CopySystemRolesToTenant(ctx, db, tenantID); err != nil {
		log.Fatalf("Failed to copy roles to tenant %d: %v", tenantID, err)
	}
	
	fmt.Printf("✓ Successfully copied all system roles to tenant %d\n", tenantID)
}
