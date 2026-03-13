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
              AND sc.user_id = s.user_id
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
