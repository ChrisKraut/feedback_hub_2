-- Schema for User, Role, and Organization Management System
-- AI-hint: Database schema following PostgreSQL best practices with proper constraints,
-- indexes, and foreign key relationships for the multi-tenant organization system.

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Roles table (organization-scoped)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(name, organization_id)
);

-- Users table (no longer organization-scoped, can belong to multiple orgs)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User-Organizations junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS user_organizations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, organization_id)
);

-- Ideas table (organization-scoped)
CREATE TABLE IF NOT EXISTS ideas (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_organizations_user_id ON user_organizations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_organization_id ON user_organizations(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_active ON user_organizations(is_active);
CREATE INDEX IF NOT EXISTS idx_roles_name_organization ON roles(name, organization_id);
CREATE INDEX IF NOT EXISTS idx_roles_organization_id ON roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_ideas_creator_user_id ON ideas(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_ideas_organization_id ON ideas(organization_id);
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
CREATE INDEX IF NOT EXISTS idx_ideas_updated_at ON ideas(updated_at);

-- Comments for documentation
COMMENT ON TABLE organizations IS 'Business entities that contain users, roles, and ideas';
COMMENT ON TABLE roles IS 'System roles for role-based access control within organizations';
COMMENT ON TABLE users IS 'Application users who can belong to multiple organizations';
COMMENT ON TABLE ideas IS 'Feedback ideas submitted by users within organizations';

COMMENT ON COLUMN organizations.name IS 'Human-readable name of the organization';
COMMENT ON COLUMN organizations.slug IS 'URL-friendly unique identifier for the organization';
COMMENT ON COLUMN organizations.description IS 'Optional description of the organization';
COMMENT ON COLUMN organizations.settings IS 'JSON configuration settings for the organization';

COMMENT ON COLUMN roles.name IS 'Role name within the organization (e.g., Super User, Product Owner, Contributor)';
COMMENT ON COLUMN roles.organization_id IS 'Foreign key reference to the organization this role belongs to';
COMMENT ON COLUMN users.email IS 'User email address for authentication (globally unique)';
COMMENT ON TABLE user_organizations IS 'Junction table linking users to organizations with roles and status';
COMMENT ON COLUMN user_organizations.user_id IS 'Foreign key reference to the user';
COMMENT ON COLUMN user_organizations.organization_id IS 'Foreign key reference to the organization';
COMMENT ON COLUMN user_organizations.role_id IS 'Foreign key reference to the user role within the organization';
COMMENT ON COLUMN user_organizations.is_active IS 'Whether the user is active in this organization';
COMMENT ON COLUMN user_organizations.joined_at IS 'When the user joined this organization';
COMMENT ON COLUMN ideas.title IS 'Short descriptive title for the idea';
COMMENT ON COLUMN ideas.content IS 'Rich text content describing the feedback idea';
COMMENT ON COLUMN ideas.creator_user_id IS 'Foreign key reference to the user who created the idea';
COMMENT ON COLUMN ideas.organization_id IS 'Foreign key reference to the organization this idea belongs to';

-- Constraints to ensure data integrity
ALTER TABLE organizations ADD CONSTRAINT chk_organizations_name_not_empty CHECK (trim(name) != '');
ALTER TABLE organizations ADD CONSTRAINT chk_organizations_slug_not_empty CHECK (trim(slug) != '');
ALTER TABLE organizations ADD CONSTRAINT chk_organizations_name_length CHECK (length(name) <= 255);
ALTER TABLE organizations ADD CONSTRAINT chk_organizations_slug_length CHECK (length(slug) <= 100);
ALTER TABLE organizations ADD CONSTRAINT chk_organizations_slug_format CHECK (slug ~* '^[a-z0-9]+(?:-[a-z0-9]+)*$');

ALTER TABLE roles ADD CONSTRAINT chk_roles_name_not_empty CHECK (trim(name) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_email_not_empty CHECK (trim(email) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_name_not_empty CHECK (trim(name) != '');
ALTER TABLE users ADD CONSTRAINT chk_users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

ALTER TABLE user_organizations ADD CONSTRAINT chk_user_organizations_active_not_null CHECK (is_active IS NOT NULL);

ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_not_empty CHECK (trim(title) != '');
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_content_not_empty CHECK (trim(content) != '');
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_length CHECK (length(title) <= 255);

-- Trigger to automatically update organizations updated_at timestamp
CREATE OR REPLACE FUNCTION update_organizations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_organizations_updated_at();

-- Trigger to automatically update user_organizations updated_at timestamp
CREATE OR REPLACE FUNCTION update_user_organizations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_organizations_updated_at
    BEFORE UPDATE ON user_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_user_organizations_updated_at();

-- Trigger to automatically update ideas updated_at timestamp
CREATE OR REPLACE FUNCTION update_ideas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_ideas_updated_at
    BEFORE UPDATE ON ideas
    FOR EACH ROW
    EXECUTE FUNCTION update_ideas_updated_at();
