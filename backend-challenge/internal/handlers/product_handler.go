package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/repository"
	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/service"
	"github.com/go-chi/chi/v5"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service *service.ProductService
	logger  *slog.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(service *service.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// ListProducts handles GET /api/product
// Returns all available products as per OpenAPI spec
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	products, err := h.service.ListProducts(ctx)
	if err != nil {
		h.logger.Error("failed to list products", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.writeJSON(w, http.StatusOK, products)
}

// GetProduct handles GET /api/product/{productId}
// Returns a single product or error as per OpenAPI spec:
// - 200: successful operation
// - 400: Invalid ID supplied
// - 404: Product not found
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productID := chi.URLParam(r, "productId")

	// Validate that productId is provided
	if productID == "" {
		h.logger.Warn("product ID is required")
		h.writeError(w, http.StatusBadRequest, "Invalid ID supplied")
		return
	}

	// Validate that productId is numeric (as per OpenAPI spec: type: integer, format: int64)
	if _, err := strconv.ParseInt(productID, 10, 64); err != nil {
		h.logger.Warn("invalid product ID format", "productId", productID, "error", err)
		h.writeError(w, http.StatusBadRequest, "Invalid ID supplied")
		return
	}

	product, err := h.service.GetProduct(ctx, productID)
	if err != nil {
		if err == repository.ErrProductNotFound {
			h.logger.Info("product not found", "productId", productID)
			h.writeError(w, http.StatusNotFound, "Product not found")
			return
		}

		h.logger.Error("failed to get product", "productId", productID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// writeJSON writes a JSON response
func (h *ProductHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

// writeError writes an error response
func (h *ProductHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode error response", "error", err)
	}
}
