-- ==================== Bonus ====================

-- name: CreateBonus :one
INSERT INTO bonuses (
    user_id, student_id, bonus_type_id, points, occurred_at, evaluation_label
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(student_id),
    sqlc.arg(bonus_type_id),
    sqlc.arg(points),
    COALESCE(sqlc.narg(occurred_at)::timestamptz, NOW()),
    sqlc.narg(evaluation_label)::text
)
RETURNING
    id, user_id, student_id, bonus_type_id, points, created_at, occurred_at, evaluation_label, used_at,
    (SELECT first_name FROM students WHERE students.id = student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = student_id) AS student_last_name,
    (SELECT name FROM bonus_types WHERE bonus_types.id = bonus_type_id) AS bonus_type_name;

-- name: GetBonusByUser :one
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.occurred_at, b.evaluation_label, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.id = sqlc.arg(id) AND b.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountBonusesByUser :one
SELECT COUNT(*)
FROM bonuses b
WHERE b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(student_id)::uuid IS NULL OR b.student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(bonus_type_id)::uuid IS NULL OR b.bonus_type_id = sqlc.narg(bonus_type_id)::uuid)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
  AND (sqlc.narg(created_from)::date IS NULL OR b.occurred_at >= sqlc.narg(created_from)::date)
  AND (sqlc.narg(created_to)::date IS NULL OR b.occurred_at < (sqlc.narg(created_to)::date + INTERVAL '1 day'))
  AND (
    sqlc.narg(classroom_id)::uuid IS NULL
    OR EXISTS (
      SELECT 1
      FROM student_classrooms sc
      JOIN classrooms c ON c.id = sc.classroom_id
      WHERE sc.student_id = b.student_id
        AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        AND c.user_id = b.user_id
    )
  );

-- name: ListBonusesByUser :many
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.occurred_at, b.evaluation_label, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(student_id)::uuid IS NULL OR b.student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(bonus_type_id)::uuid IS NULL OR b.bonus_type_id = sqlc.narg(bonus_type_id)::uuid)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
  AND (sqlc.narg(created_from)::date IS NULL OR b.occurred_at >= sqlc.narg(created_from)::date)
  AND (sqlc.narg(created_to)::date IS NULL OR b.occurred_at < (sqlc.narg(created_to)::date + INTERVAL '1 day'))
  AND (
    sqlc.narg(classroom_id)::uuid IS NULL
    OR EXISTS (
      SELECT 1
      FROM student_classrooms sc
      JOIN classrooms c ON c.id = sc.classroom_id
      WHERE sc.student_id = b.student_id
        AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        AND c.user_id = b.user_id
    )
  )
ORDER BY b.occurred_at DESC, b.id DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountBonusesByStudent :one
SELECT COUNT(*)
FROM bonuses
WHERE student_id = sqlc.arg(student_id)
  AND user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (used_at IS NOT NULL) = sqlc.narg(used)::boolean);

-- name: ListBonusesByStudent :many
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.occurred_at, b.evaluation_label, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.student_id = sqlc.arg(student_id)
  AND b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
ORDER BY b.occurred_at DESC, b.id DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UseBonus :one
UPDATE bonuses
SET used_at = NOW()
WHERE bonuses.id = sqlc.arg(id) AND bonuses.user_id = sqlc.arg(user_id) AND bonuses.used_at IS NULL
RETURNING
    bonuses.id, bonuses.user_id, bonuses.student_id, bonuses.bonus_type_id, bonuses.points, bonuses.created_at, bonuses.occurred_at, bonuses.evaluation_label, bonuses.used_at,
    (SELECT first_name FROM students WHERE students.id = bonuses.student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = bonuses.student_id) AS student_last_name,
    (SELECT name FROM bonus_types WHERE bonus_types.id = bonuses.bonus_type_id) AS bonus_type_name;

-- name: UpdateBonusByUser :one
UPDATE bonuses
SET
    occurred_at = COALESCE(sqlc.narg(occurred_at)::timestamptz, occurred_at),
    evaluation_label = CASE
        WHEN sqlc.arg(evaluation_label_set)::boolean THEN sqlc.narg(evaluation_label)::text
        ELSE evaluation_label
    END
WHERE bonuses.id = sqlc.arg(id) AND bonuses.user_id = sqlc.arg(user_id)
RETURNING
    bonuses.id, bonuses.user_id, bonuses.student_id, bonuses.bonus_type_id, bonuses.points, bonuses.created_at, bonuses.occurred_at, bonuses.evaluation_label, bonuses.used_at,
    (SELECT first_name FROM students WHERE students.id = bonuses.student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = bonuses.student_id) AS student_last_name,
    (SELECT name FROM bonus_types WHERE bonus_types.id = bonuses.bonus_type_id) AS bonus_type_name;

-- name: DeleteBonusByUser :execrows
DELETE FROM bonuses
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
