package middleware

import (
	"net/http"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/config"
)

// APIKeyAuth middleware validates API key from header
// According to OpenAPI spec, API key is passed in "api_key" header
func APIKeyAuth(cfg config.AuthConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("api_key")

			if apiKey == "" {
				http.Error(w, "Unauthorized: API key required", http.StatusUnauthorized)
				return
			}

			// Validate API key
			valid := false
			for _, validKey := range cfg.APIKeys {
				if apiKey == validKey {
					valid = true
					break
				}
			}

			if !valid {
				http.Error(w, "Forbidden: Invalid API key", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
