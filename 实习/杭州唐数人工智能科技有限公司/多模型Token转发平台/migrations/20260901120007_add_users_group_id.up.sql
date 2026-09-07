-- Add group_id to users to align the migration baseline with the Ent schema.
ALTER TABLE users ADD COLUMN IF NOT EXISTS group_id UUID;

CREATE INDEX IF NOT EXISTS user_group_id ON users(group_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_group_fk'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_group_fk
            FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL;
    END IF;
END $$;
