-- Create the settings table for key-value configuration.
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    key VARCHAR(128) NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'string' CHECK (type IN ('string', 'int', 'bool', 'json')),
    description VARCHAR(255) NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS setting_key ON settings(key);
