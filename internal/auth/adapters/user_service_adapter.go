package adapters

import (
	"context"
	"errors"
	"strings"

	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	sharederrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
	usersapp "github.com/elzestia/go-boilerplate/internal/users/application"
	usersdomain "github.com/elzestia/go-boilerplate/internal/users/domain"
)

// UserServiceAdapter is an Anti-Corruption Layer that adapts the Users context
// for use by the Auth context. It prevents Auth from depending on User domain entities.
type UserServiceAdapter struct {
	userService usersapp.Service
}

// NewUserServiceAdapter creates a new user service adapter
func NewUserServiceAdapter(userService usersapp.Service) *UserServiceAdapter {
	return &UserServiceAdapter{
		userService: userService,
	}
}

// GetUserByEmail retrieves a user by email and converts to UserInfo
func (a *UserServiceAdapter) GetUserByEmail(ctx context.Context, email string) (*ports.UserInfo, error) {
	user, err := a.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, a.translateError(err)
	}

	return &ports.UserInfo{
		ID:        user.ID(),
		UID:       user.UID(),
		Email:     user.Email().String(),
		Name:      user.Name(),
		Username:  user.Username().String(),
		AvatarURL: user.AvatarURL(),
	}, nil
}

// GetUserByID retrieves a user by ID and converts to UserInfo
func (a *UserServiceAdapter) GetUserByID(ctx context.Context, id int64) (*ports.UserInfo, error) {
	user, err := a.userService.GetByID(ctx, id)
	if err != nil {
		return nil, a.translateError(err)
	}

	return &ports.UserInfo{
		ID:        user.ID(),
		UID:       user.UID(),
		Email:     user.Email().String(),
		Name:      user.Name(),
		Username:  user.Username().String(),
		AvatarURL: user.AvatarURL(),
	}, nil
}

// CreateUser creates a new user and converts to UserInfo
func (a *UserServiceAdapter) CreateUser(ctx context.Context, email, name, username string) (*ports.UserInfo, error) {
	user, err := a.userService.Create(ctx, email, name, username)
	if err != nil {
		return nil, a.translateError(err)
	}

	return &ports.UserInfo{
		ID:        user.ID(),
		UID:       user.UID(),
		Email:     user.Email().String(),
		Name:      user.Name(),
		Username:  user.Username().String(),
		AvatarURL: user.AvatarURL(),
	}, nil
}

func (a *UserServiceAdapter) GetByUID(ctx context.Context, uid string) (*ports.UserInfo, error) {
	user, err := a.userService.GetByUID(ctx, uid)
	if err != nil {
		return nil, a.translateError(err)
	}

	return &ports.UserInfo{
		ID:        user.ID(),
		UID:       user.UID(),
		Email:     user.Email().String(),
		Name:      user.Name(),
		Username:  user.Username().String(),
		AvatarURL: user.AvatarURL(),
	}, nil
}

// GetByUsername retrieves a user by username and converts to UserInfo
func (a *UserServiceAdapter) GetByUsername(ctx context.Context, username string) (*ports.UserInfo, error) {
	user, err := a.userService.GetByUsername(ctx, username)
	if err != nil {
		return nil, a.translateError(err)
	}

	return &ports.UserInfo{
		ID:        user.ID(),
		UID:       user.UID(),
		Email:     user.Email().String(),
		Name:      user.Name(),
		Username:  user.Username().String(),
		AvatarURL: user.AvatarURL(),
	}, nil
}

// translateError maps Users domain errors to Auth domain errors.
// This prevents error leakage across domain boundaries.
//
// Error translation rules:
// - usersdomain.ErrUserNotFound → authdomain.ErrUserNotFound
// - usersdomain.ErrUserAlreadyExists → authdomain.ErrUserAlreadyExists
// - Validation errors → Pass through (shared errors)
// - Not found errors → authdomain.ErrUserNotFound
// - Others → Pass through
func (a *UserServiceAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific User domain errors
	switch {
	case errors.Is(err, usersdomain.ErrUserNotFound):
		return authdomain.ErrUserNotFound

	case errors.Is(err, usersdomain.ErrEmailRegistered):
		return authdomain.ErrEmailAlreadyUsed

	case errors.Is(err, usersdomain.ErrUsernameTaken):
		return authdomain.ErrUsernameAlreadyUsed

	// Check shared error types by category
	case sharederrors.IsNotFoundError(err):
		// Check if it's a user-related not found error
		if strings.Contains(strings.ToLower(err.Error()), "user") {
			return authdomain.ErrUserNotFound
		}
		return err // Pass through other not found errors

	default:
		// Pass through all other errors (validation, infrastructure, internal, etc.)
		return err
	}
}

// Ensure UserServiceAdapter implements ports.UserService
var _ ports.UserService = (*UserServiceAdapter)(nil)
