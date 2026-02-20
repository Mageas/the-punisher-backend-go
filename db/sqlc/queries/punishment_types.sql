-- ==================== PunishmentType ====================

-- name: CreatePunishmentType :one
INSERT INTO punishment_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetPunishmentTypeByUser :one
SELECT id, user_id, name, created_at, updated_at
FROM punishment_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPunishmentTypesByUser :one
SELECT COUNT(*) FROM punishment_types WHERE user_id = sqlc.arg(user_id);

-- name: ListPunishmentTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM punishment_types
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdatePunishmentTypeByUser :one
UPDATE punishment_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeletePunishmentTypeByUser :execrows
DELETE FROM punishment_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
