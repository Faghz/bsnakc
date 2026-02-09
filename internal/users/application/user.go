package application

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/users/domain"
)

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByUID retrieves a user by UID
func (s *UserService) GetByUID(ctx context.Context, uid string) (*domain.User, error) {
	return s.repo.FindByUID(ctx, uid)
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.FindByUsername(ctx, username)
}

// Create creates a new user with validation
func (s *UserService) Create(ctx context.Context, email, name, username string) (*domain.User, error) {
	s.logger.Debug(ctx, "Creating user", ports.String("username", username))

	// Check if user already exists by email
	existing, err := s.repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		s.logger.Warn(ctx, "User creation failed: email already exists")
		return nil, domain.ErrEmailRegistered
	}

	// Check if username already exists
	existingByUsername, err := s.repo.FindByUsername(ctx, username)
	if err == nil && existingByUsername != nil {
		s.logger.Warn(ctx, "User creation failed: username already exists", ports.String("username", username))
		return nil, domain.ErrUsernameTaken
	}

	// Create user (validation happens in domain factory)
	newUser, err := domain.NewUser(email, name, username)
	if err != nil {
		s.logger.Error(ctx, "User validation failed", ports.Error(err))
		return nil, err
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		s.logger.Error(ctx, "Failed to create user in database", ports.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Debug(ctx, "User created successfully", ports.Int64("user_id", newUser.ID()), ports.String("username", username))
	return newUser, nil
}

// Update updates an existing user
func (s *UserService) Update(ctx context.Context, u *domain.User) error {
	return s.repo.Update(ctx, u)
}

// Delete deletes a user
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
