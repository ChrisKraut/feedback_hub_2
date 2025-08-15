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
	"strings"
	"time"

	"feedback_hub_2/internal/application"
	"feedback_hub_2/internal/domain/auth"
	authinfra "feedback_hub_2/internal/infrastructure/auth"
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

	// AI-hint: Ensure database schema exists before proceeding
	schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer schemaCancel()
	if err := persistence.EnsureSchema(schemaCtx, pool); err != nil {
		log.Fatalf("failed to ensure database schema: %v", err)
	}

	// AI-hint: Initialize domain services and repositories following DDD pattern
	// Create repositories
	roleRepo := persistence.NewRoleRepository(pool)
	userRepo := persistence.NewUserRepository(pool)

	// Create authorization service
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
		log.Fatalf("failed to initialize system: %v", err)
	}

	// Create HTTP handlers
	roleHandler := httpiface.NewRoleHandler(roleService)
	userHandler := httpiface.NewUserHandler(userService)
	authHandler := httpiface.NewAuthHandler(userService, roleService, jwtService, passwordService)

	// Create authentication middleware
	authMiddleware := httpiface.NewAuthMiddleware(userService, jwtService)

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

	// AI-hint: Authentication routes (no auth required for these)
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

	// AI-hint: Role management API endpoints with authentication
	mux.HandleFunc("/roles", authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			roleHandler.ListRoles(w, r)
		case http.MethodPost:
			roleHandler.CreateRole(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Method not allowed"}`))
		}
	}))

	mux.HandleFunc("/roles/", authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			roleHandler.GetRole(w, r)
		case http.MethodPut:
			roleHandler.UpdateRole(w, r)
		case http.MethodDelete:
			roleHandler.DeleteRole(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Method not allowed"}`))
		}
	}))

	// AI-hint: User management API endpoints with authentication
	mux.HandleFunc("/users", authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Method not allowed"}`))
		}
	}))

	mux.HandleFunc("/users/", authMiddleware.RequireAuthFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a role update endpoint
		if strings.HasSuffix(r.URL.Path, "/role") && r.Method == http.MethodPut {
			userHandler.UpdateUserRole(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method Not Allowed","message":"Method not allowed"}`))
		}
	}))

	// AI-hint: Apply CORS and logging middleware to all routes
	handler := httpiface.CORS(httpiface.LoggingMiddleware(mux))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
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
