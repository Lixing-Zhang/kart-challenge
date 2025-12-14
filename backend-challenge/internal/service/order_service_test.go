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

			if len(order.Products) != len(tt.req.Items) {
				t.Errorf("CreateOrder() products count = %d, want %d", len(order.Products), len(tt.req.Items))
			}

			if order.Total <= 0 {
				t.Errorf("CreateOrder() total = %f, want > 0", order.Total)
			}
		})
	}
}

func TestOrderService_CalculateDiscount(t *testing.T) {
	productRepo := repository.NewInMemoryProductRepository()
	orderService := NewOrderService(productRepo, nil)

	tests := []struct {
		name         string
		couponCode   string
		subtotal     float64
		items        []models.OrderItem
		wantDiscount float64
	}{
		{
			name:       "HAPPYHOURS - 18% discount",
			couponCode: "HAPPYHOURS",
			subtotal:   100.0,
			items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
			wantDiscount: 18.0,
		},
		{
			name:       "BUYGETONE - lowest item free",
			couponCode: "BUYGETONE",
			subtotal:   50.0,
			items: []models.OrderItem{
				{ProductID: "1", Quantity: 1}, // $12.99
				{ProductID: "2", Quantity: 1}, // $10.99
			},
			wantDiscount: 10.99, // Lowest priced item
		},
		{
			name:       "unknown coupon code",
			couponCode: "INVALID",
			subtotal:   100.0,
			items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
			wantDiscount: 0.0,
		},
		{
			name:       "empty coupon code",
			couponCode: "",
			subtotal:   100.0,
			items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
			wantDiscount: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build product map
			productMap := make(map[int64]models.Product)
			if len(tt.items) > 0 {
				product, _ := productRepo.GetByID(context.Background(), 1)
				productMap[1] = *product
				product2, _ := productRepo.GetByID(context.Background(), 2)
				productMap[2] = *product2
			}

			discount := orderService.calculateDiscount(tt.couponCode, tt.subtotal, tt.items, productMap)

			if discount != tt.wantDiscount {
				t.Errorf("calculateDiscount() = %f, want %f", discount, tt.wantDiscount)
			}
		})
	}
}
