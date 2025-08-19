package interfaces

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"feedback_hub_2/internal/organization/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test data helpers
func createTestOrganization() *domain.Organization {
	id := uuid.New()
	now := time.Now()
	return &domain.Organization{
		ID:          id,
		Name:        "Test Organization",
		Slug:        "test-org",
		Description: "A test organization",
		Settings:    map[string]any{"theme": "dark"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// TestCreateOrganization tests the organization creation endpoint
func TestCreateOrganization(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"description": "A test organization",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "name is required",
		},
		{
			name: "empty name",
			requestBody: map[string]interface{}{
				"name":        "",
				"description": "A test organization",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (we're just testing validation)
			handler := &OrganizationHandler{}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateOrganization(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestGetOrganization tests the organization retrieval endpoint
func TestGetOrganization(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid UUID",
			organizationID: "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid organization ID",
		},
		{
			name:           "empty path",
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid URL path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (we're just testing validation)
			handler := &OrganizationHandler{}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/organizations/"+tt.organizationID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetOrganization(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestUpdateOrganization tests the organization update endpoint
func TestUpdateOrganization(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid UUID",
			organizationID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"name": "Updated Org",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid organization ID",
		},
		{
			name:           "empty path",
			organizationID: "",
			requestBody: map[string]interface{}{
				"name": "Updated Org",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid URL path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (we're just testing validation)
			handler := &OrganizationHandler{}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/organizations/"+tt.organizationID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.UpdateOrganization(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestDeleteOrganization tests the organization deletion endpoint
func TestDeleteOrganization(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid UUID",
			organizationID: "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid organization ID",
		},
		{
			name:           "empty path",
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid URL path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (we're just testing validation)
			handler := &OrganizationHandler{}

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/organizations/"+tt.organizationID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.DeleteOrganization(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestGetOrganizationBySlug tests the organization retrieval by slug endpoint
func TestGetOrganizationBySlug(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing slug in path",
			slug:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid URL path",
		},
		{
			name:           "valid path structure",
			slug:           "test",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (we're just testing validation)
			handler := &OrganizationHandler{}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/organizations/slug/"+tt.slug, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetOrganizationBySlug(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestListOrganizations tests the organization listing endpoint
func TestListOrganizations(t *testing.T) {
	t.Run("default pagination", func(t *testing.T) {
		// Create handler with nil service (we're just testing validation)
		handler := &OrganizationHandler{}

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		handler.ListOrganizations(w, req)

		// Assertions - should fail due to nil service, but we're testing the structure
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestSearchOrganizations tests the organization search endpoint
func TestSearchOrganizations(t *testing.T) {
	t.Run("default pagination", func(t *testing.T) {
		// Create handler with nil service (we're just testing validation)
		handler := &OrganizationHandler{}

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/organizations/search", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Call handler
		handler.SearchOrganizations(w, req)

		// Assertions - should fail due to nil service, but we're testing the structure
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
