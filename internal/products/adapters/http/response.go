package http

import (
	"github.com/elzestia/go-boilerplate/internal/shared/response"
)

// Swagger error response aliases
type ErrorResponse = response.ErrorResponse
type BadRequestResponse = response.BadRequestResponse
type ValidationErrorResponse = response.ValidationErrorResponse
type UnauthorizedResponse = response.UnauthorizedResponse
type NotFoundResponse = response.NotFoundResponse
type ConflictResponse = response.ConflictResponse

// ProductTypeResponse represents a product type in API responses
type ProductTypeResponse struct {
	ID        string `json:"id" example:"0192f4a1-b0e7-7f3a-9c1d-4e5f6a7b8c9d"`
	Name      string `json:"name" example:"Snacks"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ProductTypeListResponse wraps a list of product types for swagger documentation
type ProductTypeListResponse struct {
	Data       []ProductTypeResponse `json:"data"`
	Pagination *response.Pagination  `json:"pagination,omitempty"`
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	ID            string `json:"id" example:"0192f4a1-c2d3-7e4f-8a9b-0c1d2e3f4a5b"`
	Name          string `json:"name" example:"Chocolate Bar"`
	ProductTypeID string `json:"product_type_id" example:"0192f4a1-b0e7-7f3a-9c1d-4e5f6a7b8c9d"`
	CreatedAt     string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt     string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ProductListResponse wraps a list of products for swagger documentation
type ProductListResponse struct {
	Data       []ProductResponse    `json:"data"`
	Pagination *response.Pagination `json:"pagination,omitempty"`
}
