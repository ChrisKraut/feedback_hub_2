package http

import (
	"feedback_hub_2/internal/application"
	"net/http"
)

// AuthMiddleware provides authentication functionality for HTTP requests.
// AI-hint: Authentication middleware that validates user identity and adds user context.
// In production, this would validate JWT tokens or session cookies.
type AuthMiddleware struct {
	userService *application.UserService
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
// AI-hint: Factory method for auth middleware with dependency injection of user service.
func NewAuthMiddleware(userService *application.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// RequireAuth is a middleware that requires authentication for the wrapped handler.
// AI-hint: Authentication middleware that validates user identity from headers.
// For simplicity, this implementation uses a "X-User-ID" header (in production, use JWT/sessions).
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from header (simplified authentication)
		// In production, this would validate JWT tokens or session cookies
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate that the user exists
		user, err := m.userService.GetUser(r.Context(), userID)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid authentication")
			return
		}

		// Add user ID to request context
		ctx := setUserIDInContext(r.Context(), user.ID)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// RequireAuthFunc is a middleware function that requires authentication for the wrapped handler function.
// AI-hint: Function wrapper version of RequireAuth for direct handler function wrapping.
func (m *AuthMiddleware) RequireAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from header (simplified authentication)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate that the user exists
		user, err := m.userService.GetUser(r.Context(), userID)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid authentication")
			return
		}

		// Add user ID to request context
		ctx := setUserIDInContext(r.Context(), user.ID)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

// CORS middleware to handle Cross-Origin Resource Sharing.
// AI-hint: CORS middleware for browser-based API access and development.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests for debugging and monitoring.
// AI-hint: Request logging middleware for development and debugging.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple logging - in production, use structured logging
		method := r.Method
		path := r.URL.Path
		userAgent := r.Header.Get("User-Agent")

		// Extract user ID if available for logging
		userID := getUserIDFromContext(r.Context())
		if userID == "" {
			userID = "anonymous"
		}

		// Log the request (in production, use proper logger)
		println(method, path, userAgent, "user:", userID)

		next.ServeHTTP(w, r)
	})
}
