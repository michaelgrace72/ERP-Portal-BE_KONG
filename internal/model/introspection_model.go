package model

// IntrospectionRequest is the request from Kong to validate a reference token
type IntrospectionRequest struct {
	Token string `json:"token" binding:"required"`
}

// IntrospectionResponse is returned to Kong with session context
type IntrospectionResponse struct {
	Active      bool     `json:"active"`
	Sub         string   `json:"sub,omitempty"`          // User ID
	TenantID    int64    `json:"tenant_id,omitempty"`    // Current tenant ID
	UserID      int64    `json:"user_id,omitempty"`      // User ID
	RoleID      int64    `json:"role_id,omitempty"`      // Current role ID
	RoleName    string   `json:"role_name,omitempty"`    // Role name
	Permissions []string `json:"permissions,omitempty"`  // Permissions array
	Exp         int64    `json:"exp,omitempty"`          // Expiration timestamp
}

// IntrospectionHeaders are the headers Kong should inject into upstream requests
type IntrospectionHeaders struct {
	XTenantID    string `header:"X-Tenant-ID"`
	XUserID      string `header:"X-User-ID"`
	XRoleID      string `header:"X-Role-ID"`
	XPermissions string `header:"X-Permissions"` // Comma-separated
}
