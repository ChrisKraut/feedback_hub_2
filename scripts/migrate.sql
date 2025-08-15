-- Migration script for User and Role Management System
-- AI-hint: Database migration script that creates the schema for the user roles domain.
-- This script is idempotent and can be run multiple times safely.

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
COMMENT ON COLUMN roles.name IS 'Unique role name (e.g., Super User, Product Owner, Contributor)';
COMMENT ON COLUMN users.email IS 'Unique user email address for authentication';
COMMENT ON COLUMN users.role_id IS 'Foreign key reference to the user role';

-- Constraints to ensure data integrity
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_roles_name_not_empty') THEN
        ALTER TABLE roles ADD CONSTRAINT chk_roles_name_not_empty CHECK (trim(name) != '');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_email_not_empty') THEN
        ALTER TABLE users ADD CONSTRAINT chk_users_email_not_empty CHECK (trim(email) != '');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_name_not_empty') THEN
        ALTER TABLE users ADD CONSTRAINT chk_users_name_not_empty CHECK (trim(name) != '');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_email_format') THEN
        ALTER TABLE users ADD CONSTRAINT chk_users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
    END IF;
END $$;
