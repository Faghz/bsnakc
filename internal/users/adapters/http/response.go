package http

import (
	"github.com/elzestia/go-boilerplate/internal/shared/response"
)

// Swagger error response aliases
type ErrorResponse = response.ErrorResponse
type UnauthorizedResponse = response.UnauthorizedResponse
type NotFoundResponse = response.NotFoundResponse

// UserResponse represents the user data in responses
type UserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"johndoe@example.com"`
	Name      string `json:"name" example:"Johm Doe"`
	Username  string `json:"username" example:"johndoe"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://api.dicebear.com/7.x/avataaars/svg?seed=johndoe"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-03-10T15:45:00Z"`
}
