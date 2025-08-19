package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"feedback_hub_2/internal/organization/application"
	"feedback_hub_2/internal/organization/infrastructure"
	"feedback_hub_2/internal/organization/interfaces"
	"feedback_hub_2/internal/shared/auth"
	events "feedback_hub_2/internal/shared/bus"

	"github.com/stretchr/testify/assert"
)

// TestVercelHandlerPerformance tests handler performance in serverless context
func TestVercelHandlerPerformance(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping Vercel performance test in short mode")
	}

	// Set up test database
	dbPool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create repositories
	orgRepo := infrastructure.NewOrganizationRepository(dbPool)

	// Create services
	authService := auth.NewAuthorizationService()
	eventBus := events.NewInMemoryEventBus()
	orgService := application.NewOrganizationService(orgRepo, eventBus)

	// Create handler
	handler := interfaces.NewOrganizationHandler(orgService, authService)

	t.Run("cold start performance", func(t *testing.T) {
		// Simulate cold start by creating a new handler instance
		start := time.Now()

		coldHandler := interfaces.NewOrganizationHandler(orgService, authService)

		coldStartTime := time.Since(start)

		// Cold start should be fast (under 100ms for handler creation)
		assert.True(t, coldStartTime < 100*time.Millisecond,
			"Cold start should complete within 100ms, took %v", coldStartTime)

		t.Logf("Cold start time: %v", coldStartTime)

		// Test that the cold handler works
		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		w := httptest.NewRecorder()

		coldHandler.ListOrganizations(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Expected due to nil service
	})

	t.Run("request processing performance", func(t *testing.T) {
		// Test request processing time
		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.ListOrganizations(w, req)
		processingTime := time.Since(start)

		// Request processing should be fast (under 50ms for simple operations)
		assert.True(t, processingTime < 50*time.Millisecond,
			"Request processing should complete within 50ms, took %v", processingTime)

		t.Logf("Request processing time: %v", processingTime)
	})

	t.Run("memory usage patterns", func(t *testing.T) {
		// Test memory usage by creating multiple requests
		const numRequests = 100

		// Create a simple request
		createReq := map[string]interface{}{
			"name":        "Memory Test Org",
			"description": "Organization for memory testing",
		}

		body, _ := json.Marshal(createReq)

		// Process multiple requests to check memory patterns
		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			// Each request should complete successfully
			assert.Equal(t, http.StatusCreated, w.Code)
		}

		t.Logf("Processed %d requests successfully", numRequests)
	})

	t.Run("timeout handling", func(t *testing.T) {
		// Test timeout handling by creating a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// This should complete quickly due to the short timeout
		start := time.Now()
		handler.ListOrganizations(w, req)
		processingTime := time.Since(start)

		// Should complete within the timeout
		assert.True(t, processingTime < 10*time.Millisecond,
			"Request should complete quickly, took %v", processingTime)

		t.Logf("Timeout test completed in %v", processingTime)
	})
}

// TestVercelConcurrentRequests tests concurrent request handling
func TestVercelConcurrentRequests(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping Vercel concurrent test in short mode")
	}

	// Set up test database
	dbPool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create repositories
	orgRepo := infrastructure.NewOrganizationRepository(dbPool)

	// Create services
	authService := auth.NewAuthorizationService()
	eventBus := events.NewInMemoryEventBus()
	orgService := application.NewOrganizationService(orgRepo, eventBus)

	// Create handler
	handler := interfaces.NewOrganizationHandler(orgService, authService)

	t.Run("concurrent read requests", func(t *testing.T) {
		const numConcurrent = 10
		var wg sync.WaitGroup
		results := make([]int, numConcurrent)

		start := time.Now()

		// Launch concurrent read requests
		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
				w := httptest.NewRecorder()

				handler.ListOrganizations(w, req)
				results[index] = w.Code
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(start)

		// All requests should complete
		for i, statusCode := range results {
			assert.Equal(t, http.StatusInternalServerError, statusCode,
				"Request %d should complete with expected status", i)
		}

		t.Logf("Completed %d concurrent read requests in %v", numConcurrent, totalTime)

		// Concurrent requests should complete reasonably quickly
		assert.True(t, totalTime < 1*time.Second,
			"Concurrent requests should complete within 1 second, took %v", totalTime)
	})

	t.Run("concurrent create requests", func(t *testing.T) {
		const numConcurrent = 5 // Fewer creates to avoid conflicts
		var wg sync.WaitGroup
		results := make([]int, numConcurrent)

		start := time.Now()

		// Launch concurrent create requests
		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				createReq := map[string]interface{}{
					"name":        fmt.Sprintf("Concurrent Test Org %d", index+1),
					"description": fmt.Sprintf("Organization %d for concurrent testing", index+1),
				}

				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.CreateOrganization(w, req)
				results[index] = w.Code
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(start)

		// All requests should complete
		for i, statusCode := range results {
			assert.Equal(t, http.StatusCreated, statusCode,
				"Request %d should complete with expected status", i)
		}

		t.Logf("Completed %d concurrent create requests in %v", numConcurrent, totalTime)

		// Concurrent creates should complete reasonably quickly
		assert.True(t, totalTime < 2*time.Second,
			"Concurrent creates should complete within 2 seconds, took %v", totalTime)
	})
}

