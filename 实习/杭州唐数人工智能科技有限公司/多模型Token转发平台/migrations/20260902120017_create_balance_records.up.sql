CREATE TABLE IF NOT EXISTS balance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    call_log_id UUID,
    related_order_id UUID,
    type VARCHAR(20) NOT NULL CHECK (type IN ('recharge', 'consume', 'refund', 'admin_adjust')),
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL DEFAULT 0,
    remark VARCHAR(500),
    CONSTRAINT balancerecord_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT balancerecord_calllog_fk FOREIGN KEY (call_log_id) REFERENCES call_logs(id) ON DELETE SET NULL,
    CONSTRAINT balancerecord_order_fk FOREIGN KEY (related_order_id) REFERENCES recharge_orders(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS balancerecord_user_id ON balance_records(user_id);
CREATE INDEX IF NOT EXISTS balancerecord_call_log_id ON balance_records(call_log_id);
CREATE INDEX IF NOT EXISTS balancerecord_related_order_id ON balance_records(related_order_id);
CREATE INDEX IF NOT EXISTS balancerecord_created_at ON balance_records(created_at);
