package persistence

import (
	"context"
	"errors"
	roledomain "feedback_hub_2/internal/role/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleRepository implements the role.Repository interface using PostgreSQL.
// AI-hint: Persistence layer implementation for role domain using pgx for PostgreSQL.
// Provides CRUD operations and handles database-specific error translation.
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository creates a new RoleRepository instance.
// AI-hint: Factory method for role repository with dependency injection of DB pool.
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{
		pool: pool,
	}
}

// Create inserts a new role into the database.
// AI-hint: Implements role creation with proper error handling and constraint validation.
func (r *RoleRepository) Create(ctx interface{}, roleEntity *roledomain.Role) error {
	context := ctx.(context.Context)

	query := `
		INSERT INTO roles (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(context, query, roleEntity.ID, roleEntity.Name, roleEntity.CreatedAt, roleEntity.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return roledomain.ErrRoleNameAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retrieves a role by its ID.
// AI-hint: Single role retrieval with proper error handling for not found cases.
func (r *RoleRepository) GetByID(ctx interface{}, id string) (*roledomain.Role, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, name, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	var roleEntity roledomain.Role
	err := r.pool.QueryRow(context, query, id).Scan(
		&roleEntity.ID,
		&roleEntity.Name,
		&roleEntity.CreatedAt,
		&roleEntity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, roledomain.ErrRoleNotFound
		}
		return nil, err
	}

	return &roleEntity, nil
}

// GetByName retrieves a role by its name.
// AI-hint: Name-based lookup for role resolution in authorization flows.
func (r *RoleRepository) GetByName(ctx interface{}, name string) (*roledomain.Role, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, name, created_at, updated_at
		FROM roles
		WHERE name = $1
	`

	var roleEntity roledomain.Role
	err := r.pool.QueryRow(context, query, name).Scan(
		&roleEntity.ID,
		&roleEntity.Name,
		&roleEntity.CreatedAt,
		&roleEntity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, roledomain.ErrRoleNotFound
		}
		return nil, err
	}

	return &roleEntity, nil
}

// Update modifies an existing role in the database.
// AI-hint: Role update with optimistic locking and constraint validation.
func (r *RoleRepository) Update(ctx interface{}, roleEntity *roledomain.Role) error {
	context := ctx.(context.Context)

	query := `
		UPDATE roles
		SET name = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(context, query, roleEntity.ID, roleEntity.Name, roleEntity.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return roledomain.ErrRoleNameAlreadyExists
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return roledomain.ErrRoleNotFound
	}

	return nil
}

// Delete removes a role from the database.
// AI-hint: Role deletion with business rule validation and foreign key handling.
func (r *RoleRepository) Delete(ctx interface{}, id string) error {
	context := ctx.(context.Context)

	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.pool.Exec(context, query, id)
	if err != nil {
		// Check for foreign key constraint violation (users still assigned to this role)
		if isForeignKeyViolation(err) {
			return errors.New("cannot delete role with assigned users")
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return roledomain.ErrRoleNotFound
	}

	return nil
}

// List retrieves all roles from the database.
// AI-hint: Complete role listing for administrative operations and role selection UIs.
func (r *RoleRepository) List(ctx interface{}) ([]*roledomain.Role, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, name, created_at, updated_at
		FROM roles
		ORDER BY name
	`

	rows, err := r.pool.Query(context, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*roledomain.Role
	for rows.Next() {
		var roleEntity roledomain.Role
		err := rows.Scan(
			&roleEntity.ID,
			&roleEntity.Name,
			&roleEntity.CreatedAt,
			&roleEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &roleEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// Exists checks if a role with the given name already exists.
// AI-hint: Existence check for role name uniqueness validation.
func (r *RoleRepository) Exists(ctx interface{}, name string) (bool, error) {
	context := ctx.(context.Context)

	query := `SELECT 1 FROM roles WHERE name = $1`

	var exists int
	err := r.pool.QueryRow(context, query, name).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// EnsurePredefinedRoles creates the predefined roles if they don't exist.
// AI-hint: System initialization method to ensure required roles exist on startup.
func (r *RoleRepository) EnsurePredefinedRoles(ctx interface{}) error {
	context := ctx.(context.Context)

	for _, roleName := range roledomain.PredefinedRoles {
		exists, err := r.Exists(context, roleName)
		if err != nil {
			return err
		}

		if !exists {
			roleID := uuid.New().String()
			newRole, err := roledomain.NewRole(roleID, roleName)
			if err != nil {
				return err
			}

			if err := r.Create(context, newRole); err != nil {
				return err
			}
		}
	}

	return nil
}
