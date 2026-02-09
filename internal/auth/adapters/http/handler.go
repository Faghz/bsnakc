package http

import (
	"net/http"

	"github.com/elzestia/go-boilerplate/internal/adapters/http/middleware"
	"github.com/elzestia/go-boilerplate/internal/auth/adapters/oauth"
	"github.com/elzestia/go-boilerplate/internal/auth/application"
	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/validation"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService     application.Service
	tokenService    ports.TokenService
	discordProvider oauth.Provider
	validator       validation.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(route *echo.Group, authService application.Service, discordProvider oauth.Provider, tokenService ports.TokenService, validator validation.Validator) {
	handler := &AuthHandler{
		authService:     authService,
		tokenService:    tokenService,
		discordProvider: discordProvider,
		validator:       validator,
	}

	// Public routes - Auth
	authGroup := route.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.RefreshTokens)
		authGroup.POST("/logout", handler.Logout)
		authGroup.GET("/discord", handler.DiscordLogin)
		authGroup.GET("/callback/discord", handler.DiscordCallback)
	}

	// Protected routes - Auth
	authProtected := route.Group("/auth", middleware.AuthMiddleware(tokenService))
	{
		authProtected.POST("/logout-all", handler.LogoutAllDevices)
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	result, err := h.authService.RegisterLocal(c.Request().Context(), req.Email, req.Password, req.Name, req.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ToAuthResponse(result))
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	result, err := h.authService.LoginLocal(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToAuthResponse(result))
}

// DiscordLogin godoc
// @Summary Discord OAuth login
// @Description Redirect to Discord OAuth
// @Tags auth
// @Accept json
// @Produce json
// @Success 302
// @Router /v1/auth/discord [get]
func (h *AuthHandler) DiscordLogin(c echo.Context) error {
	authURL := h.discordProvider.GetAuthURL("random-state-value")
	return c.Redirect(http.StatusFound, authURL)
}

// DiscordCallback godoc
// @Summary Discord OAuth callback
// @Description Handle Discord OAuth callback
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Router /v1/auth/discord/callback [get]
func (h *AuthHandler) DiscordCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return errors.NewBadRequestError("missing authorization code")
	}

	// Register or login user via SSO
	result, err := h.authService.DiscordSSO(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToAuthResponse(result))
}

// RefreshTokens godoc
// @Summary Refresh access token
// @Description Exchange a refresh token for a new token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	result, err := h.authService.RefreshTokens(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToAuthResponse(result))
}

// Logout godoc
// @Summary Logout user
// @Description Revoke a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout request"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// LogoutAllDevices godoc
// @Summary Logout from all devices
// @Description Revoke all refresh tokens for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAllDevices(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	if err := h.authService.LogoutAllDevices(c.Request().Context(), userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
