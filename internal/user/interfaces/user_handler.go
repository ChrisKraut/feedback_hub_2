package interfaces

import (
	"encoding/json"
	"feedback_hub_2/internal/shared/web"
	userapp "feedback_hub_2/internal/user/application"
	"feedback_hub_2/internal/user/domain"
	"net/http"
	"strings"
)

// UserHandler handles HTTP requests for user management operations.
// AI-hint: HTTP transport layer for user operations following REST conventions.
// Provides proper error handling, status codes, and JSON responses.
type UserHandler struct {
	userService *userapp.UserService
}

// NewUserHandler creates a new UserHandler instance.
// AI-hint: Factory method for user handler with dependency injection of user service.
func NewUserHandler(userService *userapp.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request body for creating a user.
// AI-hint: DTO for user creation API with validation-friendly structure.
type CreateUserRequest struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	RoleID string `json:"role_id"`
}

// UserResponse represents the response body for user operations.
// AI-hint: DTO for user API responses with consistent structure.
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	RoleID    string `json:"role_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UpdateUserRequest represents the request body for updating a user.
// AI-hint: DTO for user update API with validation-friendly structure.
type UpdateUserRequest struct {
	Name string `json:"name"`
}

// UpdateUserRoleRequest represents the request body for updating a user's role.
// AI-hint: DTO for user role update API with validation-friendly structure.
type UpdateUserRoleRequest struct {
	RoleID string `json:"role_id"`
}

// CreateUser handles POST /users requests.
// AI-hint: User creation endpoint with authorization, validation, and proper error handling.
//
// @Summary Create a new user
// @Description Create a new user with role assignment (authorization rules apply)
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User creation request"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Email) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Email is required")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}
	if strings.TrimSpace(req.RoleID) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role ID is required")
		return
	}

	// Create the user
	newUser, err := h.userService.CreateUser(r.Context(), req.Email, req.Name, req.RoleID, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrEmailAlreadyExists:
			web.WriteErrorResponse(w, http.StatusConflict, "Email already exists")
		default:
			if strings.Contains(err.Error(), "invalid role ID") {
				web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid role ID")
			} else {
				web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			}
		}
		return
	}

	// Return the created user
	response := UserResponse{
		ID:        newUser.ID,
		Email:     newUser.Email,
		Name:      newUser.Name,
		RoleID:    newUser.RoleID,
		CreatedAt: newUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: newUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUser handles GET /users/{id} requests.
// AI-hint: User retrieval endpoint with proper error handling for not found cases.
//
// @Summary Get a user by ID
// @Description Get a user by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract user ID from URL path
	targetUserID := web.ExtractIDFromPath(r.URL.Path, "/users/")
	if targetUserID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Get the user
	foundUser, err := h.userService.GetUser(r.Context(), targetUserID)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the user
	response := UserResponse{
		ID:        foundUser.ID,
		Email:     foundUser.Email,
		Name:      foundUser.Name,
		RoleID:    foundUser.RoleID,
		CreatedAt: foundUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: foundUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListUsers handles GET /users requests.
// AI-hint: User listing endpoint with proper JSON array response handling.
//
// @Summary List all users
// @Description Get all users in the system
// @Tags users
// @Produce json
// @Success 200 {array} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get all users
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Convert to response format
	var responses []UserResponse
	for _, user := range users {
		responses = append(responses, UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			RoleID:    user.RoleID,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UpdateUser handles PUT /users/{id} requests.
// AI-hint: User update endpoint with proper authorization and validation.
//
// @Summary Update a user
// @Description Update a user's name (email is immutable)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User update request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract user ID from URL path
	targetUserID := web.ExtractIDFromPath(r.URL.Path, "/users/")
	if targetUserID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Name) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Update the user
	updatedUser, err := h.userService.UpdateUser(r.Context(), targetUserID, req.Name, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrUserNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the updated user
	response := UserResponse{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		Name:      updatedUser.Name,
		RoleID:    updatedUser.RoleID,
		CreatedAt: updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUserRole handles PUT /users/{id}/role requests.
// AI-hint: User role update endpoint with Super User authorization requirement.
//
// @Summary Update a user's role
// @Description Update a user's role (Super User only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param role body UpdateUserRoleRequest true "User role update request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract user ID from URL path
	targetUserID := web.ExtractIDFromPath(r.URL.Path, "/users/")
	if targetUserID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.RoleID) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Role ID is required")
		return
	}

	// Update the user's role
	updatedUser, err := h.userService.UpdateUserRole(r.Context(), targetUserID, req.RoleID, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrUserNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			if strings.Contains(err.Error(), "invalid role ID") {
				web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid role ID")
			} else {
				web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			}
		}
		return
	}

	// Return the updated user
	response := UserResponse{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		Name:      updatedUser.Name,
		RoleID:    updatedUser.RoleID,
		CreatedAt: updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE /users/{id} requests.
// AI-hint: User deletion endpoint with proper authorization and error handling.
//
// @Summary Delete a user
// @Description Delete a user (authorization rules apply)
// @Tags users
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract user ID from URL path
	targetUserID := web.ExtractIDFromPath(r.URL.Path, "/users/")
	if targetUserID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Delete the user
	err := h.userService.DeleteUser(r.Context(), targetUserID, userID)
	if err != nil {
		switch err {
		case domain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		case domain.ErrUserNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}
