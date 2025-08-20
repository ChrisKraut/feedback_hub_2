package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"feedback_hub_2/internal/shared/web"

	_ "feedback_hub_2/docs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents a configured HTTP server with all dependencies initialized.
// AI-hint: Public API wrapper that encapsulates the internal application structure.
// This allows external packages like Vercel functions to use the application without
// directly importing internal packages.
type Server struct {
	dbPool      *pgxpool.Pool
	initialized bool
}

// NewServer creates a new Server instance but doesn't initialize it yet.
// AI-hint: Factory method for server creation. Initialization is done separately
// to allow for proper error handling in serverless environments.
func NewServer() *Server {
	return &Server{}
}

// Initialize sets up all dependencies and ensures the server is ready to handle requests.
// AI-hint: One-time initialization that sets up database connections and basic services.
// Safe to call multiple times (idempotent).
func (s *Server) Initialize(ctx context.Context) error {
	if s.initialized {
		return nil
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Log database connection attempt (without exposing sensitive credentials)
	log.Printf("Attempting database connection (URL length: %d chars, first 10: %s)", len(dbURL), dbURL[:min(10, len(dbURL))])

	// Create database connection pool
	var err error
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Disable prepared statement caching for serverless environments
	// This prevents "prepared statement already exists" errors
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	s.dbPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	// Test the connection
	if err := s.dbPool.Ping(ctx); err != nil {
		return err
	}

	s.initialized = true
	return nil
}

// Close cleans up resources.
// AI-hint: Cleanup method for graceful shutdown, primarily closes database connections.
func (s *Server) Close() {
	if s.dbPool != nil {
		s.dbPool.Close()
	}
}

// rootHandler provides API information at the root endpoint
// AI-hint: Simple endpoint that provides API metadata and documentation links.
//
// @Summary Get API information
// @Description Get basic API information and status
// @Tags root
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Feedback Hub API","version":"1.0","status":"ready","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// statusHandler provides API status information
// AI-hint: Status endpoint for monitoring and health checks.
//
// @Summary Get API status
// @Description Get current API status and timestamp
// @Tags status
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/status [get]
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"API is running","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// Handler creates and returns the complete HTTP handler for the application.
// AI-hint: Main HTTP handler that sets up all routes and middleware.
// This is the entry point for both regular servers and serverless functions.
func (s *Server) Handler() http.Handler {
	if !s.initialized {
		panic("server not initialized - call Initialize() first")
	}

	mux := http.NewServeMux()

	// AI-hint: Root handler provides API information
	mux.HandleFunc("/", s.rootHandler)
	// AI-hint: Health endpoint for monitoring
	mux.HandleFunc("/healthz", web.HealthHandler)
	// AI-hint: Swagger UI for API documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// AI-hint: Basic API endpoints (placeholder for now)
	mux.HandleFunc("/api/status", s.statusHandler)

	// AI-hint: Return the configured mux
	return mux
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
