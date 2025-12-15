-- name: CreateUser :one
INSERT INTO users (
    email, first_name, last_name, password_hash
) VALUES (
    LOWER(sqlc.arg(email)), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(password_hash)
)
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1
);

-- name: GetUserCredentialsByEmailForAuth :one
SELECT id, email, password_hash FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1;

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
WHERE user_id = sqlc.arg(user_id) AND token = sqlc.arg(token) AND revoked_at IS NULL LIMIT 1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = sqlc.arg(token)
RETURNING id, user_id, token, revoked_at, expires_at;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = sqlc.arg(token);

-- name: ListRefreshTokensByUserId :many
SELECT id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;

-- -- name: GetUser :one
-- SELECT id, email, first_name, last_name, created_at, updated_at
-- FROM users
-- WHERE id = sqlc.arg(id) LIMIT 1;

-- -- name: UpdateUser :one
-- UPDATE users
-- SET
--     email = COALESCE(LOWER(sqlc.arg(email)), email),
--     first_name = COALESCE(sqlc.narg(first_name), first_name),
--     last_name = COALESCE(sqlc.narg(last_name), last_name),
--     updated_at = NOW()
-- WHERE id = sqlc.arg(id)
-- RETURNING id, email, first_name, last_name, created_at, updated_at;

-- -- name: UpdateUserPassword :one
-- UPDATE users
-- SET
--     password_hash = sqlc.arg(password_hash),
--     updated_at = NOW()
-- WHERE id = sqlc.arg(id)
-- RETURNING id, updated_at;

-- -- name: GetUserPasswordByEmailForAuth :one
-- SELECT password_hash FROM users
-- WHERE email = LOWER(sqlc.arg(email)) LIMIT 1;