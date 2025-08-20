package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"feedback_hub_2/internal/shared/web"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Feedback Hub API","version":"1.0","status":"ready","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
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

	// AI-hint: Basic API endpoints (placeholder for now)
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"API is running","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// AI-hint: Swagger endpoint temporarily disabled until docs are regenerated
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		// Check if they want the raw JSON or the UI
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"message": "API Documentation",
				"status": "available",
				"current_endpoints": [
					{
						"path": "/",
						"method": "GET",
						"description": "API root information and status"
					},
					{
						"path": "/healthz",
						"method": "GET", 
						"description": "Health check endpoint"
					},
					{
						"path": "/api/status",
						"method": "GET",
						"description": "API status and timestamp"
					}
				],
				"note": "Swagger UI documentation is being updated for the new API structure. Use the endpoints above to interact with the API."
			}`))
			return
		}

		// Provide a nice HTML interface
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Feedback Hub API Documentation</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 30px; }
        .endpoint { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 6px; padding: 20px; margin-bottom: 20px; }
        .method { display: inline-block; background: #007bff; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; margin-right: 10px; }
        .path { font-family: monospace; font-size: 16px; color: #495057; }
        .description { color: #6c757d; margin-top: 10px; }
        .note { background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 6px; padding: 15px; margin-top: 20px; color: #856404; }
        .json-link { margin-top: 20px; }
        .json-link a { color: #007bff; text-decoration: none; }
        .json-link a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Feedback Hub API Documentation</h1>
        
        <div class="endpoint">
            <span class="method">GET</span>
            <span class="path">/</span>
            <div class="description">API root information and status</div>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span>
            <span class="path">/healthz</span>
            <div class="description">Health check endpoint for monitoring</div>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span>
            <span class="path">/api/status</span>
            <div class="description">API status and current timestamp</div>
        </div>
        
        <div class="note">
            <strong>Note:</strong> Swagger UI documentation is being updated for the new API structure. 
            Use the endpoints above to interact with the API.
        </div>
        
        <div class="json-link">
            <a href="/swagger/?format=json">View API documentation as JSON</a>
        </div>
    </div>
</body>
</html>`))
	})

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
