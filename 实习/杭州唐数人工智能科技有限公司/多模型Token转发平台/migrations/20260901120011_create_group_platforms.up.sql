CREATE TABLE IF NOT EXISTS group_platforms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    group_id UUID NOT NULL,
    platform_id UUID NOT NULL,
    CONSTRAINT gp_group_fk FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT gp_platform_fk FOREIGN KEY (platform_id) REFERENCES platforms(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS gp_group_platform ON group_platforms(group_id, platform_id);
CREATE INDEX IF NOT EXISTS gp_group_id ON group_platforms(group_id);
CREATE INDEX IF NOT EXISTS gp_platform_id ON group_platforms(platform_id);
