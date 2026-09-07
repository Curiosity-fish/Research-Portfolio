CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_type VARCHAR(20) NOT NULL CHECK (actor_type IN ('admin', 'user', 'system')),
    actor_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id UUID,
    details JSONB NOT NULL DEFAULT '{}',
    ip VARCHAR(50),
    user_agent VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS auditlog_actor_type ON audit_logs(actor_type);
CREATE INDEX IF NOT EXISTS auditlog_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS auditlog_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS auditlog_created_at ON audit_logs(created_at);
