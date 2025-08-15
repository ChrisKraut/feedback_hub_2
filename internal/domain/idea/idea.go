package idea

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Idea represents a feedback idea in the system.
// AI-hint: Core domain entity for feedback ideas with business logic and invariants.
// Enforces title/content validation and maintains creator relationship integrity.
type Idea struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatorUserID uuid.UUID `json:"creator_user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewIdea creates a new Idea with validation.
// AI-hint: Factory method that enforces business rules during idea creation.
// Validates title and content requirements, generates UUID, and sets timestamps.
func NewIdea(title, content string, creatorUserID uuid.UUID) (*Idea, error) {
	if title == "" {
		return nil, errors.New("idea title cannot be empty")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("idea title cannot be empty")
	}
	if content == "" {
		return nil, errors.New("idea content cannot be empty")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("idea content cannot be empty")
	}
	if creatorUserID == uuid.Nil {
		return nil, errors.New("creator user ID cannot be empty")
	}

	now := time.Now()
	return &Idea{
		ID:            uuid.New(),
		Title:         strings.TrimSpace(title),
		Content:       strings.TrimSpace(content),
		CreatorUserID: creatorUserID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// NewIdeaWithID creates a new Idea with a specific ID (for testing/imports).
// AI-hint: Factory method for scenarios requiring specific ID assignment.
// Useful for testing, data migration, or external system integration.
func NewIdeaWithID(id uuid.UUID, title, content string, creatorUserID uuid.UUID) (*Idea, error) {
	if id == uuid.Nil {
		return nil, errors.New("idea ID cannot be empty")
	}
	if title == "" {
		return nil, errors.New("idea title cannot be empty")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("idea title cannot be empty")
	}
	if content == "" {
		return nil, errors.New("idea content cannot be empty")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("idea content cannot be empty")
	}
	if creatorUserID == uuid.Nil {
		return nil, errors.New("creator user ID cannot be empty")
	}

	now := time.Now()
	return &Idea{
		ID:            id,
		Title:         strings.TrimSpace(title),
		Content:       strings.TrimSpace(content),
		CreatorUserID: creatorUserID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// UpdateTitle updates the idea's title with validation.
// AI-hint: Domain method that maintains business invariants during updates.
// Ensures title remains non-empty and updates the modification timestamp.
func (i *Idea) UpdateTitle(title string) error {
	if title == "" {
		return errors.New("idea title cannot be empty")
	}
	if strings.TrimSpace(title) == "" {
		return errors.New("idea title cannot be empty")
	}
	i.Title = strings.TrimSpace(title)
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateContent updates the idea's content with validation.
// AI-hint: Domain method that maintains business invariants during updates.
// Ensures content remains non-empty and updates the modification timestamp.
func (i *Idea) UpdateContent(content string) error {
	if content == "" {
		return errors.New("idea content cannot be empty")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("idea content cannot be empty")
	}
	i.Content = strings.TrimSpace(content)
	i.UpdatedAt = time.Now()
	return nil
}

// Repository defines the interface for idea persistence operations.
// AI-hint: Repository pattern interface for dependency inversion.
// Keeps domain logic independent of persistence implementation.
type Repository interface {
	Save(ctx interface{}, idea *Idea) error
	FindByID(ctx interface{}, id uuid.UUID) (*Idea, error)
	FindByCreatorUserID(ctx interface{}, creatorUserID uuid.UUID) ([]*Idea, error)
	FindAll(ctx interface{}) ([]*Idea, error)
	Delete(ctx interface{}, id uuid.UUID) error
}

// Service defines the business operations for idea management.
// AI-hint: Domain service interface for complex business operations
// that don't naturally belong to a single entity or require coordination.
type Service interface {
	CreateIdea(ctx interface{}, title, content string, creatorUserID uuid.UUID) (*Idea, error)
	GetIdea(ctx interface{}, id uuid.UUID) (*Idea, error)
	UpdateIdea(ctx interface{}, id uuid.UUID, title, content string, updatedByUserID uuid.UUID) (*Idea, error)
	DeleteIdea(ctx interface{}, id uuid.UUID, deletedByUserID uuid.UUID) error
	ListIdeas(ctx interface{}) ([]*Idea, error)
	ListIdeasByCreator(ctx interface{}, creatorUserID uuid.UUID) ([]*Idea, error)
}

// Error types for the idea domain.
// AI-hint: Domain-specific errors for clear error handling and business rules.
var (
	ErrIdeaNotFound    = errors.New("idea not found")
	ErrInvalidIdeaData = errors.New("invalid idea data")
	ErrUnauthorized    = errors.New("unauthorized operation")
	ErrCreatorNotFound = errors.New("creator user not found")
)
