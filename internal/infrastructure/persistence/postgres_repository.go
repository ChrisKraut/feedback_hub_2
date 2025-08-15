package persistence

import (
	"context"
	"fmt"
	"log"

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

// EnsureSchema ensures the database schema exists by running migrations.
// AI-hint: Auto-migration function that creates required tables if they don't exist.
// This prevents the "relation does not exist" errors on fresh databases.
func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Ensuring database schema exists...")

	// Check if users table exists with password_hash column (indicates new schema)
	var columnCount int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'users' 
		AND column_name = 'password_hash'
	`).Scan(&columnCount)
	if err != nil {
		return fmt.Errorf("failed to check schema version: %w", err)
	}

	if columnCount > 0 {
		log.Println("Database schema is up to date, skipping migration")
		return nil
	}

	// Check if we need to migrate existing schema or create from scratch
	var userTableExists int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&userTableExists)
	if err != nil {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}

	if userTableExists > 0 {
		log.Println("Migrating existing schema to add password support...")
		// Add password_hash column to existing users table
		_, err = pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)`)
		if err != nil {
			return fmt.Errorf("failed to add password_hash column: %w", err)
		}
		log.Println("Schema migration completed successfully")
		return nil
	}

	log.Println("Creating database schema from scratch...")

	// Enable UUID extension
	_, err = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Create roles table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create roles table: %w", err)
	}

	// Create users table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255), -- Optional for OAuth users
			role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id)`,
		`CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name)`,
	}

	for _, indexSQL := range indexes {
		_, err = pool.Exec(ctx, indexSQL)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database schema created successfully")
	return nil
}
