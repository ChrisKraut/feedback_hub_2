-- Migration script to add ideas table to the PostgreSQL schema
-- AI-hint: Database migration following PostgreSQL best practices with proper constraints,
-- indexes, and foreign key relationships for the ideas domain.

-- Ideas table
CREATE TABLE IF NOT EXISTS ideas (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ideas_creator_user_id ON ideas(creator_user_id);
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
CREATE INDEX IF NOT EXISTS idx_ideas_updated_at ON ideas(updated_at);

-- Comments for documentation
COMMENT ON TABLE ideas IS 'Feedback ideas submitted by users';
COMMENT ON COLUMN ideas.title IS 'Short descriptive title for the idea';
COMMENT ON COLUMN ideas.content IS 'Rich text content describing the feedback idea';
COMMENT ON COLUMN ideas.creator_user_id IS 'Foreign key reference to the user who created the idea';

-- Constraints to ensure data integrity
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_not_empty CHECK (trim(title) != '');
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_content_not_empty CHECK (trim(content) != '');
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_title_length CHECK (length(title) <= 255);

-- Trigger to automatically update updated_at timestamp
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
