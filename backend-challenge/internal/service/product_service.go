package service

import (
	"context"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
)

// ProductService handles business logic for products
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

// ListProducts returns all available products
func (s *ProductService) ListProducts(ctx context.Context) ([]models.Product, error) {
	return s.repo.GetAll(ctx)
}

// GetProduct returns a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.GetByID(ctx, id)
}
