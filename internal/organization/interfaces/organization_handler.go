package interfaces

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"feedback_hub_2/internal/organization/application"
	"feedback_hub_2/internal/organization/domain"
	"feedback_hub_2/internal/shared/auth"

	"github.com/google/uuid"
)

// OrganizationHandler handles HTTP requests for organization management.
// AI-hint: HTTP interface layer that translates HTTP requests to application service calls,
// handles request/response DTOs, and provides proper HTTP status codes and error handling.
type OrganizationHandler struct {
	service *application.OrganizationService
	auth    *auth.AuthorizationService
}

// NewOrganizationHandler creates a new organization handler instance.
// AI-hint: Factory method for creating organization handler with dependency injection
// of the organization service and authorization service.
func NewOrganizationHandler(service *application.OrganizationService, auth *auth.AuthorizationService) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
		auth:    auth,
	}
}

// CreateOrganizationRequest represents the request body for creating an organization.
// AI-hint: DTO for organization creation requests with validation tags.
type CreateOrganizationRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Slug        string                 `json:"slug,omitempty"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// CreateOrganizationResponse represents the response for organization creation.
// AI-hint: DTO for organization creation responses with proper JSON tags.
type CreateOrganizationResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// UpdateOrganizationRequest represents the request body for updating an organization.
// AI-hint: DTO for organization update requests with optional fields.
type UpdateOrganizationRequest struct {
	Name        string                 `json:"name,omitempty"`
	Slug        string                 `json:"slug,omitempty"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// OrganizationResponse represents the response for organization operations.
