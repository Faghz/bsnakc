package http

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	Name      string `json:"name" validate:"omitempty,min=2,max=100" example:"Johm Doe-Martinez"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url" example:"https://avatars.githubusercontent.com/u/1234567?v=4"`
}
