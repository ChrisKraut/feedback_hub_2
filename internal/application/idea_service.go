package application

import (
	"context"
	"feedback_hub_2/internal/domain/idea"
	"feedback_hub_2/internal/domain/user"

	"github.com/google/uuid"
)

// IdeaService implements the idea.Service interface and coordinates idea management operations.
// AI-hint: Application service that orchestrates idea business logic with authorization checks.
// Enforces business rules about who can perform what operations on ideas.
type IdeaService struct {
	ideaRepo idea.Repository
	userRepo user.Repository
}

// NewIdeaService creates a new IdeaService instance.
// AI-hint: Factory method for idea service with dependency injection of repositories.
func NewIdeaService(ideaRepo idea.Repository, userRepo user.Repository) *IdeaService {
	return &IdeaService{
		ideaRepo: ideaRepo,
		userRepo: userRepo,
	}
}

// CreateIdea creates a new idea with validation and user association.
// AI-hint: Idea creation with business rules - validates idea text and associates with authenticated user.
func (s *IdeaService) CreateIdea(ctx interface{}, ideaText, userID string) (*idea.Idea, error) {
	context := ctx.(context.Context)

	// Validate that the user exists
	_, err := s.userRepo.GetByID(context, userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, idea.ErrInvalidIdeaData
		}
		return nil, err
	}

	// Create the idea with a new UUID
	ideaID := uuid.New().String()
	newIdea, err := idea.NewIdea(ideaID, ideaText, userID)
	if err != nil {
		return nil, err
	}

	// Save the idea to the repository
	if err := s.ideaRepo.Create(context, newIdea); err != nil {
		return nil, err
	}

	return newIdea, nil
}

// GetIdea retrieves an idea by ID.
// AI-hint: Idea retrieval with proper error handling for not found cases.
func (s *IdeaService) GetIdea(ctx interface{}, id string) (*idea.Idea, error) {
	context := ctx.(context.Context)
	return s.ideaRepo.GetByID(context, id)
}

// GetIdeasByUser retrieves all ideas submitted by a specific user.
// AI-hint: User-specific idea listing with proper error handling.
func (s *IdeaService) GetIdeasByUser(ctx interface{}, userID string) ([]*idea.Idea, error) {
	context := ctx.(context.Context)
	return s.ideaRepo.GetByUserID(context, userID)
}

// UpdateIdea updates an idea's text with authorization checks.
// AI-hint: Idea update with permission validation - only the idea creator can update.
func (s *IdeaService) UpdateIdea(ctx interface{}, id, ideaText, userID string) (*idea.Idea, error) {
	context := ctx.(context.Context)

	// Get the existing idea
	existingIdea, err := s.ideaRepo.GetByID(context, id)
	if err != nil {
		return nil, err
	}

	// Check authorization - only the creator can update their idea
	if existingIdea.UserID != userID {
		return nil, idea.ErrUnauthorized
	}

	// Update the idea text
	if err := existingIdea.UpdateIdeaText(ideaText); err != nil {
		return nil, err
	}

	// Save the updated idea
	if err := s.ideaRepo.Update(context, existingIdea); err != nil {
		return nil, err
	}

	return existingIdea, nil
}

// DeleteIdea deletes an idea with authorization checks.
// AI-hint: Idea deletion with permission validation - only the idea creator can delete.
func (s *IdeaService) DeleteIdea(ctx interface{}, id, userID string) error {
	context := ctx.(context.Context)

	// Get the existing idea
	existingIdea, err := s.ideaRepo.GetByID(context, id)
	if err != nil {
		return err
	}

	// Check authorization - only the creator can delete their idea
	if existingIdea.UserID != userID {
		return idea.ErrUnauthorized
	}

	// Delete the idea
	return s.ideaRepo.Delete(context, id)
}

// ListIdeas retrieves all ideas ordered by creation date (most recent first).
// AI-hint: Idea listing with proper ordering for chronological display.
func (s *IdeaService) ListIdeas(ctx interface{}) ([]*idea.Idea, error) {
	context := ctx.(context.Context)
	return s.ideaRepo.List(context)
}
