-- ==================== RefreshToken ====================

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token, user_agent, client_ip, expires_at
) VALUES (
    sqlc.arg(user_id), sqlc.arg(token), sqlc.arg(user_agent), sqlc.arg(client_ip), sqlc.arg(expires_at)
)
RETURNING id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id) AND token = sqlc.arg(token) AND revoked_at IS NULL AND expires_at > NOW() LIMIT 1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = sqlc.arg(token) AND revoked_at IS NULL
RETURNING id, user_id, token, revoked_at, expires_at;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = sqlc.arg(token);

-- name: DeleteRefreshTokensByUserId :execrows
DELETE FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id);

-- name: ListRefreshTokensByUserId :many
SELECT id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;
