-- ==================== Student ====================

-- name: CreateStudent :one
INSERT INTO students (
    user_id, first_name, last_name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(first_name), sqlc.arg(last_name)
)
RETURNING
    id, user_id, first_name, last_name, created_at, updated_at,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = students.id
          AND b.user_id = students.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = students.id
          AND p.user_id = students.user_id
    ), 0)::bigint AS penalty_count;

-- name: GetStudentByUser :one
SELECT
    s.id, s.user_id, s.first_name, s.last_name, s.created_at, s.updated_at,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = s.id
          AND b.user_id = s.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = s.id
          AND p.user_id = s.user_id
    ), 0)::bigint AS penalty_count
FROM students s
WHERE s.id = sqlc.arg(id) AND s.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: ListStudentsByUserForImport :many
SELECT
    s.id, s.first_name, s.last_name
FROM students s
WHERE s.user_id = sqlc.arg(user_id)
ORDER BY s.created_at ASC, s.id ASC;

-- name: CountStudentsByUser :one
SELECT COUNT(*)
FROM students s
WHERE s.user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  );

-- name: ListStudentsByUser :many
SELECT
    s.id, s.user_id, s.first_name, s.last_name, s.created_at, s.updated_at,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = s.id
          AND b.user_id = s.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = s.id
          AND p.user_id = s.user_id
    ), 0)::bigint AS penalty_count
FROM students s
WHERE s.user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  )
ORDER BY s.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateStudentByUser :one
UPDATE students
SET
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    updated_at = NOW()
WHERE students.id = sqlc.arg(id) AND students.user_id = sqlc.arg(user_id)
RETURNING
    students.id, students.user_id, students.first_name, students.last_name, students.created_at, students.updated_at,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = students.id
          AND b.user_id = students.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = students.id
          AND p.user_id = students.user_id
    ), 0)::bigint AS penalty_count;

-- name: DeleteStudentByUser :execrows
DELETE FROM students
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: DeleteAllStudentsByUser :execrows
DELETE FROM students
WHERE user_id = sqlc.arg(user_id);
