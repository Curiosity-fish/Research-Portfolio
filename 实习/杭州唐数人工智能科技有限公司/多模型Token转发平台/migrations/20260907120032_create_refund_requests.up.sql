CREATE TABLE IF NOT EXISTS refund_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    reason VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    CONSTRAINT refundrequest_order_fk FOREIGN KEY (order_id) REFERENCES recharge_orders(id) ON DELETE CASCADE,
    CONSTRAINT refundrequest_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT refundrequest_reviewer_fk FOREIGN KEY (reviewed_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS refundrequest_order_id ON refund_requests(order_id);
CREATE INDEX IF NOT EXISTS refundrequest_user_id ON refund_requests(user_id);
CREATE INDEX IF NOT EXISTS refundrequest_status ON refund_requests(status);
CREATE INDEX IF NOT EXISTS refundrequest_created_at ON refund_requests(created_at);
