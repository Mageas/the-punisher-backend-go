-- ==================== Dashboard ====================

-- name: GetDashboardKpis :one
WITH filtered_students AS (
    SELECT s.id
    FROM students s
    WHERE s.user_id = sqlc.arg(user_id)
      AND (
        sqlc.narg(classroom_id)::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM student_classrooms sc
            WHERE sc.student_id = s.id
              AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        )
    )
)
SELECT
    COUNT(*)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        JOIN filtered_students fs ON fs.id = b.student_id
        WHERE b.user_id = sqlc.arg(user_id)
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        JOIN filtered_students fs ON fs.id = b.student_id
        WHERE b.user_id = sqlc.arg(user_id)
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM bonuses b
        JOIN filtered_students fs ON fs.id = b.student_id
        WHERE b.user_id = sqlc.arg(user_id)
          AND b.used_at IS NULL
    ), 0)::bigint AS unused_bonus_count,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        JOIN filtered_students fs ON fs.id = p.student_id
        WHERE p.user_id = sqlc.arg(user_id)
    ), 0)::bigint AS penalty_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        JOIN filtered_students fs ON fs.id = p.student_id
        WHERE p.user_id = sqlc.arg(user_id)
    ), 0)::bigint AS total_punishment_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        JOIN filtered_students fs ON fs.id = p.student_id
        WHERE p.user_id = sqlc.arg(user_id)
          AND p.resolved_at IS NULL
          AND p.due_at < NOW()
    ), 0)::bigint AS overdue_punishment_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        JOIN filtered_students fs ON fs.id = p.student_id
        WHERE p.user_id = sqlc.arg(user_id)
          AND p.resolved_at IS NULL
    ), 0)::bigint AS pending_punishment_count
FROM filtered_students;

-- name: ListDashboardRecentPenalties :many
WITH filtered_students AS (
    SELECT s.id
    FROM students s
    WHERE s.user_id = sqlc.arg(user_id)
      AND (
        sqlc.narg(classroom_id)::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM student_classrooms sc
            WHERE sc.student_id = s.id
              AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        )
    )
)
SELECT
    p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS penalty_type_name
FROM penalties p
JOIN filtered_students fs ON fs.id = p.student_id
JOIN students s ON s.id = p.student_id
JOIN penalty_types pt ON pt.id = p.penalty_type_id
WHERE p.user_id = sqlc.arg(user_id)
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit);

-- name: ListDashboardRecentBonuses :many
WITH filtered_students AS (
    SELECT s.id
    FROM students s
    WHERE s.user_id = sqlc.arg(user_id)
      AND (
        sqlc.narg(classroom_id)::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM student_classrooms sc
            WHERE sc.student_id = s.id
              AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        )
    )
)
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN filtered_students fs ON fs.id = b.student_id
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.user_id = sqlc.arg(user_id)
ORDER BY b.created_at DESC
LIMIT sqlc.arg(query_limit);

-- name: ListDashboardPendingPunishments :many
WITH filtered_students AS (
    SELECT s.id
    FROM students s
    WHERE s.user_id = sqlc.arg(user_id)
      AND (
        sqlc.narg(classroom_id)::uuid IS NULL
        OR EXISTS (
            SELECT 1
            FROM student_classrooms sc
            WHERE sc.student_id = s.id
              AND sc.classroom_id = sqlc.narg(classroom_id)::uuid
        )
    )
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.automated, p.created_at, p.due_at, p.resolved_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    pt.name AS punishment_type_name,
    r.name AS triggering_rule_name
FROM punishments p
JOIN filtered_students fs ON fs.id = p.student_id
JOIN students s ON s.id = p.student_id
JOIN punishment_types pt ON pt.id = p.punishment_type_id
LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
WHERE p.user_id = sqlc.arg(user_id)
  AND p.resolved_at IS NULL
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit);
