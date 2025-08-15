package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"feedback_hub_2/internal/application"
	"feedback_hub_2/internal/domain/auth"
	"feedback_hub_2/internal/infrastructure/persistence"
	authinfra "feedback_hub_2/internal/infrastructure/auth"
	httpiface "feedback_hub_2/internal/interfaces/http"
	_ "feedback_hub_2/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

// AI-hint: Global variables for connection pooling in serverless environment
var (
	dbPool           *pgxpool.Pool
	initOnce         bool
	roleHandler      *httpiface.RoleHandler
	userHandler      *httpiface.UserHandler
	authHandler      *httpiface.AuthHandler
	authMiddleware   *httpiface.AuthMiddleware
)

// initializeServices sets up database connection and services once per cold start
// AI-hint: One-time initialization for serverless environment to reuse connections
func initializeServices() error {
	if initOnce {
		return nil
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Create database connection pool
	var err error
	dbPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return err
	}

	// Test the connection
	if err := dbPool.Ping(context.Background()); err != nil {
		return err
	}

	// Ensure database schema exists
	if err := persistence.EnsureSchema(context.Background(), dbPool); err != nil {
		return err
	}

	// Create repositories
	roleRepo := persistence.NewRoleRepository(dbPool)
	userRepo := persistence.NewUserRepository(dbPool)

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
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	if err := bootstrapService.Initialize(initCtx); err != nil {
		return err
	}

	// Create HTTP handlers
	roleHandler = httpiface.NewRoleHandler(roleService)
	userHandler = httpiface.NewUserHandler(userService)
	authHandler = httpiface.NewAuthHandler(userService, roleService, jwtService, passwordService)

	// Create authentication middleware
	authMiddleware = httpiface.NewAuthMiddleware(userService, jwtService)

	initOnce = true
	return nil
}

// rootHandler provides API information at the root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Feedback Hub API","version":"1.0","docs":"/swagger/"}`))
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

	// Set up CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set up routing
	mux := http.NewServeMux()

	// AI-hint: Root handler provides API information
	mux.HandleFunc("/", rootHandler)
	// AI-hint: Health endpoint for monitoring
	mux.HandleFunc("/healthz", httpiface.HealthHandler)
	// AI-hint: Swagger UI for API documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// AI-hint: Authentication routes (no auth required)
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Register(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Logout(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only POST allowed"}`))
		}
	})

	// AI-hint: Authenticated route to get current user info
	mux.HandleFunc("/auth/me", authMiddleware.RequireAuthFunc(authHandler.Me))

	// AI-hint: Role management routes (authenticated)
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authMiddleware.RequireAuthFunc(roleHandler.ListRoles)(w, r)
		case http.MethodPost:
			authMiddleware.RequireAuthFunc(roleHandler.CreateRole)(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET and POST allowed"}`))
		}
	})

	mux.HandleFunc("/roles/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path (assuming format /roles/{id})
		path := r.URL.Path
		if len(path) > 7 { // "/roles/" = 7 chars
			switch r.Method {
			case http.MethodGet:
				authMiddleware.RequireAuthFunc(roleHandler.GetRole)(w, r)
			case http.MethodPut:
				authMiddleware.RequireAuthFunc(roleHandler.UpdateRole)(w, r)
			case http.MethodDelete:
				authMiddleware.RequireAuthFunc(roleHandler.DeleteRole)(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET, PUT, and DELETE allowed"}`))
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not Found","message":"Role ID required"}`))
		}
	})

	// AI-hint: User management routes (authenticated)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authMiddleware.RequireAuthFunc(userHandler.ListUsers)(w, r)
		case http.MethodPost:
			authMiddleware.RequireAuthFunc(userHandler.CreateUser)(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET and POST allowed"}`))
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := r.URL.Path
		if len(path) > 7 { // "/users/" = 7 chars
			// Check if it's a role update path
			if len(path) > 12 && path[len(path)-5:] == "/role" {
				// Handle /users/{id}/role
				if r.Method == http.MethodPut {
					authMiddleware.RequireAuthFunc(userHandler.UpdateUserRole)(w, r)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusMethodNotAllowed)
					w.Write([]byte(`{"error":"Method Not Allowed","message":"Only PUT allowed"}`))
				}
			} else {
				// Handle /users/{id}
				switch r.Method {
				case http.MethodGet:
					authMiddleware.RequireAuthFunc(userHandler.GetUser)(w, r)
				case http.MethodPut:
					authMiddleware.RequireAuthFunc(userHandler.UpdateUser)(w, r)
				case http.MethodDelete:
					authMiddleware.RequireAuthFunc(userHandler.DeleteUser)(w, r)
				default:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusMethodNotAllowed)
					w.Write([]byte(`{"error":"Method Not Allowed","message":"Only GET, PUT, and DELETE allowed"}`))
				}
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not Found","message":"User ID required"}`))
		}
	})

	// AI-hint: Serve the request using the configured mux
	mux.ServeHTTP(w, r)
}
