-- ==================== Rule ====================

-- name: CreateRule :one
INSERT INTO rules (
    user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, due_at_after_days, due_at_mode, due_at_after_lessons
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(resulting_punishment_type_id), sqlc.arg(penalty_type_id), sqlc.arg(threshold), sqlc.arg(mode), sqlc.arg(is_active), sqlc.arg(due_at_after_days), sqlc.arg(due_at_mode), sqlc.narg(due_at_after_lessons)
)
RETURNING
    id, user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, created_at, updated_at, due_at_after_days, due_at_mode, due_at_after_lessons,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalty_type_id) AS penalty_type_name,
    (SELECT name FROM punishment_types WHERE punishment_types.id = resulting_punishment_type_id) AS resulting_punishment_type_name;

-- name: GetRuleByUser :one
SELECT
    r.id, r.user_id, r.name, r.resulting_punishment_type_id, r.penalty_type_id, r.threshold, r.mode, r.is_active, r.created_at, r.updated_at, r.due_at_after_days, r.due_at_mode, r.due_at_after_lessons,
    pt.name AS penalty_type_name,
    put.name AS resulting_punishment_type_name
FROM rules r
JOIN penalty_types pt ON pt.id = r.penalty_type_id
JOIN punishment_types put ON put.id = r.resulting_punishment_type_id
WHERE r.id = sqlc.arg(id) AND r.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountRulesByUser :one
SELECT COUNT(*) FROM rules WHERE user_id = sqlc.arg(user_id);

-- name: ListRulesByUser :many
SELECT
    r.id, r.user_id, r.name, r.resulting_punishment_type_id, r.penalty_type_id, r.threshold, r.mode, r.is_active, r.created_at, r.updated_at, r.due_at_after_days, r.due_at_mode, r.due_at_after_lessons,
    pt.name AS penalty_type_name,
    put.name AS resulting_punishment_type_name
FROM rules r
JOIN penalty_types pt ON pt.id = r.penalty_type_id
JOIN punishment_types put ON put.id = r.resulting_punishment_type_id
WHERE r.user_id = sqlc.arg(user_id)
ORDER BY r.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListActiveRulesByUserAndPenaltyType :many
SELECT id, user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, created_at, updated_at, due_at_after_days, due_at_mode, due_at_after_lessons
FROM rules
WHERE user_id = sqlc.arg(user_id)
  AND penalty_type_id = sqlc.arg(penalty_type_id)
  AND is_active = TRUE
ORDER BY created_at DESC;

-- name: UpdateRuleByUser :one
UPDATE rules
SET
    name = sqlc.arg(name),
    resulting_punishment_type_id = sqlc.arg(resulting_punishment_type_id),
    penalty_type_id = sqlc.arg(penalty_type_id),
    threshold = sqlc.arg(threshold),
    mode = sqlc.arg(mode),
    is_active = sqlc.arg(is_active),
    due_at_after_days = sqlc.arg(due_at_after_days),
    due_at_mode = sqlc.arg(due_at_mode),
    due_at_after_lessons = sqlc.narg(due_at_after_lessons),
    updated_at = NOW()
WHERE rules.id = sqlc.arg(id) AND rules.user_id = sqlc.arg(user_id)
RETURNING
    rules.id, rules.user_id, rules.name, rules.resulting_punishment_type_id, rules.penalty_type_id, rules.threshold, rules.mode, rules.is_active, rules.created_at, rules.updated_at, rules.due_at_after_days, rules.due_at_mode, rules.due_at_after_lessons,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalty_type_id) AS penalty_type_name,
    (SELECT name FROM punishment_types WHERE punishment_types.id = resulting_punishment_type_id) AS resulting_punishment_type_name;

-- name: DeleteRuleByUser :execrows
DELETE FROM rules
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
