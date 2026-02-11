package domain

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/google/uuid"
)

// Product represents a product entity
type Product struct {
	id             int64
	uid            string
	name           string
	productTypeID  int64
	productTypeUID string
	createdAt      time.Time
	createdBy      string
	updatedAt      time.Time
	updatedBy      string
	deletedAt      types.NullTime
	deletedBy      types.NullString
}

// NewProduct creates a new product entity with validation
func NewProduct(name, createdBy string, productTypeID int64) (*Product, error) {
	if err := validateProductName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	uid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Product{
		uid:           uid.String(),
		name:          name,
		productTypeID: productTypeID,
		createdAt:     now,
		createdBy:     createdBy,
		updatedAt:     now,
		updatedBy:     createdBy,
	}, nil
}

// ReconstructProduct reconstructs a product from persistence
func ReconstructProduct(
	id int64,
	uid string,
	name string,
	productTypeID int64,
	productTypeUID string,
	createdAt time.Time,
	createdBy string,
	updatedAt time.Time,
	updatedBy string,
	deletedAt types.NullTime,
	deletedBy types.NullString,
) *Product {
	return &Product{
		id:             id,
		uid:            uid,
		name:           name,
		productTypeID:  productTypeID,
		productTypeUID: productTypeUID,
		createdAt:      createdAt,
		createdBy:      createdBy,
		updatedAt:      updatedAt,
		updatedBy:      updatedBy,
		deletedAt:      deletedAt,
		deletedBy:      deletedBy,
	}
}

// Getters

func (p *Product) ID() int64              { return p.id }
func (p *Product) UID() string            { return p.uid }
func (p *Product) Name() string           { return p.name }
func (p *Product) ProductTypeID() int64   { return p.productTypeID }
func (p *Product) ProductTypeUID() string { return p.productTypeUID }
func (p *Product) CreatedAt() time.Time   { return p.createdAt }
func (p *Product) CreatedBy() string      { return p.createdBy }
func (p *Product) UpdatedAt() time.Time   { return p.updatedAt }
func (p *Product) UpdatedBy() string      { return p.updatedBy }
func (p *Product) DeletedAt() types.NullTime    { return p.deletedAt }
func (p *Product) DeletedBy() types.NullString   { return p.deletedBy }

// SetInternalID sets the internal database ID (called by repository after INSERT)
func (p *Product) SetInternalID(id int64) {
	p.id = id
}

// SetName updates the product name with validation
func (p *Product) SetName(name string) error {
	if err := validateProductName(name); err != nil {
		return err
	}
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

// SetProductTypeID updates the product type reference
func (p *Product) SetProductTypeID(productTypeID int64) {
	p.productTypeID = productTypeID
	p.updatedAt = time.Now()
}

// MarkDeleted marks the product as deleted (soft delete)
func (p *Product) MarkDeleted(deletedBy string) {
	now := time.Now()
	p.deletedAt = types.NewNullTime(now)
	if deletedBy != "" {
		p.deletedBy = types.NewNullString(deletedBy)
	}
	p.updatedAt = now
}

// IsDeleted checks if the product is soft-deleted
func (p *Product) IsDeleted() bool {
	return p.deletedAt.Valid
}

func validateProductName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return ErrInvalidProductName
	}
	return nil
}
