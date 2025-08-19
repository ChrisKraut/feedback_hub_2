package infrastructure

import (
	"context"
	"fmt"
	"time"

	"feedback_hub_2/internal/organization/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserOrganizationRepository implements the user-organization repository interface using PostgreSQL.
// AI-hint: PostgreSQL implementation of the user-organization repository with optimized
// connection pooling for Vercel serverless functions and Supabase integration.
type UserOrganizationRepository struct {
	pool *pgxpool.Pool
}

// NewUserOrganizationRepository creates a new user-organization repository instance.
// AI-hint: Factory method for creating user-organization repository with dependency injection
// of the database connection pool.
func NewUserOrganizationRepository(pool *pgxpool.Pool) *UserOrganizationRepository {
	return &UserOrganizationRepository{
		pool: pool,
	}
}

// Create stores a new user-organization relationship in the repository.
// AI-hint: Creates a new user-organization relationship record with validation and constraint checking.
// Handles uniqueness constraints and database constraints.
func (r *UserOrganizationRepository) Create(ctx context.Context, userOrg *domain.UserOrganization) error {
	if userOrg == nil {
		return domain.ErrUserOrganizationInvalidData
	}

	if err := userOrg.Validate(); err != nil {
		return fmt.Errorf("invalid user-organization data: %w", err)
	}

	// Check if relationship already exists
	exists, err := r.Exists(ctx, userOrg.UserID, userOrg.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check relationship existence: %w", err)
	}
	if exists {
		return domain.ErrUserOrganizationAlreadyExists
	}

	// Check user organization limit
	count, err := r.CountByUser(ctx, userOrg.UserID)
	if err != nil {
		return fmt.Errorf("failed to check user organization count: %w", err)
	}
	if count >= domain.MaxOrganizationsPerUser {
		return domain.ErrUserOrganizationLimitExceeded
	}

	// Check organization user limit
	count, err = r.CountByOrganization(ctx, userOrg.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check organization user count: %w", err)
	}
	if count >= domain.MaxUsersPerOrganization {
		return domain.ErrOrganizationUserLimitExceeded
	}

	query := `
		INSERT INTO user_organizations (id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.pool.Exec(ctx, query,
		userOrg.ID,
		userOrg.UserID,
		userOrg.OrganizationID,
		userOrg.RoleID,
		userOrg.IsActive,
		userOrg.JoinedAt,
		userOrg.CreatedAt,
		userOrg.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user-organization relationship: %w", err)
	}

	return nil
}

// GetByID retrieves a user-organization relationship by its unique identifier.
// AI-hint: Primary lookup method for user-organization relationships using UUID. Returns the relationship
// or an error if not found. Optimized for connection pooling.
func (r *UserOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserOrganization, error) {
	if id == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE id = $1
	`

	var userOrg domain.UserOrganization

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&userOrg.ID,
		&userOrg.UserID,
		&userOrg.OrganizationID,
		&userOrg.RoleID,
		&userOrg.IsActive,
		&userOrg.JoinedAt,
		&userOrg.CreatedAt,
		&userOrg.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get user-organization relationship: %w", err)
	}

	return &userOrg, nil
}

