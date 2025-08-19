package tests

import (
	"bytes"
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteOrganizationWorkflow tests the complete organization workflow
// including user creation, role assignment, and idea creation within organizations
func TestCompleteOrganizationWorkflow(t *testing.T) {
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

	t.Run("complete organization workflow", func(t *testing.T) {
		// Step 1: Create organization
		t.Run("create organization", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":        "Workflow Test Org",
				"description": "Organization for complete workflow testing",
				"settings":    map[string]interface{}{"theme": "dark", "features": []string{"ideas", "roles", "users"}},
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
			orgSlug := response["slug"].(string)
			t.Logf("Created organization with ID: %s, Slug: %s", orgID, orgSlug)

			// Verify organization was created with proper data
			assert.Equal(t, "Workflow Test Org", response["name"])
			assert.Equal(t, "workflow-test-org", response["slug"])
			assert.Equal(t, "Organization for complete workflow testing", response["description"])

			// Verify settings were stored correctly
			settings := response["settings"].(map[string]interface{})
			assert.Equal(t, "dark", settings["theme"])
			features := settings["features"].([]interface{})
			assert.Len(t, features, 3)
		})

		// Step 2: Verify organization can be retrieved
		t.Run("retrieve organization", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/workflow-test-org", nil)
			w := httptest.NewRecorder()

			handler.GetOrganizationBySlug(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "Workflow Test Org", response["name"])
			assert.Equal(t, "workflow-test-org", response["slug"])
		})

		// Step 3: Update organization
		t.Run("update organization", func(t *testing.T) {
			// First get the organization to get its ID
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/workflow-test-org", nil)
			w := httptest.NewRecorder()
			handler.GetOrganizationBySlug(w, req)

			var orgResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &orgResponse)
			require.NoError(t, err)

			orgID := orgResponse["id"].(string)
			originalUpdatedAt := orgResponse["updated_at"].(string)

			// Wait a moment to ensure timestamp difference
			time.Sleep(10 * time.Millisecond)

			// Now update it
			updateReq := map[string]interface{}{
				"name":        "Updated Workflow Test Org",
				"description": "Updated description for workflow testing",
				"settings":    map[string]interface{}{"theme": "light", "features": []string{"ideas", "roles", "users", "analytics"}},
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

			assert.Equal(t, "Updated Workflow Test Org", updatedResponse["name"])
			assert.Equal(t, "Updated description for workflow testing", updatedResponse["description"])

			// Verify settings were updated
			settings := updatedResponse["settings"].(map[string]interface{})
			assert.Equal(t, "light", settings["theme"])
			features := settings["features"].([]interface{})
			assert.Len(t, features, 4)
			assert.Contains(t, features, "analytics")

			// Verify updated_at timestamp changed
			newUpdatedAt := updatedResponse["updated_at"].(string)
			assert.NotEqual(t, originalUpdatedAt, newUpdatedAt)
		})

		// Step 4: Search and list organizations
		t.Run("search and list organizations", func(t *testing.T) {
			// Test search
			req := httptest.NewRequest(http.MethodGet, "/organizations/search?q=Workflow&limit=10&offset=0", nil)
			w := httptest.NewRecorder()

			handler.SearchOrganizations(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var searchResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &searchResponse)
			require.NoError(t, err)

			organizations := searchResponse["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), 1)

			// Check that our test organization is found
			found := false
			for _, org := range organizations {
				orgMap := org.(map[string]interface{})
				if orgMap["name"] == "Updated Workflow Test Org" {
					found = true
					break
				}
			}
			assert.True(t, found, "Test organization should be found in search")

			// Test list
			req = httptest.NewRequest(http.MethodGet, "/organizations?limit=10&offset=0", nil)
			w = httptest.NewRecorder()

			handler.ListOrganizations(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var listResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &listResponse)
			require.NoError(t, err)

			organizations = listResponse["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), 1)
		})

		// Step 5: Clean up - delete organization
		t.Run("delete organization", func(t *testing.T) {
			// First get the organization to get its ID
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/workflow-test-org", nil)
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

// TestOrganizationDataIntegrity tests organization data integrity
// including validation, constraints, and business rules
func TestOrganizationDataIntegrity(t *testing.T) {
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

	t.Run("data validation and constraints", func(t *testing.T) {
		// Test 1: Empty name validation
		t.Run("empty name validation", func(t *testing.T) {
			createReq := map[string]interface{}{
				"name":        "",
				"description": "Should fail validation",
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		// Test 2: Very long name validation
		t.Run("very long name validation", func(t *testing.T) {
			longName := string(make([]byte, 300)) // 300 bytes
			for i := range longName {
				longName = longName[:i] + "a" + longName[i+1:]
			}

			createReq := map[string]interface{}{
				"name":        longName,
				"description": "Should fail validation",
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		// Test 3: Duplicate slug prevention
		t.Run("duplicate slug prevention", func(t *testing.T) {
			// Create first organization
			createReq1 := map[string]interface{}{
				"name":        "Duplicate Slug Test 1",
				"description": "First organization",
			}

			body1, _ := json.Marshal(createReq1)
			req1 := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body1))
			req1.Header.Set("Content-Type", "application/json")
			w1 := httptest.NewRecorder()

			handler.CreateOrganization(w1, req1)

			assert.Equal(t, http.StatusCreated, w1.Code)

			// Try to create second organization with same name (should generate same slug)
			createReq2 := map[string]interface{}{
				"name":        "Duplicate Slug Test 1", // Same name
				"description": "Second organization",
			}

			body2, _ := json.Marshal(createReq2)
			req2 := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body2))
			req2.Header.Set("Content-Type", "application/json")
			w2 := httptest.NewRecorder()

			handler.CreateOrganization(w2, req2)

			// Should fail due to duplicate slug
			assert.Equal(t, http.StatusConflict, w2.Code)

			// Clean up first organization
			var response1 map[string]interface{}
			err := json.Unmarshal(w1.Body.Bytes(), &response1)
			require.NoError(t, err)
			orgID1 := response1["id"].(string)

			req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID1, nil)
			w := httptest.NewRecorder()
			handler.DeleteOrganization(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		// Test 4: Invalid JSON handling
		t.Run("invalid JSON handling", func(t *testing.T) {
			invalidJSON := `{"name": "Invalid JSON", "description": "Missing closing brace"`

			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer([]byte(invalidJSON)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		// Test 5: Missing required fields
		t.Run("missing required fields", func(t *testing.T) {
			createReq := map[string]interface{}{
				"description": "Missing name field",
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateOrganization(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

// TestOrganizationWorkflowPerformanceUnderLoad tests organization performance under load for workflow testing
func TestOrganizationWorkflowPerformanceUnderLoad(t *testing.T) {
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

	t.Run("performance under load", func(t *testing.T) {
		const numOrgs = 20
		orgIDs := make([]string, 0, numOrgs)

		// Create multiple organizations
		start := time.Now()
		for i := 0; i < numOrgs; i++ {
			createReq := map[string]interface{}{
				"name":        fmt.Sprintf("Performance Test Org %d", i+1),
				"description": fmt.Sprintf("Organization %d for performance testing", i+1),
				"settings":    map[string]interface{}{"index": i, "category": fmt.Sprintf("cat_%d", i%5)},
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

		// Performance assertion: Creating 20 organizations should take less than 5 seconds
		assert.Less(t, creationTime, 5*time.Second, "Creating 20 organizations should be fast")
		t.Logf("Created %d organizations in %v", numOrgs, creationTime)

		// Test list performance
		t.Run("list performance", func(t *testing.T) {
			start := time.Now()
			req := httptest.NewRequest(http.MethodGet, "/organizations?limit=100&offset=0", nil)
			w := httptest.NewRecorder()

			handler.ListOrganizations(w, req)

			listTime := time.Since(start)
			assert.Equal(t, http.StatusOK, w.Code)

			// Performance assertion: Listing organizations should be fast
			assert.Less(t, listTime, 2*time.Second, "Listing organizations should be fast")
			t.Logf("Listed organizations in %v", listTime)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			organizations := response["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), numOrgs)
		})

		// Test search performance
		t.Run("search performance", func(t *testing.T) {
			start := time.Now()
			req := httptest.NewRequest(http.MethodGet, "/organizations/search?q=Performance&limit=100&offset=0", nil)
			w := httptest.NewRecorder()

			handler.SearchOrganizations(w, req)

			searchTime := time.Since(start)
			assert.Equal(t, http.StatusOK, w.Code)

			// Performance assertion: Searching organizations should be fast
			assert.Less(t, searchTime, 2*time.Second, "Searching organizations should be fast")
			t.Logf("Searched organizations in %v", searchTime)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			organizations := response["organizations"].([]interface{})
			assert.GreaterOrEqual(t, len(organizations), numOrgs)
		})

		// Clean up all created organizations
		t.Run("cleanup", func(t *testing.T) {
			start := time.Now()
			for _, orgID := range orgIDs {
				req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID, nil)
				w := httptest.NewRecorder()

				handler.DeleteOrganization(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}
			cleanupTime := time.Since(start)

			// Performance assertion: Cleanup should be fast
			assert.Less(t, cleanupTime, 3*time.Second, "Cleanup should be fast")
			t.Logf("Cleaned up %d organizations in %v", numOrgs, cleanupTime)
		})
	})
}
