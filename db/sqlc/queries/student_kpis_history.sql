-- ==================== StudentKpisHistory ====================

-- name: GetStudentKpis :one
SELECT
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = sqlc.arg(student_id)
          AND b.user_id = sqlc.arg(user_id)
          AND b.used_at IS NULL
    ), 0)::double precision AS available_bonus_points,
    COALESCE((
        SELECT SUM(b.points)
        FROM bonuses b
        WHERE b.student_id = sqlc.arg(student_id)
          AND b.user_id = sqlc.arg(user_id)
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM bonuses b
        WHERE b.student_id = sqlc.arg(student_id)
          AND b.user_id = sqlc.arg(user_id)
          AND b.used_at IS NULL
    ), 0)::bigint AS active_bonus_count,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
    ), 0)::bigint AS penalty_count,
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
    ), 0)::bigint AS total_penalty_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
    ), 0)::bigint AS total_punishment_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
          AND p.resolved_at IS NULL
          AND p.due_at < NOW()
    ), 0)::bigint AS overdue_punishment_count,
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
          AND p.resolved_at IS NULL
    ), 0)::bigint AS pending_punishment_count;

-- name: ListStudentHistory :many
SELECT
    history.type,
    history.id,
    history.created_at,
    history.occurred_at,
    history.evaluation_label,
    history.penalty_type_id,
    history.penalty_type_name,
    history.bonus_type_id,
    history.bonus_type_name,
    history.points,
    history.used_at,
    history.punishment_type_id,
    history.punishment_type_name,
    history.triggering_rule_id,
    history.triggering_rule_name,
    history.automated,
    history.due_at,
    history.resolved_at
FROM (
    SELECT
        'punishment'::text AS type,
        p.id,
        p.created_at,
        p.occurred_at,
        p.evaluation_label,
        NULL::uuid AS penalty_type_id,
        NULL::text AS penalty_type_name,
        NULL::uuid AS bonus_type_id,
        NULL::text AS bonus_type_name,
        NULL::double precision AS points,
        NULL::timestamptz AS used_at,
        CASE WHEN TRUE THEN p.punishment_type_id ELSE NULL::uuid END AS punishment_type_id,
        CASE WHEN TRUE THEN pt.name ELSE NULL::text END AS punishment_type_name,
        p.triggering_rule_id,
        r.name AS triggering_rule_name,
        CASE WHEN TRUE THEN p.automated ELSE NULL::boolean END AS automated,
        CASE WHEN TRUE THEN p.due_at ELSE NULL::timestamptz END AS due_at,
        p.resolved_at
    FROM punishments p
    JOIN punishment_types pt ON pt.id = p.punishment_type_id AND pt.user_id = p.user_id
    LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
    WHERE p.student_id = sqlc.arg(student_id)
      AND p.user_id = sqlc.arg(user_id)

    UNION ALL

    SELECT
        'penalty'::text AS type,
        p.id,
        p.created_at,
        p.occurred_at,
        p.evaluation_label,
        p.penalty_type_id,
        pt.name AS penalty_type_name,
        NULL::uuid AS bonus_type_id,
        NULL::text AS bonus_type_name,
        NULL::double precision AS points,
        NULL::timestamptz AS used_at,
        p.penalty_type_id AS punishment_type_id,
        pt.name AS punishment_type_name,
        NULL::uuid AS triggering_rule_id,
        NULL::text AS triggering_rule_name,
        FALSE AS automated,
        p.occurred_at AS due_at,
        NULL::timestamptz AS resolved_at
    FROM penalties p
    JOIN penalty_types pt ON pt.id = p.penalty_type_id AND pt.user_id = p.user_id
    WHERE p.student_id = sqlc.arg(student_id)
      AND p.user_id = sqlc.arg(user_id)

    UNION ALL

    SELECT
        'bonus'::text AS type,
        b.id,
        b.created_at,
        b.occurred_at,
        b.evaluation_label,
        NULL::uuid AS penalty_type_id,
        NULL::text AS penalty_type_name,
        b.bonus_type_id,
        bt.name AS bonus_type_name,
        b.points,
        b.used_at,
        b.bonus_type_id AS punishment_type_id,
        bt.name AS punishment_type_name,
        NULL::uuid AS triggering_rule_id,
        NULL::text AS triggering_rule_name,
        FALSE AS automated,
        b.occurred_at AS due_at,
        NULL::timestamptz AS resolved_at
    FROM bonuses b
    JOIN bonus_types bt ON bt.id = b.bonus_type_id AND bt.user_id = b.user_id
    WHERE b.student_id = sqlc.arg(student_id)
      AND b.user_id = sqlc.arg(user_id)
) history
ORDER BY history.occurred_at DESC, history.id DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountStudentHistory :one
SELECT (
    COALESCE((
        SELECT COUNT(*)
        FROM punishments p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
    ), 0)
    +
    COALESCE((
        SELECT COUNT(*)
        FROM penalties p
        WHERE p.student_id = sqlc.arg(student_id)
          AND p.user_id = sqlc.arg(user_id)
    ), 0)
    +
    COALESCE((
        SELECT COUNT(*)
        FROM bonuses b
        WHERE b.student_id = sqlc.arg(student_id)
          AND b.user_id = sqlc.arg(user_id)
    ), 0)
)::bigint;