// GetByUserAndOrganization retrieves a user-organization relationship by user and organization IDs.
// AI-hint: Alternative lookup method using the composite key of user and organization.
// Useful for checking if a user belongs to a specific organization.
func (r *UserOrganizationRepository) GetByUserAndOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*domain.UserOrganization, error) {
	if userID == uuid.Nil || organizationID == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE user_id = $1 AND organization_id = $2
	`

	var userOrg domain.UserOrganization

	err := r.pool.QueryRow(ctx, query, userID, organizationID).Scan(
		&userOrg.ID,
		&userOrg.UserID,
		&userOrg.OrganizationID,
		&userOrg.RoleID,
		&userOrg.IsActive,
		&userOrg.JoinedAt,
		&userOrg.CreatedAt,
		&userOrg.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get user-organization relationship: %w", err)
	}

	return &userOrg, nil
}

// GetByUser retrieves all organizations a user belongs to.
// AI-hint: Retrieves all user-organization relationships for a specific user.
// Useful for user dashboard and organization switching functionality.
func (r *UserOrganizationRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserOrganization, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user-organization relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*domain.UserOrganization
	for rows.Next() {
		var userOrg domain.UserOrganization
		err := rows.Scan(
			&userOrg.ID,
			&userOrg.UserID,
			&userOrg.OrganizationID,
			&userOrg.RoleID,
			&userOrg.IsActive,
			&userOrg.JoinedAt,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user-organization relationship: %w", err)
		}
		relationships = append(relationships, &userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user-organization relationships: %w", err)
	}

	return relationships, nil
}

// GetByOrganization retrieves all users in a specific organization.
// AI-hint: Retrieves all user-organization relationships for a specific organization.
// Useful for organization management and user listing within organizations.
func (r *UserOrganizationRepository) GetByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.UserOrganization, error) {
	if organizationID == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE organization_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user-organization relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*domain.UserOrganization
	for rows.Next() {
		var userOrg domain.UserOrganization
		err := rows.Scan(
			&userOrg.ID,
			&userOrg.UserID,
			&userOrg.OrganizationID,
			&userOrg.RoleID,
			&userOrg.IsActive,
			&userOrg.JoinedAt,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user-organization relationship: %w", err)
		}
		relationships = append(relationships, &userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user-organization relationships: %w", err)
	}

	return relationships, nil
}

// GetActiveByUser retrieves all active user-organization relationships for a user.
// AI-hint: Retrieves only active relationships for a user. Useful for filtering
// out inactive or suspended relationships.
func (r *UserOrganizationRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserOrganization, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE user_id = $1 AND is_active = true
		ORDER BY joined_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user-organization relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*domain.UserOrganization
	for rows.Next() {
		var userOrg domain.UserOrganization
		err := rows.Scan(
			&userOrg.ID,
			&userOrg.UserID,
			&userOrg.OrganizationID,
			&userOrg.RoleID,
			&userOrg.IsActive,
			&userOrg.JoinedAt,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user-organization relationship: %w", err)
		}
		relationships = append(relationships, &userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over active user-organization relationships: %w", err)
	}

	return relationships, nil
}

// GetActiveByOrganization retrieves all active users in a specific organization.
// AI-hint: Retrieves only active user relationships in an organization. Useful for
// filtering out inactive or suspended users.
func (r *UserOrganizationRepository) GetActiveByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.UserOrganization, error) {
	if organizationID == uuid.Nil {
		return nil, domain.ErrUserOrganizationInvalidData
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		WHERE organization_id = $1 AND is_active = true
		ORDER BY joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user-organization relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*domain.UserOrganization
	for rows.Next() {
		var userOrg domain.UserOrganization
		err := rows.Scan(
			&userOrg.ID,
			&userOrg.UserID,
			&userOrg.OrganizationID,
			&userOrg.RoleID,
			&userOrg.IsActive,
			&userOrg.JoinedAt,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user-organization relationship: %w", err)
		}
		relationships = append(relationships, &userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over active user-organization relationships: %w", err)
	}

	return relationships, nil
}

// Update modifies an existing user-organization relationship in the repository.
// AI-hint: Updates an existing relationship. Should validate that the relationship
// exists and handle optimistic locking if needed.
func (r *UserOrganizationRepository) Update(ctx context.Context, userOrg *domain.UserOrganization) error {
	if userOrg == nil {
		return domain.ErrUserOrganizationInvalidData
	}

	if err := userOrg.Validate(); err != nil {
		return fmt.Errorf("invalid user-organization data: %w", err)
	}

	// Check if relationship exists
	exists, err := r.Exists(ctx, userOrg.UserID, userOrg.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to check relationship existence: %w", err)
	}
	if !exists {
		return domain.ErrUserOrganizationNotFound
	}

	query := `
		UPDATE user_organizations
		SET role_id = $1, is_active = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.pool.Exec(ctx, query,
		userOrg.RoleID,
		userOrg.IsActive,
		userOrg.UpdatedAt,
		userOrg.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user-organization relationship: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserOrganizationNotFound
	}

	return nil
}

// UpdateRole updates the user's role within an organization.
// AI-hint: Convenience method for updating just the role of a user-organization relationship.
// Should be more efficient than full updates when only the role changes.
func (r *UserOrganizationRepository) UpdateRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	if userID == uuid.Nil || organizationID == uuid.Nil || roleID == uuid.Nil {
		return domain.ErrUserOrganizationInvalidData
	}

	query := `
		UPDATE user_organizations
		SET role_id = $1, updated_at = $2
		WHERE user_id = $3 AND organization_id = $4
	`

	result, err := r.pool.Exec(ctx, query, roleID, time.Now(), userID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update user-organization role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserOrganizationNotFound
	}

	return nil
}

// UpdateActiveStatus updates the active status of a user-organization relationship.
// AI-hint: Convenience method for activating/deactivating user-organization relationships.
// Useful for temporary suspensions or reactivations.
func (r *UserOrganizationRepository) UpdateActiveStatus(ctx context.Context, userID, organizationID uuid.UUID, isActive bool) error {
	if userID == uuid.Nil || organizationID == uuid.Nil {
		return domain.ErrUserOrganizationInvalidData
	}

	query := `
		UPDATE user_organizations
		SET is_active = $1, updated_at = $2
		WHERE user_id = $3 AND organization_id = $4
	`

	result, err := r.pool.Exec(ctx, query, isActive, time.Now(), userID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update user-organization active status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserOrganizationNotFound
	}

	return nil
}

