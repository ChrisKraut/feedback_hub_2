package persistence

import (
	"context"
	"errors"
	"feedback_hub_2/internal/domain/idea"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdeaRepository implements the idea.Repository interface using PostgreSQL.
// AI-hint: Persistence layer implementation for idea domain using pgx for PostgreSQL.
// Provides CRUD operations with proper error handling and constraint validation.
type IdeaRepository struct {
	pool *pgxpool.Pool
}

// NewIdeaRepository creates a new IdeaRepository instance.
// AI-hint: Factory method for idea repository with dependency injection of DB pool.
func NewIdeaRepository(pool *pgxpool.Pool) *IdeaRepository {
	return &IdeaRepository{
		pool: pool,
	}
}

// Save inserts a new idea or updates an existing one in the database.
// AI-hint: Upsert operation that handles both creation and updates.
// Uses ON CONFLICT to handle duplicate ID scenarios gracefully.
func (r *IdeaRepository) Save(ctx interface{}, ideaEntity *idea.Idea) error {
	context := ctx.(context.Context)

	query := `
		INSERT INTO ideas (id, title, content, creator_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(context, query,
		ideaEntity.ID, ideaEntity.Title, ideaEntity.Content, ideaEntity.CreatorUserID,
		ideaEntity.CreatedAt, ideaEntity.UpdatedAt,
	)
	if err != nil {
		// Check for foreign key constraint violation (invalid creator_user_id)
		if isForeignKeyViolation(err) {
			return idea.ErrCreatorNotFound
		}
		return err
	}

	return nil
}

// FindByID retrieves an idea by its ID.
// AI-hint: Single idea retrieval with proper error handling for not found cases.
func (r *IdeaRepository) FindByID(ctx interface{}, id uuid.UUID) (*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, title, content, creator_user_id, created_at, updated_at
		FROM ideas
		WHERE id = $1
	`

	var ideaEntity idea.Idea
	err := r.pool.QueryRow(context, query, id).Scan(
		&ideaEntity.ID,
		&ideaEntity.Title,
		&ideaEntity.Content,
		&ideaEntity.CreatorUserID,
		&ideaEntity.CreatedAt,
		&ideaEntity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, idea.ErrIdeaNotFound
		}
		return nil, err
	}

	return &ideaEntity, nil
}

// FindByCreatorUserID retrieves all ideas created by a specific user.
// AI-hint: Collection retrieval filtered by creator with proper error handling.
func (r *IdeaRepository) FindByCreatorUserID(ctx interface{}, creatorUserID uuid.UUID) ([]*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, title, content, creator_user_id, created_at, updated_at
		FROM ideas
		WHERE creator_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(context, query, creatorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ideas []*idea.Idea
	for rows.Next() {
		var ideaEntity idea.Idea
		err := rows.Scan(
			&ideaEntity.ID,
			&ideaEntity.Title,
			&ideaEntity.Content,
			&ideaEntity.CreatorUserID,
			&ideaEntity.CreatedAt,
			&ideaEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ideas = append(ideas, &ideaEntity)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ideas, nil
}

// FindAll retrieves all ideas from the database.
// AI-hint: Collection retrieval with proper error handling and ordering.
func (r *IdeaRepository) FindAll(ctx interface{}) ([]*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, title, content, creator_user_id, created_at, updated_at
		FROM ideas
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(context, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ideas []*idea.Idea
	for rows.Next() {
		var ideaEntity idea.Idea
		err := rows.Scan(
			&ideaEntity.ID,
			&ideaEntity.Title,
			&ideaEntity.Content,
			&ideaEntity.CreatorUserID,
			&ideaEntity.CreatedAt,
			&ideaEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ideas = append(ideas, &ideaEntity)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ideas, nil
}

// Delete removes an idea from the database by its ID.
// AI-hint: Soft or hard delete operation with proper error handling.
func (r *IdeaRepository) Delete(ctx interface{}, id uuid.UUID) error {
	context := ctx.(context.Context)

	query := `DELETE FROM ideas WHERE id = $1`

	result, err := r.pool.Exec(context, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return idea.ErrIdeaNotFound
	}

	return nil
}
