package interfaces

import (
	"encoding/json"
	roleapp "feedback_hub_2/internal/role/application"
	"feedback_hub_2/internal/role/domain"
	"feedback_hub_2/internal/shared/web"
	"net/http"
	"strings"
)

// RoleHandler handles HTTP requests for role management operations.
// AI-hint: HTTP transport layer for role operations following REST conventions.
// Provides proper error handling, status codes, and JSON responses.
type RoleHandler struct {
	roleService *roleapp.RoleService
}

// NewRoleHandler creates a new RoleHandler instance.
// AI-hint: Factory method for role handler with dependency injection of role service.
func NewRoleHandler(roleService *roleapp.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRoleRequest represents the request body for creating a role.
// AI-hint: DTO for role creation API with validation-friendly structure.
type CreateRoleRequest struct {
	Name string `json:"name"`
}

// RoleResponse represents the response body for role operations.
// AI-hint: DTO for role API responses with consistent structure.
type RoleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UpdateRoleRequest represents the request body for updating a role.
// AI-hint: DTO for role update API with validation-friendly structure.
type UpdateRoleRequest struct {
	Name string `json:"name"`
}

// CreateRole handles POST /roles requests.
// AI-hint: Role creation endpoint with proper authorization, validation, and error handling.
//
// @Summary Create a new role
// @Description Create a new role (Super User only)
// @Tags roles
// @Accept json
// @Produce json
// @Param role body CreateRoleRequest true "Role creation request"
// @Success 201 {object} RoleResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /roles [post]
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Name) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Create the role
	newRole, err := h.roleService.CreateRole(r.Context(), req.Name, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrRoleNameAlreadyExists:
			web.WriteErrorResponse(w, http.StatusConflict, "Role name already exists")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the created role
	response := RoleResponse{
		ID:        newRole.ID,
		Name:      newRole.Name,
		CreatedAt: newRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: newRole.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetRole handles GET /roles/{id} requests.
// AI-hint: Role retrieval endpoint with proper error handling for not found cases.
//
// @Summary Get a role by ID
// @Description Get a role by its ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} RoleResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract role ID from URL path
	roleID := web.ExtractIDFromPath(r.URL.Path, "/roles/")
	if roleID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role ID is required")
		return
	}

	// Get the role
	foundRole, err := h.roleService.GetRole(r.Context(), roleID)
	if err != nil {
		switch err {
		case domain.ErrRoleNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the role
	response := RoleResponse{
		ID:        foundRole.ID,
		Name:      foundRole.Name,
		CreatedAt: foundRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: foundRole.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListRoles handles GET /roles requests.
// AI-hint: Role listing endpoint with proper JSON array response handling.
//
// @Summary List all roles
// @Description Get all roles in the system
// @Tags roles
// @Produce json
// @Success 200 {array} RoleResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /roles [get]
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get all roles
	roles, err := h.roleService.ListRoles(r.Context())
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Convert to response format
	var responses []RoleResponse
	for _, role := range roles {
		responses = append(responses, RoleResponse{
			ID:        role.ID,
			Name:      role.Name,
			CreatedAt: role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UpdateRole handles PUT /roles/{id} requests.
// AI-hint: Role update endpoint with proper authorization and business rule validation.
//
// @Summary Update a role
// @Description Update a role's name (Super User only)
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body UpdateRoleRequest true "Role update request"
// @Success 200 {object} RoleResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract role ID from URL path
	roleID := web.ExtractIDFromPath(r.URL.Path, "/roles/")
	if roleID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role ID is required")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Name) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Update the role
	updatedRole, err := h.roleService.UpdateRole(r.Context(), roleID, req.Name, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrRoleNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		case domain.ErrRoleNameAlreadyExists:
			web.WriteErrorResponse(w, http.StatusConflict, "Role name already exists")
		case domain.ErrCannotModifySuperUserRole:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Cannot modify Super User role")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the updated role
	response := RoleResponse{
		ID:        updatedRole.ID,
		Name:      updatedRole.Name,
		CreatedAt: updatedRole.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: updatedRole.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteRole handles DELETE /roles/{id} requests.
// AI-hint: Role deletion endpoint with business rule enforcement and proper error handling.
//
// @Summary Delete a role
// @Description Delete a role (Super User only, cannot delete Super User role)
// @Tags roles
// @Param id path string true "Role ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract role ID from URL path
	roleID := web.ExtractIDFromPath(r.URL.Path, "/roles/")
	if roleID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role ID is required")
		return
	}

	// Delete the role
	err := h.roleService.DeleteRole(r.Context(), roleID, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrRoleNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		case domain.ErrCannotDeleteSuperUserRole:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete Super User role")
		case domain.ErrInvalidRoleData:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete role with assigned users")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}
