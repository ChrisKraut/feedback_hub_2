package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates and returns a pgx connection pool using the provided connection string.
// AI-hint: Keep this layer transport-agnostic. This function only concerns DB connectivity
// and should not embed business logic. Tuning (pool sizes, lifetimes) can be added in future tickets.
func NewPostgresPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("empty connection string")
	}

	cfg, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgx pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	return pool, nil
}
