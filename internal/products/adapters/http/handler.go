package http

import (
	"net/http"
	"strconv"

	"github.com/elzestia/go-boilerplate/internal/adapters/http/middleware"
	authports "github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/products/application"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/validation"
	"github.com/labstack/echo/v4"
)

// ProductHandler handles product and product type endpoints
type ProductHandler struct {
	productTypeService application.ProductTypeService
	productService     application.ProductService
	validator          validation.Validator
}

// NewProductHandler creates a new product handler and registers routes
func NewProductHandler(
	route *echo.Group,
	productTypeService application.ProductTypeService,
	productService application.ProductService,
	tokenService authports.TokenService,
	validator validation.Validator,
) {
	handler := &ProductHandler{
		productTypeService: productTypeService,
		productService:     productService,
		validator:          validator,
	}

	auth := middleware.AuthMiddleware(tokenService)

	// Product Types
	ptGroup := route.Group("/product-types", auth)
	{
		ptGroup.POST("", handler.CreateProductType)
		ptGroup.GET("", handler.ListProductTypes)
		ptGroup.GET("/:id", handler.GetProductType)
		ptGroup.PUT("/:id", handler.UpdateProductType)
		ptGroup.DELETE("/:id", handler.DeleteProductType)
	}

	// Products
	pGroup := route.Group("/products", auth)
	{
		pGroup.POST("", handler.CreateProduct)
		pGroup.GET("", handler.ListProducts)
		pGroup.GET("/:id", handler.GetProduct)
		pGroup.PUT("/:id", handler.UpdateProduct)
		pGroup.DELETE("/:id", handler.DeleteProduct)
	}
}

// Product Type handlers

// CreateProductType godoc
// @Summary Create a new product type
// @Description Create a new product type with a unique name
// @Tags product-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateProductTypeRequest true "Create product type request"
// @Success 201 {object} ProductTypeResponse "Product type created successfully"
// @Failure 400 {object} ValidationErrorResponse "Name must be between 2 and 100 characters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 409 {object} ConflictResponse "Product type name already exists"
// @Router /v1/product-types [post]
func (h *ProductHandler) CreateProductType(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	var req CreateProductTypeRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	pt, err := h.productTypeService.Create(c.Request().Context(), req.Name, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ToProductTypeResponse(pt))
}

// ListProductTypes godoc
// @Summary List all product types
// @Description Get a list of all active product types
// @Tags product-types
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} ProductTypeListResponse "List of product types"
// @Failure 400 {object} ValidationErrorResponse "Invalid pagination parameters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Router /v1/product-types [get]
func (h *ProductHandler) ListProductTypes(c echo.Context) error {
	page := 1
	perPage := 20

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			return errors.NewBadRequestError("invalid page parameter")
		}
	}

	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		} else {
			return errors.NewBadRequestError("invalid per_page parameter")
		}
	}

	pts, total, err := h.productTypeService.GetAllPaginated(c.Request().Context(), page, perPage)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductTypeListResponse(pts, page, perPage, total))
}

// GetProductType godoc
// @Summary Get a product type
// @Description Get a product type by its ID
// @Tags product-types
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product Type ID (UUID)"
// @Success 200 {object} ProductTypeResponse "Product type details"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product type not found"
// @Router /v1/product-types/{id} [get]
func (h *ProductHandler) GetProductType(c echo.Context) error {
	uid := c.Param("id")

	pt, err := h.productTypeService.GetByUID(c.Request().Context(), uid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductTypeResponse(pt))
}

// UpdateProductType godoc
// @Summary Update a product type
// @Description Update an existing product type's name
// @Tags product-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product Type ID (UUID)"
// @Param request body UpdateProductTypeRequest true "Update product type request"
// @Success 200 {object} ProductTypeResponse "Product type updated successfully"
// @Failure 400 {object} ValidationErrorResponse "Name must be between 2 and 100 characters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product type not found"
// @Failure 409 {object} ConflictResponse "Product type name already exists"
// @Router /v1/product-types/{id} [put]
func (h *ProductHandler) UpdateProductType(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	uid := c.Param("id")

	var req UpdateProductTypeRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	pt, err := h.productTypeService.Update(c.Request().Context(), uid, req.Name, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductTypeResponse(pt))
}

// DeleteProductType godoc
// @Summary Delete a product type
// @Description Soft-delete a product type by its ID
// @Tags product-types
// @Security BearerAuth
// @Param id path string true "Product Type ID (UUID)"
// @Success 204 "Product type deleted successfully"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product type not found"
// @Router /v1/product-types/{id} [delete]
func (h *ProductHandler) DeleteProductType(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	uid := c.Param("id")

	if err := h.productTypeService.Delete(c.Request().Context(), uid, userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Product handlers

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with a name and product type
// @Tags products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Create product request"
// @Success 201 {object} ProductResponse "Product created successfully"
// @Failure 400 {object} ValidationErrorResponse "Name must be between 2 and 100 characters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Referenced product type not found"
// @Router /v1/products [post]
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	p, err := h.productService.Create(c.Request().Context(), req.Name, req.ProductTypeID, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, ToProductResponse(p))
}

// ListProducts godoc
// @Summary List all products
// @Description Get a list of all active products with optional search and filtering
// @Tags products
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" minimum(1) maximum(100)
// @Param search query string false "Search by product name (case-insensitive partial match)" maxlength(100)
// @Param product_type_uid query string false "Filter by product type UID" format(uuid)
// @Success 200 {object} ProductListResponse "List of products"
// @Failure 400 {object} ValidationErrorResponse "Invalid parameters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Router /v1/products [get]
func (h *ProductHandler) ListProducts(c echo.Context) error {
	var req FindProductsRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request parameters").WithCause(err)
	}

	req.SetDefaults()
	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	products, total, err := h.productService.FindPaginated(c.Request().Context(), application.ProductSearchCommand{
		Page:           req.Page,
		PerPage:        req.PerPage,
		Search:         req.Search,
		ProductTypeUID: req.ProductTypeUID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductListResponse(products, req.Page, req.PerPage, total))
}

// GetProduct godoc
// @Summary Get a product
// @Description Get a product by its ID
// @Tags products
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} ProductResponse "Product details"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product not found"
// @Router /v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	uid := c.Param("id")

	p, err := h.productService.GetByUID(c.Request().Context(), uid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductResponse(p))
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product's name and product type
// @Tags products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body UpdateProductRequest true "Update product request"
// @Success 200 {object} ProductResponse "Product updated successfully"
// @Failure 400 {object} ValidationErrorResponse "Name must be between 2 and 100 characters"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product or referenced product type not found"
// @Router /v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	uid := c.Param("id")

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("invalid request body").WithCause(err)
	}

	if err := h.validator.Validate(&req); err != nil {
		return err
	}

	p, err := h.productService.Update(c.Request().Context(), uid, req.Name, req.ProductTypeID, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ToProductResponse(p))
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Soft-delete a product by its ID
// @Tags products
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 204 "Product deleted successfully"
// @Failure 401 {object} UnauthorizedResponse "Missing or invalid authorization token"
// @Failure 404 {object} NotFoundResponse "Product not found"
// @Router /v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return errors.NewAuthenticationError("user not authenticated")
	}

	uid := c.Param("id")

	if err := h.productService.Delete(c.Request().Context(), uid, userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
