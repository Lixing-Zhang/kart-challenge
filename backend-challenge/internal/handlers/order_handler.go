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
		WriteError(w, http.StatusBadRequest, "Invalid request body", h.log)
		return
	}

	// Validate and create order
	order, err := h.orderService.CreateOrder(r.Context(), req)
	if err != nil {
		h.log.Error("failed to create order", "error", err)

		switch err {
		case service.ErrEmptyOrder:
			WriteError(w, http.StatusBadRequest, "Order must contain at least one item", h.log)
		case service.ErrInvalidQuantity:
			WriteError(w, http.StatusBadRequest, "Quantity must be positive", h.log)
		case service.ErrInvalidProduct:
			WriteError(w, http.StatusBadRequest, "Invalid product", h.log)
		case service.ErrInvalidCoupon:
			WriteError(w, http.StatusBadRequest, "Coupon code is not valid", h.log)
		default:
			WriteError(w, http.StatusInternalServerError, "Internal server error", h.log)
		}
		return
	}

	// Return successful response
	WriteJSON(w, http.StatusOK, order, h.log)
	h.log.Info("order created successfully", "order_id", order.ID, "items_count", len(order.Items))
}