// Delete removes a user-organization relationship from the repository.
// AI-hint: Removes a user-organization relationship. Should handle cascading deletes
// for related entities if needed.
func (r *UserOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrUserOrganizationInvalidData
	}

	query := `DELETE FROM user_organizations WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user-organization relationship: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserOrganizationNotFound
	}

	return nil
}

// DeleteByUserAndOrganization removes a user-organization relationship by user and organization IDs.
// AI-hint: Alternative delete method using the composite key. Useful for removing
// users from organizations without knowing the relationship ID.
func (r *UserOrganizationRepository) DeleteByUserAndOrganization(ctx context.Context, userID, organizationID uuid.UUID) error {
	if userID == uuid.Nil || organizationID == uuid.Nil {
		return domain.ErrUserOrganizationInvalidData
	}

	query := `DELETE FROM user_organizations WHERE user_id = $1 AND organization_id = $2`

	result, err := r.pool.Exec(ctx, query, userID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete user-organization relationship: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserOrganizationNotFound
	}

	return nil
}

// List retrieves a paginated list of user-organization relationships.
// AI-hint: Retrieves relationships with pagination support. Useful for admin
// interfaces and bulk operations.
func (r *UserOrganizationRepository) List(ctx context.Context, limit, offset int) ([]*domain.UserOrganization, error) {
	// Apply limits
	if limit <= 0 {
		limit = domain.DefaultUserOrganizationListLimit
	}
	if limit > domain.MaxUserOrganizationListLimit {
		limit = domain.MaxUserOrganizationListLimit
	}
	if offset < 0 {
		offset = domain.DefaultUserOrganizationListOffset
	}

	query := `
		SELECT id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at
		FROM user_organizations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user-organization relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*domain.UserOrganization
	for rows.Next() {
		var userOrg domain.UserOrganization
		err := rows.Scan(
			&userOrg.ID,
			&userOrg.UserID,
			&userOrg.OrganizationID,
			&userOrg.RoleID,
			&userOrg.IsActive,
			&userOrg.JoinedAt,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user-organization relationship: %w", err)
		}
		relationships = append(relationships, &userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user-organization relationships: %w", err)
	}

	return relationships, nil
}

// Count returns the total number of user-organization relationships in the repository.
// AI-hint: Utility method for pagination and statistics. Should be efficient
// and not require loading all relationship data.
func (r *UserOrganizationRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM user_organizations`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user-organization relationships: %w", err)
	}

	return count, nil
}

// CountByUser returns the number of organizations a user belongs to.
// AI-hint: Utility method for user statistics. Useful for enforcing limits
// on the number of organizations a user can join.
func (r *UserOrganizationRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil {
		return 0, domain.ErrUserOrganizationInvalidData
	}

	query := `SELECT COUNT(*) FROM user_organizations WHERE user_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user organizations: %w", err)
	}

	return count, nil
}

// CountByOrganization returns the number of users in a specific organization.
// AI-hint: Utility method for organization statistics. Useful for organization
// size limits and analytics.
func (r *UserOrganizationRepository) CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int, error) {
	if organizationID == uuid.Nil {
		return 0, domain.ErrUserOrganizationInvalidData
	}

	query := `SELECT COUNT(*) FROM user_organizations WHERE organization_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, organizationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organization users: %w", err)
	}

	return count, nil
}

// Exists checks if a user-organization relationship exists.
// AI-hint: Efficient method for checking relationship existence without
// loading the full relationship data. Useful for validation and business logic.
func (r *UserOrganizationRepository) Exists(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || organizationID == uuid.Nil {
		return false, domain.ErrUserOrganizationInvalidData
	}

	query := `SELECT EXISTS(SELECT 1 FROM user_organizations WHERE user_id = $1 AND organization_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, organizationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user-organization relationship existence: %w", err)
	}

	return exists, nil
}

// ExistsActive checks if an active user-organization relationship exists.
// AI-hint: Efficient method for checking active relationship existence.
// Useful for access control and permission checking.
func (r *UserOrganizationRepository) ExistsActive(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || organizationID == uuid.Nil {
		return false, domain.ErrUserOrganizationInvalidData
	}

	query := `SELECT EXISTS(SELECT 1 FROM user_organizations WHERE user_id = $1 AND organization_id = $2 AND is_active = true)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, organizationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active user-organization relationship existence: %w", err)
	}

	return exists, nil
}
