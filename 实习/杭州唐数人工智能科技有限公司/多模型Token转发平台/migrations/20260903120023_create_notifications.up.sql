CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID,
    type VARCHAR(20) NOT NULL DEFAULT 'announcement' CHECK (type IN ('announcement', 'system')),
    title VARCHAR(200) NOT NULL,
    content VARCHAR(5000) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT notification_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS notification_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS notification_type ON notifications(type);
CREATE INDEX IF NOT EXISTS notification_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS notification_created_at ON notifications(created_at);
