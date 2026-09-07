CREATE TABLE IF NOT EXISTS call_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    token_id UUID NOT NULL,
    platform_id UUID NOT NULL,
    account_id UUID NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT '',
    prompt_tokens BIGINT NOT NULL DEFAULT 0,
    completion_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    status_code INT NOT NULL DEFAULT 0,
    error_msg VARCHAR(500),
    CONSTRAINT calllog_account_fk FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS calllog_user_id ON call_logs(user_id);
CREATE INDEX IF NOT EXISTS calllog_token_id ON call_logs(token_id);
CREATE INDEX IF NOT EXISTS calllog_account_id ON call_logs(account_id);
CREATE INDEX IF NOT EXISTS calllog_platform_id ON call_logs(platform_id);
CREATE INDEX IF NOT EXISTS calllog_created_at ON call_logs(created_at);
