package application

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"golang.org/x/sync/errgroup"
)

// Create creates a new product
func (s *ProductServiceImpl) Create(ctx context.Context, name, productTypeUID, createdBy string) (*domain.Product, error) {
	s.logger.Debug(ctx, "Creating product", sharedports.String("name", name))

	// Resolve product type UID to internal ID
	pt, err := s.productTypeRepo.FindByUID(ctx, productTypeUID)
	if err != nil {
		return nil, err
	}

	p, err := domain.NewProduct(name, createdBy, pt.ID())
	if err != nil {
		return nil, err
	}

	if err := s.productRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	s.logger.Debug(ctx, "Product created successfully", sharedports.String("uid", p.UID()))

	// Re-fetch to get joined product type UID
	return s.productRepo.FindByUID(ctx, p.UID())
}

// GetByUID gets a product by UID
func (s *ProductServiceImpl) GetByUID(ctx context.Context, uid string) (*domain.Product, error) {
	return s.productRepo.FindByUID(ctx, uid)
}

// GetAll gets all products (non-deleted)
func (s *ProductServiceImpl) GetAll(ctx context.Context) ([]*domain.Product, error) {
	return s.productRepo.FindAll(ctx)
}

// GetAllPaginated gets products with pagination
func (s *ProductServiceImpl) GetAllPaginated(ctx context.Context, page, perPage int) ([]*domain.Product, int64, error) {
	offset := (page - 1) * perPage
	return s.productRepo.FindAllPaginated(ctx, offset, perPage)
}

// FindPaginated gets products with pagination, search, and filtering using concurrent data fetching
func (s *ProductServiceImpl) FindPaginated(ctx context.Context, params ProductSearchCommand) ([]*domain.Product, int64, error) {
	criteria := domain.ProductSearchCriteria{
		Offset:         (params.Page - 1) * params.PerPage,
		Limit:          params.PerPage,
		Search:         params.Search,
		ProductTypeUID: params.ProductTypeUID,
	}

	var products []*domain.Product
	var count int64

	// Use errgroup for concurrent execution
	g, gctx := errgroup.WithContext(ctx)

	// Fetch products concurrently
	g.Go(func() error {
		p, err := s.productRepo.FindPaginatedList(gctx, criteria)
		if err != nil {
			return err
		}
		products = p
		return nil
	})

	// Fetch count concurrently
	g.Go(func() error {
		c, err := s.productRepo.FindPaginatedCount(gctx, criteria)
		if err != nil {
			return err
		}
		count = c
		return nil
	})

	// Wait for both goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return products, count, nil
}

// Update updates a product
func (s *ProductServiceImpl) Update(ctx context.Context, uid, name, productTypeUID, updatedBy string) (*domain.Product, error) {
	p, err := s.productRepo.FindByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if err := p.SetName(name); err != nil {
		return nil, err
	}

	// Resolve product type UID to internal ID
	pt, err := s.productTypeRepo.FindByUID(ctx, productTypeUID)
	if err != nil {
		return nil, err
	}
	p.SetProductTypeID(pt.ID())

	if err := s.productRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	// Re-fetch to get joined product type UID
	return s.productRepo.FindByUID(ctx, p.UID())
}

// Delete soft-deletes a product
func (s *ProductServiceImpl) Delete(ctx context.Context, uid, deletedBy string) error {
	p, err := s.productRepo.FindByUID(ctx, uid)
	if err != nil {
		return err
	}

	p.MarkDeleted(deletedBy)

	return s.productRepo.Update(ctx, p)
}
