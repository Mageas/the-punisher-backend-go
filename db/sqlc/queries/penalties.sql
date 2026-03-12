-- ==================== Penalty ====================

-- name: CreatePenalty :one
INSERT INTO penalties (
    user_id, student_id, penalty_type_id, occurred_at, evaluation_label
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(student_id),
    sqlc.arg(penalty_type_id),
    COALESCE(sqlc.narg(occurred_at)::timestamptz, NOW()),
    COALESCE(sqlc.narg(evaluation_label)::text, '')
)
RETURNING
    id, user_id, student_id, penalty_type_id, created_at, occurred_at, evaluation_label,
    (SELECT first_name FROM students WHERE students.id = penalties.student_id AND students.user_id = penalties.user_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = penalties.student_id AND students.user_id = penalties.user_id) AS student_last_name,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalties.penalty_type_id AND penalty_types.user_id = penalties.user_id) AS penalty_type_name;

-- name: GetPenaltyByUser :one
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at, p.occurred_at, p.evaluation_label,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id AND s.user_id = p.user_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id AND pt.user_id = p.user_id
WHERE p.id = sqlc.arg(id) AND p.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPenaltiesByUser :one
SELECT COUNT(*)
FROM penalties p
WHERE p.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(student_id)::uuid IS NULL OR p.student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(penalty_type_id)::uuid IS NULL OR p.penalty_type_id = sqlc.narg(penalty_type_id)::uuid)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR p.occurred_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR p.occurred_at < sqlc.narg(created_to)::timestamptz)
  AND (
    sqlc.narg(classroom_id)::uuid IS NULL
    OR EXISTS (
      SELECT 1
      FROM student_classrooms sc
      JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
      WHERE sc.student_id = p.student_id
        AND sc.user_id = p.user_id
        AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        AND c.user_id = p.user_id
    )
  );

-- name: ListPenaltiesByUser :many
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at, p.occurred_at, p.evaluation_label,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id AND s.user_id = p.user_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id AND pt.user_id = p.user_id
WHERE p.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(student_id)::uuid IS NULL OR p.student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(penalty_type_id)::uuid IS NULL OR p.penalty_type_id = sqlc.narg(penalty_type_id)::uuid)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR p.occurred_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR p.occurred_at < sqlc.narg(created_to)::timestamptz)
  AND (
    sqlc.narg(classroom_id)::uuid IS NULL
    OR EXISTS (
      SELECT 1
      FROM student_classrooms sc
      JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
      WHERE sc.student_id = p.student_id
        AND sc.user_id = p.user_id
        AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        AND c.user_id = p.user_id
    )
  )
ORDER BY p.occurred_at DESC, p.id DESC
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
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at, p.occurred_at, p.evaluation_label,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN students s ON s.id = p.student_id AND s.user_id = p.user_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id AND pt.user_id = p.user_id
WHERE p.student_id = sqlc.arg(student_id) AND p.user_id = sqlc.arg(user_id)
ORDER BY p.occurred_at DESC, p.id DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdatePenaltyByUser :one
UPDATE penalties
SET
    occurred_at = COALESCE(sqlc.narg(occurred_at)::timestamptz, occurred_at),
    evaluation_label = COALESCE(sqlc.narg(evaluation_label)::text, evaluation_label)
WHERE penalties.id = sqlc.arg(id) AND penalties.user_id = sqlc.arg(user_id)
RETURNING
    penalties.id, penalties.user_id, penalties.student_id, penalties.penalty_type_id, penalties.created_at, penalties.occurred_at, penalties.evaluation_label,
    (SELECT first_name FROM students WHERE students.id = penalties.student_id AND students.user_id = penalties.user_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = penalties.student_id AND students.user_id = penalties.user_id) AS student_last_name,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalties.penalty_type_id AND penalty_types.user_id = penalties.user_id) AS penalty_type_name;

-- name: DeletePenaltyByUser :execrows
DELETE FROM penalties
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
