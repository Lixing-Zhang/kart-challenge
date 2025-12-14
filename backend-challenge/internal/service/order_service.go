package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
)

var (
	ErrInvalidProduct  = errors.New("invalid product")
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrEmptyOrder      = errors.New("order must contain at least one item")
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

	// Validate items and fetch products
	products := make([]models.Product, 0, len(req.Items))
	productMap := make(map[int64]models.Product)
	
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		productID, err := strconv.ParseInt(item.ProductID, 10, 64)
		if err != nil {
			return nil, ErrInvalidProduct
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, ErrInvalidProduct
		}

		products = append(products, *product)
		productMap[productID] = *product
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range req.Items {
		productID, _ := strconv.ParseInt(item.ProductID, 10, 64)
		product := productMap[productID]
		subtotal += product.Price * float64(item.Quantity)
	}

	// Validate coupon and calculate discount
	discount := 0.0
	if req.CouponCode != "" && s.couponValidator != nil {
		if s.couponValidator.IsValid(ctx, req.CouponCode) {
			discount = s.calculateDiscount(req.CouponCode, subtotal, req.Items, productMap)
		}
	}

	total := subtotal - discount

	// Generate order ID (simple implementation - in production use UUID)
	orderID := generateOrderID()

	order := &models.Order{
		ID:       orderID,
		Items:    req.Items,
		Products: products,
		Total:    total,
		Discount: discount,
	}

	return order, nil
}

// calculateDiscount calculates discount based on coupon code
func (s *OrderService) calculateDiscount(couponCode string, subtotal float64, items []models.OrderItem, productMap map[int64]models.Product) float64 {
	// Known coupon codes from requirements
	switch couponCode {
	case "HAPPYHOURS":
		// 18% discount on order total
		return subtotal * 0.18
		
	case "BUYGETONE":
		// Give lowest priced item for free
		minPrice := -1.0
		for _, item := range items {
			productID, _ := strconv.ParseInt(item.ProductID, 10, 64)
			product := productMap[productID]
			if minPrice < 0 || product.Price < minPrice {
				minPrice = product.Price
			}
		}
		if minPrice > 0 {
			return minPrice
		}
		return 0.0
		
	default:
		// Unknown coupon code - no discount
		return 0.0
	}
}

// generateOrderID generates a simple order ID
// In production, use UUID or similar
func generateOrderID() string {
	// Simple implementation - in production use uuid.New().String()
	return "ORD-" + strconv.FormatInt(123456789, 10)
}
