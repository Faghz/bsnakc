package application

import (
	"github.com/elzestia/go-boilerplate/internal/identities/application/ports"
	"github.com/elzestia/go-boilerplate/internal/identities/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
)

// IdentityService implements the Service port and handles identity business logic.
// This is a rich service that orchestrates identity operations including password management.
type IdentityService struct {
	repo           domain.Repository
	passwordHasher domain.PasswordHasher
	logger         sharedports.Logger
}

var _ ports.Service = (*IdentityService)(nil)

// Service is the main service interface for the identities context.
// It's defined in ports package following standard port organization.
type Service = ports.Service

// NewIdentityService creates a new identity service.
func NewIdentityService(
	repo domain.Repository,
	passwordHasher domain.PasswordHasher,
	logger sharedports.Logger,
) *IdentityService {
	return &IdentityService{
		repo:           repo,
		passwordHasher: passwordHasher,
		logger:         logger.With(sharedports.String("domain", "identities")),
	}
}
