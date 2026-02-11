package bootstrap

import (
	"context"
	"fmt"

	authadapters "github.com/elzestia/go-boilerplate/internal/auth/adapters"
	authoauth "github.com/elzestia/go-boilerplate/internal/auth/adapters/oauth"
	authpaseto "github.com/elzestia/go-boilerplate/internal/auth/adapters/paseto"
	authpostgres "github.com/elzestia/go-boilerplate/internal/auth/adapters/postgres"
	authredis "github.com/elzestia/go-boilerplate/internal/auth/adapters/redis"
	authapp "github.com/elzestia/go-boilerplate/internal/auth/application"
	authports "github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	identitiesapp "github.com/elzestia/go-boilerplate/internal/identities/application"
	identitydomain "github.com/elzestia/go-boilerplate/internal/identities/domain"
	productsapp "github.com/elzestia/go-boilerplate/internal/products/application"
	productsports "github.com/elzestia/go-boilerplate/internal/products/application"
	"github.com/elzestia/go-boilerplate/internal/shared/adapters/postgres"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/elzestia/go-boilerplate/internal/shared/crypto"
	"github.com/elzestia/go-boilerplate/internal/shared/validation"
	usersapp "github.com/elzestia/go-boilerplate/internal/users/application"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Services struct {
	UserService        usersapp.Service
	AuthService        authapp.Service
	TokenService       authports.TokenService
	RefreshTokenRepo   authdomain.RefreshTokenRepository
	DiscordProvider    authoauth.Provider
	Validator          validation.Validator
	ProductTypeService productsports.ProductTypeService
	ProductService     productsports.ProductService
}

func initServices(db *sqlx.DB, redis *redis.Client, repos *Repositories, cfg *config.Config, logger ports.Logger) *Services {

	// Infrastructure adapters
	txManager := postgres.NewTransactionManager(db, logger)
	var passwordHasher identitydomain.PasswordHasher = crypto.NewPasswordHasher()

	// Initialize refresh token repositories
	// Postgres is the source of truth, Redis is the cache layer
	postgresRefreshTokenRepo := authpostgres.NewRefreshTokenRepository(db, logger)
	refreshTokenRepo := authredis.NewCachedRefreshTokenRepository(postgresRefreshTokenRepo, redis, logger)

	// Initialize PASETO token service
	tokenService, err := authpaseto.NewTokenService(
		cfg.PASETO.SymmetricKey,
		cfg.PASETO.AccessTokenDuration,
		cfg.PASETO.RefreshTokenDuration,
		refreshTokenRepo,
	)
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize token service", ports.Error(err))
		panic(fmt.Sprintf("failed to initialize token service: %v", err))
	}

	discordProvider := authoauth.NewDiscordProvider(cfg.OAuth.Discord)

	// Initialize validator
	validator := validation.NewValidator()

	// Application services
	userSvc := usersapp.NewUserService(repos.UserRepo, logger)
	userServiceAdapter := authadapters.NewUserServiceAdapter(userSvc)

	// Create identity service
	identitySvc := identitiesapp.NewIdentityService(
		repos.IdentityRepo,
		passwordHasher,
		logger,
	)
	identityServiceAdapter := authadapters.NewIdentityServiceAdapter(identitySvc)

	// Initialize auth service with transaction management
	authSvc := authapp.NewAuthService(
		txManager,              // Transaction management via port
		userServiceAdapter,     // User service port (ACL)
		identityServiceAdapter, // Identity service port (ACL)
		tokenService,           // Token service port
		refreshTokenRepo,       // Refresh token repository
		logger,                 // Logger port
		discordProvider,        // OAuth provider
	)

	// Product services
	productTypeSvc := productsapp.NewProductTypeService(repos.ProductTypeRepo, logger)
	productSvc := productsapp.NewProductService(repos.ProductRepo, repos.ProductTypeRepo, logger)

	return &Services{
		UserService:        userSvc,
		AuthService:        authSvc,
		TokenService:       tokenService,
		RefreshTokenRepo:   refreshTokenRepo,
		DiscordProvider:    discordProvider,
		Validator:          validator,
		ProductTypeService: productTypeSvc,
		ProductService:     productSvc,
	}
}
