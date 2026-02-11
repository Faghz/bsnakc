package http

import (
	"net/http"

	"github.com/elzestia/go-boilerplate/internal/adapters/http/middleware"
	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/validation"
	"github.com/elzestia/go-boilerplate/internal/users/application"
	"github.com/labstack/echo/v4"
)

// UserHandler handles user endpoints
type UserHandler struct {
	userService application.Service
	validator   validation.Validator
}

// NewUserHandler creates a new user handler
func NewUserHandler(route *echo.Group, userService application.Service, tokenService ports.TokenService, validator validation.Validator) {
	handler := &UserHandler{
		userService: userService,
		validator:   validator,
	}

	usersGroup := route.Group("/users", middleware.AuthMiddleware(tokenService))
	{
		usersGroup.GET("/me", handler.GetMe)
	}
}

// GetMe godoc
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserResponse "Current user details"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "User not found"
// @Router /v1/users/me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	u, err := h.userService.GetByUID(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToUserResponse(u))
}
