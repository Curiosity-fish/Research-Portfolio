-- Roll back the users table and its indexes.
DROP INDEX IF EXISTS user_status;
DROP INDEX IF EXISTS user_department_id;
DROP INDEX IF EXISTS user_phone;
DROP INDEX IF EXISTS user_email;
DROP INDEX IF EXISTS user_username;
DROP TABLE IF EXISTS users;
