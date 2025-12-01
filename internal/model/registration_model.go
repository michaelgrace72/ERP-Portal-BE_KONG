package model

// RegisterWithTenantRequest represents the request to register a new user with a tenant
type RegisterWithTenantRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name" binding:"required"`
	CompanyName string `json:"company_name" binding:"required"`
}

// RegisterWithTenantResponse represents the response after successful registration
type RegisterWithTenantResponse struct {
	UserID     int64  `json:"user_id"`
	UserUUID   string `json:"user_uuid"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	TenantID   int64  `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	Role       string `json:"role"`
}

// TenantInfo represents basic tenant information
type TenantInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Domain   string `json:"domain,omitempty"`
	IsActive bool   `json:"is_active"`
}

// MembershipInfo represents user's membership in a tenant
type MembershipInfo struct {
	TenantID   int64  `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	RoleID     int64  `json:"role_id"`
	RoleName   string `json:"role_name"`
}
