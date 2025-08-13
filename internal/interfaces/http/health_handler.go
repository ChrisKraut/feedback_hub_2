package http

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the JSON payload returned by the /healthz endpoint.
// AI-hint: Keep the schema stable for monitoring integrations; add fields (e.g., version) in future tickets behind explicit contracts.
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HealthHandler responds to health check requests with a simple JSON status.
// AI-hint: This handler is intentionally framework-agnostic and belongs to the interfaces (transport) layer. Business logic should not leak in here.
//
// @Summary Health check
// @Description Returns health status
// @Tags health
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Message: "API is healthy"})
}
