package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/service"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/pkg/logger"
	"github.com/go-chi/chi/v5"
)

func TestListProducts(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	log := logger.New("error")
	handler := NewProductHandler(svc, log)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/product", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListProducts(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var products []models.Product
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(products) == 0 {
		t.Error("expected products to be returned")
	}

	// Verify we have the expected number of products
	if len(products) != 10 {
		t.Errorf("expected 10 products, got %d", len(products))
	}
}

func TestGetProduct_Success(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	log := logger.New("error")
	handler := NewProductHandler(svc, log)

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Get("/api/product/{productId}", handler.GetProduct)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/product/1", nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var product models.Product
	if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if product.ID != 1 {
		t.Errorf("expected product ID 1, got %d", product.ID)
	}

	if product.Name != "Chicken Waffle" {
		t.Errorf("expected product name 'Chicken Waffle', got %s", product.Name)
	}

	if product.Price != 12.99 {
		t.Errorf("expected product price 12.99, got %f", product.Price)
	}

	if product.Category != "Waffle" {
		t.Errorf("expected product category 'Waffle', got %s", product.Category)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	log := logger.New("error")
	handler := NewProductHandler(svc, log)

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Get("/api/product/{productId}", handler.GetProduct)

	// Create request with non-existent ID
	req := httptest.NewRequest(http.MethodGet, "/api/product/999", nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response["error"] != "Product not found" {
		t.Errorf("expected error message 'Product not found', got %s", response["error"])
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	log := logger.New("error")
	handler := NewProductHandler(svc, log)

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Get("/api/product/{productId}", handler.GetProduct)

	// Test invalid ID formats
	testCases := []struct {
		name string
		id   string
	}{
		{"letters", "invalid"},
		{"special chars", "abc@123"},
		{"float", "12.34"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with invalid ID format
			req := httptest.NewRequest(http.MethodGet, "/api/product/"+tc.id, nil)
			w := httptest.NewRecorder()

			// Execute
			r.ServeHTTP(w, req)

			// Assert
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for ID %s, got %d", tc.id, w.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if response["error"] != "Invalid ID supplied" {
				t.Errorf("expected error message 'Invalid ID supplied', got %s", response["error"])
			}
		})
	}
}

func TestGetProduct_MultipleProducts(t *testing.T) {
	// Setup
	repo := repository.NewInMemoryProductRepository()
	svc := service.NewProductService(repo)
	log := logger.New("error")
	handler := NewProductHandler(svc, log)

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Get("/api/product/{productId}", handler.GetProduct)

	// Test multiple product IDs
	testCases := []struct {
		id       string
		expectedID int64
		name     string
		category string
	}{
		{"1", 1, "Chicken Waffle", "Waffle"},
		{"4", 4, "Caesar Salad", "Salad"},
		{"7", 7, "Margherita Pizza", "Pizza"},
		{"10", 10, "Classic Burger", "Burger"},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/product/"+tc.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			var product models.Product
			if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if product.ID != tc.expectedID {
				t.Errorf("expected product ID %d, got %d", tc.expectedID, product.ID)
			}

			if product.Name != tc.name {
				t.Errorf("expected product name '%s', got %s", tc.name, product.Name)
			}

			if product.Category != tc.category {
				t.Errorf("expected product category '%s', got %s", tc.category, product.Category)
			}
		})
	}
}
