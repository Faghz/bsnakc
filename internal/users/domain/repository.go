package domain

import (
	"context"
)

// Repository defines the interface for user data operations (output port)
type Repository interface {
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUID(ctx context.Context, uid string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}
