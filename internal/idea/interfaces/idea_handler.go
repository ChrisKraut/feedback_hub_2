package interfaces

import (
	"encoding/json"
	ideaapp "feedback_hub_2/internal/idea/application"
	ideadomain "feedback_hub_2/internal/idea/domain"
	"feedback_hub_2/internal/shared/web"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// IdeaHandler handles HTTP requests for idea management operations.
// AI-hint: HTTP transport layer for idea operations following REST conventions.
// Provides proper error handling, status codes, and JSON responses.
type IdeaHandler struct {
	ideaService *ideaapp.IdeaApplicationService
}

// NewIdeaHandler creates a new IdeaHandler instance.
// AI-hint: Factory method for idea handler with dependency injection of idea service.
func NewIdeaHandler(ideaService *ideaapp.IdeaApplicationService) *IdeaHandler {
	return &IdeaHandler{
		ideaService: ideaService,
	}
}

// CreateIdeaRequest represents the request body for creating an idea.
// AI-hint: DTO for idea creation API with validation-friendly structure.
type CreateIdeaRequest struct {
	Title   string `json:"title" example:"Improve user dashboard"`
	Content string `json:"content" example:"The current dashboard could be enhanced with better data visualization and filtering options."`
}

// CreateIdeaResponse represents the response body for idea creation.
// AI-hint: DTO for idea creation API responses with consistent structure.
type CreateIdeaResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateIdeaRequest represents the request body for updating an idea.
// AI-hint: DTO for idea update API with validation-friendly structure.
type UpdateIdeaRequest struct {
	Title   string `json:"title" example:"Improved user dashboard"`
	Content string `json:"content" example:"The dashboard has been enhanced with better data visualization and filtering options."`
}

// CreateIdea handles POST /ideas requests.
// AI-hint: Idea creation endpoint with authentication, validation, and proper error handling.
//
// @Summary Create a new idea
// @Description Create a new feedback idea (authentication required)
// @Tags ideas
// @Accept json
// @Produce json
// @Param idea body CreateIdeaRequest true "Idea creation request"
// @Success 201 {object} CreateIdeaResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /ideas [post]
func (h *IdeaHandler) CreateIdea(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse the request body
	var req CreateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Title) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Call the application service
	newIdea, err := h.ideaService.CreateIdea(r.Context(), req.Title, req.Content, userID)
	if err != nil {
		switch err {
		case ideadomain.ErrInvalidIdeaData:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid idea data")
		case ideadomain.ErrCreatorNotFound:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Creator user not found")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the created idea ID
	response := CreateIdeaResponse{
		ID: newIdea.ID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateIdea handles PUT /ideas/{ideaId} requests.
// AI-hint: Idea update endpoint with authentication, authorization, validation, and proper error handling.
//
// @Summary Update an existing idea
// @Description Update the title and content of an existing feedback idea (authentication required, creator only)
// @Tags ideas
// @Accept json
// @Produce json
// @Param ideaId path string true "Idea ID" format(uuid)
// @Param idea body UpdateIdeaRequest true "Idea update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /ideas/{ideaId} [put]
func (h *IdeaHandler) UpdateIdea(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract idea ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 || pathParts[0] != "ideas" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid URL path")
		return
	}

	ideaIDStr := pathParts[1]
	ideaID, err := uuid.Parse(ideaIDStr)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid idea ID format")
		return
	}

	// Parse the request body
	var req UpdateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Title) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Call the application service
	updatedIdea, err := h.ideaService.UpdateIdea(r.Context(), ideaID, req.Title, req.Content, userID)
	if err != nil {
		switch err {
		case ideadomain.ErrInvalidIdeaData:
			web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid idea data")
		case ideadomain.ErrIdeaNotFound:
			web.WriteErrorResponse(w, http.StatusNotFound, "Idea not found")
		case ideadomain.ErrUnauthorized:
			web.WriteErrorResponse(w, http.StatusForbidden, "Only the creator can update this idea")
		default:
			web.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Idea updated successfully",
		"idea_id": updatedIdea.ID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
