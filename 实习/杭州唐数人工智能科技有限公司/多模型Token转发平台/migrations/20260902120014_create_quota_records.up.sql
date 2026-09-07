CREATE TABLE IF NOT EXISTS quota_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    token_id UUID NOT NULL,
    call_log_id UUID,
    type VARCHAR(30) NOT NULL CHECK (type IN ('pre_deduct', 'consume', 'refund', 'admin_adjust', 'request_approved')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    quota_after BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT quotarecord_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT quotarecord_token_fk FOREIGN KEY (token_id) REFERENCES user_tokens(id) ON DELETE CASCADE,
    CONSTRAINT quotarecord_calllog_fk FOREIGN KEY (call_log_id) REFERENCES call_logs(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS quotarecord_user_id ON quota_records(user_id);
CREATE INDEX IF NOT EXISTS quotarecord_token_id ON quota_records(token_id);
CREATE INDEX IF NOT EXISTS quotarecord_call_log_id ON quota_records(call_log_id);
CREATE INDEX IF NOT EXISTS quotarecord_created_at ON quota_records(created_at);
