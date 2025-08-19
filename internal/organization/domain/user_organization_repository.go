package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// UserOrganizationRepository defines the interface for user-organization relationship persistence operations.
// AI-hint: Repository interface following DDD patterns. Defines all CRUD operations
// needed for user-organization relationship management. Implementations should handle database-specific
// concerns while maintaining domain consistency.
type UserOrganizationRepository interface {
	// Create stores a new user-organization relationship in the repository.
	// AI-hint: Creates a new user-organization relationship record. Should validate that the
	// user-organization combination is unique and handle any database constraints.
	Create(ctx context.Context, userOrg *UserOrganization) error

	// GetByID retrieves a user-organization relationship by its unique identifier.
	// AI-hint: Primary lookup method for user-organization relationships. Returns the relationship
	// or an error if not found.
	GetByID(ctx context.Context, id uuid.UUID) (*UserOrganization, error)

	// GetByUserAndOrganization retrieves a user-organization relationship by user and organization IDs.
	// AI-hint: Alternative lookup method using the composite key of user and organization.
	// Useful for checking if a user belongs to a specific organization.
	GetByUserAndOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*UserOrganization, error)

	// GetByUser retrieves all organizations a user belongs to.
	// AI-hint: Retrieves all user-organization relationships for a specific user.
	// Useful for user dashboard and organization switching functionality.
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*UserOrganization, error)

	// GetByOrganization retrieves all users in a specific organization.
	// AI-hint: Retrieves all user-organization relationships for a specific organization.
	// Useful for organization management and user listing within organizations.
	GetByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*UserOrganization, error)

	// GetActiveByUser retrieves all active user-organization relationships for a user.
	// AI-hint: Retrieves only active relationships for a user. Useful for filtering
	// out inactive or suspended relationships.
	GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*UserOrganization, error)

	// GetActiveByOrganization retrieves all active users in a specific organization.
	// AI-hint: Retrieves only active user relationships in an organization. Useful for
	// filtering out inactive or suspended users.
	GetActiveByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*UserOrganization, error)

	// Update modifies an existing user-organization relationship in the repository.
	// AI-hint: Updates an existing relationship. Should validate that the relationship
	// exists and handle optimistic locking if needed.
	Update(ctx context.Context, userOrg *UserOrganization) error

	// UpdateRole updates the user's role within an organization.
	// AI-hint: Convenience method for updating just the role of a user-organization relationship.
	// Should be more efficient than full updates when only the role changes.
	UpdateRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error

	// UpdateActiveStatus updates the active status of a user-organization relationship.
	// AI-hint: Convenience method for activating/deactivating user-organization relationships.
	// Useful for temporary suspensions or reactivations.
	UpdateActiveStatus(ctx context.Context, userID, organizationID uuid.UUID, isActive bool) error

	// Delete removes a user-organization relationship from the repository.
	// AI-hint: Removes a user-organization relationship. Should handle cascading deletes
	// for related entities if needed.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndOrganization removes a user-organization relationship by user and organization IDs.
	// AI-hint: Alternative delete method using the composite key. Useful for removing
	// users from organizations without knowing the relationship ID.
	DeleteByUserAndOrganization(ctx context.Context, userID, organizationID uuid.UUID) error

	// List retrieves a paginated list of user-organization relationships.
	// AI-hint: Retrieves relationships with pagination support. Useful for admin
	// interfaces and bulk operations.
	List(ctx context.Context, limit, offset int) ([]*UserOrganization, error)

	// Count returns the total number of user-organization relationships in the repository.
	// AI-hint: Utility method for pagination and statistics. Should be efficient
	// and not require loading all relationship data.
	Count(ctx context.Context) (int, error)

	// CountByUser returns the number of organizations a user belongs to.
	// AI-hint: Utility method for user statistics. Useful for enforcing limits
	// on the number of organizations a user can join.
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// CountByOrganization returns the number of users in a specific organization.
	// AI-hint: Utility method for organization statistics. Useful for organization
	// size limits and analytics.
	CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int, error)

	// Exists checks if a user-organization relationship exists.
	// AI-hint: Efficient method for checking relationship existence without
	// loading the full relationship data. Useful for validation and business logic.
	Exists(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)

	// ExistsActive checks if an active user-organization relationship exists.
	// AI-hint: Efficient method for checking active relationship existence.
	// Useful for access control and permission checking.
	ExistsActive(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)
}

// Repository error types for user-organization operations.
// AI-hint: Domain-specific error types that provide clear information about
// what went wrong during repository operations. These errors should be
// wrapped with additional context when propagated up the call stack.

// ErrUserOrganizationNotFound is returned when a user-organization relationship cannot be found.
var ErrUserOrganizationNotFound = errors.New("user-organization relationship not found")

// ErrUserOrganizationAlreadyExists is returned when trying to create a relationship that already exists.
var ErrUserOrganizationAlreadyExists = errors.New("user-organization relationship already exists")

// ErrUserOrganizationInvalidData is returned when the user-organization relationship data is invalid or malformed.
var ErrUserOrganizationInvalidData = errors.New("invalid user-organization relationship data")

// ErrUserOrganizationInactive is returned when trying to perform operations on an inactive relationship.
var ErrUserOrganizationInactive = errors.New("user-organization relationship is inactive")

// ErrUserOrganizationLimitExceeded is returned when a user tries to join more organizations than allowed.
var ErrUserOrganizationLimitExceeded = errors.New("user organization limit exceeded")

// ErrOrganizationUserLimitExceeded is returned when an organization reaches its maximum user capacity.
var ErrOrganizationUserLimitExceeded = errors.New("organization user limit exceeded")

// Repository operation constants for pagination and limits.
// AI-hint: Constants that define reasonable limits for repository operations
// to prevent performance issues and ensure consistent behavior.

const (
	// DefaultUserOrganizationListLimit is the default number of user-organization relationships to return in list operations.
	DefaultUserOrganizationListLimit = 50

	// MaxUserOrganizationListLimit is the maximum number of user-organization relationships that can be returned in a single list operation.
	MaxUserOrganizationListLimit = 1000

	// DefaultUserOrganizationListOffset is the default offset for pagination.
	DefaultUserOrganizationListOffset = 0

	// MaxOrganizationsPerUser is the maximum number of organizations a user can belong to.
	MaxOrganizationsPerUser = 10

	// MaxUsersPerOrganization is the maximum number of users that can belong to an organization.
	MaxUsersPerOrganization = 10000
)
