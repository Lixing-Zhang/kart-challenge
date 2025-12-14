package handlers

import (
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
		WriteError(w, http.StatusInternalServerError, "Internal server error", h.logger)
		return
	}

	WriteJSON(w, http.StatusOK, products, h.logger)
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
		WriteError(w, http.StatusBadRequest, "Invalid ID supplied", h.logger)
		return
	}

	// Validate that productId is numeric and convert to int64
	productIDInt, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		h.logger.Warn("invalid product ID format", "productId", productID, "error", err)
		WriteError(w, http.StatusBadRequest, "Invalid ID supplied", h.logger)
		return
	}

	// Validate that productId is positive
	if productIDInt <= 0 {
		h.logger.Warn("product ID must be positive", "productId", productIDInt)
		WriteError(w, http.StatusBadRequest, "Invalid ID supplied", h.logger)
		return
	}
	product, err := h.service.GetProduct(ctx, productIDInt)
	if err != nil {
		if err == repository.ErrProductNotFound {
			h.logger.Info("product not found", "productId", productID)
			WriteError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}

		h.logger.Error("failed to get product", "productId", productID, "error", err)
		WriteError(w, http.StatusInternalServerError, "Internal server error", h.logger)
		return
	}

	WriteJSON(w, http.StatusOK, product, h.logger)
}
