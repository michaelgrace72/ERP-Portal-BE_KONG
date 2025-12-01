package model

// GetUserProfileResponse returns user's complete profile with all affiliations
type GetUserProfileResponse struct {
	UserID      int64                `json:"user_id"`
	UUID        string               `json:"uuid"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Email       string               `json:"email"`
	Avatar      string               `json:"avatar,omitempty"`
	IsActive    bool                 `json:"is_active"`
	IsVerified  bool                 `json:"is_verified"`
	Memberships []MembershipResponse `json:"memberships"`
}

// MembershipResponse represents a user's membership in a tenant
type MembershipResponse struct {
	MembershipID int64    `json:"membership_id"`
	TenantID     int64    `json:"tenant_id"`
	TenantName   string   `json:"tenant_name"`
	TenantSlug   string   `json:"tenant_slug"`
	RoleID       int64    `json:"role_id"`
	RoleName     string   `json:"role_name"`
	Permissions  []string `json:"permissions"`
}

// CreateUserResponse after creating a user
type CreateUserResponse struct {
	UserID   int64  `json:"user_id"`
	UUID     string `json:"uuid"`
	Code     string `json:"code"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// AssignUserToTenantRequest to add a user to a tenant with a role
type AssignUserToTenantRequest struct {
	UserID   int64 `json:"user_id" binding:"required"`
	TenantID int64 `json:"tenant_id" binding:"required"`
	RoleID   int64 `json:"role_id" binding:"required"`
}

// AssignUserToTenantResponse after assigning user to tenant
type AssignUserToTenantResponse struct {
	MembershipID int64  `json:"membership_id"`
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	TenantID     int64  `json:"tenant_id"`
	TenantName   string `json:"tenant_name"`
	RoleID       int64  `json:"role_id"`
	RoleName     string `json:"role_name"`
}

// RemoveUserFromTenantRequest to remove a user from a tenant
type RemoveUserFromTenantRequest struct {
	UserID   int64 `json:"user_id" binding:"required"`
	TenantID int64 `json:"tenant_id" binding:"required"`
}

// UpdateUserRoleRequest to change user's role in a tenant
type UpdateUserRoleRequest struct {
	MembershipID int64 `json:"membership_id" binding:"required"`
	RoleID       int64 `json:"role_id" binding:"required"`
}

// GetTenantMembersResponse lists all members of a tenant
type GetTenantMembersResponse struct {
	TenantID   int64                     `json:"tenant_id"`
	TenantName string                    `json:"tenant_name"`
	Members    []TenantMemberResponse    `json:"members"`
}

// TenantMemberResponse represents a member in a tenant
type TenantMemberResponse struct {
	MembershipID int64  `json:"membership_id"`
	UserID       int64  `json:"user_id"`
	UserUUID     string `json:"user_uuid"`
	UserCode     string `json:"user_code"`
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`
	RoleID       int64  `json:"role_id"`
	RoleName     string `json:"role_name"`
	JoinedAt     string `json:"joined_at"`
}

// UpdateTenantRequest for tenant owner to update tenant details
type UpdateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

// GetTenantRolesResponse lists all roles available in a tenant
type GetTenantRolesResponse struct {
	TenantID int64              `json:"tenant_id"`
	Roles    []TenantRoleDetail `json:"roles"`
}

// TenantRoleDetail represents detailed role information
type TenantRoleDetail struct {
	RoleID          int64    `json:"role_id"`
	RoleName        string   `json:"role_name"`
	Description     string   `json:"description"`
	PermissionCount int      `json:"permission_count"`
	Permissions     []string `json:"permissions,omitempty"`
}
