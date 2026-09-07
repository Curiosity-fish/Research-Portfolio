-- Create the user_tokens table for external API keys.
CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    token_last4 VARCHAR(10) NOT NULL,
    quota_limit BIGINT,
    quota_used BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT user_tokens_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS usertoken_token_hash ON user_tokens(token_hash);
CREATE INDEX IF NOT EXISTS usertoken_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS usertoken_expires_at ON user_tokens(expires_at);
CREATE INDEX IF NOT EXISTS usertoken_is_enabled ON user_tokens(is_enabled);
