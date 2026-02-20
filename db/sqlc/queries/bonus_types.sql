-- ==================== BonusType ====================

-- name: CreateBonusType :one
INSERT INTO bonus_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetBonusTypeByUser :one
SELECT id, user_id, name, created_at, updated_at
FROM bonus_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountBonusTypesByUser :one
SELECT COUNT(*) FROM bonus_types WHERE user_id = sqlc.arg(user_id);

-- name: ListBonusTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM bonus_types
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateBonusTypeByUser :one
UPDATE bonus_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeleteBonusTypeByUser :execrows
DELETE FROM bonus_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
