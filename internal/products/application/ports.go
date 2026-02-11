package application

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
)

// ProductTypeService is the input port for product type operations
type ProductTypeService interface {
	Create(ctx context.Context, name, createdBy string) (*domain.ProductType, error)
	GetByUID(ctx context.Context, uid string) (*domain.ProductType, error)
	GetAll(ctx context.Context) ([]*domain.ProductType, error)
	GetAllPaginated(ctx context.Context, page, perPage int) ([]*domain.ProductType, int64, error)
	Update(ctx context.Context, uid, name, updatedBy string) (*domain.ProductType, error)
	Delete(ctx context.Context, uid, deletedBy string) error
}

// ProductService is the input port for product operations
type ProductService interface {
	Create(ctx context.Context, name, productTypeUID, createdBy string) (*domain.Product, error)
	GetByUID(ctx context.Context, uid string) (*domain.Product, error)
	GetAll(ctx context.Context) ([]*domain.Product, error)
	GetAllPaginated(ctx context.Context, page, perPage int) ([]*domain.Product, int64, error)
	FindPaginated(ctx context.Context, params ProductSearchCommand) ([]*domain.Product, int64, error)
	Update(ctx context.Context, uid, name, productTypeUID, updatedBy string) (*domain.Product, error)
	Delete(ctx context.Context, uid, deletedBy string) error
}
