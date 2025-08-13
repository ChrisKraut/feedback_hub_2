// @title Feedback Hub API
// @version 1.0
// @description API documentation for Feedback Hub.
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"feedback_hub_2/internal/infrastructure/persistence"
	httpiface "feedback_hub_2/internal/interfaces/http"
	"feedback_hub_2/pkg/config"

	_ "feedback_hub_2/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// rootHandler responds to the root path with a simple welcome message.
// AI-hint: Keep this handler minimal for health checks and smoke tests; expand with routing/middleware in future tickets.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Welcome"))
}

// main initializes and starts the HTTP server.
// AI-hint: Reads PORT from environment (defaults to 8080), sets conservative timeouts, and starts the server. Future work: structured logging, graceful shutdown with context, and DI for handlers.
func main() {
	// AI-hint: Load environment from .env for local dev; safe no-op in production.
	if err := config.Load(); err != nil {
		log.Printf("config load warning: %v", err)
	}

	// AI-hint: Initialize database connection pool early and verify connectivity with Ping.
	dbURL := config.DatabaseURL()
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set; cannot start without database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := persistence.NewPostgresPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Printf("database connection established")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	// AI-hint: Root handler is minimal; in future, migrate to a router and middleware stack.
	mux.HandleFunc("/", rootHandler)
	// AI-hint: Health endpoint is part of interfaces layer to keep transport concerns separate from domain logic.
	mux.HandleFunc("/healthz", httpiface.HealthHandler)
	// AI-hint: Serve Swagger UI for interactive API docs at /swagger/.
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
