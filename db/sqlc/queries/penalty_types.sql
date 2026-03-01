-- ==================== PenaltyType ====================

-- name: CreatePenaltyType :one
INSERT INTO penalty_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetPenaltyTypeByUser :one
SELECT id, user_id, name, created_at, updated_at
FROM penalty_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPenaltyTypesByUser :one
SELECT COUNT(*)
FROM penalty_types
WHERE user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR name ILIKE '%' || sqlc.narg(search)::text || '%'
  );

-- name: ListPenaltyTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM penalty_types
WHERE user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR name ILIKE '%' || sqlc.narg(search)::text || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdatePenaltyTypeByUser :one
UPDATE penalty_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeletePenaltyTypeByUser :execrows
DELETE FROM penalty_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