// AI-hint: DTO for organization responses used across multiple endpoints.
type OrganizationResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// ListOrganizationsResponse represents the response for listing organizations.
// AI-hint: DTO for organization listing responses with pagination metadata.
type ListOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// ErrorResponse represents an error response.
// AI-hint: Standard error response format for consistent API error handling.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateOrganization handles POST /organizations requests.
// AI-hint: HTTP endpoint for creating new organizations with validation and business logic.
// @Summary Create a new organization
// @Description Create a new organization with the provided details
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} CreateOrganizationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "name is required")
		return
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Create organization using service
	org, err := h.service.CreateOrganizationWithSlug(r.Context(), req.Name, req.Description, req.Settings)
	if err != nil {
		switch err {
		case domain.ErrOrganizationSlugAlreadyExists:
			respondWithError(w, http.StatusConflict, "Conflict", "organization slug already exists")
		case domain.ErrInvalidOrganizationData:
			respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization data")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		}
		return
	}

	// Convert to response DTO
	response := CreateOrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Settings:    org.Settings,
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetOrganization handles GET /organizations/{id} requests.
// AI-hint: HTTP endpoint for retrieving organizations by ID with proper error handling.
// @Summary Get organization by ID
// @Description Retrieve an organization by its unique identifier
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} OrganizationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	// Extract organization ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid URL path")
		return
	}

	orgID := pathParts[1]
	if orgID == "" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "organization ID is required")
		return
	}

	// Parse UUID
	id, err := uuid.Parse(orgID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization ID")
		return
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Get organization using service
	org, err := h.service.GetOrganizationByID(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			respondWithError(w, http.StatusNotFound, "Not Found", "organization not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		}
		return
	}

	// Convert to response DTO
	response := OrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Settings:    org.Settings,
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOrganizationBySlug handles GET /organizations/slug/{slug} requests.
// AI-hint: HTTP endpoint for retrieving organizations by slug with proper error handling.
// @Summary Get organization by slug
// @Description Retrieve an organization by its unique slug
// @Tags organizations
// @Accept json
// @Produce json
// @Param slug path string true "Organization slug"
// @Success 200 {object} OrganizationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations/slug/{slug} [get]
func (h *OrganizationHandler) GetOrganizationBySlug(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[1] != "slug" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid URL path")
		return
	}

	slug := pathParts[2]
	if slug == "" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "slug is required")
		return
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Get organization using service
	org, err := h.service.GetOrganizationBySlug(r.Context(), slug)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			respondWithError(w, http.StatusNotFound, "Not Found", "organization not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		}
		return
	}

	// Convert to response DTO
	response := OrganizationResponse{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Settings:    org.Settings,
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateOrganization handles PUT /organizations/{id} requests.
// AI-hint: HTTP endpoint for updating existing organizations with validation and business logic.
// @Summary Update organization
// @Description Update an existing organization with new details
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body UpdateOrganizationRequest true "Organization update details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// Extract organization ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid URL path")
		return
	}

	orgID := pathParts[1]
	if orgID == "" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "organization ID is required")
		return
	}

	// Parse UUID
	id, err := uuid.Parse(orgID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization ID")
		return
	}

	// Parse request body
	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Update organization using service
	err = h.service.UpdateOrganization(r.Context(), id, req.Name, req.Slug, req.Description, req.Settings)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			respondWithError(w, http.StatusNotFound, "Not Found", "organization not found")
		case domain.ErrOrganizationSlugAlreadyExists:
			respondWithError(w, http.StatusConflict, "Conflict", "organization slug already exists")
		case domain.ErrInvalidOrganizationData:
			respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization data")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		}
		return
	}

	// Return success response
	response := map[string]string{"message": "organization updated successfully"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteOrganization handles DELETE /organizations/{id} requests.
// AI-hint: HTTP endpoint for deleting organizations with proper cleanup and validation.
// @Summary Delete organization
// @Description Delete an organization from the system
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// Extract organization ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid URL path")
		return
	}

	orgID := pathParts[1]
	if orgID == "" {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "organization ID is required")
		return
	}

	// Parse UUID
	id, err := uuid.Parse(orgID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", "invalid organization ID")
		return
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Delete organization using service
	err = h.service.DeleteOrganization(r.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrOrganizationNotFound:
			respondWithError(w, http.StatusNotFound, "Not Found", "organization not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		}
		return
	}

	// Return success response
	response := map[string]string{"message": "organization deleted successfully"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListOrganizations handles GET /organizations requests.
// AI-hint: HTTP endpoint for listing organizations with pagination support.
// @Summary List organizations
// @Description Retrieve a paginated list of organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Param limit query int false "Number of organizations to return (default: 20, max: 100)"
// @Param offset query int false "Number of organizations to skip (default: 0)"
// @Success 200 {object} ListOrganizationsResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations [get]
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Parse offset
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Enforce maximum limit
	if limit > 100 {
		limit = 100
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Get organizations using service
	organizations, err := h.service.ListOrganizations(r.Context(), limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		return
	}

	// Convert to response DTOs
	var orgResponses []OrganizationResponse
	for _, org := range organizations {
		orgResponses = append(orgResponses, OrganizationResponse{
			ID:          org.ID.String(),
			Name:        org.Name,
			Slug:        org.Slug,
			Description: org.Description,
			Settings:    org.Settings,
			CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Get total count
	total, err := h.service.CountOrganizations(r.Context())
	if err != nil {
		// If we can't get the count, just return what we have
		total = len(orgResponses)
	}

	// Create response
	response := ListOrganizationsResponse{
		Organizations: orgResponses,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SearchOrganizations handles GET /organizations/search requests.
// AI-hint: HTTP endpoint for searching organizations with flexible criteria and pagination.
// @Summary Search organizations
// @Description Search for organizations by name, description, or slug
// @Tags organizations
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param limit query int false "Number of organizations to return (default: 20, max: 100)"
// @Param offset query int false "Number of organizations to skip (default: 0)"
// @Success 200 {object} ListOrganizationsResponse
// @Failure 500 {object} ErrorResponse
// @Router /organizations/search [get]
func (h *OrganizationHandler) SearchOrganizations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Parse offset
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Enforce maximum limit
	if limit > 100 {
		limit = 100
	}

	// Check if service is available
	if h.service == nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "service not available")
		return
	}

	// Search organizations using service
	organizations, err := h.service.SearchOrganizations(r.Context(), query, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", "internal server error")
		return
	}

	// Convert to response DTOs
	var orgResponses []OrganizationResponse
	for _, org := range organizations {
		orgResponses = append(orgResponses, OrganizationResponse{
			ID:          org.ID.String(),
			Name:        org.Name,
			Slug:        org.Slug,
			Description: org.Description,
			Settings:    org.Settings,
			CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Create response
	response := ListOrganizationsResponse{
		Organizations: orgResponses,
		Total:         len(orgResponses),
		Limit:         limit,
		Offset:        offset,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// respondWithError is a helper function to send error responses.
// AI-hint: Utility function for consistent error response formatting across all endpoints.
func respondWithError(w http.ResponseWriter, statusCode int, errorType, message string) {
	response := ErrorResponse{
		Error:   errorType,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
