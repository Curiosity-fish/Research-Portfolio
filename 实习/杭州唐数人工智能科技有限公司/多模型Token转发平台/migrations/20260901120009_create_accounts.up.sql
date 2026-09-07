CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    platform_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    api_key_encrypted VARCHAR(1000) NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    max_rpm INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    error_count INT NOT NULL DEFAULT 0,
    CONSTRAINT account_platform_fk FOREIGN KEY (platform_id) REFERENCES platforms(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS account_platform_id ON accounts(platform_id);
CREATE INDEX IF NOT EXISTS account_status ON accounts(status);
