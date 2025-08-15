package http

import (
	"encoding/json"
	"feedback_hub_2/internal/application"
	"feedback_hub_2/internal/domain/idea"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// IdeaHandler handles HTTP requests for idea management operations.
// AI-hint: HTTP transport layer for idea operations following REST conventions.
// Provides proper error handling, status codes, and JSON responses.
type IdeaHandler struct {
	ideaService *application.IdeaApplicationService
}

// NewIdeaHandler creates a new IdeaHandler instance.
// AI-hint: Factory method for idea handler with dependency injection of idea service.
func NewIdeaHandler(ideaService *application.IdeaApplicationService) *IdeaHandler {
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
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse the request body
	var req CreateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.Title) == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Convert string user ID to UUID
	creatorUserID, err := uuid.Parse(userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	// Create the idea command
	cmd := application.CreateIdeaCommand{
		Title:         strings.TrimSpace(req.Title),
		Content:       strings.TrimSpace(req.Content),
		CreatorUserID: creatorUserID,
	}

	// Call the application service
	newIdeaID, err := h.ideaService.CreateIdea(r.Context(), cmd)
	if err != nil {
		switch err {
		case idea.ErrInvalidIdeaData:
			writeErrorResponse(w, http.StatusBadRequest, "Invalid idea data")
		case idea.ErrCreatorNotFound:
			writeErrorResponse(w, http.StatusBadRequest, "Creator user not found")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the created idea ID
	response := CreateIdeaResponse{
		ID: newIdeaID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
