package application

import (
	"github.com/elzestia/go-boilerplate/internal/auth/adapters/oauth"
	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
)

// AuthService implements the Auth input port using transactional operations
type AuthService struct {
	// Transaction management via port
	txManager sharedports.TransactionManager

	// Repository dependencies via domain interfaces
	refreshTokenRepo authdomain.RefreshTokenRepository

	// Port-based dependencies (ACL adapters)
	userService     ports.UserService     // Anti-Corruption Layer for user operations
	identityService ports.IdentityService // Anti-Corruption Layer for identity operations
	tokenService    ports.TokenService    // PASETO adapter
	logger          sharedports.Logger    // Logger port

	// OAuth provider adapters
	discordProvider oauth.Provider
}

// Ensure AuthService implements Service interface
var _ ports.Service = (*AuthService)(nil)

type Service = ports.Service

// NewAuthService creates a new auth service
func NewAuthService(
	txManager sharedports.TransactionManager,
	userService ports.UserService,
	identityService ports.IdentityService,
	tokenService ports.TokenService,
	refreshTokenRepo authdomain.RefreshTokenRepository,
	logger sharedports.Logger,
	discordProvider oauth.Provider,
) *AuthService {
	return &AuthService{
		txManager:        txManager,
		userService:      userService,
		identityService:  identityService,
		tokenService:     tokenService,
		refreshTokenRepo: refreshTokenRepo,
		logger:           logger.With(sharedports.String("domain", "auth")),
		discordProvider:  discordProvider,
	}
}
