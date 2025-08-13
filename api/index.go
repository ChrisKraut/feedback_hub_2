package handler

import (
	"encoding/json"
	"net/http"

	"feedback_hub_2/pkg/config"
	// _ "feedback_hub_2/docs" // Temporarily commented out
	// httpSwagger "github.com/swaggo/http-swagger" // Temporarily commented out
)

// HealthResponse represents the JSON payload returned by the /healthz endpoint.
// AI-hint: Duplicated here to avoid importing internal packages in serverless context.
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// healthHandler responds to health check requests with a simple JSON status.
// AI-hint: Kept minimal and framework-agnostic for serverless usage.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Message: "API is healthy"})
}

// Handler is the main Vercel serverless function handler.
// AI-hint: This function handles all HTTP requests in the Vercel serverless environment.
// It sets up routing and middleware similar to the main.go server but as a single handler function.
// Database initialization is omitted for serverless efficiency - add back when database endpoints are needed.
func Handler(w http.ResponseWriter, r *http.Request) {
	// AI-hint: Load environment configuration for Vercel deployment
	if err := config.Load(); err != nil {
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}

	// AI-hint: Set up routing using http.ServeMux to handle different endpoints
	mux := http.NewServeMux()

	// AI-hint: Root handler provides a simple welcome message
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Welcome"))
	})

	// AI-hint: Health endpoint for monitoring and load balancer checks
	mux.HandleFunc("/healthz", healthHandler)

	// AI-hint: Swagger UI for interactive API documentation
	// mux.Handle("/swagger/", httpSwagger.WrapHandler) // Temporarily commented out

	// AI-hint: Serve the request using the configured mux
	mux.ServeHTTP(w, r)
}
