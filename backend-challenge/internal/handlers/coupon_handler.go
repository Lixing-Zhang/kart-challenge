package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// couponValidator is the interface for coupon validation
type couponValidator interface {
	IsValid(ctx context.Context, code string) bool
	GetStats() map[string]interface{}
}

// CouponHandler handles HTTP requests for coupon validation
type CouponHandler struct {
	validator couponValidator
}

// NewCouponHandler creates a new CouponHandler
func NewCouponHandler(validator couponValidator) *CouponHandler {
	return &CouponHandler{
		validator: validator,
	}
}

// ValidateCoupon handles GET /api/coupon/{couponCode}
// Validates if the provided coupon code is valid according to the business rules
func (h *CouponHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	couponCode := chi.URLParam(r, "couponCode")
	
	// Validate the coupon
	isValid := h.validator.IsValid(r.Context(), couponCode)
	
	if isValid {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  true,
			"coupon": couponCode,
		})
	} else {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"valid":   false,
			"coupon":  couponCode,
			"message": "Coupon not found or invalid",
		})
	}
}

// GetStats handles GET /api/coupon/stats (for debugging/monitoring)
func (h *CouponHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.validator.GetStats()
	writeJSON(w, http.StatusOK, stats)
}
