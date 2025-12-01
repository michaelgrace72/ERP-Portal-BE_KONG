package model

// SessionValue represents the "fat" session object stored in Redis
type SessionValue struct {
	UserID      int64    `json:"uid"`
	UserUUID    string   `json:"uuid"`
	TenantID    int64    `json:"tid"`
	TenantSlug  string   `json:"tenant_slug"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Scope       string   `json:"scope"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
}

// PhantomLoginRequest represents the login credentials for phantom token
type PhantomLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TenantID *int64 `json:"tenant_id,omitempty"` // Optional: for multi-tenant users
}

// PhantomLoginResponse returns the reference token (phantom token)
type PhantomLoginResponse struct {
	AccessToken string          `json:"access_token"` // This is the reference token
	ExpiresIn   int             `json:"expires_in"`   // Seconds until expiration
	TokenType   string          `json:"token_type"`   // Always "Bearer"
	User        UserSessionInfo `json:"user"`
	Tenant      TenantInfo      `json:"tenant"`
}

// UserSessionInfo contains basic user info for the session
type UserSessionInfo struct {
	ID    int64  `json:"id"`
	UUID  string `json:"uuid"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// TenantSelectionResponse is returned when user has multiple tenants
type TenantSelectionResponse struct {
	Message      string              `json:"message"`
	Tenants      []TenantMembership  `json:"tenants"`
	RequiresChoice bool              `json:"requires_choice"`
}

// TenantMembership extends TenantInfo with role information
type TenantMembership struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

// SelectTenantRequest for multi-tenant user to select active tenant
type SelectTenantRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TenantID int64  `json:"tenant_id" binding:"required"`
}

// MembershipWithTenant represents a user's membership with tenant details
type MembershipWithTenant struct {
	TenantID    int64    `json:"tenant_id"`
	TenantName  string   `json:"tenant_name"`
	TenantSlug  string   `json:"tenant_slug"`
	RoleID      int64    `json:"role_id"`
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
}

// SessionContext represents the full context for a user session
type SessionContext struct {
	User        *UserSessionInfo
	Tenant      *TenantInfo
	Roles       []string
	Permissions []string
	Scope       string
}

// RefreshSessionRequest for refreshing the reference token
type RefreshSessionRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest for invalidating the session
type LogoutRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}
