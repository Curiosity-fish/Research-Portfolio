CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name VARCHAR(100) NOT NULL,
    upstream_name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'chat',
    input_price BIGINT NOT NULL DEFAULT 0,
    output_price BIGINT NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX IF NOT EXISTS aimodel_name ON ai_models(name);
CREATE INDEX IF NOT EXISTS aimodel_is_enabled ON ai_models(is_enabled);
