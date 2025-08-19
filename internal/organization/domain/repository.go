package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// OrganizationRepository defines the interface for organization persistence operations.
// AI-hint: Repository interface following DDD patterns. Defines all CRUD operations
// needed for organization management. Implementations should handle database-specific
// concerns while maintaining domain consistency.
type OrganizationRepository interface {
	// Create stores a new organization in the repository.
	// AI-hint: Creates a new organization record. Should validate that the slug
	// is unique and handle any database constraints.
	Create(ctx context.Context, org *Organization) error

	// GetByID retrieves an organization by its unique identifier.
	// AI-hint: Primary lookup method for organizations. Returns the organization
	// or an error if not found.
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)

	// GetBySlug retrieves an organization by its unique slug.
	// AI-hint: Alternative lookup method using the human-readable slug.
	// Useful for URL-based routing and user-friendly identifiers.
	GetBySlug(ctx context.Context, slug string) (*Organization, error)

	// Update modifies an existing organization in the repository.
	// AI-hint: Updates an existing organization. Should validate that the slug
	// remains unique if changed and handle optimistic locking if needed.
	Update(ctx context.Context, org *Organization) error

	// Delete removes an organization from the repository.
	// AI-hint: Removes an organization. Should handle cascading deletes for
	// related entities (users, roles, ideas) or return an error if dependencies exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a paginated list of organizations.
	// AI-hint: Retrieves organizations with pagination support. Useful for
	// admin interfaces and bulk operations.
	List(ctx context.Context, limit, offset int) ([]*Organization, error)

	// Count returns the total number of organizations in the repository.
	// AI-hint: Utility method for pagination and statistics. Should be efficient
	// and not require loading all organization data.
	Count(ctx context.Context) (int, error)
}

// Repository error types for organization operations.
// AI-hint: Domain-specific error types that provide clear information about
// what went wrong during repository operations. These errors should be
// wrapped with additional context when propagated up the call stack.

// ErrOrganizationNotFound is returned when an organization cannot be found.
var ErrOrganizationNotFound = errors.New("organization not found")

// ErrOrganizationAlreadyExists is returned when trying to create an organization that already exists.
var ErrOrganizationAlreadyExists = errors.New("organization already exists")

// ErrOrganizationSlugAlreadyExists is returned when trying to create or update an organization with a slug that already exists.
var ErrOrganizationSlugAlreadyExists = errors.New("organization slug already exists")

// ErrInvalidOrganizationData is returned when the organization data is invalid or malformed.
var ErrInvalidOrganizationData = errors.New("invalid organization data")

// ErrOrganizationHasDependencies is returned when trying to delete an organization that has dependent entities.
var ErrOrganizationHasDependencies = errors.New("cannot delete organization with dependent entities")

// ErrOrganizationInactive is returned when trying to perform operations on an inactive organization.
var ErrOrganizationInactive = errors.New("organization is inactive")

// Repository operation constants for pagination and limits.
// AI-hint: Constants that define reasonable limits for repository operations
// to prevent performance issues and ensure consistent behavior.

const (
	// DefaultListLimit is the default number of organizations to return in list operations.
	DefaultListLimit = 50

	// MaxListLimit is the maximum number of organizations that can be returned in a single list operation.
	MaxListLimit = 1000

	// DefaultListOffset is the default offset for pagination.
	DefaultListOffset = 0
)