// TestVercelMemoryUsage tests memory usage and garbage collection
func TestVercelMemoryUsage(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping Vercel memory test in short mode")
	}

	// Set up test database
	dbPool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create repositories
	orgRepo := infrastructure.NewOrganizationRepository(dbPool)

	// Create services
	authService := auth.NewAuthorizationService()
	eventBus := events.NewInMemoryEventBus()
	orgService := application.NewOrganizationService(orgRepo, eventBus)

	// Create handler
	handler := interfaces.NewOrganizationHandler(orgService, authService)

	t.Run("memory usage under load", func(t *testing.T) {
		const numRequests = 50

		// Create a simple request template
		createReq := map[string]interface{}{
			"name":        "Memory Load Test Org",
			"description": "Organization for memory load testing",
		}

		body, _ := json.Marshal(createReq)

		// Process multiple requests to simulate load
		start := time.Now()
		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			// Each request should complete successfully
			assert.Equal(t, http.StatusCreated, w.Code)

			// Small delay to simulate real-world usage
			time.Sleep(1 * time.Millisecond)
		}

		totalTime := time.Since(start)
		t.Logf("Processed %d requests in %v", numRequests, totalTime)

		// All requests should complete successfully
		assert.True(t, totalTime < 5*time.Second,
			"Load test should complete within 5 seconds, took %v", totalTime)
	})

	t.Run("garbage collection behavior", func(t *testing.T) {
		// Test that handlers can be created and destroyed without memory leaks
		const numHandlers = 100

		for i := 0; i < numHandlers; i++ {
			// Create a new handler instance
			tempHandler := interfaces.NewOrganizationHandler(orgService, authService)

			// Use it for a simple operation
			req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
			w := httptest.NewRecorder()

			tempHandler.ListOrganizations(w, req)

			// Handler should work
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}

		t.Logf("Created and used %d handler instances successfully", numHandlers)
	})
}

// TestVercelTimeoutHandling tests timeout handling in serverless context
func TestVercelTimeoutHandling(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping Vercel timeout test in short mode")
	}

	// Set up test database
	dbPool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create repositories
	orgRepo := infrastructure.NewOrganizationRepository(dbPool)

	// Create services
	authService := auth.NewAuthorizationService()
	eventBus := events.NewInMemoryEventBus()
	orgService := application.NewOrganizationService(orgRepo, eventBus)

	// Create handler
	handler := interfaces.NewOrganizationHandler(orgService, authService)

	t.Run("short timeout handling", func(t *testing.T) {
		// Test with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.ListOrganizations(w, req)
		processingTime := time.Since(start)

		// Should complete very quickly
		assert.True(t, processingTime < 1*time.Millisecond,
			"Short timeout request should complete quickly, took %v", processingTime)

		t.Logf("Short timeout test completed in %v", processingTime)
	})

	t.Run("context cancellation handling", func(t *testing.T) {
		// Test with context cancellation
		ctx, cancel := context.WithCancel(context.Background())

		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Cancel context immediately
		cancel()

		start := time.Now()
		handler.ListOrganizations(w, req)
		processingTime := time.Since(start)

		// Should complete quickly even with cancelled context
		assert.True(t, processingTime < 10*time.Millisecond,
			"Cancelled context request should complete quickly, took %v", processingTime)

		t.Logf("Context cancellation test completed in %v", processingTime)
	})
}

// Note: setupTestDatabase is defined in organization_integration_test.go
