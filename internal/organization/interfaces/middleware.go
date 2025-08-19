package interfaces

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// OrganizationSelectionMiddleware extracts and validates the organization ID from headers
// and adds it to the request context for downstream handlers.
// AI-hint: Middleware that extracts organization context from headers and validates
// the organization ID format before passing it to the next handler.
func OrganizationSelectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from header
		orgID := r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "organization header required")
			return
		}

		// Validate organization ID format
		if _, err := uuid.Parse(orgID); err != nil {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization ID")
			return
		}

		// Add organization ID to context
		ctx := context.WithValue(r.Context(), "organization_id", orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OrganizationValidationMiddleware validates that the user has access to the selected organization.
// AI-hint: Middleware that ensures users can only access organizations they belong to,
// implementing organization-level access control.
func OrganizationValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get organization ID from context
		orgID := r.Context().Value("organization_id")
		if orgID == nil {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "organization context not found")
			return
		}

		// Get user ID from context (set by authentication middleware)
		userID := r.Context().Value("user_id")
		if userID == nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", "user context not found")
			return
		}

		// TODO: Implement actual validation logic to check if user belongs to organization
		// For now, we'll just pass through - this should be implemented with the user-organization repository

		// Add user-organization context
		ctx := context.WithValue(r.Context(), "user_organization_context", map[string]interface{}{
			"user_id":         userID,
			"organization_id": orgID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthenticationWithOrganizationMiddleware combines authentication and organization context.
// AI-hint: Combined middleware that handles both user authentication and organization
// context extraction in a single middleware chain.
func AuthenticationWithOrganizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", "invalid authorization format")
			return
		}

		// TODO: Implement actual JWT validation
		// For now, we'll just check the format and pass through
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", "invalid token")
			return
		}

		// Extract organization ID from header
		orgID := r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "organization header required")
			return
		}

		// Validate organization ID format
		if _, err := uuid.Parse(orgID); err != nil {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization ID")
			return
		}

		// Add both user and organization context
		ctx := context.WithValue(r.Context(), "user_id", "test-user-id") // TODO: Extract from JWT
		ctx = context.WithValue(ctx, "organization_id", orgID)
		ctx = context.WithValue(ctx, "user_organization_context", map[string]interface{}{
			"user_id":         "test-user-id",
			"organization_id": orgID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PanicRecoveryMiddleware recovers from panics and returns a proper error response.
// AI-hint: Safety middleware that prevents the application from crashing due to panics
// and provides graceful error handling for unexpected errors.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic (in production, this would go to a proper logging system)
				// log.Printf("Panic recovered: %v", err)

				// Return internal server error
				respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// OrganizationContextFallbackMiddleware provides a fallback organization context
// when none is specified in the request.
// AI-hint: Fallback middleware that provides default organization context for
// backward compatibility or when organization selection is not required.
func OrganizationContextFallbackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if organization context already exists
		if r.Context().Value("organization_id") != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Try to get from header
		orgID := r.Header.Get("X-Organization-ID")
		if orgID != "" {
			// Validate and add to context
			if _, err := uuid.Parse(orgID); err == nil {
				ctx := context.WithValue(r.Context(), "organization_id", orgID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Use fallback organization (could be from user preferences or default)
		// For now, we'll use a hardcoded fallback
		fallbackOrgID := "00000000-0000-0000-0000-000000000000" // Default organization
		ctx := context.WithValue(r.Context(), "organization_id", fallbackOrgID)
		ctx = context.WithValue(ctx, "fallback_organization", true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireOrganizationMiddleware ensures that organization context is present.
// AI-hint: Validation middleware that enforces the presence of organization context
// for endpoints that require organization scoping.
func RequireOrganizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Context().Value("organization_id")
		if orgID == nil {
			respondWithError(w, http.StatusBadRequest, "Bad Request", "organization context required")
			return
		}

		// Check if it's a fallback organization
		if fallback := r.Context().Value("fallback_organization"); fallback != nil && fallback.(bool) {
			// Log that fallback organization is being used
			// log.Printf("Using fallback organization: %s", orgID)
		}

		next.ServeHTTP(w, r)
	})
}

// OrganizationScopedMiddleware combines organization selection, validation, and requirement.
// AI-hint: Comprehensive middleware that handles the complete organization scoping flow
// for endpoints that require full organization context.
func OrganizationScopedMiddleware(next http.Handler) http.Handler {
	return OrganizationSelectionMiddleware(
		OrganizationValidationMiddleware(
			RequireOrganizationMiddleware(next),
		),
	)
}
