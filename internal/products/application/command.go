package application

// ProductSearchParams contains parameters for product search and filtering
type ProductSearchCommand struct {
	Page           int
	PerPage        int
	Search         string
	ProductTypeUID string
}
