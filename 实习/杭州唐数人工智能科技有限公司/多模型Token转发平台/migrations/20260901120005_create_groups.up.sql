-- Create the groups table for AI relay routing and add a group_id to users.
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX IF NOT EXISTS group_code ON groups(code);
CREATE INDEX IF NOT EXISTS group_status ON groups(status);

ALTER TABLE users ADD COLUMN IF NOT EXISTS group_id UUID;
ALTER TABLE users ADD CONSTRAINT users_group_fk
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS user_group_id ON users(group_id);
