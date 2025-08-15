package http

import (
	"encoding/json"
	"feedback_hub_2/internal/application"
	"feedback_hub_2/internal/domain/idea"
	"net/http"
	"strings"
)

// IdeaHandler handles HTTP requests for idea management operations.
// AI-hint: HTTP transport layer for idea operations following REST conventions.
// Provides proper error handling, status codes, and JSON responses.
type IdeaHandler struct {
	ideaService *application.IdeaService
}

// NewIdeaHandler creates a new IdeaHandler instance.
// AI-hint: Factory method for idea handler with dependency injection of idea service.
func NewIdeaHandler(ideaService *application.IdeaService) *IdeaHandler {
	return &IdeaHandler{
		ideaService: ideaService,
	}
}

// CreateIdeaRequest represents the request body for creating an idea.
// AI-hint: DTO for idea creation API with validation-friendly structure.
type CreateIdeaRequest struct {
	IdeaText string `json:"idea_text"`
}

// IdeaResponse represents the response body for idea operations.
// AI-hint: DTO for idea API responses with consistent structure.
type IdeaResponse struct {
	ID        string `json:"id"`
	IdeaText  string `json:"idea_text"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// CreateIdea handles POST /ideas requests.
// AI-hint: Idea creation endpoint with authentication, validation, and proper error handling.
//
// @Summary Create a new idea
// @Description Create a new idea with markdown-formatted text (authentication required)
// @Tags ideas
// @Accept json
// @Produce json
// @Param idea body CreateIdeaRequest true "Idea creation request"
// @Success 201 {object} IdeaResponse
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

	var req CreateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if strings.TrimSpace(req.IdeaText) == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Idea text is required")
		return
	}

	// Create the idea
	newIdea, err := h.ideaService.CreateIdea(r.Context(), req.IdeaText, userID)
	if err != nil {
		switch err {
		case idea.ErrInvalidIdeaData:
			writeErrorResponse(w, http.StatusBadRequest, "Invalid idea data")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the created idea
	response := IdeaResponse{
		ID:        newIdea.ID,
		IdeaText:  newIdea.IdeaText,
		UserID:    newIdea.UserID,
		CreatedAt: newIdea.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetIdea handles GET /ideas/{id} requests.
// AI-hint: Idea retrieval endpoint with proper error handling for not found cases.
//
// @Summary Get an idea by ID
// @Description Get an idea by its ID
// @Tags ideas
// @Produce json
// @Param id path string true "Idea ID"
// @Success 200 {object} IdeaResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /ideas/{id} [get]
func (h *IdeaHandler) GetIdea(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract idea ID from URL path
	ideaID := extractIDFromPath(r.URL.Path, "/ideas/")
	if ideaID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Idea ID is required")
		return
	}

	// Get the idea
	foundIdea, err := h.ideaService.GetIdea(r.Context(), ideaID)
	if err != nil {
		switch err {
		case idea.ErrIdeaNotFound:
			writeErrorResponse(w, http.StatusNotFound, "Idea not found")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return the idea
	response := IdeaResponse{
		ID:        foundIdea.ID,
		IdeaText:  foundIdea.IdeaText,
		UserID:    foundIdea.UserID,
		CreatedAt: foundIdea.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListIdeas handles GET /ideas requests.
// AI-hint: Idea listing endpoint with proper JSON array response handling.
//
// @Summary List all ideas
// @Description Get all ideas ordered by creation date (most recent first)
// @Tags ideas
// @Produce json
// @Success 200 {array} IdeaResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /ideas [get]
func (h *IdeaHandler) ListIdeas(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get all ideas
	ideas, err := h.ideaService.ListIdeas(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Convert to response format
	var responses []IdeaResponse
	for _, idea := range ideas {
		responses = append(responses, IdeaResponse{
			ID:        idea.ID,
			IdeaText:  idea.IdeaText,
			UserID:    idea.UserID,
			CreatedAt: idea.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetIdeasByUser handles GET /users/{id}/ideas requests.
// AI-hint: User-specific idea listing endpoint with proper error handling.
//
// @Summary Get ideas by user
// @Description Get all ideas submitted by a specific user
// @Tags ideas
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {array} IdeaResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /users/{id}/ideas [get]
func (h *IdeaHandler) GetIdeasByUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by authentication middleware)
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract user ID from URL path
	targetUserID := extractIDFromPath(r.URL.Path, "/users/")
	if targetUserID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Remove "/ideas" suffix if present
	if strings.HasSuffix(targetUserID, "/ideas") {
		targetUserID = strings.TrimSuffix(targetUserID, "/ideas")
	}

	// Get ideas by user
	ideas, err := h.ideaService.GetIdeasByUser(r.Context(), targetUserID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Convert to response format
	var responses []IdeaResponse
	for _, idea := range ideas {
		responses = append(responses, IdeaResponse{
			ID:        idea.ID,
			IdeaText:  idea.IdeaText,
			UserID:    idea.UserID,
			CreatedAt: idea.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
