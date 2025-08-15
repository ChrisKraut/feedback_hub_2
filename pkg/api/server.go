package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"feedback_hub_2/internal/application"
	"feedback_hub_2/internal/domain/auth"
	authinfra "feedback_hub_2/internal/infrastructure/auth"
	"feedback_hub_2/internal/infrastructure/persistence"
	httpiface "feedback_hub_2/internal/interfaces/http"

	_ "feedback_hub_2/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents a configured HTTP server with all dependencies initialized.
// AI-hint: Public API wrapper that encapsulates the internal application structure.
// This allows external packages like Vercel functions to use the application without
// directly importing internal packages.
type Server struct {
	dbPool         *pgxpool.Pool
	roleHandler    *httpiface.RoleHandler
	userHandler    *httpiface.UserHandler
	authHandler    *httpiface.AuthHandler
	authMiddleware *httpiface.AuthMiddleware
	initialized    bool
}

// NewServer creates a new Server instance but doesn't initialize it yet.
// AI-hint: Factory method for server creation. Initialization is done separately
// to allow for proper error handling in serverless environments.
func NewServer() *Server {
	return &Server{}
}

// Initialize sets up all dependencies and ensures the server is ready to handle requests.
// AI-hint: One-time initialization that sets up database connections, repositories,
// services, and handlers. Safe to call multiple times (idempotent).
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
	s.dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return err
	}

	// Test the connection
	if err := s.dbPool.Ping(ctx); err != nil {
		return err
	}

	// Ensure database schema exists
	if err := persistence.EnsureSchema(ctx, s.dbPool); err != nil {
		return err
	}

	// Create repositories
	roleRepo := persistence.NewRoleRepository(s.dbPool)
	userRepo := persistence.NewUserRepository(s.dbPool)

	// Create domain services
	authService := auth.NewAuthorizationService()

	// Create authentication services
	jwtService := authinfra.NewJWTService()
	passwordService := authinfra.NewPasswordService()

	// Create application services
	roleService := application.NewRoleService(roleRepo, userRepo, authService)
	userService := application.NewUserService(userRepo, roleRepo, authService)

	// Create bootstrap service and initialize system
	bootstrapService := application.NewBootstrapService(roleService, userService)
	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	defer initCancel()

	if err := bootstrapService.Initialize(initCtx); err != nil {
		return err
	}

	// Create HTTP handlers
	s.roleHandler = httpiface.NewRoleHandler(roleService)
	s.userHandler = httpiface.NewUserHandler(userService)
	s.authHandler = httpiface.NewAuthHandler(userService, roleService, jwtService, passwordService)

	// Create authentication middleware
	s.authMiddleware = httpiface.NewAuthMiddleware(userService, jwtService)

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
	w.Write([]byte(`{"message":"Feedback Hub API","version":"1.0","docs":"/swagger/"}`))
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
	mux.HandleFunc("/healthz", httpiface.HealthHandler)
	// AI-hint: Swagger UI for API documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// AI-hint: Authentication routes (no auth required)
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.authHandler.Login(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.authHandler.Register(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.authHandler.Logout(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	// AI-hint: Authenticated route to get current user info
	mux.HandleFunc("/auth/me", s.authMiddleware.RequireAuthFunc(s.authHandler.Me))

	// AI-hint: Role management routes (authenticated)
	mux.HandleFunc("/roles", s.authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.roleHandler.ListRoles(w, r)
		case http.MethodPost:
			s.roleHandler.CreateRole(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET and POST allowed"}`))
		}
	}))

	mux.HandleFunc("/roles/", s.authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.roleHandler.GetRole(w, r)
		case http.MethodPut:
			s.roleHandler.UpdateRole(w, r)
		case http.MethodDelete:
			s.roleHandler.DeleteRole(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET, PUT, and DELETE allowed"}`))
		}
	}))

	// AI-hint: User management routes (authenticated)
	mux.HandleFunc("/users", s.authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.userHandler.ListUsers(w, r)
		case http.MethodPost:
			s.userHandler.CreateUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET and POST allowed"}`))
		}
	}))

	mux.HandleFunc("/users/", s.authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a role update endpoint
		if strings.HasSuffix(r.URL.Path, "/role") && r.Method == http.MethodPut {
			s.userHandler.UpdateUserRole(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.userHandler.GetUser(w, r)
		case http.MethodPut:
			s.userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			s.userHandler.DeleteUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET, PUT, and DELETE allowed"}`))
		}
	}))

	// AI-hint: Apply CORS middleware to enable cross-origin requests
	return httpiface.CORS(mux)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
