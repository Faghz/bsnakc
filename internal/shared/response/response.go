package response

// ErrorResponse represents a standardized JSON error response
type ErrorResponse struct {
	Error   string                 `json:"error" example:"VALIDATION_ERROR"`
	Message string                 `json:"message" example:"Validation failed for one or more fields"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object,string" example:"field:error description"`
}

// BadRequestResponse represents a 400 Bad Request error (invalid input or validation failure)
type BadRequestResponse struct {
	Error   string                 `json:"error" example:"bad_request"`
	Message string                 `json:"message" example:"invalid request body"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object,string" example:"name:must be between 2 and 100 characters"`
}

// ValidationErrorResponse represents a 400 Validation error (struct field validation failure)
type ValidationErrorResponse struct {
	Error   string                 `json:"error" example:"validation_error"`
	Message string                 `json:"message" example:"invalid product type name"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object,string" example:"name:must be between 2 and 100 characters"`
}

// UnauthorizedResponse represents a 401 Unauthorized error (missing or invalid token)
type UnauthorizedResponse struct {
	Error   string                 `json:"error" example:"unauthorized"`
	Message string                 `json:"message" example:"missing or invalid authorization token"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NotFoundResponse represents a 404 Not Found error (resource does not exist)
type NotFoundResponse struct {
	Error   string                 `json:"error" example:"not_found"`
	Message string                 `json:"message" example:"product type not found"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object,string" example:"resource:product type"`
}

// ConflictResponse represents a 409 Conflict error (duplicate resource)
type ConflictResponse struct {
	Error   string                 `json:"error" example:"conflict"`
	Message string                 `json:"message" example:"product type name already exists"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Pagination contains pagination metadata for list responses
type Pagination struct {
	Page       int   `json:"page" example:"1"`
	PerPage    int   `json:"per_page" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}

// PaginatedResponse wraps a list response with pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// NewPaginatedResponse creates a paginated response from data and pagination info
func NewPaginatedResponse(data interface{}, page, perPage int, total int64) PaginatedResponse {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
