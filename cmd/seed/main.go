package main

import (
	"context"
	"fmt"
	"log"

	"go-gin-clean/internal/entity"
	"go-gin-clean/pkg/config"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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

	// Run seeders
	if err := SeedSystemRolesAndPermissions(ctx, db); err != nil {
		log.Fatalf("Failed to seed roles and permissions: %v", err)
	}

	fmt.Println("✓ Seeding completed successfully!")
}

// SeedSystemRolesAndPermissions seeds default roles and permissions
// Creates a special "system" tenant (ID = 0 or special slug) for role templates
func SeedSystemRolesAndPermissions(ctx context.Context, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Create or get system tenant for role templates
		var systemTenant entity.Tenant
		if err := tx.Where("slug = ?", "system").First(&systemTenant).Error; err != nil {
			// Create system tenant
			systemTenant = entity.Tenant{
				Name:     "System",
				Slug:     "system",
				Config:   "{}",
				IsActive: true,
			}
			if err := tx.Create(&systemTenant).Error; err != nil {
				return fmt.Errorf("failed to create system tenant: %w", err)
			}
			fmt.Println("✓ Created system tenant for role templates")
		}

		// Define system roles (these will be templates)
		systemRoles := []struct {
			Name        string
			Description string
			Permissions []Permission
		}{
		{
			Name:        "Super Administrator",
			Description: "Full system access across all tenants",
			Permissions: getAllPermissions(),
		},
		{
			Name:        "Tenant Owner",
			Description: "Full access within their tenant",
			Permissions: getAllPermissions(),
		},
		{
			Name:        "Administrator",
			Description: "Administrative access within tenant",
			Permissions: getAllPermissions(),
		},
		{
			Name:        "Manager",
			Description: "Management-level access with limited admin capabilities",
			Permissions: getManagerPermissions(),
		},
		{
			Name:        "Editor",
			Description: "Can create and edit content",
			Permissions: getEditorPermissions(),
		},
		{
			Name:        "Viewer",
			Description: "Read-only access",
			Permissions: getViewerPermissions(),
		},
		}

		// Use the system tenant ID for all role templates
		systemTenantID := systemTenant.ID

		for _, roleData := range systemRoles {
			// Check if role already exists
			var existingRole entity.TenantRole
			if err := tx.Where("tenant_id = ? AND name = ?", systemTenantID, roleData.Name).First(&existingRole).Error; err == nil {
				fmt.Printf("⊘ Role '%s' already exists, skipping\n", roleData.Name)
				continue
			}

			// Create role
			role := entity.TenantRole{
				TenantID:    systemTenantID,
				Name:        roleData.Name,
				Description: roleData.Description,
			}

			if err := tx.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", roleData.Name, err)
			}

			// Create permissions for this role
			for _, perm := range roleData.Permissions {
				permission := entity.Permission{
					RoleID:   role.ID,
					Resource: perm.Resource,
					Action:   perm.Action,
				}

				if err := tx.Create(&permission).Error; err != nil {
					return fmt.Errorf("failed to create permission for role %s: %w", roleData.Name, err)
				}
			}

			fmt.Printf("✓ Created role '%s' with %d permissions\n", roleData.Name, len(roleData.Permissions))
		}

		return nil
	})
}

// Permission represents a permission structure for seeding
type Permission struct {
	Resource string
	Action   string
}

// getAllPermissions returns all possible permissions (for admin roles)
func getAllPermissions() []Permission {
	resources := []string{
		"portal",
		"erp.accounts_receivable",
		"erp.general_ledger",
		"erp.accounts_payable",
		"erp.manufacturing",
		"erp.hrm",
		"erp.inventory",
		"erp.fixed_asset",
		"erp.sales",
		"erp.cash_bank",
		"erp.purchasing",
		"erp.taxation",
		"erp.scheduling",
	}

	actions := []string{"create", "read", "update", "delete", "list", "export"}

	var permissions []Permission
	for _, resource := range resources {
		for _, action := range actions {
			permissions = append(permissions, Permission{
				Resource: resource,
				Action:   action,
			})
		}
	}

	return permissions
}

// getManagerPermissions returns permissions for Manager role (all except delete)
func getManagerPermissions() []Permission {
	resources := []string{
		"portal",
		"erp.accounts_receivable",
		"erp.general_ledger",
		"erp.accounts_payable",
		"erp.manufacturing",
		"erp.hrm",
		"erp.inventory",
		"erp.fixed_asset",
		"erp.sales",
		"erp.cash_bank",
		"erp.purchasing",
		"erp.taxation",
		"erp.scheduling",
	}

	actions := []string{"create", "read", "update", "list", "export"}

	var permissions []Permission
	for _, resource := range resources {
		for _, action := range actions {
			permissions = append(permissions, Permission{
				Resource: resource,
				Action:   action,
			})
		}
	}

	return permissions
}

// getEditorPermissions returns permissions for Editor role (create/read/update)
func getEditorPermissions() []Permission {
	resources := []string{
		"portal",
		"erp.accounts_receivable",
		"erp.general_ledger",
		"erp.accounts_payable",
		"erp.manufacturing",
		"erp.hrm",
		"erp.inventory",
		"erp.fixed_asset",
		"erp.sales",
		"erp.cash_bank",
		"erp.purchasing",
		"erp.taxation",
		"erp.scheduling",
	}

	actions := []string{"create", "read", "update"}

	var permissions []Permission
	for _, resource := range resources {
		for _, action := range actions {
			permissions = append(permissions, Permission{
				Resource: resource,
				Action:   action,
			})
		}
	}

	return permissions
}

// getViewerPermissions returns permissions for Viewer role (read/list/export only)
func getViewerPermissions() []Permission {
	resources := []string{
		"portal",
		"erp.accounts_receivable",
		"erp.general_ledger",
		"erp.accounts_payable",
		"erp.manufacturing",
		"erp.hrm",
		"erp.inventory",
		"erp.fixed_asset",
		"erp.sales",
		"erp.cash_bank",
		"erp.purchasing",
		"erp.taxation",
		"erp.scheduling",
	}

	actions := []string{"read", "list", "export"}

	var permissions []Permission
	for _, resource := range resources {
		for _, action := range actions {
			permissions = append(permissions, Permission{
				Resource: resource,
				Action:   action,
			})
		}
	}

	return permissions
}
