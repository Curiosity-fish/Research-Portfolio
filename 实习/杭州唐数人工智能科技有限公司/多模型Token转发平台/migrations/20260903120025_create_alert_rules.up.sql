CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name VARCHAR(200) NOT NULL,
    metric VARCHAR(20) NOT NULL CHECK (metric IN ('balance_low', 'quota_low', 'error_rate', 'cost_spike')),
    threshold BIGINT NOT NULL CHECK (threshold > 0),
    enabled BOOLEAN NOT NULL DEFAULT true,
    description VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS alertrule_metric ON alert_rules(metric);
CREATE INDEX IF NOT EXISTS alertrule_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS alertrule_created_at ON alert_rules(created_at);
