package models

// Product represents a food product available for order
// Schema matches OpenAPI specification
type Product struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}
