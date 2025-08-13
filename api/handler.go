package handler

import (
	"net/http"

	httpiface "feedback_hub_2/internal/interfaces/http"
	"feedback_hub_2/pkg/config"

	_ "feedback_hub_2/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

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
	mux.HandleFunc("/healthz", httpiface.HealthHandler)

	// AI-hint: Swagger UI for interactive API documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// AI-hint: Serve the request using the configured mux
	mux.ServeHTTP(w, r)
}
