-- Roll back the user_tokens table and its indexes.
DROP INDEX IF EXISTS usertoken_is_enabled;
DROP INDEX IF EXISTS usertoken_expires_at;
DROP INDEX IF EXISTS usertoken_user_id;
DROP INDEX IF EXISTS usertoken_token_hash;
DROP TABLE IF EXISTS user_tokens;
