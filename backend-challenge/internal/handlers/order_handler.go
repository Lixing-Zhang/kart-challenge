package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/models"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/service"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService *service.OrderService
	log          *slog.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *service.OrderService, log *slog.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		log:          log,
	}
}

// CreateOrder handles POST /api/order
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.OrderRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode order request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and create order
	order, err := h.orderService.CreateOrder(r.Context(), req)
	if err != nil {
		h.log.Error("failed to create order", "error", err)

		switch err {
		case service.ErrEmptyOrder:
			http.Error(w, "Order must contain at least one item", http.StatusBadRequest)
		case service.ErrInvalidQuantity:
			http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		case service.ErrInvalidProduct:
			http.Error(w, "Invalid product", http.StatusBadRequest)
		case service.ErrInvalidCoupon:
			http.Error(w, "Coupon code is not valid", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.log.Error("failed to encode order response", "error", err)
		// Cannot send error response - headers already written
		return
	}

	h.log.Info("order created successfully", "order_id", order.ID, "items_count", len(order.Items))
}
