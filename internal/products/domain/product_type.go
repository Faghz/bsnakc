package domain

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/google/uuid"
)

// ProductType represents a product type (category) entity
type ProductType struct {
	id        int64
	uid       string
	name      string
	createdAt time.Time
	createdBy string
	updatedAt time.Time
	updatedBy string
	deletedAt types.NullTime
	deletedBy types.NullString
}

// NewProductType creates a new product type entity with validation
func NewProductType(name, createdBy string) (*ProductType, error) {
	if err := validateProductTypeName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	uid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &ProductType{
		uid:       uid.String(),
		name:      name,
		createdAt: now,
		createdBy: createdBy,
		updatedAt: now,
		updatedBy: createdBy,
	}, nil
}

// ReconstructProductType reconstructs a product type from persistence
func ReconstructProductType(
	id int64,
	uid string,
	name string,
	createdAt time.Time,
	createdBy string,
	updatedAt time.Time,
	updatedBy string,
	deletedAt types.NullTime,
	deletedBy types.NullString,
) *ProductType {
	return &ProductType{
		id:        id,
		uid:       uid,
		name:      name,
		createdAt: createdAt,
		createdBy: createdBy,
		updatedAt: updatedAt,
		updatedBy: updatedBy,
		deletedAt: deletedAt,
		deletedBy: deletedBy,
	}
}

// Getters

func (pt *ProductType) ID() int64         { return pt.id }
func (pt *ProductType) UID() string       { return pt.uid }
func (pt *ProductType) Name() string      { return pt.name }
func (pt *ProductType) CreatedAt() time.Time { return pt.createdAt }
func (pt *ProductType) CreatedBy() string  { return pt.createdBy }
func (pt *ProductType) UpdatedAt() time.Time { return pt.updatedAt }
func (pt *ProductType) UpdatedBy() string  { return pt.updatedBy }
func (pt *ProductType) DeletedAt() types.NullTime   { return pt.deletedAt }
func (pt *ProductType) DeletedBy() types.NullString  { return pt.deletedBy }

// SetInternalID sets the internal database ID (called by repository after INSERT)
func (pt *ProductType) SetInternalID(id int64) {
	pt.id = id
}

// SetName updates the product type name with validation
func (pt *ProductType) SetName(name string) error {
	if err := validateProductTypeName(name); err != nil {
		return err
	}
	pt.name = name
	pt.updatedAt = time.Now()
	return nil
}

// MarkDeleted marks the product type as deleted (soft delete)
func (pt *ProductType) MarkDeleted(deletedBy string) {
	now := time.Now()
	pt.deletedAt = types.NewNullTime(now)
	if deletedBy != "" {
		pt.deletedBy = types.NewNullString(deletedBy)
	}
	pt.updatedAt = now
}

// IsDeleted checks if the product type is soft-deleted
func (pt *ProductType) IsDeleted() bool {
	return pt.deletedAt.Valid
}

func validateProductTypeName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return ErrInvalidProductTypeName
	}
	return nil
}
