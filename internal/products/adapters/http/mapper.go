package http

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/response"
)

func ToProductTypeResponse(pt *domain.ProductType) *ProductTypeResponse {
	return &ProductTypeResponse{
		ID:        pt.UID(),
		Name:      pt.Name(),
		CreatedAt: pt.CreatedAt().Format(time.RFC3339),
		UpdatedAt: pt.UpdatedAt().Format(time.RFC3339),
	}
}

func ToProductTypeListResponse(pts []*domain.ProductType, page, perPage int, total int64) *ProductTypeListResponse {
	items := make([]ProductTypeResponse, len(pts))
	for i, pt := range pts {
		items[i] = *ToProductTypeResponse(pt)
	}

	resp := &ProductTypeListResponse{
		Data: items,
	}

	if page > 0 && perPage > 0 {
		totalPages := int(total) / perPage
		if int(total)%perPage > 0 {
			totalPages++
		}

		resp.Pagination = &response.Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		}
	}

	return resp
}

func ToProductResponse(p *domain.Product) *ProductResponse {
	return &ProductResponse{
		ID:            p.UID(),
		Name:          p.Name(),
		ProductTypeID: p.ProductTypeUID(),
		CreatedAt:     p.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt().Format(time.RFC3339),
	}
}

func ToProductListResponse(products []*domain.Product, page, perPage int, total int64) *ProductListResponse {
	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = *ToProductResponse(p)
	}

	resp := &ProductListResponse{
		Data: items,
	}

	if page > 0 && perPage > 0 {
		totalPages := int(total) / perPage
		if int(total)%perPage > 0 {
			totalPages++
		}

		resp.Pagination = &response.Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		}
	}

	return resp
}
