CREATE TABLE IF NOT EXISTS quota_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    token_id UUID NOT NULL,
    requested_amount BIGINT NOT NULL CHECK (requested_amount > 0),
    reason VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    CONSTRAINT quotarequest_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT quotarequest_token_fk FOREIGN KEY (token_id) REFERENCES user_tokens(id) ON DELETE CASCADE,
    CONSTRAINT quotarequest_reviewer_fk FOREIGN KEY (reviewed_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS quotarequest_user_id ON quota_requests(user_id);
CREATE INDEX IF NOT EXISTS quotarequest_token_id ON quota_requests(token_id);
CREATE INDEX IF NOT EXISTS quotarequest_status ON quota_requests(status);
CREATE INDEX IF NOT EXISTS quotarequest_created_at ON quota_requests(created_at);
