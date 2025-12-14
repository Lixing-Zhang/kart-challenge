package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode JSON response", "error", err)
	}
}

// WriteError writes an error response in JSON format
func WriteError(w http.ResponseWriter, status int, message string, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode error response", "error", err)
	}
}
