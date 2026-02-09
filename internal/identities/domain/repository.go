package domain

import (
	"context"
)

// Repository defines the interface for identity data operations (output port)
type Repository interface {
	Create(ctx context.Context, identity *Identity) error
	FindByProviderAndUserID(ctx context.Context, provider Provider, userID int64) (*Identity, error)
	FindByProviderAndProviderID(ctx context.Context, provider Provider, providerID string) (*Identity, error)
	Update(ctx context.Context, identity *Identity) error
	Delete(ctx context.Context, id, actor string) error
}
