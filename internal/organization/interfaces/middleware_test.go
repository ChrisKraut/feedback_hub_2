package interfaces

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOrganizationSelectionMiddleware tests the organization selection middleware
func TestOrganizationSelectionMiddleware(t *testing.T) {
	t.Run("missing organization header", func(t *testing.T) {
		// Create a simple handler that just returns success
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware (we'll implement this later)
		middleware := OrganizationSelectionMiddleware(nextHandler)

		// Create request without organization header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should return 400 Bad Request for missing organization
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "organization header required")
	})

	t.Run("invalid organization header", func(t *testing.T) {
		// Create a simple handler that just returns success
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := OrganizationSelectionMiddleware(nextHandler)

		// Create request with invalid organization header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Organization-ID", "invalid-uuid")
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should return 400 Bad Request for invalid organization ID
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})

	t.Run("valid organization header", func(t *testing.T) {
		// Create a simple handler that checks for organization context
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if organization context was set
			orgID := r.Context().Value("organization_id")
			if orgID == nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("organization context not set"))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := OrganizationSelectionMiddleware(nextHandler)

		// Create request with valid organization header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Organization-ID", "123e4567-e89b-12d3-a456-426614174000")
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should pass through to next handler
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

// TestOrganizationValidationMiddleware tests the organization validation middleware
func TestOrganizationValidationMiddleware(t *testing.T) {
	t.Run("user not in organization", func(t *testing.T) {
		// Create a simple handler
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := OrganizationValidationMiddleware(nextHandler)

		// Create request with organization context but user not in org
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := req.Context()
		// Set up organization context but no user context
		ctx = context.WithValue(ctx, "organization_id", "123e4567-e89b-12d3-a456-426614174000")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should return 401 Unauthorized for missing user context
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "user context not found")
	})

	t.Run("user in organization", func(t *testing.T) {
		// Create a simple handler
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := OrganizationValidationMiddleware(nextHandler)

		// Create request with valid organization and user context
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, "organization_id", "123e4567-e89b-12d3-a456-426614174000")
		ctx = context.WithValue(ctx, "user_id", "test-user-id")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should pass through to next handler
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

// TestAuthenticationWithOrganizationContext tests authentication with organization context
func TestAuthenticationWithOrganizationContext(t *testing.T) {
	t.Run("missing authentication", func(t *testing.T) {
		// Create a simple handler
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := AuthenticationWithOrganizationMiddleware(nextHandler)

		// Create request without authentication
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authentication required")
	})

	t.Run("valid authentication with organization", func(t *testing.T) {
		// Create a simple handler
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware
		middleware := AuthenticationWithOrganizationMiddleware(nextHandler)

		// Create request with valid authentication and organization
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Organization-ID", "123e4567-e89b-12d3-a456-426614174000")
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should pass through to next handler
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

// TestErrorHandlingAndFallbacks tests error handling and fallbacks in middleware
func TestErrorHandlingAndFallbacks(t *testing.T) {
	t.Run("middleware panic recovery", func(t *testing.T) {
		// Create a handler that panics
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		// Create middleware with panic recovery
		middleware := PanicRecoveryMiddleware(nextHandler)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Call middleware - should not panic
		assert.NotPanics(t, func() {
			middleware.ServeHTTP(w, req)
		})

		// Should return 500 Internal Server Error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")
	})

	t.Run("organization context fallback", func(t *testing.T) {
		// Create a simple handler
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Create middleware with fallback
		middleware := OrganizationContextFallbackMiddleware(nextHandler)

		// Create request without organization header
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Call middleware
		middleware.ServeHTTP(w, req)

		// Should use fallback organization and pass through
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}
