package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lixing-Zhang/kart-challenge/backend-challenge/internal/config"
)

func TestAPIKeyAuth(t *testing.T) {
	cfg := config.AuthConfig{
		APIKeys: []string{"apitest", "testkey123"},
	}

	// Create a test handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Wrap with auth middleware
	authHandler := APIKeyAuth(cfg)(testHandler)

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid API key - apitest",
			apiKey:         "apitest",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid API key - testkey123",
			apiKey:         "testkey123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid API key",
			apiKey:         "wrongkey",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/order", nil)
			if tt.apiKey != "" {
				req.Header.Set("api_key", tt.apiKey)
			}

			w := httptest.NewRecorder()
			authHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				if w.Body.String() != "success" {
					t.Errorf("body = %s, want success", w.Body.String())
				}
			}
		})
	}
}
