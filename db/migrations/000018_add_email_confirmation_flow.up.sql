ALTER TABLE users
    ADD COLUMN email_verified_at TIMESTAMPTZ NULL;

UPDATE users
SET email_verified_at = created_at
WHERE email_verified_at IS NULL;

CREATE TABLE email_confirmation_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_confirmation_tokens_user_id
    ON email_confirmation_tokens(user_id);

CREATE INDEX idx_email_confirmation_tokens_expires_at
    ON email_confirmation_tokens(expires_at);
