package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"feedback_hub_2/internal/organization/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrganizationRepository implements the organization repository interface using PostgreSQL.
// AI-hint: PostgreSQL implementation of the organization repository with optimized
// connection pooling for Vercel serverless functions and Supabase integration.
type OrganizationRepository struct {
	pool *pgxpool.Pool
}

// NewOrganizationRepository creates a new organization repository instance.
// AI-hint: Factory method for creating organization repository with dependency injection
// of the database connection pool.
func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{
		pool: pool,
	}
}

// Create stores a new organization in the repository.
// AI-hint: Creates a new organization record with validation and constraint checking.
// Handles slug uniqueness and database constraints.
func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if org == nil {
		return domain.ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return fmt.Errorf("invalid organization data: %w", err)
	}

	// Check if slug already exists
	existing, err := r.GetBySlug(ctx, org.Slug)
	if err != nil && err != domain.ErrOrganizationNotFound {
		return fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if existing != nil {
		return domain.ErrOrganizationSlugAlreadyExists
	}

	// Convert settings to JSONB
	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO organizations (id, name, slug, description, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.pool.Exec(ctx, query,
		org.ID,
		org.Name,
		org.Slug,
		org.Description,
		settingsJSON,
		org.CreatedAt,
		org.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

// GetByID retrieves an organization by its unique identifier.
// AI-hint: Primary lookup method for organizations using UUID. Returns the organization
// or an error if not found. Optimized for connection pooling.
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidOrganizationData
	}

	query := `
		SELECT id, name, slug, description, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org domain.Organization
	var settingsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Description,
		&settingsJSON,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by ID: %w", err)
	}

	// Parse settings JSONB
	if err := json.Unmarshal(settingsJSON, &org.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &org, nil
}

// GetBySlug retrieves an organization by its unique slug.
// AI-hint: Alternative lookup method using the human-readable slug. Useful for
// URL-based routing and user-friendly identifiers.
func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	if slug == "" {
		return nil, domain.ErrInvalidOrganizationData
	}

	query := `
		SELECT id, name, slug, description, settings, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`

	var org domain.Organization
	var settingsJSON []byte

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Description,
		&settingsJSON,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by slug: %w", err)
	}

	// Parse settings JSONB
	if err := json.Unmarshal(settingsJSON, &org.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &org, nil
}

// Update modifies an existing organization in the repository.
// AI-hint: Updates an existing organization with validation and constraint checking.
// Ensures slug uniqueness if changed and handles optimistic locking if needed.
func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	if org == nil {
		return domain.ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return fmt.Errorf("invalid organization data: %w", err)
	}

	// Check if organization exists
	existing, err := r.GetByID(ctx, org.ID)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			return domain.ErrOrganizationNotFound
		}
		return fmt.Errorf("failed to check organization existence: %w", err)
	}

	// Check if slug changed and if new slug already exists
	if existing.Slug != org.Slug {
		slugOrg, err := r.GetBySlug(ctx, org.Slug)
		if err != nil && err != domain.ErrOrganizationNotFound {
			return fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if slugOrg != nil && slugOrg.ID != org.ID {
			return domain.ErrOrganizationSlugAlreadyExists
		}
	}

	// Convert settings to JSONB
	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE organizations
		SET name = $1, slug = $2, description = $3, settings = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.pool.Exec(ctx, query,
		org.Name,
		org.Slug,
		org.Description,
		settingsJSON,
		org.UpdatedAt,
		org.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}

	return nil
}

// Delete removes an organization from the repository.
// AI-hint: Removes an organization with cascade behavior for related entities.
// Returns an error if dependencies exist or if the organization is not found.
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidOrganizationData
	}

	// Check if organization exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			return domain.ErrOrganizationNotFound
		}
		return fmt.Errorf("failed to check organization existence: %w", err)
	}

	query := `DELETE FROM organizations WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}

	return nil
}

// List retrieves a paginated list of organizations.
// AI-hint: Retrieves organizations with pagination support. Useful for admin
// interfaces and bulk operations. Optimized for performance with many organizations.
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	if limit < 0 || offset < 0 {
		return nil, domain.ErrInvalidOrganizationData
	}

	// Apply reasonable limits for performance
	if limit > domain.MaxListLimit {
		limit = domain.MaxListLimit
	}
	if limit == 0 {
		limit = domain.DefaultListLimit
	}

	query := `
		SELECT id, name, slug, description, settings, created_at, updated_at
		FROM organizations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*domain.Organization

	for rows.Next() {
		var org domain.Organization
		var settingsJSON []byte

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.Description,
			&settingsJSON,
			&org.CreatedAt,
			&org.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}

		// Parse settings JSONB
		if err := json.Unmarshal(settingsJSON, &org.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		organizations = append(organizations, &org)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during organization list iteration: %w", err)
	}

	return organizations, nil
}

// Count returns the total number of organizations in the repository.
// AI-hint: Utility method for pagination and statistics. Should be efficient
// and not require loading all organization data.
func (r *OrganizationRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM organizations`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	return count, nil
}

// Close closes the repository and releases resources.
// AI-hint: Cleanup method for proper resource management, especially important
// for connection pooling in serverless environments.
func (r *OrganizationRepository) Close() error {
	if r.pool != nil {
		r.pool.Close()
	}
	return nil
}

// Ping checks if the database connection is alive.
// AI-hint: Health check method for monitoring database connectivity,
// useful for Vercel serverless function health checks.
func (r *OrganizationRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return fmt.Errorf("repository not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.Ping(ctx)
}

// GetConnectionStats returns connection pool statistics.
// AI-hint: Monitoring method for connection pool health and performance,
// useful for debugging connection issues in production.
func (r *OrganizationRepository) GetConnectionStats() *pgxpool.Stat {
	if r.pool == nil {
		return nil
	}
	return r.pool.Stat()
}
