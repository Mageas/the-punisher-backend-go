DROP INDEX IF EXISTS idx_email_confirmation_tokens_expires_at;
DROP INDEX IF EXISTS idx_email_confirmation_tokens_user_id;
DROP TABLE IF EXISTS email_confirmation_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified_at;
