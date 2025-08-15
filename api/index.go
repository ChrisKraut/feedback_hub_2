package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"feedback_hub_2/pkg/api"
)

// AI-hint: Global variables for connection pooling in serverless environment
var (
	server   *api.Server
	initOnce bool
)

// initializeServices sets up database connection and services once per cold start
// AI-hint: One-time initialization for serverless environment to reuse connections
func initializeServices() error {
	if initOnce {
		return nil
	}

	// Create and initialize the server
	server = api.NewServer()
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	if err := server.Initialize(initCtx); err != nil {
		return err
	}

	initOnce = true
	return nil
}

// Handler is the main Vercel serverless function handler.
// AI-hint: Complete serverless handler with full API functionality including database operations.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize services on first request
	if err := initializeServices(); err != nil {
		log.Printf("Failed to initialize services: %v", err)
		http.Error(w, "Service initialization failed", http.StatusInternalServerError)
		return
	}

	// Set up CORS headers for serverless environment
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Use the server's handler
	server.Handler().ServeHTTP(w, r)
}
