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

	// Acquire a connection from the pool for schema operations
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire database connection for schema check: %w", err)
	}
	defer conn.Release()

	// Check if organizations table exists (indicates new organization schema)
	var orgTableExists int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organizations'").Scan(&orgTableExists)
	if err != nil {
		return fmt.Errorf("failed to check if organizations table exists: %w", err)
	}

	if orgTableExists > 0 {
		log.Println("Organization schema exists, checking for user_organizations table...")

		// Check if user_organizations table exists
		var userOrgTableExists int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'user_organizations'").Scan(&userOrgTableExists)
		if err != nil {
			return fmt.Errorf("failed to check if user_organizations table exists: %w", err)
		}

		if userOrgTableExists > 0 {
			log.Println("Organization schema is up to date, skipping migration")
			return nil
		}

		log.Println("Adding user_organizations table to existing schema...")

		// Create user_organizations table
		_, err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS user_organizations (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
				is_active BOOLEAN NOT NULL DEFAULT true,
				joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				UNIQUE(user_id, organization_id)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create user_organizations table: %w", err)
		}

		// Create user_organizations indexes
		userOrgIndexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_user_id ON user_organizations(user_id)`,
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_organization_id ON user_organizations(organization_id)`,
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_active ON user_organizations(is_active)`,
		}

		for _, indexSQL := range userOrgIndexes {
			_, err = conn.Exec(ctx, indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create user_organizations index: %w", err)
			}
		}

		log.Println("User_organizations table added to existing schema successfully")
		return nil
	}

	// Check if users table exists with password_hash column (indicates old schema)
	var columnCount int
	err = conn.QueryRow(ctx, `
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
		log.Println("Old database schema exists, migrating to organization schema...")

		// Create organizations table
		_, err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS organizations (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				name VARCHAR(255) NOT NULL,
				slug VARCHAR(100) NOT NULL UNIQUE,
				description TEXT,
				settings JSONB DEFAULT '{}',
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create organizations table: %w", err)
		}

		// Create default organization for existing data
		_, err = conn.Exec(ctx, `
			INSERT INTO organizations (id, name, slug, description, settings, created_at, updated_at)
			VALUES (uuid_generate_v4(), 'Default Organization', 'default', 'Default organization for existing data', '{}', NOW(), NOW())
		`)
		if err != nil {
			return fmt.Errorf("failed to create default organization: %w", err)
		}

		// Get the default organization ID
		var defaultOrgID string
		err = conn.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = 'default'").Scan(&defaultOrgID)
		if err != nil {
			return fmt.Errorf("failed to get default organization ID: %w", err)
		}

		// Add organization_id column to roles table
		_, err = conn.Exec(ctx, `ALTER TABLE roles ADD COLUMN IF NOT EXISTS organization_id UUID`)
		if err != nil {
			return fmt.Errorf("failed to add organization_id column to roles: %w", err)
		}

		// Update existing roles to use default organization
		_, err = conn.Exec(ctx, `UPDATE roles SET organization_id = $1 WHERE organization_id IS NULL`, defaultOrgID)
		if err != nil {
			return fmt.Errorf("failed to update existing roles: %w", err)
		}

		// Make organization_id NOT NULL
		_, err = conn.Exec(ctx, `ALTER TABLE roles ALTER COLUMN organization_id SET NOT NULL`)
		if err != nil {
			return fmt.Errorf("failed to make organization_id NOT NULL: %w", err)
		}

		// Add foreign key constraint
		_, err = conn.Exec(ctx, `ALTER TABLE roles ADD CONSTRAINT fk_roles_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE`)
		if err != nil {
			return fmt.Errorf("failed to add foreign key constraint to roles: %w", err)
		}

		// Add organization_id column to ideas table
		_, err = conn.Exec(ctx, `ALTER TABLE ideas ADD COLUMN IF NOT EXISTS organization_id UUID`)
		if err != nil {
			return fmt.Errorf("failed to add organization_id column to ideas: %w", err)
		}

		// Update existing ideas to use default organization
		_, err = conn.Exec(ctx, `UPDATE ideas SET organization_id = $1 WHERE organization_id IS NULL`, defaultOrgID)
		if err != nil {
			return fmt.Errorf("failed to update existing ideas: %w", err)
		}

		// Make organization_id NOT NULL
		_, err = conn.Exec(ctx, `ALTER TABLE ideas ALTER COLUMN organization_id SET NOT NULL`)
		if err != nil {
			return fmt.Errorf("failed to make organization_id NOT NULL: %w", err)
		}

		// Add foreign key constraint
		_, err = conn.Exec(ctx, `ALTER TABLE ideas ADD CONSTRAINT fk_ideas_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE`)
		if err != nil {
			return fmt.Errorf("failed to add foreign key constraint to ideas: %w", err)
		}

		// Create user_organizations table
		_, err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS user_organizations (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
				is_active BOOLEAN NOT NULL DEFAULT true,
				joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				UNIQUE(user_id, organization_id)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create user_organizations table: %w", err)
		}

		// Create user_organizations indexes
		userOrgIndexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_user_id ON user_organizations(user_id)`,
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_organization_id ON user_organizations(organization_id)`,
			`CREATE INDEX IF NOT EXISTS idx_user_organizations_active ON user_organizations(is_active)`,
		}

		for _, indexSQL := range userOrgIndexes {
			_, err = conn.Exec(ctx, indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create user_organizations index: %w", err)
			}
		}

		// Migrate existing user-role relationships to user_organizations
		_, err = conn.Exec(ctx, `
			INSERT INTO user_organizations (id, user_id, organization_id, role_id, is_active, joined_at, created_at, updated_at)
			SELECT 
				uuid_generate_v4(),
				u.id,
				$1,
				u.role_id,
				true,
				u.created_at,
				u.created_at,
				u.created_at
			FROM users u
			WHERE u.role_id IS NOT NULL
		`, defaultOrgID)
		if err != nil {
			return fmt.Errorf("failed to migrate existing user-role relationships: %w", err)
		}

		// Remove role_id column from users table (no longer needed)
		_, err = conn.Exec(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS role_id`)
		if err != nil {
			return fmt.Errorf("failed to remove role_id column from users: %w", err)
		}

		// Create organization indexes
		orgIndexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug)`,
			`CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at)`,
		}

		for _, indexSQL := range orgIndexes {
			_, err = conn.Exec(ctx, indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create organization index: %w", err)
			}
		}

		// Create role indexes
		roleIndexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_roles_name_organization ON roles(name, organization_id)`,
			`CREATE INDEX IF NOT EXISTS idx_roles_organization_id ON roles(organization_id)`,
		}

		for _, indexSQL := range roleIndexes {
			_, err = conn.Exec(ctx, indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create role index: %w", err)
			}
		}

		// Create idea indexes
		ideaIndexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_ideas_creator_user_id ON ideas(creator_user_id)`,
			`CREATE INDEX IF NOT EXISTS idx_ideas_organization_id ON ideas(organization_id)`,
			`CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at)`,
			`CREATE INDEX IF NOT EXISTS idx_ideas_updated_at ON ideas(updated_at)`,
		}

		for _, indexSQL := range ideaIndexes {
			_, err = conn.Exec(ctx, indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create idea index: %w", err)
			}
		}

		// Add constraints
		constraints := []string{
			`ALTER TABLE organizations ADD CONSTRAINT chk_organizations_name_not_empty CHECK (trim(name) != '')`,
			`ALTER TABLE organizations ADD CONSTRAINT chk_organizations_slug_not_empty CHECK (trim(slug) != '')`,
			`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_not_empty CHECK (trim(title) != '')`,
			`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_content_not_empty CHECK (trim(content) != '')`,
			`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_length CHECK (length(title) <= 255)`,
		}

		for _, constraintSQL := range constraints {
			_, err = conn.Exec(ctx, constraintSQL)
			if err != nil {
				return fmt.Errorf("failed to add constraint: %w", err)
			}
		}

		log.Println("Schema migration to organization schema completed successfully")
		return nil
	}

	// Check if we need to migrate existing schema or create from scratch
	var userTableExists int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Scan(&userTableExists)
	if err != nil {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}

	if userTableExists > 0 {
		log.Println("Migrating existing schema to add password support...")
		// Add password_hash column to existing users table
		_, err = conn.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)`)
		if err != nil {
			return fmt.Errorf("failed to add password_hash column: %w", err)
		}
		log.Println("Schema migration completed successfully")
		return nil
	}

	log.Println("Creating database schema from scratch...")

	// Enable UUID extension
	_, err = conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Create organizations table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS organizations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			settings JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create organizations table: %w", err)
	}

	// Create default organization
	_, err = conn.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, description, settings, created_at, updated_at)
		VALUES (uuid_generate_v4(), 'Default Organization', 'default', 'Default organization for new installations', '{}', NOW(), NOW())
	`)
	if err != nil {
		return fmt.Errorf("failed to create default organization: %w", err)
	}

	// Get the default organization ID
	var defaultOrgID string
	err = conn.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = 'default'").Scan(&defaultOrgID)
	if err != nil {
		return fmt.Errorf("failed to get default organization ID: %w", err)
	}

	// Create roles table (organization-scoped)
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL,
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(name, organization_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create roles table: %w", err)
	}

	// Create default roles for the default organization
	defaultRoles := []string{"Super User", "Product Owner", "Contributor"}
	for _, roleName := range defaultRoles {
		_, err = conn.Exec(ctx, `
			INSERT INTO roles (id, name, organization_id, created_at, updated_at)
			VALUES (uuid_generate_v4(), $1, $2, NOW(), NOW())
		`, roleName, defaultOrgID)
		if err != nil {
			return fmt.Errorf("failed to create default role %s: %w", roleName, err)
		}
	}

	// Create users table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255), -- Optional for OAuth users
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create user_organizations table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_organizations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
			is_active BOOLEAN NOT NULL DEFAULT true,
			joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, organization_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_organizations table: %w", err)
	}

	// Create ideas table (organization-scoped)
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ideas (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ideas table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_roles_name_organization ON roles(name, organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_roles_organization_id ON roles(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_organizations_user_id ON user_organizations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_organizations_organization_id ON user_organizations(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_organizations_active ON user_organizations(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_ideas_creator_user_id ON ideas(creator_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ideas_organization_id ON ideas(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_ideas_updated_at ON ideas(updated_at)`,
	}

	for _, indexSQL := range indexes {
		_, err = conn.Exec(ctx, indexSQL)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Add constraints
	constraints := []string{
		`ALTER TABLE organizations ADD CONSTRAINT chk_organizations_name_not_empty CHECK (trim(name) != '')`,
		`ALTER TABLE organizations ADD CONSTRAINT chk_organizations_slug_not_empty CHECK (trim(slug) != '')`,
		`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_not_empty CHECK (trim(title) != '')`,
		`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_content_not_empty CHECK (trim(content) != '')`,
		`ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_length CHECK (length(title) <= 255)`,
	}

	for _, constraintSQL := range constraints {
		_, err = conn.Exec(ctx, constraintSQL)
		if err != nil {
			return fmt.Errorf("failed to add constraint: %w", err)
		}
	}

	log.Println("Database schema created successfully")
	return nil
}
