package main

import (
	"context"
	"feedback_hub_2/pkg/api"
	"log"
	"net/http"
	"os"
	"time"
)

// Global variables for connection pooling in serverless environment
var (
	server   *api.Server
	initOnce bool
)

// initializeServices sets up database connection and services once per cold start
func initializeServices() error {
	if initOnce {
		return nil
	}

	log.Printf("Initializing services...")
	log.Printf("Environment: %s", os.Getenv("VERCEL_ENV"))
	log.Printf("Go version: %s", os.Getenv("GOVERSION"))

	// Create and initialize the server
	server = api.NewServer()
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	if err := server.Initialize(initCtx); err != nil {
		log.Printf("Failed to initialize server: %v", err)
		return err
	}

	log.Printf("Services initialized successfully")
	initOnce = true
	return nil
}

// Handler is the main Vercel serverless function handler
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize services on first request
	if err := initializeServices(); err != nil {
		log.Printf("Failed to initialize services: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Service initialization failed","details":"Check environment variables and database connectivity"}`))
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

// main function for local development
func main() {
	// This will only run locally, not on Vercel
	server := api.NewServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the server
	ctx := context.Background()
	if err := server.Initialize(ctx); err != nil {
		panic(err)
	}
	defer server.Close()

	// Start the server
	http.ListenAndServe(":"+port, server.Handler())
}
