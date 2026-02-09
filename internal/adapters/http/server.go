package http

import (
	"fmt"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/elzestia/go-boilerplate/internal/adapters/http/middleware"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// NewServer creates a new Echo server instance
// If otelEnabled is true, OpenTelemetry tracing middleware is added first
func NewServer(cfg *config.Config, logger ports.Logger, otelEnabled bool) *echo.Echo {
	e := echo.New()

	// Hide Echo banner
	e.HideBanner = true

	// Custom error handler
	e.HTTPErrorHandler = middleware.ErrorHandler(logger)

	// IMPORTANT: OTel middleware must be FIRST to propagate trace context
	// This ensures trace IDs are available in logger and all subsequent middleware
	if otelEnabled {
		e.Use(middleware.OTelMiddleware(cfg.OTel.ServiceName))
	}

	// Middleware (order matters!)
	e.Use(middleware.Logger(logger)) // Uses trace context if OTel is enabled
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.CORS.AllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	// Scalar documentation
	e.GET("/docs", scalarHandler)

	return e
}

func scalarHandler(e echo.Context) error {
	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		// SpecURL: "https://generator3.swagger.io/openapi.json",// allow external URL or local path file
		SpecURL: "./docs/swagger.json",
		CustomOptions: scalar.CustomOptions{
			PageTitle: "HTTP Boilerplate API Reference",
		},
		DarkMode: true,
	})

	if err != nil {
		fmt.Printf("%v", err)
	}

	return e.HTML(200, htmlContent)
}
