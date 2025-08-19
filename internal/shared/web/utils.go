package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorResponse represents the standard error response format.
// AI-hint: Consistent error response structure for all API endpoints.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// WriteErrorResponse writes a standardized error response to the HTTP response writer.
// AI-hint: Centralized error response helper for consistent API error handling.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// ExtractIDFromPath extracts an ID from a URL path given a prefix.
// AI-hint: URL path parsing helper for RESTful resource ID extraction.
// Example: ExtractIDFromPath("/users/123", "/users/") returns "123"
func ExtractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	id := strings.TrimPrefix(path, prefix)
	// Remove any trailing slashes or query parameters
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	if idx := strings.Index(id, "?"); idx != -1 {
		id = id[:idx]
	}

	return strings.TrimSpace(id)
}

// ContextKey type for context keys to avoid collisions.
// AI-hint: Type-safe context key for storing user information.
type ContextKey string

const (
	// UserIDContextKey is the context key for storing the authenticated user ID.
	UserIDContextKey ContextKey = "user_id"
)

// GetUserIDFromContext retrieves the user ID from the request context.
// AI-hint: Context extraction helper for authenticated user identification.
// Returns empty string if no user ID is found in context.
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDContextKey).(string); ok {
		return userID
	}
	return ""
}

// SetUserIDInContext adds the user ID to the request context.
// AI-hint: Context injection helper for authentication middleware.
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}
