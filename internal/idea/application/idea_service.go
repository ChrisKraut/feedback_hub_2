package application

import (
	"context"
	ideadomain "feedback_hub_2/internal/idea/domain"
	"feedback_hub_2/internal/shared/queries"

	"github.com/google/uuid"
)

// CreateIdeaCommand carries the data needed to create a new idea.
// AI-hint: Command object that encapsulates all data required for idea creation.
// Follows CQRS pattern for clear separation of commands and queries.
type CreateIdeaCommand struct {
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatorUserID uuid.UUID `json:"creator_user_id"`
}

// UpdateIdeaCommand carries the data needed to update an existing idea.
// AI-hint: Command object that encapsulates all data required for idea updates.
// Follows CQRS pattern and includes authorization data for security.
type UpdateIdeaCommand struct {
	IdeaID  uuid.UUID `json:"idea_id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	UserID  uuid.UUID `json:"user_id"`
}

// IdeaApplicationService implements the idea.ApplicationService interface and coordinates idea management operations.
// AI-hint: Application service that orchestrates idea business logic.
// Uses domain events for cross-domain communication instead of direct dependencies.
type IdeaApplicationService struct {
	ideaRepo    ideadomain.Repository
	userQueries queries.UserQueries
}

// NewIdeaApplicationService creates a new IdeaApplicationService instance.
// AI-hint: Factory method for idea service with dependency injection of repositories.
func NewIdeaApplicationService(ideaRepo ideadomain.Repository, userQueries queries.UserQueries) *IdeaApplicationService {
	return &IdeaApplicationService{
		ideaRepo:    ideaRepo,
		userQueries: userQueries,
	}
}

// CreateIdea creates a new idea with validation checks.
// AI-hint: Idea creation with business rule enforcement.
func (s *IdeaApplicationService) CreateIdea(ctx interface{}, title, content string, creatorUserID string) (*ideadomain.Idea, error) {
	context := ctx.(context.Context)

	// Validate that the creator user exists using shared queries
	_, err := s.userQueries.GetUserByID(context, creatorUserID)
	if err != nil {
		return nil, ideadomain.ErrCreatorNotFound
	}

	// Create the idea
	creatorUUID := uuid.MustParse(creatorUserID)
	newIdea, err := ideadomain.NewIdea(title, content, creatorUUID)
	if err != nil {
		return nil, err
	}

	if err := s.ideaRepo.Save(context, newIdea); err != nil {
		return nil, err
	}

	return newIdea, nil
}

// UpdateIdea updates an existing idea with validation checks.
// AI-hint: Idea update with business rule enforcement.
func (s *IdeaApplicationService) UpdateIdea(ctx interface{}, ideaID uuid.UUID, title, content string, updatedByUserID string) (*ideadomain.Idea, error) {
	context := ctx.(context.Context)

	// Validate that the updater user exists using shared queries
	_, err := s.userQueries.GetUserByID(context, updatedByUserID)
	if err != nil {
		return nil, ideadomain.ErrCreatorNotFound
	}

	// Get the existing idea
	existingIdea, err := s.ideaRepo.FindByID(context, ideaID)
	if err != nil {
		return nil, err
	}

	// Update the idea using domain methods
	if err := existingIdea.UpdateTitle(title); err != nil {
		return nil, err
	}
	if err := existingIdea.UpdateContent(content); err != nil {
		return nil, err
	}

	if err := s.ideaRepo.Update(context, existingIdea); err != nil {
		return nil, err
	}

	return existingIdea, nil
}

// GetIdea retrieves an idea by ID.
func (s *IdeaApplicationService) GetIdea(ctx interface{}, ideaID uuid.UUID) (*ideadomain.Idea, error) {
	context := ctx.(context.Context)
	return s.ideaRepo.FindByID(context, ideaID)
}

// GetIdeasByCreator retrieves all ideas created by a specific user.
func (s *IdeaApplicationService) GetIdeasByCreator(ctx interface{}, creatorUserID string) ([]*ideadomain.Idea, error) {
	context := ctx.(context.Context)

	// Validate that the creator user exists using shared queries
	_, err := s.userQueries.GetUserByID(context, creatorUserID)
	if err != nil {
		return nil, ideadomain.ErrCreatorNotFound
	}

	return s.ideaRepo.FindByCreatorUserID(context, uuid.MustParse(creatorUserID))
}

// GetAllIdeas retrieves all ideas.
func (s *IdeaApplicationService) GetAllIdeas(ctx interface{}) ([]*ideadomain.Idea, error) {
	context := ctx.(context.Context)
	return s.ideaRepo.FindAll(context)
}
