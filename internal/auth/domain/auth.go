package domain

import "time"

// AuthResponse represents the result of an authentication operation
type AuthResponse struct {
	UserID             string
	Email              string
	Name               string
	Username           string
	AvatarURL          string
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
}
