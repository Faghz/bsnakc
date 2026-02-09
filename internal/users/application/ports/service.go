package ports

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/users/domain"
)

// Service is the input port for user operations (consumed by HTTP adapters)
type Service interface {
	Create(ctx context.Context, email, name, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUID(ctx context.Context, uid string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}
