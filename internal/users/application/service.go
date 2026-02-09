package application

import (
	sharedPorts "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/users/application/ports"
	"github.com/elzestia/go-boilerplate/internal/users/domain"
)

// UserService implements the Service port and handles user business logic
type UserService struct {
	repo   domain.Repository
	logger sharedPorts.Logger
}

// Ensure UserService implements Service interface
var _ ports.Service = (*UserService)(nil)

type Service = ports.Service

// NewUserService creates a new user service
func NewUserService(repo domain.Repository, logger sharedPorts.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger.With(sharedPorts.String("domain", "users")),
	}
}
