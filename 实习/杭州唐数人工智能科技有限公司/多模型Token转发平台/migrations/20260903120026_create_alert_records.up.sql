CREATE TABLE IF NOT EXISTS alert_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rule_id UUID,
    metric VARCHAR(20) NOT NULL,
    user_id UUID,
    token_id UUID,
    account_id UUID,
    triggered_value BIGINT NOT NULL DEFAULT 0,
    message VARCHAR(500) NOT NULL,
    is_resolved BOOLEAN NOT NULL DEFAULT false,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    CONSTRAINT alertrecord_rule_fk FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE SET NULL,
    CONSTRAINT alertrecord_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT alertrecord_token_fk FOREIGN KEY (token_id) REFERENCES user_tokens(id) ON DELETE CASCADE,
    CONSTRAINT alertrecord_account_fk FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CONSTRAINT alertrecord_resolver_fk FOREIGN KEY (resolved_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS alertrecord_rule_id ON alert_records(rule_id);
CREATE INDEX IF NOT EXISTS alertrecord_metric ON alert_records(metric);
CREATE INDEX IF NOT EXISTS alertrecord_user_id ON alert_records(user_id);
CREATE INDEX IF NOT EXISTS alertrecord_is_resolved ON alert_records(is_resolved);
CREATE INDEX IF NOT EXISTS alertrecord_created_at ON alert_records(created_at);
