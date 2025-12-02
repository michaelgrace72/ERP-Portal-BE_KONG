package infrastructure

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/kong"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/gateway/session"
	"go-gin-clean/internal/repository"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/pkg/config"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler             http.UserHandler
	OauthHandler            http.OAuthHandler
	RegistrationHandler     http.RegistrationHandler
	AuthHandler             http.AuthHandler
	UserManagementHandler   http.UserManagementHandler
	IntrospectionHandler    http.IntrospectionHandler
	JWTService              security.JWTService
	OAuthService            security.OAuthService
	SessionService          *session.SessionService
}

func NewContainer(db *gorm.DB, ch *amqp091.Channel, cfg *config.Config) *Container {
	// Init repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	tenantRoleRepo := repository.NewTenantRoleRepository(db)
	membershipRepo := repository.NewMembershipRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// Init services
	jwtService := security.NewJWTService(&cfg.JWT)
	passwordService := security.NewBcryptService()
	oauthService := security.NewOAuthService(&cfg.OAuth)
	aesService := security.NewAESService(&cfg.AES)
	cloudinaryService := media.NewCloudinaryService(&cfg.Cloudinary)
	localStorageService := media.NewLocalStorageService("")
	redisService := cache.NewRedisService(&cfg.Redis)
	
	// Init Kong client
	kongClient := kong.NewKongAdminClient(cfg.Kong.AdminURL, cfg.Kong.Timeout)
	
	// Init session service
	sessionTTL := 30 * time.Minute // 30 minutes session expiration
	sessionService := session.NewSessionService(redisService.GetClient(), sessionTTL)

	// init message publisher
	userPublisher := messaging.NewUserPublisher(ch)

	// Init use cases
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, jwtService, passwordService, oauthService, aesService, cloudinaryService, localStorageService, redisService, userPublisher)
	registrationUseCase := usecase.NewRegistrationUseCase(db, userRepo, tenantRepo, tenantRoleRepo, membershipRepo, passwordService, kongClient)
	authUseCase := usecase.NewAuthUseCase(userRepo, membershipRepo, tenantRepo, tenantRoleRepo, permissionRepo, passwordService, sessionService, sessionTTL)
	userManagementUseCase := usecase.NewUserManagementUseCase(db, userRepo, tenantRepo, tenantRoleRepo, membershipRepo, permissionRepo, passwordService)
	introspectionUseCase := usecase.NewIntrospectionUseCase(sessionService)

	// Init handlers
	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)
	registrationHandler := http.NewRegistrationHandler(registrationUseCase)
	authHandler := http.NewAuthHandler(authUseCase)
	userManagementHandler := http.NewUserManagementHandler(userManagementUseCase)
	introspectionHandler := http.NewIntrospectionHandler(introspectionUseCase)

	return &Container{
		UserHandler:           *userHandler,
		OauthHandler:          *oauthHandler,
		RegistrationHandler:   *registrationHandler,
		AuthHandler:           *authHandler,
		UserManagementHandler: *userManagementHandler,
		IntrospectionHandler:  *introspectionHandler,
		JWTService:            *jwtService,
		OAuthService:          *oauthService,
		SessionService:        sessionService,
	}
}
