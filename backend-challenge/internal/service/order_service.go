package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/google/uuid"
)

var (
	ErrInvalidProduct  = errors.New("invalid product")
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrEmptyOrder      = errors.New("order must contain at least one item")
	ErrInvalidCoupon   = errors.New("coupon code is not valid")
)

// CouponValidator interface for coupon validation
// Will be implemented when coupon feature is merged
type CouponValidator interface {
	IsValid(ctx context.Context, code string) bool
}

// OrderService handles order business logic
type OrderService struct {
	productRepo     ProductRepository
	couponValidator CouponValidator
}

// ProductRepository interface for product data access
type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Product, error)
}

// NewOrderService creates a new order service
func NewOrderService(productRepo ProductRepository, couponValidator CouponValidator) *OrderService {
	return &OrderService{
		productRepo:     productRepo,
		couponValidator: couponValidator,
	}
}

// CreateOrder creates a new order with optional coupon validation
func (s *OrderService) CreateOrder(ctx context.Context, req models.OrderRequest) (*models.Order, error) {
	// Validate request
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	// Validate items and fetch products (deduplicated)
	productMap := make(map[int64]models.Product)

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		productID, err := strconv.ParseInt(item.ProductID, 10, 64)
		if err != nil {
			return nil, ErrInvalidProduct
		}

		// Skip if we've already fetched this product
		if _, exists := productMap[productID]; exists {
			continue
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, ErrInvalidProduct
		}

		productMap[productID] = *product
	}

	// Convert map to slice for response
	products := make([]models.Product, 0, len(productMap))
	for _, product := range productMap {
		products = append(products, product)
	}

	// Validate coupon if provided
	if req.CouponCode != "" && s.couponValidator != nil {
		if !s.couponValidator.IsValid(ctx, req.CouponCode) {
			return nil, ErrInvalidCoupon
		}
	}

	// Generate order ID using UUID
	orderID := generateOrderID()

	order := &models.Order{
		ID:       orderID,
		Items:    req.Items,
		Products: products,
	}

	return order, nil
}

// generateOrderID generates a unique order ID using UUID
func generateOrderID() string {
	return uuid.New().String()
}
