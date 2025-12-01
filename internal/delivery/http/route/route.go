package route

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/delivery/http/middleware"
	"go-gin-clean/internal/gateway/security"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userHandler *http.UserHandler,
	oauthHandler *http.OAuthHandler,
	registrationHandler *http.RegistrationHandler,
	authHandler *http.AuthHandler,
	userManagementHandler *http.UserManagementHandler,
	jwtService *security.JWTService,
) {
	// Setup handlers
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup CORS
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (auth)
		auth := api.Group("/auth")
		{
			// Legacy JWT-based login (keep for backward compatibility)
			auth.POST("/login", userHandler.Login)
			
			// New phantom token authentication endpoints
			auth.POST("/phantom-login", authHandler.Login)
			auth.POST("/select-tenant", authHandler.SelectTenant)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshSession)
			auth.GET("/session", authHandler.GetSession)
			
			// Registration
			auth.POST("/register", registrationHandler.RegisterWithTenant)
			
			// Other auth endpoints
			auth.POST("/refresh-token", userHandler.RefreshToken)
			auth.POST("/verify-email", userHandler.VerifyEmail)
			auth.POST("/reset-password", userHandler.ResetPassword)
			auth.POST("/send-reset-password", userHandler.SendResetPassword)
			auth.POST("/resend-verification", userHandler.SendVerifyEmail)
		}

		oauth := auth.Group("/oauth2")
		{
			oauth.POST("/url", oauthHandler.GetLoginURL)
			oauth.GET("/:provider/callback", oauthHandler.CallBack)
		}

		profile := api.Group("/profile")
		profile.Use(authMiddleware.RequireAuth())
		{
			profile.GET("", userHandler.Profile)
			profile.PUT("", userHandler.UpdateProfile)
			profile.PUT("/change-password", userHandler.ChangePassword)
			profile.POST("/logout", userHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			// User management - Get own profile
			users.GET("/me", userManagementHandler.GetMyProfile)
			
			// Admin: Create user (for invitation)
			users.POST("", userManagementHandler.CreateUser)
			
			// Legacy user endpoints
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:code", userHandler.GetUserByCode)
			users.PUT("/:code", userHandler.UpdateUser)
			users.PUT("/:code/change-status", userHandler.ChangeStatus)
			users.DELETE("/:code", userHandler.DeleteUser)
		}

		// Membership management
		memberships := api.Group("/memberships")
		memberships.Use(authMiddleware.RequireAuth())
		{
			memberships.POST("", userManagementHandler.AssignUserToTenant)
			memberships.DELETE("", userManagementHandler.RemoveUserFromTenant)
			memberships.PUT("/:id/role", userManagementHandler.UpdateUserRole)
		}

		// Tenant management
		tenants := api.Group("/tenants")
		tenants.Use(authMiddleware.RequireAuth())
		{
			tenants.GET("/:id/members", userManagementHandler.GetTenantMembers)
			tenants.GET("/:id/roles", userManagementHandler.GetTenantRoles)
			tenants.PUT("/:id", userManagementHandler.UpdateTenant)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})
}
