-- ==================== EmailConfirmationToken ====================

-- name: CreateEmailConfirmationToken :one
INSERT INTO email_confirmation_tokens (
    user_id,
    token_hash,
    expires_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at)
)
RETURNING id, user_id, token_hash, expires_at, used_at, created_at;

-- name: GetEmailConfirmationTokenByHash :one
SELECT id, user_id, token_hash, expires_at, used_at, created_at
FROM email_confirmation_tokens
WHERE token_hash = sqlc.arg(token_hash)
LIMIT 1;

-- name: MarkEmailConfirmationTokenUsedByID :execrows
UPDATE email_confirmation_tokens
SET used_at = NOW()
WHERE id = sqlc.arg(id)
  AND used_at IS NULL;

-- name: InvalidateEmailConfirmationTokensByUserID :execrows
UPDATE email_confirmation_tokens
SET used_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND used_at IS NULL;
