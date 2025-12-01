package infrastructure

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/kong"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/repository"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/pkg/config"

	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler         http.UserHandler
	OauthHandler        http.OAuthHandler
	RegistrationHandler http.RegistrationHandler
	JWTService          security.JWTService
	OAuthService        security.OAuthService
}

func NewContainer(db *gorm.DB, ch *amqp091.Channel, cfg *config.Config) *Container {
	// Init repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	tenantRoleRepo := repository.NewTenantRoleRepository(db)
	membershipRepo := repository.NewMembershipRepository(db)

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

	// init message publisher
	userPublisher := messaging.NewUserPublisher(ch)

	// Init use cases
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, jwtService, passwordService, oauthService, aesService, cloudinaryService, localStorageService, redisService, userPublisher)
	registrationUseCase := usecase.NewRegistrationUseCase(db, userRepo, tenantRepo, tenantRoleRepo, membershipRepo, passwordService, kongClient)

	// Init handlers
	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)
	registrationHandler := http.NewRegistrationHandler(registrationUseCase)

	return &Container{
		UserHandler:         *userHandler,
		OauthHandler:        *oauthHandler,
		RegistrationHandler: *registrationHandler,
		JWTService:          *jwtService,
		OAuthService:        *oauthService,
	}
}
