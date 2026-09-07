CREATE TABLE IF NOT EXISTS recharge_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
    provider VARCHAR(20) NOT NULL DEFAULT 'mock' CHECK (provider IN ('mock')),
    provider_order_id VARCHAR(255),
    paid_at TIMESTAMPTZ,
    CONSTRAINT rechargeorder_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS rechargeorder_user_id ON recharge_orders(user_id);
CREATE INDEX IF NOT EXISTS rechargeorder_status ON recharge_orders(status);
CREATE UNIQUE INDEX IF NOT EXISTS rechargeorder_provider_order_id ON recharge_orders(provider_order_id);
CREATE INDEX IF NOT EXISTS rechargeorder_created_at ON recharge_orders(created_at);
