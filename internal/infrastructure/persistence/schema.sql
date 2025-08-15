-- Schema for User, Role, and Idea Management System
-- AI-hint: Database schema following PostgreSQL best practices with proper constraints,
-- indexes, and foreign key relationships for the user roles and ideas domains.

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- Comments for documentation
COMMENT ON TABLE roles IS 'System roles for role-based access control';
COMMENT ON TABLE users IS 'Application users with assigned roles';
COMMENT ON TABLE ideas IS 'User-submitted ideas with markdown-formatted text';
COMMENT ON COLUMN roles.name IS 'Unique role name (e.g., Super User, Product Owner, Contributor)';
COMMENT ON COLUMN users.email IS 'Unique user email address for authentication';
COMMENT ON COLUMN users.role_id IS 'Foreign key reference to the user role';
COMMENT ON COLUMN ideas.idea_text IS 'Markdown-formatted idea text content';
COMMENT ON COLUMN ideas.user_id IS 'Foreign key reference to the user who submitted the idea';
COMMENT ON COLUMN ideas.created_at IS 'Timestamp when the idea was created';

-- Ideas table
CREATE TABLE IF NOT EXISTS ideas (
    id UUID PRIMARY KEY,
    idea_text TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ideas_user_id ON ideas(user_id);
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at DESC);

-- Constraints to ensure data integrity
ALTER TABLE roles ADD CONSTRAINT chk_roles_name_not_empty CHECK (trim(name) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_email_not_empty CHECK (trim(email) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_name_not_empty CHECK (trim(name) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_text_not_empty CHECK (trim(idea_text) != '');
