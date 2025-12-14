package repository

import (
	"context"
	"errors"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
}

// InMemoryProductRepository implements ProductRepository with in-memory storage
type InMemoryProductRepository struct {
	products map[string]models.Product
}

// NewInMemoryProductRepository creates a new in-memory product repository with seed data
func NewInMemoryProductRepository() *InMemoryProductRepository {
	// Seed data based on OpenAPI spec examples
	products := map[string]models.Product{
		"1":  {ID: "1", Name: "Chicken Waffle", Price: 12.99, Category: "Waffle"},
		"2":  {ID: "2", Name: "Belgian Waffle", Price: 10.99, Category: "Waffle"},
		"3":  {ID: "3", Name: "Chocolate Waffle", Price: 11.99, Category: "Waffle"},
		"4":  {ID: "4", Name: "Caesar Salad", Price: 8.99, Category: "Salad"},
		"5":  {ID: "5", Name: "Greek Salad", Price: 9.49, Category: "Salad"},
		"6":  {ID: "6", Name: "Garden Salad", Price: 7.99, Category: "Salad"},
		"7":  {ID: "7", Name: "Margherita Pizza", Price: 14.99, Category: "Pizza"},
		"8":  {ID: "8", Name: "Pepperoni Pizza", Price: 16.99, Category: "Pizza"},
		"9":  {ID: "9", Name: "Veggie Pizza", Price: 15.49, Category: "Pizza"},
		"10": {ID: "10", Name: "Classic Burger", Price: 13.99, Category: "Burger"},
	}

	return &InMemoryProductRepository{
		products: products,
	}
}

// GetAll returns all products
func (r *InMemoryProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	products := make([]models.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

// GetByID returns a product by its ID
func (r *InMemoryProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product, exists := r.products[id]
	if !exists {
		return nil, ErrProductNotFound
	}
	return &product, nil
}
