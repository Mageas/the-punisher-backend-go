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