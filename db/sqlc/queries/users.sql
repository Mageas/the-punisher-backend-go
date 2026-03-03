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

-- name: GetUserEmailVerificationStateByID :one
SELECT id, email_verified_at
FROM users
WHERE id = sqlc.arg(id) LIMIT 1;

-- name: GetUserEmailVerificationStateByEmail :one
SELECT id, email, first_name, email_verified_at
FROM users
WHERE email = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: VerifyUserEmailByID :execrows
UPDATE users
SET email_verified_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND email_verified_at IS NULL;

-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1
);

-- name: GetUserCredentialsByEmailForAuth :one
SELECT id, email, password_hash, email_verified_at
FROM users
WHERE email = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: GetUserPasswordCredentialsByIDForAuth :one
SELECT id, password_hash, password_changed_at
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateUserPasswordByID :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash),
    password_changed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id);
