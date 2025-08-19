package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"feedback_hub_2/internal/organization/application"
	"feedback_hub_2/internal/organization/infrastructure"
	"feedback_hub_2/internal/organization/interfaces"
	"feedback_hub_2/internal/shared/auth"
	events "feedback_hub_2/internal/shared/bus"
	"feedback_hub_2/internal/shared/persistence"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganizationIntegration tests the complete organization CRUD flow
func TestOrganizationIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
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

	t.Run("complete organization lifecycle", func(t *testing.T) {
		// Test 1: Create organization
		t.Run("create organization", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":        "Test Integration Org",
				"description": "Organization for integration testing",
				"settings":    map[string]interface{}{"theme": "light"},
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Store organization ID for subsequent tests
			orgID := response["id"].(string)
			t.Logf("Created organization with ID: %s", orgID)
		})

		// Test 2: Get organization by slug
		t.Run("get organization by slug", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/test-integration-org", nil)
			w := httptest.NewRecorder()

			handler.GetOrganizationBySlug(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "Test Integration Org", response["name"])
			assert.Equal(t, "test-integration-org", response["slug"])
		})

		// Test 3: Update organization
		t.Run("update organization", func(t *testing.T) {
			// First get the organization to get its ID
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/test-integration-org", nil)
			w := httptest.NewRecorder()
			handler.GetOrganizationBySlug(w, req)

			var orgResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &orgResponse)
			require.NoError(t, err)

			orgID := orgResponse["id"].(string)

			// Now update it
			updateReq := map[string]interface{}{
				"name":        "Updated Integration Org",
				"description": "Updated description for integration testing",
			}

			body, _ := json.Marshal(updateReq)
			req = httptest.NewRequest(http.MethodPut, "/organizations/"+orgID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()

			handler.UpdateOrganization(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify the update
			req = httptest.NewRequest(http.MethodGet, "/organizations/"+orgID, nil)
			w = httptest.NewRecorder()
			handler.GetOrganization(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var updatedResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &updatedResponse)
			require.NoError(t, err)

			assert.Equal(t, "Updated Integration Org", updatedResponse["name"])
			assert.Equal(t, "Updated description for integration testing", updatedResponse["description"])
		})

		// Test 4: List organizations
		t.Run("list organizations", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/organizations?limit=10&offset=0", nil)
			w := httptest.NewRecorder()

			handler.ListOrganizations(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			organizations := response["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), 1)

			// Check that our test organization is in the list
			found := false
			for _, org := range organizations {
				orgMap := org.(map[string]interface{})
				if orgMap["name"] == "Updated Integration Org" {
					found = true
					break
				}
			}
			assert.True(t, found, "Test organization should be in the list")
		})

		// Test 5: Search organizations
		t.Run("search organizations", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/organizations/search?q=Integration&limit=10&offset=0", nil)
			w := httptest.NewRecorder()

			handler.SearchOrganizations(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			organizations := response["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), 1)

			// Check that our test organization is found
			found := false
			for _, org := range organizations {
				orgMap := org.(map[string]interface{})
				if orgMap["name"] == "Updated Integration Org" {
					found = true
					break
				}
			}
			assert.True(t, found, "Test organization should be found in search")
		})

		// Test 6: Delete organization
		t.Run("delete organization", func(t *testing.T) {
			// First get the organization to get its ID
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/test-integration-org", nil)
			w := httptest.NewRecorder()
			handler.GetOrganizationBySlug(w, req)

			var orgResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &orgResponse)
			require.NoError(t, err)

			orgID := orgResponse["id"].(string)

			// Now delete it
			req = httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID, nil)
			w = httptest.NewRecorder()

			handler.DeleteOrganization(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify it's deleted
			req = httptest.NewRequest(http.MethodGet, "/organizations/"+orgID, nil)
			w = httptest.NewRecorder()
			handler.GetOrganization(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

// TestOrganizationLifecycleEvents tests organization lifecycle events
func TestOrganizationLifecycleEvents(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
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

	t.Run("organization creation events", func(t *testing.T) {
		// Create organization
		createReq := map[string]interface{}{
			"name":        "Event Test Org",
			"description": "Organization for event testing",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrganization(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify organization was created with proper timestamps
		createdAt := response["created_at"].(string)
		updatedAt := response["updated_at"].(string)

		// Parse timestamps
		createdTime, err := time.Parse(time.RFC3339, createdAt)
		require.NoError(t, err)
		updatedTime, err := time.Parse(time.RFC3339, updatedAt)
		require.NoError(t, err)

		// Verify timestamps are recent
		now := time.Now()
		assert.True(t, now.Sub(createdTime) < 5*time.Second)
		assert.True(t, now.Sub(updatedTime) < 5*time.Second)

		// Verify created_at and updated_at are the same for new organization
		assert.Equal(t, createdTime, updatedTime)

		// Clean up
		orgID := response["id"].(string)
		req = httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID, nil)
		w = httptest.NewRecorder()
		handler.DeleteOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("organization update events", func(t *testing.T) {
		// Create organization
		createReq := map[string]interface{}{
			"name":        "Update Event Test Org",
			"description": "Organization for update event testing",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrganization(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		orgID := response["id"].(string)
		originalUpdatedAt := response["updated_at"].(string)

		// Wait a moment to ensure timestamp difference
		time.Sleep(100 * time.Millisecond)

		// Update organization
		updateReq := map[string]interface{}{
			"name": "Updated Event Test Org",
		}

		body, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, "/organizations/"+orgID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		handler.UpdateOrganization(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify updated_at changed
		req = httptest.NewRequest(http.MethodGet, "/organizations/"+orgID, nil)
		w = httptest.NewRecorder()
		handler.GetOrganization(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &updatedResponse)
		require.NoError(t, err)

		updatedUpdatedAt := updatedResponse["updated_at"].(string)
		assert.NotEqual(t, originalUpdatedAt, updatedUpdatedAt, "updated_at should change after update")

		// Clean up
		req = httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID, nil)
		w = httptest.NewRecorder()
		handler.DeleteOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestCrossDomainOrganizationInteractions tests cross-domain organization interactions
func TestCrossDomainOrganizationInteractions(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
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

	t.Run("multiple organizations isolation", func(t *testing.T) {
		// Create first organization
		org1Req := map[string]interface{}{
			"name":        "Isolation Test Org 1",
			"description": "First organization for isolation testing",
		}

		body, _ := json.Marshal(org1Req)
		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateOrganization(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var org1Response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &org1Response)
		require.NoError(t, err)

		org1ID := org1Response["id"].(string)
		org1Slug := org1Response["slug"].(string)

		// Create second organization
		org2Req := map[string]interface{}{
			"name":        "Isolation Test Org 2",
			"description": "Second organization for isolation testing",
		}

		body, _ = json.Marshal(org2Req)
		req = httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		handler.CreateOrganization(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var org2Response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &org2Response)
		require.NoError(t, err)

		org2ID := org2Response["id"].(string)
		org2Slug := org2Response["slug"].(string)

		// Verify organizations are different
		assert.NotEqual(t, org1ID, org2ID)
		assert.NotEqual(t, org1Slug, org2Slug)

		// Verify each organization can be retrieved independently
		req = httptest.NewRequest(http.MethodGet, "/organizations/"+org1ID, nil)
		w = httptest.NewRecorder()
		handler.GetOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedOrg1 map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &retrievedOrg1)
		require.NoError(t, err)
		assert.Equal(t, "Isolation Test Org 1", retrievedOrg1["name"])

		req = httptest.NewRequest(http.MethodGet, "/organizations/"+org2ID, nil)
		w = httptest.NewRecorder()
		handler.GetOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedOrg2 map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &retrievedOrg2)
		require.NoError(t, err)
		assert.Equal(t, "Isolation Test Org 2", retrievedOrg2["name"])

		// Clean up
		req = httptest.NewRequest(http.MethodDelete, "/organizations/"+org1ID, nil)
		w = httptest.NewRecorder()
		handler.DeleteOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		req = httptest.NewRequest(http.MethodDelete, "/organizations/"+org2ID, nil)
		w = httptest.NewRecorder()
		handler.DeleteOrganization(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestOrganizationPerformanceUnderLoad tests organization performance under load
func TestOrganizationPerformanceUnderLoad(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
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

	t.Run("create multiple organizations", func(t *testing.T) {
		const numOrgs = 10
		orgIDs := make([]string, 0, numOrgs)

		// Create multiple organizations
		start := time.Now()
		for i := 0; i < numOrgs; i++ {
			createReq := map[string]interface{}{
				"name":        fmt.Sprintf("Performance Test Org %d", i+1),
				"description": fmt.Sprintf("Organization %d for performance testing", i+1),
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			orgIDs = append(orgIDs, response["id"].(string))
		}
		creationTime := time.Since(start)

		// Verify all organizations were created
		assert.Equal(t, numOrgs, len(orgIDs))

		// Test listing performance
		start = time.Now()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations?limit=%d&offset=0", numOrgs), nil)
		w := httptest.NewRecorder()

		handler.ListOrganizations(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		listingTime := time.Since(start)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		organizations := response["organizations"].([]interface{})
		assert.Equal(t, numOrgs, len(organizations))

		// Log performance metrics
		t.Logf("Created %d organizations in %v", numOrgs, creationTime)
		t.Logf("Listed %d organizations in %v", numOrgs, listingTime)
		t.Logf("Average creation time: %v", creationTime/time.Duration(numOrgs))
		t.Logf("Average listing time per org: %v", listingTime/time.Duration(numOrgs))

		// Performance assertions (adjust thresholds as needed)
		assert.True(t, creationTime < 5*time.Second, "Organization creation should complete within 5 seconds")
		assert.True(t, listingTime < 1*time.Second, "Organization listing should complete within 1 second")

		// Clean up
		for _, orgID := range orgIDs {
			req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID, nil)
			w := httptest.NewRecorder()
			handler.DeleteOrganization(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})
}

// setupTestDatabase sets up a test database connection
func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	// This would typically use testcontainers or a test database
	// For now, we'll use the main database connection
	// In a real test environment, you'd want to use a separate test database

	// Get database URL from environment
	dbURL := "postgres://postgres:postgres@localhost:5432/feedback_hub_test"
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Create connection pool
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	// Ensure schema exists
	err = persistence.EnsureSchema(ctx, dbPool)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		dbPool.Close()
	}

	return dbPool, cleanup
}
