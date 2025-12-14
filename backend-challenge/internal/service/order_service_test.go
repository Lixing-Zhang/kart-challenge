package service

import (
	"context"
	"testing"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
)

func TestOrderService_CreateOrder(t *testing.T) {
	productRepo := repository.NewInMemoryProductRepository()
	orderService := NewOrderService(productRepo, nil) // No coupon validator for basic tests

	tests := []struct {
		name    string
		req     models.OrderRequest
		wantErr error
	}{
		{
			name: "valid order with single item",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 2},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid order with multiple items",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
					{ProductID: "2", Quantity: 3},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty order",
			req: models.OrderRequest{
				Items: []models.OrderItem{},
			},
			wantErr: ErrEmptyOrder,
		},
		{
			name: "invalid quantity - zero",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 0},
				},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "invalid quantity - negative",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: -1},
				},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "invalid product ID - non-numeric",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "invalid", Quantity: 1},
				},
			},
			wantErr: ErrInvalidProduct,
		},
		{
			name: "invalid product ID - not found",
			req: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "99999", Quantity: 1},
				},
			},
			wantErr: ErrInvalidProduct,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := orderService.CreateOrder(context.Background(), tt.req)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateOrder() unexpected error = %v", err)
				return
			}

			if order == nil {
				t.Error("CreateOrder() returned nil order")
				return
			}

			if order.ID == "" {
				t.Error("CreateOrder() order ID is empty")
			}

			if len(order.Items) != len(tt.req.Items) {
				t.Errorf("CreateOrder() items count = %d, want %d", len(order.Items), len(tt.req.Items))
			}

			// Products should be deduplicated - count unique product IDs
			uniqueProductIDs := make(map[string]bool)
			for _, item := range tt.req.Items {
				uniqueProductIDs[item.ProductID] = true
			}
			expectedProductCount := len(uniqueProductIDs)

			if len(order.Products) != expectedProductCount {
				t.Errorf("CreateOrder() products count = %d, want %d (deduplicated)", len(order.Products), expectedProductCount)
			}
		})
	}
}

func TestOrderService_CreateOrder_DuplicateProducts(t *testing.T) {
	productRepo := repository.NewInMemoryProductRepository()
	orderService := NewOrderService(productRepo, nil)

	// Order with duplicate product IDs
	req := models.OrderRequest{
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 2},
			{ProductID: "1", Quantity: 3},
			{ProductID: "2", Quantity: 1},
		},
	}

	order, err := orderService.CreateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateOrder() unexpected error = %v", err)
	}

	// Should have 3 items
	if len(order.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(order.Items))
	}

	// Should have 2 unique products (deduplicated)
	if len(order.Products) != 2 {
		t.Errorf("products count = %d, want 2 (deduplicated)", len(order.Products))
	}

	// Verify products are unique
	productIDs := make(map[int64]bool)
	for _, product := range order.Products {
		if productIDs[product.ID] {
			t.Errorf("duplicate product ID %d in products array", product.ID)
		}
		productIDs[product.ID] = true
	}
}
