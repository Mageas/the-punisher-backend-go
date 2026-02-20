-- ==================== User ====================

-- name: CreateUser :one
INSERT INTO users (
    email, first_name, last_name, password_hash
) VALUES (
    LOWER(sqlc.arg(email)), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(password_hash)
)
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users
WHERE id = sqlc.arg(id) LIMIT 1;

-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1
);

-- name: GetUserCredentialsByEmailForAuth :one
SELECT id, email, password_hash FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1;
