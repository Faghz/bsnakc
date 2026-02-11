package application

import (
	sharedPorts "github.com/elzestia/go-boilerplate/internal/shared/application/ports"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
)

// ProductTypeServiceImpl implements the ProductTypeService port
type ProductTypeServiceImpl struct {
	productTypeRepo domain.ProductTypeRepository
	logger          sharedPorts.Logger
}

// Ensure ProductTypeServiceImpl implements ProductTypeService
var _ ProductTypeService = (*ProductTypeServiceImpl)(nil)

// ProductServiceImpl implements the ProductService port
type ProductServiceImpl struct {
	productRepo     domain.ProductRepository
	productTypeRepo domain.ProductTypeRepository
	logger          sharedPorts.Logger
}

// Ensure ProductServiceImpl implements ProductService
var _ ProductService = (*ProductServiceImpl)(nil)

// NewProductTypeService creates a new product type service
func NewProductTypeService(repo domain.ProductTypeRepository, logger sharedPorts.Logger) *ProductTypeServiceImpl {
	return &ProductTypeServiceImpl{
		productTypeRepo: repo,
		logger:          logger.With(sharedPorts.String("domain", "product_types")),
	}
}

// NewProductService creates a new product service
func NewProductService(productRepo domain.ProductRepository, productTypeRepo domain.ProductTypeRepository, logger sharedPorts.Logger) *ProductServiceImpl {
	return &ProductServiceImpl{
		productRepo:     productRepo,
		productTypeRepo: productTypeRepo,
		logger:          logger.With(sharedPorts.String("domain", "products")),
	}
}
