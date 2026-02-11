package application

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// Create creates a new product type
func (s *ProductTypeServiceImpl) Create(ctx context.Context, name, createdBy string) (*domain.ProductType, error) {
	s.logger.Debug(ctx, "Creating product type", ports.String("name", name))

	// Check uniqueness by name
	existing, err := s.productTypeRepo.FindByName(ctx, name)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrProductTypeTaken
	}

	pt, err := domain.NewProductType(name, createdBy)
	if err != nil {
		return nil, err
	}

	if err := s.productTypeRepo.Create(ctx, pt); err != nil {
		return nil, err
	}

	s.logger.Debug(ctx, "Product type created successfully", ports.String("uid", pt.UID()))
	return pt, nil
}

// GetByUID gets a product type by UID
func (s *ProductTypeServiceImpl) GetByUID(ctx context.Context, uid string) (*domain.ProductType, error) {
	return s.productTypeRepo.FindByUID(ctx, uid)
}

// GetAll gets all product types (non-deleted)
func (s *ProductTypeServiceImpl) GetAll(ctx context.Context) ([]*domain.ProductType, error) {
	return s.productTypeRepo.FindAll(ctx)
}

// GetAllPaginated gets product types with pagination
func (s *ProductTypeServiceImpl) GetAllPaginated(ctx context.Context, page, perPage int) ([]*domain.ProductType, int64, error) {
	offset := (page - 1) * perPage
	return s.productTypeRepo.FindAllPaginated(ctx, offset, perPage)
}

// Update updates a product type
func (s *ProductTypeServiceImpl) Update(ctx context.Context, uid, name, updatedBy string) (*domain.ProductType, error) {
	pt, err := s.productTypeRepo.FindByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Check name uniqueness if changed
	if pt.Name() != name {
		existing, err := s.productTypeRepo.FindByName(ctx, name)
		if err != nil && !errors.IsNotFoundError(err) {
			return nil, err
		}
		if existing != nil {
			return nil, domain.ErrProductTypeTaken
		}
	}

	if err := pt.SetName(name); err != nil {
		return nil, err
	}

	if err := s.productTypeRepo.Update(ctx, pt); err != nil {
		return nil, err
	}

	return pt, nil
}

// Delete soft-deletes a product type
func (s *ProductTypeServiceImpl) Delete(ctx context.Context, uid, deletedBy string) error {
	pt, err := s.productTypeRepo.FindByUID(ctx, uid)
	if err != nil {
		return err
	}

	pt.MarkDeleted(deletedBy)

	return s.productTypeRepo.Update(ctx, pt)
}
