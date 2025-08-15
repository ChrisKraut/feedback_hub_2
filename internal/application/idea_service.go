package application

import (
	"context"
	"feedback_hub_2/internal/domain/idea"
	"feedback_hub_2/internal/domain/user"

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

// IdeaApplicationService handles business logic for idea operations.
// AI-hint: Application service that orchestrates domain entities and repositories.
// Implements business rules and validation logic for idea management.
type IdeaApplicationService struct {
	ideaRepo idea.Repository
	userRepo user.Repository
}

// NewIdeaApplicationService creates a new IdeaApplicationService instance.
// AI-hint: Factory method with dependency injection for repositories.
// Ensures proper initialization and dependency management.
func NewIdeaApplicationService(ideaRepo idea.Repository, userRepo user.Repository) *IdeaApplicationService {
	return &IdeaApplicationService{
		ideaRepo: ideaRepo,
		userRepo: userRepo,
	}
}

// CreateIdea handles the creation of a new idea with validation.
// AI-hint: Core business logic method that validates input, creates domain entity,
// and persists it through the repository layer.
func (s *IdeaApplicationService) CreateIdea(ctx context.Context, cmd CreateIdeaCommand) (uuid.UUID, error) {
	// Validate the command
	if err := s.validateCreateIdeaCommand(cmd); err != nil {
		return uuid.Nil, err
	}

	// Verify that the creator user exists
	if err := s.verifyCreatorExists(ctx, cmd.CreatorUserID); err != nil {
		return uuid.Nil, err
	}

	// Create the idea domain entity
	newIdea, err := idea.NewIdea(cmd.Title, cmd.Content, cmd.CreatorUserID)
	if err != nil {
		return uuid.Nil, err
	}

	// Persist the idea
	if err := s.ideaRepo.Save(ctx, newIdea); err != nil {
		return uuid.Nil, err
	}

	return newIdea.ID, nil
}

// validateCreateIdeaCommand validates the create idea command.
// AI-hint: Private validation method that enforces business rules for idea creation.
// Centralizes validation logic for maintainability and consistency.
func (s *IdeaApplicationService) validateCreateIdeaCommand(cmd CreateIdeaCommand) error {
	if cmd.Title == "" {
		return idea.ErrInvalidIdeaData
	}
	if cmd.Content == "" {
		return idea.ErrInvalidIdeaData
	}
	if cmd.CreatorUserID == uuid.Nil {
		return idea.ErrInvalidIdeaData
	}
	return nil
}

// verifyCreatorExists verifies that the creator user exists in the system.
// AI-hint: Business rule validation that ensures referential integrity.
// Prevents creation of ideas with non-existent users.
func (s *IdeaApplicationService) verifyCreatorExists(ctx context.Context, creatorUserID uuid.UUID) error {
	_, err := s.userRepo.GetByID(ctx, creatorUserID.String())
	if err != nil {
		if err == user.ErrUserNotFound {
			return idea.ErrCreatorNotFound
		}
		return err
	}
	return nil
}
