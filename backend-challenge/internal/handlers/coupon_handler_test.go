package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// mockValidator implements a simple mock validator for testing
type mockValidator struct {
	validCoupons map[string]bool
}

func (m *mockValidator) IsValid(ctx context.Context, code string) bool {
	return m.validCoupons[code]
}

func (m *mockValidator) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_files":   3,
		"file_sizes":    []int{100, 200, 150},
		"total_coupons": 450,
	}
}

func TestCouponHandler_ValidateCoupon(t *testing.T) {
	// Create a mock validator with some test data
	mockVal := &mockValidator{
		validCoupons: map[string]bool{
			"HAPPYHRS": true,
			"FIFTYOFF": true,
		},
	}
	
	tests := []struct {
		name           string
		couponCode     string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "valid coupon",
			couponCode:     "HAPPYHRS",
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "invalid coupon - too short",
			couponCode:     "SHORT",
			expectedStatus: http.StatusNotFound,
			expectedValid:  false,
		},
		{
			name:           "invalid coupon - does not exist",
			couponCode:     "NOTEXIST",
			expectedStatus: http.StatusNotFound,
			expectedValid:  false,
		},
		{
			name:           "empty coupon code",
			couponCode:     "",
			expectedStatus: http.StatusNotFound,
			expectedValid:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Always use mock validator for these tests
			h := NewCouponHandler(mockVal)
			
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/coupon/"+tt.couponCode, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("couponCode", tt.couponCode)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			// Create response recorder
			rr := httptest.NewRecorder()
			
			// Execute handler
			h.ValidateCoupon(rr, req)
			
			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			
			// Parse response
			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			
			// Check valid field
			valid, ok := response["valid"].(bool)
			if !ok {
				t.Fatalf("valid field is not a boolean")
			}
			
			if valid != tt.expectedValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectedValid, valid)
			}
			
			// Check coupon field
			responseCoupon, ok := response["coupon"].(string)
			if !ok {
				t.Fatalf("coupon field is not a string")
			}
			
			if responseCoupon != tt.couponCode {
				t.Errorf("expected coupon=%q, got coupon=%q", tt.couponCode, responseCoupon)
			}
		})
	}
}

func TestCouponHandler_GetStats(t *testing.T) {
	mockVal := &mockValidator{
		validCoupons: map[string]bool{},
	}
	
	handler := NewCouponHandler(mockVal)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/coupon/stats", nil)
	rr := httptest.NewRecorder()
	
	// Execute handler
	handler.GetStats(rr, req)
	
	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	
	// Parse response
	var stats map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	
	// Verify stats content
	totalFiles, ok := stats["total_files"].(float64)
	if !ok {
		t.Fatalf("total_files is not a number")
	}
	
	if int(totalFiles) != 3 {
		t.Errorf("expected total_files=3, got %v", totalFiles)
	}
	
	totalCoupons, ok := stats["total_coupons"].(float64)
	if !ok {
		t.Fatalf("total_coupons is not a number")
	}
	
	if int(totalCoupons) != 450 {
		t.Errorf("expected total_coupons=450, got %v", totalCoupons)
	}
}
