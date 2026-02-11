package http

type CreateProductTypeRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateProductTypeRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type CreateProductRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=100"`
	ProductTypeID string `json:"product_type_id" validate:"required,uuid"`
}

type UpdateProductRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=100"`
	ProductTypeID string `json:"product_type_id" validate:"required,uuid"`
}

type FindProductsRequest struct {
	Page           int    `query:"page" validate:"omitempty,min=1"`
	PerPage        int    `query:"per_page" validate:"omitempty,min=1,max=100"`
	Search         string `query:"search" validate:"omitempty,max=100"`
	ProductTypeUID string `query:"product_type_uid" validate:"omitempty,uuid"`
}

func (r *FindProductsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PerPage == 0 {
		r.PerPage = 10
	}
}
