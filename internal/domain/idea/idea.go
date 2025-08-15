package idea

import (
	"errors"
	"strings"
	"time"
)

// Idea represents a user-submitted idea with markdown-formatted text.
// AI-hint: Core domain entity for idea management with business rules around
// text validation, user association, and creation timestamps.
type Idea struct {
	ID        string    `json:"id"`
	IdeaText  string    `json:"idea_text"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewIdea creates a new Idea with validation.
// AI-hint: Factory method that enforces business rules during idea creation.
// Validates idea text is not empty and ensures required fields are present.
func NewIdea(id, ideaText, userID string) (*Idea, error) {
	if id == "" {
		return nil, errors.New("idea ID cannot be empty")
	}
	if strings.TrimSpace(ideaText) == "" {
		return nil, errors.New("idea text cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Sanitize the idea text to prevent XSS attacks
	sanitizedText := sanitizeMarkdown(ideaText)
	if strings.TrimSpace(sanitizedText) == "" {
		return nil, errors.New("idea text cannot be empty after sanitization")
	}

	now := time.Now()
	return &Idea{
		ID:        id,
		IdeaText:  sanitizedText,
		UserID:    userID,
		CreatedAt: now,
	}, nil
}

// UpdateIdeaText updates the idea's text with validation and sanitization.
// AI-hint: Domain method that maintains business invariants during updates.
// Ensures text remains valid and sanitized after modification.
func (i *Idea) UpdateIdeaText(ideaText string) error {
	if strings.TrimSpace(ideaText) == "" {
		return errors.New("idea text cannot be empty")
	}

	// Sanitize the new text
	sanitizedText := sanitizeMarkdown(ideaText)
	if strings.TrimSpace(sanitizedText) == "" {
		return errors.New("idea text cannot be empty after sanitization")
	}

	i.IdeaText = sanitizedText
	return nil
}

// sanitizeMarkdown removes potentially dangerous HTML/script content from markdown text.
// AI-hint: Security function that prevents XSS attacks while preserving markdown formatting.
// Strips script tags and other dangerous HTML elements while keeping safe markdown syntax.
func sanitizeMarkdown(text string) string {
	// Remove script tags and their content
	text = removeScriptTags(text)

	// Remove other potentially dangerous HTML tags
	text = removeDangerousTags(text)

	// Trim whitespace
	return strings.TrimSpace(text)
}

// removeScriptTags removes script tags and their content from the text.
// AI-hint: Security function that specifically targets script injection attempts.
func removeScriptTags(text string) string {
	// Simple script tag removal - in production, consider using a proper HTML parser
	lowerText := strings.ToLower(text)

	// Remove <script> tags and content
	scriptStart := strings.Index(lowerText, "<script")
	if scriptStart != -1 {
		scriptEnd := strings.Index(lowerText[scriptStart:], "</script>")
		if scriptEnd != -1 {
			scriptEnd += scriptStart + 8 // length of "</script>"
			text = text[:scriptStart] + text[scriptEnd:]
		} else {
			// If no closing tag, remove from start to end
			text = text[:scriptStart]
		}
	}

	return text
}

// removeDangerousTags removes other potentially dangerous HTML tags.
// AI-hint: Security function that removes HTML tags that could be used for XSS attacks.
func removeDangerousTags(text string) string {
	// Remove common dangerous tags
	dangerousTags := []string{"<iframe", "<object", "<embed", "<form", "<input", "<button"}

	for _, tag := range dangerousTags {
		lowerText := strings.ToLower(text)
		tagStart := strings.Index(lowerText, tag)
		if tagStart != -1 {
			// Find the closing > or space
			tagEnd := strings.IndexAny(text[tagStart:], " >")
			if tagEnd != -1 {
				tagEnd += tagStart + 1
				text = text[:tagStart] + text[tagEnd:]
			} else {
				// If no closing character, remove from start to end
				text = text[:tagStart]
			}
		}
	}

	return text
}

// Repository defines the interface for idea persistence operations.
// AI-hint: Repository pattern interface for dependency inversion.
// Keeps domain logic independent of persistence implementation.
type Repository interface {
	Create(ctx interface{}, idea *Idea) error
	GetByID(ctx interface{}, id string) (*Idea, error)
	GetByUserID(ctx interface{}, userID string) ([]*Idea, error)
	Update(ctx interface{}, idea *Idea) error
	Delete(ctx interface{}, id string) error
	List(ctx interface{}) ([]*Idea, error)
}

// Service defines the business operations for idea management.
// AI-hint: Domain service interface for complex business operations
// that require coordination between multiple entities or external services.
type Service interface {
	CreateIdea(ctx interface{}, ideaText, userID string) (*Idea, error)
	GetIdea(ctx interface{}, id string) (*Idea, error)
	GetIdeasByUser(ctx interface{}, userID string) ([]*Idea, error)
	UpdateIdea(ctx interface{}, id, ideaText, userID string) (*Idea, error)
	DeleteIdea(ctx interface{}, id, userID string) error
	ListIdeas(ctx interface{}) ([]*Idea, error)
}

// Error types for the idea domain.
// AI-hint: Domain-specific errors for clear error handling and business rules.
var (
	ErrIdeaNotFound    = errors.New("idea not found")
	ErrIdeaTextEmpty   = errors.New("idea text cannot be empty")
	ErrUnauthorized    = errors.New("unauthorized operation")
	ErrInvalidIdeaData = errors.New("invalid idea data")
)
