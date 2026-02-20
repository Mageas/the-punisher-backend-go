-- ==================== Penalty ====================

-- name: CreatePenalty :one
INSERT INTO penalties (
    user_id, student_id, penalty_type_id
) VALUES (
    sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(penalty_type_id)
)
RETURNING
    id, user_id, student_id, penalty_type_id, created_at,
    (SELECT first_name FROM students WHERE students.id = student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = student_id) AS student_last_name,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalty_type_id) AS penalty_type_name;

-- name: GetPenaltyByUser :one
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id
WHERE p.id = sqlc.arg(id) AND p.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPenaltiesByUser :one
SELECT COUNT(*) FROM penalties WHERE user_id = sqlc.arg(user_id);

-- name: ListPenaltiesByUser :many
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id
WHERE p.user_id = sqlc.arg(user_id)
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountPenaltiesByStudent :one
SELECT COUNT(*)
FROM penalties
WHERE student_id = sqlc.arg(student_id) AND user_id = sqlc.arg(user_id);

-- name: CountPenaltiesByStudentAndType :one
SELECT COUNT(*)
FROM penalties
WHERE student_id = sqlc.arg(student_id)
  AND user_id = sqlc.arg(user_id)
  AND penalty_type_id = sqlc.arg(penalty_type_id);

-- name: ListPenaltiesByStudent :many
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id
WHERE p.student_id = sqlc.arg(student_id) AND p.user_id = sqlc.arg(user_id)
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: DeletePenaltyByUser :execrows
DELETE FROM penalties
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
