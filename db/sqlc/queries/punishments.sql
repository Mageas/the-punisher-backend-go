-- ==================== Punishment ====================

-- name: CreatePunishment :one
WITH created_punishment AS (
    INSERT INTO punishments (
        user_id, student_id, punishment_type_id, automated, due_at
    ) VALUES (
        sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(punishment_type_id), FALSE, sqlc.arg(due_at)
    )
    RETURNING id, user_id, student_id, punishment_type_id, triggering_rule_id, automated, created_at, due_at, resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM created_punishment p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id;

-- name: CreatePunishmentFromRule :one
WITH created_punishment AS (
    INSERT INTO punishments (
        user_id, student_id, punishment_type_id, triggering_rule_id, automated, due_at
    ) VALUES (
        sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(punishment_type_id), sqlc.arg(triggering_rule_id), sqlc.arg(automated), sqlc.arg(due_at)
    )
    RETURNING id, user_id, student_id, punishment_type_id, triggering_rule_id, automated, created_at, due_at, resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM created_punishment p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id;

-- name: GetPunishmentByUser :one
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM punishments p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
WHERE p.id = sqlc.arg(id) AND p.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPunishmentsByUser :one
SELECT COUNT(*)
FROM punishments p
JOIN students s ON s.id = p.student_id AND s.user_id = p.user_id
WHERE p.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(resolved)::boolean IS NULL OR (p.resolved_at IS NOT NULL) = sqlc.narg(resolved)::boolean)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  );

-- name: ListPunishmentsByUser :many
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM punishments p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
WHERE p.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(resolved)::boolean IS NULL OR (p.resolved_at IS NOT NULL) = sqlc.narg(resolved)::boolean)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  )
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountPunishmentsByStudent :one
SELECT COUNT(*)
FROM punishments
WHERE student_id = sqlc.arg(student_id)
  AND user_id = sqlc.arg(user_id)
  AND (sqlc.narg(resolved)::boolean IS NULL OR (resolved_at IS NOT NULL) = sqlc.narg(resolved)::boolean);

-- name: ListPunishmentsByStudent :many
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM punishments p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
WHERE p.student_id = sqlc.arg(student_id)
  AND p.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(resolved)::boolean IS NULL OR (p.resolved_at IS NOT NULL) = sqlc.narg(resolved)::boolean)
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ResolvePunishment :one
WITH resolved_punishment AS (
    UPDATE punishments
    SET resolved_at = NOW()
    WHERE punishments.id = sqlc.arg(id) AND punishments.user_id = sqlc.arg(user_id) AND punishments.resolved_at IS NULL
    RETURNING punishments.id, punishments.user_id, punishments.student_id, punishments.punishment_type_id, punishments.triggering_rule_id, punishments.automated, punishments.created_at, punishments.due_at, punishments.resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM resolved_punishment p
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id;

-- name: DeletePunishmentByUser :execrows
DELETE FROM punishments
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
