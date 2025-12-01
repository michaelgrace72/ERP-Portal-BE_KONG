package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go-gin-clean/internal/entity"
	"go-gin-clean/pkg/config"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Parse command line flags
	tenantID := flag.Int64("tenant-id", 0, "Tenant ID to copy roles to")
	flag.Parse()

	if *tenantID == 0 {
		fmt.Println("Usage: go run cmd/copy-roles/main.go --tenant-id=<id>")
		fmt.Println("Example: go run cmd/copy-roles/main.go --tenant-id=1")
		os.Exit(1)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// Copy roles
	fmt.Printf("Copying system roles to tenant ID: %d\n", *tenantID)
	fmt.Println("----------------------------------------")

	if err := CopySystemRolesToTenant(ctx, db, *tenantID); err != nil {
		log.Fatalf("Failed to copy roles: %v", err)
	}

	fmt.Printf("\n✓ Successfully copied all system roles to tenant %d\n", *tenantID)
}

func CopySystemRolesToTenant(ctx context.Context, db *gorm.DB, tenantID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Get system tenant
		var systemTenant entity.Tenant
		if err := tx.Where("slug = ?", "system").First(&systemTenant).Error; err != nil {
			return fmt.Errorf("system tenant not found. Please run the seeder first: %w", err)
		}

		// Verify target tenant exists
		var targetTenant entity.Tenant
		if err := tx.First(&targetTenant, tenantID).Error; err != nil {
			return fmt.Errorf("target tenant %d not found: %w", tenantID, err)
		}

		fmt.Printf("Target tenant: %s (ID: %d)\n\n", targetTenant.Name, targetTenant.ID)

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
				fmt.Printf("⊘ Role '%s' already exists, skipping\n", systemRole.Name)
				continue
			}

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

			fmt.Printf("✓ Copied role '%s' with %d permissions\n", newRole.Name, len(systemPermissions))
		}

		return nil
	})
}
