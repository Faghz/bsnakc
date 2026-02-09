package ports

import (
	"context"
)

// UserService is the port that Auth uses to interact with the User context
// This is an Anti-Corruption Layer interface - Auth shouldn't depend on User domain
type UserService interface {
	// Read operations
	GetUserByEmail(ctx context.Context, email string) (*UserInfo, error)
	GetUserByID(ctx context.Context, id int64) (*UserInfo, error)
	GetByUID(ctx context.Context, uid string) (*UserInfo, error)
	GetByUsername(ctx context.Context, username string) (*UserInfo, error)

	// Write operations
	CreateUser(ctx context.Context, email, name, username string) (*UserInfo, error)
}
