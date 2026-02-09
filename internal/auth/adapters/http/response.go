package http

// AuthResponse represents the authentication response
type AuthResponse struct {
	User             UserResponse `json:"user"`
	AccessToken      string       `json:"access_token" example:"v2.local.Qf7FxW3K9Z2nR5mH8jY4tL6pB0dN1vS3cH7eA9gF2kR4mT6xZ8wQ"`
	ExpiresAt        string       `json:"expires_at" example:"2024-03-15T14:30:00Z"`
	RefreshToken     string       `json:"refresh_token" example:"v2.local.Ey8DhK2mT4wS7vN9xR3cF1aG5kM0pW2jX6zL9yB4nV8tQ1rC3eH"`
	RefreshExpiresAt string       `json:"refresh_expires_at" example:"2024-04-14T14:30:00Z"`
}

// UserResponse represents the user data in responses
type UserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"johndoe@example.com"`
	Name      string `json:"name" example:"Johm Doe"`
	Username  string `json:"username" example:"johndoe"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://api.dicebear.com/7.x/avataaars/svg?seed=johndoe"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error" example:"VALIDATION_ERROR"`
	Message string                 `json:"message" example:"Validation failed for one or more fields"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object,string" example:"email:must be a valid email address,password:must be at least 8 characters"`
}
