-- Roll back the group_id addition.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_group_fk;

DROP INDEX IF EXISTS user_group_id;

ALTER TABLE users DROP COLUMN IF EXISTS group_id;
