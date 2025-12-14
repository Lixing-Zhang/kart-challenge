package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/service"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/pkg/logger"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	// Setup
	productRepo := repository.NewInMemoryProductRepository()
	orderService := service.NewOrderService(productRepo, nil)
	log := logger.New("info")
	handler := NewOrderHandler(orderService, log)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *models.Order)
	}{
		{
			name: "successful order",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 2},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, order *models.Order) {
				if order.ID == "" {
					t.Error("order ID is empty")
				}
				if len(order.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(order.Items))
				}
				if order.Total <= 0 {
					t.Errorf("expected positive total, got %f", order.Total)
				}
			},
		},
		{
			name: "multiple items order",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
					{ProductID: "2", Quantity: 2},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, order *models.Order) {
				if len(order.Items) != 2 {
					t.Errorf("expected 2 items, got %d", len(order.Items))
				}
			},
		},
		{
			name: "empty order",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name: "invalid quantity",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 0},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name: "invalid product",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "99999", Quantity: 1},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body []byte
			var err error
			
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewReader(body))
			req = req.WithContext(context.Background())
			w := httptest.NewRecorder()

			// Execute
			handler.CreateOrder(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Check response if success
			if tt.expectedStatus == http.StatusOK && tt.checkResponse != nil {
				var order models.Order
				if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				tt.checkResponse(t, &order)
			}
		})
	}
}
