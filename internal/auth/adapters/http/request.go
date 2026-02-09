package http

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email_format" example:"johndoe@example.com"`
	Password string `json:"password" validate:"required,min=8,max=128,password_complexity" example:"P@ssw0rd!2024#Secure"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Username string `json:"username" validate:"required,username_format" example:"johndoe"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email_format" example:"johndoe@example.com"`
	Password string `json:"password" validate:"required" example:"P@ssw0rd!2024#Secure"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"v2.local.Ey8DhK2mT4wS7vN9xR3cF1aG5kM0pW2jX6zL9yB4nV8tQ1rC3eH"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"v2.local.Ey8DhK2mT4wS7vN9xR3cF1aG5kM0pW2jX6zL9yB4nV8tQ1rC3eH"`
}
