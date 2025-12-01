# Database Seeders

This directory contains seeders for populating the database with initial roles and permissions.

## Quick Start

### 1. Seed System Roles (First Time Setup)

```bash
make seed
```

This creates:
- A "System" tenant (slug: `system`) for role templates
- 6 default roles with their permissions
- Total: 377 permissions across all roles

### 2. Copy Roles to a Tenant

```bash
make copy-roles TENANT_ID=1
```

Or manually:

```bash
go run cmd/copy-roles/main.go --tenant-id=1
```

## What Gets Created

### System Tenant
- **Name**: System
- **Slug**: system  
- **Domain**: system.internal
- **Purpose**: Stores role templates that can be copied to any tenant

### Default Roles

| Role | Permissions | Actions Allowed |
|------|-------------|----------------|
| Super Administrator | 78 | All (create, read, update, delete, list, export) |
| Tenant Owner | 78 | All (create, read, update, delete, list, export) |
| Administrator | 78 | All (create, read, update, delete, list, export) |
| Manager | 65 | create, read, update, list, export (no delete) |
| Editor | 39 | create, read, update |
| Viewer | 39 | read, list, export |

### Resources (Microservices)

Each role has permissions for the following resources:

- `portal` - Portal service
- `erp.accounts_receivable` - Accounts Receivable
- `erp.general_ledger` - General Ledger
- `erp.accounts_payable` - Accounts Payable
- `erp.manufacturing` - Manufacturing
- `erp.hrm` - Human Resource Management
- `erp.inventory` - Inventory Management
- `erp.fixed_asset` - Fixed Asset Management
- `erp.sales` - Sales
- `erp.cash_bank` - Cash & Bank Management
- `erp.purchasing` - Purchasing
- `erp.taxation` - Taxation
- `erp.scheduling` - Scheduling

### Actions

Standard CRUD + List/Export:
- `create` - Create new records
- `read` - Read/view records
- `update` - Update existing records
- `delete` - Delete records
- `list` - List/search records
- `export` - Export data

## Usage

### 1. Run the Seeder

First, seed the system role templates:

```bash
go run cmd/seed/main.go cmd/seed/copy_roles.go
```

This creates roles with `tenant_id = 0` as templates.

### 2. Copy Roles to a Tenant

When a new tenant is created, copy the system roles to that tenant.

You can either:

**Option A: Add to your registration flow**

In your user registration use case, after creating the tenant, call:

```go
import "go-gin-clean/cmd/seed"

// After creating tenant
if err := seed.CopySystemRolesToTenant(ctx, db, newTenant.ID); err != nil {
    return err
}
```

**Option B: Run manually via CLI**

Create a simple CLI tool:

```bash
# Add a command flag to copy_roles.go
go run cmd/seed/copy_roles.go --tenant-id=1
```

### 3. Verify in Database

```sql
-- Check system roles (templates)
SELECT * FROM roles WHERE tenant_id = 0;

-- Check roles for specific tenant
SELECT * FROM roles WHERE tenant_id = 1;

-- Check permissions for a role
SELECT p.* FROM permissions p
JOIN roles r ON p.role_id = r.id
WHERE r.tenant_id = 1 AND r.name = 'Manager';
```

## Integration with Registration Flow

The current registration flow in `user_usecase.go` creates a tenant and assigns the "owner" role. You should modify it to:

1. Create the tenant
2. Copy system roles to the new tenant using `CopySystemRolesToTenant()`
3. Find the "Tenant Owner" role for this tenant
4. Create membership with that role

Example:

```go
// After creating tenant
if err := CopySystemRolesToTenant(ctx, u.db, newTenant.ID); err != nil {
    return nil, fmt.Errorf("failed to copy system roles: %w", err)
}

// Find the Tenant Owner role
var ownerRole entity.TenantRole
if err := tx.Where("tenant_id = ? AND name = ?", newTenant.ID, "Tenant Owner").
    First(&ownerRole).Error; err != nil {
    return nil, fmt.Errorf("failed to find owner role: %w", err)
}

// Create membership
membership := &entity.Membership{
    UserID:   newUser.ID,
    TenantID: newTenant.ID,
    RoleID:   ownerRole.ID,
}
```

## Permission Checking

In your middleware or use cases, check permissions like this:

```go
func (u *UserUseCase) HasPermission(ctx context.Context, userID, tenantID int64, resource, action string) (bool, error) {
    // Get user's role in tenant
    var membership entity.Membership
    if err := u.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
        First(&membership).Error; err != nil {
        return false, err
    }

    // Check if role has permission
    var count int64
    if err := u.db.Model(&entity.Permission{}).
        Where("role_id = ? AND resource = ? AND action = ?", 
            membership.RoleID, resource, action).
        Count(&count).Error; err != nil {
        return false, err
    }

    return count > 0, nil
}
```

## Notes

- System roles (tenant_id = 0) should NOT be deleted as they serve as templates
- Each tenant gets isolated copies of roles and permissions
- Modifying a tenant's role doesn't affect other tenants
- You can customize permissions per tenant after copying from templates
