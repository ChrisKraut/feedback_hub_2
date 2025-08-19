package interfaces

import (
	"feedback_hub_2/internal/shared/web"
	userapp "feedback_hub_2/internal/user/application"
	"feedback_hub_2/internal/user/infrastructure/auth"
	"net/http"
)

// AuthMiddleware provides authentication functionality for HTTP requests.
// AI-hint: JWT-based authentication middleware that validates tokens from HTTP-only cookies.
// Provides secure authentication with proper token validation and user context.
type AuthMiddleware struct {
	userService *userapp.UserService
	jwtService  *auth.JWTService
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
// AI-hint: Factory method for auth middleware with dependency injection of services.
func NewAuthMiddleware(userService *userapp.UserService, jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		jwtService:  jwtService,
	}
}

// RequireAuth is a middleware that requires authentication for the wrapped handler.
// AI-hint: JWT-based authentication middleware that validates tokens from HTTP-only cookies.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get JWT token from HTTP-only cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate JWT token
		claims, err := m.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authentication token")
			return
		}

		// Verify that the user still exists (important for user deletion/deactivation)
		user, err := m.userService.GetUser(r.Context(), claims.UserID)
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Add user ID to request context
		ctx := web.SetUserIDInContext(r.Context(), user.ID)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// RequireAuthFunc is a middleware function that requires authentication for the wrapped handler function.
// AI-hint: JWT-based function wrapper version of RequireAuth for direct handler function wrapping.
func (m *AuthMiddleware) RequireAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get JWT token from HTTP-only cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate JWT token
		claims, err := m.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authentication token")
			return
		}

		// Verify that the user still exists
		user, err := m.userService.GetUser(r.Context(), claims.UserID)
		if err != nil {
			web.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Add user ID to request context
		ctx := web.SetUserIDInContext(r.Context(), user.ID)
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
		userID := web.GetUserIDFromContext(r.Context())
		if userID == "" {
			userID = "anonymous"
		}

		// Log the request (in production, use proper logger)
		println(method, path, userAgent, "user:", userID)

		next.ServeHTTP(w, r)
	})
}
