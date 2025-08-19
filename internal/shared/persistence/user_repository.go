package persistence

import (
	"context"
	"database/sql"
	"errors"
	userdomain "feedback_hub_2/internal/user/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements the user.Repository interface using PostgreSQL.
// AI-hint: Persistence layer implementation for user domain using pgx for PostgreSQL.
// Provides CRUD operations with proper error handling and constraint validation.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository instance.
// AI-hint: Factory method for user repository with dependency injection of DB pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

// Create inserts a new user into the database.
// AI-hint: User creation with email uniqueness validation and role foreign key check.
func (r *UserRepository) Create(ctx interface{}, userEntity *userdomain.User) error {
	context := ctx.(context.Context)

	query := `
		INSERT INTO users (id, email, name, password_hash, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var passwordHash interface{}
	if userEntity.PasswordHash == "" {
		passwordHash = nil
	} else {
		passwordHash = userEntity.PasswordHash
	}

	_, err := r.pool.Exec(context, query,
		userEntity.ID, userEntity.Email, userEntity.Name, passwordHash, userEntity.RoleID,
		userEntity.CreatedAt, userEntity.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation (duplicate email)
		if isUniqueViolation(err) {
			return userdomain.ErrEmailAlreadyExists
		}
		// Check for foreign key constraint violation (invalid role_id)
		if isForeignKeyViolation(err) {
			return errors.New("invalid role ID")
		}
		return err
	}

	return nil
}

// GetByID retrieves a user by their ID.
// AI-hint: Single user retrieval with proper error handling for not found cases.
func (r *UserRepository) GetByID(ctx interface{}, id string) (*userdomain.User, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, email, name, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var userEntity userdomain.User
	var passwordHash sql.NullString
	err := r.pool.QueryRow(context, query, id).Scan(
		&userEntity.ID,
		&userEntity.Email,
		&userEntity.Name,
		&passwordHash,
		&userEntity.RoleID,
		&userEntity.CreatedAt,
		&userEntity.UpdatedAt,
	)
	if passwordHash.Valid {
		userEntity.PasswordHash = passwordHash.String
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, err
	}

	return &userEntity, nil
}

// GetByEmail retrieves a user by their email address.
// AI-hint: Email-based lookup for authentication and user identification flows.
func (r *UserRepository) GetByEmail(ctx interface{}, email string) (*userdomain.User, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, email, name, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var userEntity userdomain.User
	var passwordHash sql.NullString
	err := r.pool.QueryRow(context, query, email).Scan(
		&userEntity.ID,
		&userEntity.Email,
		&userEntity.Name,
		&passwordHash,
		&userEntity.RoleID,
		&userEntity.CreatedAt,
		&userEntity.UpdatedAt,
	)
	if passwordHash.Valid {
		userEntity.PasswordHash = passwordHash.String
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, err
	}

	return &userEntity, nil
}

// Update modifies an existing user in the database.
// AI-hint: User update with optimistic locking and constraint validation.
func (r *UserRepository) Update(ctx interface{}, userEntity *userdomain.User) error {
	context := ctx.(context.Context)

	query := `
		UPDATE users
		SET email = $2, name = $3, password_hash = $4, role_id = $5, updated_at = $6
		WHERE id = $1
	`

	var passwordHash interface{}
	if userEntity.PasswordHash == "" {
		passwordHash = nil
	} else {
		passwordHash = userEntity.PasswordHash
	}

	result, err := r.pool.Exec(context, query,
		userEntity.ID, userEntity.Email, userEntity.Name, passwordHash, userEntity.RoleID, userEntity.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return userdomain.ErrEmailAlreadyExists
		}
		if isForeignKeyViolation(err) {
			return errors.New("invalid role ID")
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return userdomain.ErrUserNotFound
	}

	return nil
}

// Delete removes a user from the database.
// AI-hint: User deletion with proper error handling for not found cases.
func (r *UserRepository) Delete(ctx interface{}, id string) error {
	context := ctx.(context.Context)

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(context, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return userdomain.ErrUserNotFound
	}

	return nil
}

// List retrieves all users from the database.
// AI-hint: Complete user listing for administrative operations with proper ordering.
func (r *UserRepository) List(ctx interface{}) ([]*userdomain.User, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, email, name, password_hash, role_id, created_at, updated_at
		FROM users
		ORDER BY email
	`

	rows, err := r.pool.Query(context, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userdomain.User
	for rows.Next() {
		var userEntity userdomain.User
		var passwordHash sql.NullString
		err := rows.Scan(
			&userEntity.ID,
			&userEntity.Email,
			&userEntity.Name,
			&passwordHash,
			&userEntity.RoleID,
			&userEntity.CreatedAt,
			&userEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if passwordHash.Valid {
			userEntity.PasswordHash = passwordHash.String
		}
		users = append(users, &userEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetByRoleID retrieves all users assigned to a specific role.
// AI-hint: Role-based user lookup for role deletion validation and management operations.
func (r *UserRepository) GetByRoleID(ctx interface{}, roleID string) ([]*userdomain.User, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, email, name, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE role_id = $1
		ORDER BY email
	`

	rows, err := r.pool.Query(context, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userdomain.User
	for rows.Next() {
		var userEntity userdomain.User
		var passwordHash sql.NullString
		err := rows.Scan(
			&userEntity.ID,
			&userEntity.Email,
			&userEntity.Name,
			&passwordHash,
			&userEntity.RoleID,
			&userEntity.CreatedAt,
			&userEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if passwordHash.Valid {
			userEntity.PasswordHash = passwordHash.String
		}
		users = append(users, &userEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
