-- Roll back the admin_users table and its indexes.
DROP INDEX IF EXISTS adminuser_status;
DROP INDEX IF EXISTS adminuser_username;
DROP TABLE IF EXISTS admin_users;
