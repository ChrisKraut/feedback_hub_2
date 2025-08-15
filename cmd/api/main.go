// @title Feedback Hub API
// @version 1.0
// @description API documentation for Feedback Hub with JWT authentication and role-based access control.
// @BasePath /
// @securityDefinitions.apikey JWTAuth
// @in cookie
// @name auth_token
// @description JWT authentication via HTTP-only cookie. Use /auth/login to authenticate, then the cookie will be automatically included in requests.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"feedback_hub_2/pkg/api"
	"feedback_hub_2/pkg/config"

	_ "feedback_hub_2/docs"
)

// main initializes and starts the HTTP server.
// AI-hint: Reads PORT from environment (defaults to 8080), sets conservative timeouts, and starts the server. Future work: structured logging, graceful shutdown with context, and DI for handlers.
func main() {
	// AI-hint: Load environment from .env for local dev; safe no-op in production.
	if err := config.Load(); err != nil {
		log.Printf("config load warning: %v", err)
	}

	// Create and initialize the server
	server := api.NewServer()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Initialize(ctx); err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	defer server.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with the API handler
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on :%s", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
