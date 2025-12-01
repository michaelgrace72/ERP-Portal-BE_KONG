package entity

import "time"

// Tenant represents a company/organization in the multi-tenant system
type Tenant struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Slug      string    `gorm:"type:varchar(100);uniqueIndex;not null"` // URL-friendly identifier
	Domain    string    `gorm:"type:varchar(100);uniqueIndex"`          // Optional custom domain
	Config    string    `gorm:"type:jsonb;default:'{}'"`                // Tenant-specific configurations
	IsActive  bool      `gorm:"default:true;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time
	IsDeleted bool `gorm:"default:false;not null"`

}

func (Tenant) TableName() string {
	return "tenants"
}

// TenantRole represents a role within a specific tenant
type TenantRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID    int64     `gorm:"not null;index:idx_tenant_role"`
	Name        string    `gorm:"type:varchar(50);not null;index:idx_tenant_role"`
	Description string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

func (TenantRole) TableName() string {
	return "roles"
}

// Membership represents the relationship between users, tenants, and roles
type Membership struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    int64     `gorm:"not null;index:idx_user_tenant"`
	TenantID  int64     `gorm:"not null;index:idx_user_tenant"`
	RoleID    int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time

	// Relations
	User   *User       `gorm:"foreignKey:UserID;references:ID"`
	Tenant *Tenant     `gorm:"foreignKey:TenantID;references:ID"`
	Role   *TenantRole `gorm:"foreignKey:RoleID;references:ID"`
}

func (Membership) TableName() string {
	return "memberships"
}
