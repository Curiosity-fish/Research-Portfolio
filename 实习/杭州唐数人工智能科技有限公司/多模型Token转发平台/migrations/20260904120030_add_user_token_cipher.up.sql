-- IF NOT EXISTS because integration tests build the same schema through Ent
-- auto-migration before this migration ever runs.
ALTER TABLE user_tokens
    ADD COLUMN IF NOT EXISTS api_key_cipher VARCHAR(512);
