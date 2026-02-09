package http

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/users/domain"
)

// ToUserResponse converts domain user to DTO UserResponse
func ToUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:        u.UID(),
		Email:     u.Email().String(),
		Name:      u.Name(),
		Username:  u.Username().String(),
		AvatarURL: u.AvatarURL(),
		CreatedAt: u.CreatedAt().Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt().Format(time.RFC3339),
	}
}
