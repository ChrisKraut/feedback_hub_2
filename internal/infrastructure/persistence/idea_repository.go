package persistence

import (
	"context"
	"feedback_hub_2/internal/domain/idea"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdeaRepository implements the idea.Repository interface using PostgreSQL.
// AI-hint: PostgreSQL implementation of idea persistence with proper error handling
// and connection management using pgxpool for connection pooling.
type IdeaRepository struct {
	db *pgxpool.Pool
}

// NewIdeaRepository creates a new IdeaRepository instance.
// AI-hint: Factory method for idea repository with dependency injection of database pool.
func NewIdeaRepository(db *pgxpool.Pool) *IdeaRepository {
	return &IdeaRepository{
		db: db,
	}
}

// Create saves a new idea to the database.
// AI-hint: Database insert operation with proper error handling and validation.
func (r *IdeaRepository) Create(ctx interface{}, idea *idea.Idea) error {
	context := ctx.(context.Context)

	query := `
		INSERT INTO ideas (id, idea_text, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(context, query, idea.ID, idea.IdeaText, idea.UserID, idea.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create idea: %w", err)
	}

	return nil
}

// GetByID retrieves an idea by its ID.
// AI-hint: Database select operation with proper error handling for not found cases.
func (r *IdeaRepository) GetByID(ctx interface{}, id string) (*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, idea_text, user_id, created_at
		FROM ideas
		WHERE id = $1
	`

	var ideaItem idea.Idea
	err := r.db.QueryRow(context, query, id).Scan(
		&ideaItem.ID,
		&ideaItem.IdeaText,
		&ideaItem.UserID,
		&ideaItem.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, idea.ErrIdeaNotFound
		}
		return nil, fmt.Errorf("failed to get idea by ID: %w", err)
	}

	return &ideaItem, nil
}

// GetByUserID retrieves all ideas submitted by a specific user.
// AI-hint: Database select operation with ordering by creation date for user-specific idea listing.
func (r *IdeaRepository) GetByUserID(ctx interface{}, userID string) ([]*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, idea_text, user_id, created_at
		FROM ideas
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ideas by user ID: %w", err)
	}
	defer rows.Close()

	var ideas []*idea.Idea
	for rows.Next() {
		var ideaItem idea.Idea
		err := rows.Scan(
			&ideaItem.ID,
			&ideaItem.IdeaText,
			&ideaItem.UserID,
			&ideaItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan idea row: %w", err)
		}
		ideas = append(ideas, &ideaItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over idea rows: %w", err)
	}

	return ideas, nil
}

// Update modifies an existing idea in the database.
// AI-hint: Database update operation with proper error handling and validation.
func (r *IdeaRepository) Update(ctx interface{}, ideaItem *idea.Idea) error {
	context := ctx.(context.Context)

	query := `
		UPDATE ideas
		SET idea_text = $2, user_id = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(context, query, ideaItem.ID, ideaItem.IdeaText, ideaItem.UserID)
	if err != nil {
		return fmt.Errorf("failed to update idea: %w", err)
	}

	if result.RowsAffected() == 0 {
		return idea.ErrIdeaNotFound
	}

	return nil
}

// Delete removes an idea from the database.
// AI-hint: Database delete operation with proper error handling and validation.
func (r *IdeaRepository) Delete(ctx interface{}, id string) error {
	context := ctx.(context.Context)

	query := `DELETE FROM ideas WHERE id = $1`

	result, err := r.db.Exec(context, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete idea: %w", err)
	}

	if result.RowsAffected() == 0 {
		return idea.ErrIdeaNotFound
	}

	return nil
}

// List retrieves all ideas ordered by creation date (most recent first).
// AI-hint: Database select operation with proper ordering for idea listing.
func (r *IdeaRepository) List(ctx interface{}) ([]*idea.Idea, error) {
	context := ctx.(context.Context)

	query := `
		SELECT id, idea_text, user_id, created_at
		FROM ideas
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list ideas: %w", err)
	}
	defer rows.Close()

	var ideas []*idea.Idea
	for rows.Next() {
		var ideaItem idea.Idea
		err := rows.Scan(
			&ideaItem.ID,
			&ideaItem.IdeaText,
			&ideaItem.UserID,
			&ideaItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan idea row: %w", err)
		}
		ideas = append(ideas, &ideaItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over idea rows: %w", err)
	}

	return ideas, nil
}

// EnsureIdeasTable creates the ideas table if it doesn't exist.
// AI-hint: Database schema initialization function for the ideas domain.
func EnsureIdeasTable(ctx context.Context, db *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS ideas (
			id UUID PRIMARY KEY,
			idea_text TEXT NOT NULL,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create ideas table: %w", err)
	}

	// Create indexes for performance
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_ideas_user_id ON ideas(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at DESC)`,
	}

	for _, indexQuery := range indexQueries {
		_, err := db.Exec(ctx, indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
