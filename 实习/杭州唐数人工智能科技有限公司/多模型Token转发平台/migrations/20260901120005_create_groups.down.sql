-- Roll back the groups table and the users.group_id column.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_group_fk;
ALTER TABLE users DROP COLUMN IF EXISTS group_id;
DROP INDEX IF EXISTS user_group_id;
DROP INDEX IF EXISTS group_status;
DROP INDEX IF EXISTS group_code;
DROP TABLE IF EXISTS groups;
