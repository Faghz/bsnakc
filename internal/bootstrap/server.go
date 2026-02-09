package bootstrap

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/adapters/http"
	authhttp "github.com/elzestia/go-boilerplate/internal/auth/adapters/http"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	usershttp "github.com/elzestia/go-boilerplate/internal/users/adapters/http"
	"github.com/labstack/echo/v4"
)

func initServer(cfg *config.Config, services *Services, logger ports.Logger) *echo.Echo {
	logger.Info(context.Background(), "Initializing HTTP server")

	// Create Echo server with OTel middleware if enabled
	server := http.NewServer(cfg, logger, cfg.OTel.Enabled)

	v1 := server.Group("/v1")
	{
		// Initialize handlers
		authhttp.NewAuthHandler(v1, services.AuthService, services.DiscordProvider, services.TokenService, services.Validator)
		usershttp.NewUserHandler(v1, services.UserService, services.TokenService, services.Validator)
	}

	return server
}
