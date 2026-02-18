-- ==================== User ====================

-- name: CreateUser :one
INSERT INTO users (
    email, first_name, last_name, password_hash
) VALUES (
    LOWER(sqlc.arg(email)), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(password_hash)
)
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users
WHERE id = sqlc.arg(id) LIMIT 1;

-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1
);

-- name: GetUserCredentialsByEmailForAuth :one
SELECT id, email, password_hash FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1;

-- ==================== RefreshToken ====================

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token, user_agent, client_ip, expires_at
) VALUES (
    sqlc.arg(user_id), sqlc.arg(token), sqlc.arg(user_agent), sqlc.arg(client_ip), sqlc.arg(expires_at)
)
RETURNING id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id) AND token = sqlc.arg(token) AND revoked_at IS NULL LIMIT 1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = sqlc.arg(token)
RETURNING id, user_id, token, revoked_at, expires_at;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = sqlc.arg(token);

-- name: ListRefreshTokensByUserId :many
SELECT id, user_id, token, user_agent, client_ip, revoked_at, expires_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;

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

-- ==================== Classroom ====================

-- name: CreateClassroom :one
INSERT INTO classrooms (
    user_id, name, year, main_teacher
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.narg(year), sqlc.narg(main_teacher)
)
RETURNING
    id, user_id, name, year, main_teacher, created_at, updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        WHERE sc.classroom_id = classrooms.id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN bonuses b ON b.student_id = s.id
        WHERE sc.classroom_id = classrooms.id
          AND b.user_id = classrooms.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN penalties p ON p.student_id = s.id
        WHERE sc.classroom_id = classrooms.id
          AND p.user_id = classrooms.user_id
    ), 0)::bigint AS total_penalty_count;

-- name: GetClassroomByUser :one
SELECT
    c.id, c.user_id, c.name, c.year, c.main_teacher, c.created_at, c.updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        WHERE sc.classroom_id = c.id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN bonuses b ON b.student_id = s.id
        WHERE sc.classroom_id = c.id
          AND b.user_id = c.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN penalties p ON p.student_id = s.id
        WHERE sc.classroom_id = c.id
          AND p.user_id = c.user_id
    ), 0)::bigint AS total_penalty_count
FROM classrooms c
WHERE c.id = sqlc.arg(id) AND c.user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountClassroomsByUser :one
SELECT COUNT(*) FROM classrooms WHERE user_id = sqlc.arg(user_id);

-- name: ListClassroomsByUser :many
SELECT
    c.id, c.user_id, c.name, c.year, c.main_teacher, c.created_at, c.updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        WHERE sc.classroom_id = c.id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN bonuses b ON b.student_id = s.id
        WHERE sc.classroom_id = c.id
          AND b.user_id = c.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN penalties p ON p.student_id = s.id
        WHERE sc.classroom_id = c.id
          AND p.user_id = c.user_id
    ), 0)::bigint AS total_penalty_count
FROM classrooms c
WHERE c.user_id = sqlc.arg(user_id)
ORDER BY c.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateClassroomByUser :one
UPDATE classrooms
SET
    name = COALESCE(sqlc.narg(name), name),
    year = COALESCE(sqlc.narg(year), year),
    main_teacher = COALESCE(sqlc.narg(main_teacher), main_teacher),
    updated_at = NOW()
WHERE classrooms.id = sqlc.arg(id) AND classrooms.user_id = sqlc.arg(user_id)
RETURNING
    classrooms.id, classrooms.user_id, classrooms.name, classrooms.year, classrooms.main_teacher, classrooms.created_at, classrooms.updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        WHERE sc.classroom_id = classrooms.id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN bonuses b ON b.student_id = s.id
        WHERE sc.classroom_id = classrooms.id
          AND b.user_id = classrooms.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc
        JOIN students s ON s.id = sc.student_id
        JOIN penalties p ON p.student_id = s.id
        WHERE sc.classroom_id = classrooms.id
          AND p.user_id = classrooms.user_id
    ), 0)::bigint AS total_penalty_count;

-- name: DeleteClassroomByUser :execrows
DELETE FROM classrooms
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- ==================== StudentClassroom ====================

-- name: AddStudentToClassroom :execrows
INSERT INTO student_classrooms (student_id, classroom_id)
SELECT sqlc.arg(student_id), sqlc.arg(classroom_id)
WHERE EXISTS (
    SELECT 1 FROM students st WHERE st.id = sqlc.arg(student_id) AND st.user_id = sqlc.arg(user_id)
) AND EXISTS (
    SELECT 1 FROM classrooms cl WHERE cl.id = sqlc.arg(classroom_id) AND cl.user_id = sqlc.arg(user_id)
);

-- name: RemoveStudentFromClassroom :execrows
DELETE FROM student_classrooms
WHERE student_id = sqlc.arg(student_id)
  AND classroom_id = sqlc.arg(classroom_id)
  AND EXISTS (
    SELECT 1 FROM classrooms cl WHERE cl.id = sqlc.arg(classroom_id) AND cl.user_id = sqlc.arg(user_id)
  );

-- name: CountStudentsByClassroom :one
SELECT COUNT(*)
FROM student_classrooms sc
JOIN classrooms c ON c.id = sc.classroom_id
WHERE sc.classroom_id = sqlc.arg(classroom_id) AND c.user_id = sqlc.arg(user_id);

-- name: ListStudentsByClassroom :many
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
JOIN student_classrooms sc ON sc.student_id = s.id
JOIN classrooms c ON c.id = sc.classroom_id
WHERE sc.classroom_id = sqlc.arg(classroom_id) AND c.user_id = sqlc.arg(user_id)
ORDER BY s.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListClassroomRefsByStudentIDs :many
SELECT
    sc.student_id,
    c.id AS classroom_id,
    c.name AS classroom_name
FROM student_classrooms sc
JOIN classrooms c ON c.id = sc.classroom_id
JOIN students s ON s.id = sc.student_id
WHERE s.user_id = sqlc.arg(user_id)
  AND sc.student_id = ANY(sqlc.arg(student_ids)::uuid[])
ORDER BY c.created_at DESC;

-- name: CountClassroomsByStudent :one
SELECT COUNT(*)
FROM student_classrooms sc
JOIN students s ON s.id = sc.student_id
WHERE sc.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id);

-- name: ListClassroomsByStudent :many
SELECT
    c.id, c.user_id, c.name, c.year, c.main_teacher, c.created_at, c.updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc2
        WHERE sc2.classroom_id = c.id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc2
        JOIN students s2 ON s2.id = sc2.student_id
        JOIN bonuses b ON b.student_id = s2.id
        WHERE sc2.classroom_id = c.id
          AND b.user_id = c.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc2
        JOIN students s2 ON s2.id = sc2.student_id
        JOIN penalties p ON p.student_id = s2.id
        WHERE sc2.classroom_id = c.id
          AND p.user_id = c.user_id
    ), 0)::bigint AS total_penalty_count
FROM classrooms c
JOIN student_classrooms sc ON sc.classroom_id = c.id
JOIN students s ON s.id = sc.student_id
WHERE sc.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id)
ORDER BY c.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListStudentsPreviewByClassroomIDs :many
SELECT
    c.id AS classroom_id,
    preview.student_id,
    preview.first_name,
    preview.last_name
FROM classrooms c
JOIN LATERAL (
    SELECT
        s.id AS student_id,
        s.first_name,
        s.last_name
    FROM student_classrooms sc
    JOIN students s ON s.id = sc.student_id
    WHERE sc.classroom_id = c.id
    ORDER BY s.created_at DESC
    LIMIT sqlc.arg(preview_limit)
) preview ON TRUE
WHERE c.user_id = sqlc.arg(user_id)
  AND c.id = ANY(sqlc.arg(classroom_ids)::uuid[])
ORDER BY c.id;

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
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
    ), 0)::bigint AS total_penalty_count,
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
    history.due_at,
    history.resolved_at
FROM (
    SELECT
        'punishment'::text AS type,
        p.id,
        p.created_at,
        '00000000-0000-0000-0000-000000000000'::uuid AS penalty_type_id,
        ''::text AS penalty_type_name,
        '00000000-0000-0000-0000-000000000000'::uuid AS bonus_type_id,
        ''::text AS bonus_type_name,
        0::double precision AS points,
        '1970-01-01T00:00:00Z'::timestamptz AS used_at,
        p.punishment_type_id,
        pt.name AS punishment_type_name,
        COALESCE(p.triggering_rule_id, '00000000-0000-0000-0000-000000000000'::uuid) AS triggering_rule_id,
        COALESCE(r.name, ''::text) AS triggering_rule_name,
        p.due_at,
        COALESCE(p.resolved_at, '1970-01-01T00:00:00Z'::timestamptz) AS resolved_at
    FROM punishments p
    JOIN punishment_types pt ON pt.id = p.punishment_type_id
    LEFT JOIN rules r ON r.id = p.triggering_rule_id AND r.user_id = p.user_id
    WHERE p.student_id = sqlc.arg(student_id)
      AND p.user_id = sqlc.arg(user_id)

    UNION ALL

    SELECT
        'penalty'::text AS type,
        p.id,
        p.created_at,
        p.penalty_type_id,
        pt.name AS penalty_type_name,
        '00000000-0000-0000-0000-000000000000'::uuid AS bonus_type_id,
        ''::text AS bonus_type_name,
        0::double precision AS points,
        '1970-01-01T00:00:00Z'::timestamptz AS used_at,
        '00000000-0000-0000-0000-000000000000'::uuid AS punishment_type_id,
        ''::text AS punishment_type_name,
        '00000000-0000-0000-0000-000000000000'::uuid AS triggering_rule_id,
        ''::text AS triggering_rule_name,
        '1970-01-01T00:00:00Z'::timestamptz AS due_at,
        '1970-01-01T00:00:00Z'::timestamptz AS resolved_at
    FROM penalties p
    JOIN penalty_types pt ON pt.id = p.penalty_type_id
    WHERE p.student_id = sqlc.arg(student_id)
      AND p.user_id = sqlc.arg(user_id)

    UNION ALL

    SELECT
        'bonus'::text AS type,
        b.id,
        b.created_at,
        '00000000-0000-0000-0000-000000000000'::uuid AS penalty_type_id,
        ''::text AS penalty_type_name,
        b.bonus_type_id,
        bt.name AS bonus_type_name,
        b.points,
        COALESCE(b.used_at, '1970-01-01T00:00:00Z'::timestamptz) AS used_at,
        '00000000-0000-0000-0000-000000000000'::uuid AS punishment_type_id,
        ''::text AS punishment_type_name,
        '00000000-0000-0000-0000-000000000000'::uuid AS triggering_rule_id,
        ''::text AS triggering_rule_name,
        '1970-01-01T00:00:00Z'::timestamptz AS due_at,
        '1970-01-01T00:00:00Z'::timestamptz AS resolved_at
    FROM bonuses b
    JOIN bonus_types bt ON bt.id = b.bonus_type_id
    WHERE b.student_id = sqlc.arg(student_id)
      AND b.user_id = sqlc.arg(user_id)
) history
ORDER BY history.created_at DESC, history.id DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- ==================== BonusType ====================

-- name: CreateBonusType :one
INSERT INTO bonus_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetBonusTypeByUser :one
SELECT id, user_id, name, created_at, updated_at
FROM bonus_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountBonusTypesByUser :one
SELECT COUNT(*) FROM bonus_types WHERE user_id = sqlc.arg(user_id);

-- name: ListBonusTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM bonus_types
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateBonusTypeByUser :one
UPDATE bonus_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeleteBonusTypeByUser :execrows
DELETE FROM bonus_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

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
SELECT COUNT(*) FROM penalty_types WHERE user_id = sqlc.arg(user_id);

-- name: ListPenaltyTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM penalty_types
WHERE user_id = sqlc.arg(user_id)
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

-- ==================== PunishmentType ====================

-- name: CreatePunishmentType :one
INSERT INTO punishment_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetPunishmentTypeByUser :one
SELECT id, user_id, name, created_at, updated_at
FROM punishment_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPunishmentTypesByUser :one
SELECT COUNT(*) FROM punishment_types WHERE user_id = sqlc.arg(user_id);

-- name: ListPunishmentTypesByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM punishment_types
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdatePunishmentTypeByUser :one
UPDATE punishment_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeletePunishmentTypeByUser :execrows
DELETE FROM punishment_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- ==================== Bonus ====================

-- name: CreateBonus :one
INSERT INTO bonuses (
    user_id, student_id, bonus_type_id, points
) VALUES (
    sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(bonus_type_id), sqlc.arg(points)
)
RETURNING
    id, user_id, student_id, bonus_type_id, points, created_at, used_at,
    (SELECT first_name FROM students WHERE students.id = student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = student_id) AS student_last_name,
    (SELECT name FROM bonus_types WHERE bonus_types.id = bonus_type_id) AS bonus_type_name;

-- name: GetBonusByUser :one
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.used_at,
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
JOIN students s ON s.id = b.student_id AND s.user_id = b.user_id
WHERE b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  );

-- name: ListBonusesByUser :many
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  )
ORDER BY b.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountBonusesByStudent :one
SELECT COUNT(*)
FROM bonuses
WHERE student_id = sqlc.arg(student_id)
  AND user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (used_at IS NOT NULL) = sqlc.narg(used)::boolean);

-- name: ListBonusesByStudent :many
SELECT
    b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.used_at,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.student_id = sqlc.arg(student_id)
  AND b.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(used)::boolean IS NULL OR (b.used_at IS NOT NULL) = sqlc.narg(used)::boolean)
ORDER BY b.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UseBonus :one
UPDATE bonuses
SET used_at = NOW()
WHERE bonuses.id = sqlc.arg(id) AND bonuses.user_id = sqlc.arg(user_id) AND bonuses.used_at IS NULL
RETURNING
    bonuses.id, bonuses.user_id, bonuses.student_id, bonuses.bonus_type_id, bonuses.points, bonuses.created_at, bonuses.used_at,
    (SELECT first_name FROM students WHERE students.id = bonuses.student_id) AS student_first_name,
    (SELECT last_name FROM students WHERE students.id = bonuses.student_id) AS student_last_name,
    (SELECT name FROM bonus_types WHERE bonus_types.id = bonuses.bonus_type_id) AS bonus_type_name;

-- name: DeleteBonusByUser :execrows
DELETE FROM bonuses
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

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

-- ==================== Rule ====================

-- name: CreateRule :one
INSERT INTO rules (
    user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, due_at_after_days
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(resulting_punishment_type_id), sqlc.arg(penalty_type_id), sqlc.arg(threshold), sqlc.arg(mode), sqlc.arg(is_active), sqlc.arg(due_at_after_days)
)
RETURNING
    id, user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, created_at, updated_at, due_at_after_days,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalty_type_id) AS penalty_type_name,
    (SELECT name FROM punishment_types WHERE punishment_types.id = resulting_punishment_type_id) AS resulting_punishment_type_name;

-- name: GetRuleByUser :one
SELECT
    r.id, r.user_id, r.name, r.resulting_punishment_type_id, r.penalty_type_id, r.threshold, r.mode, r.is_active, r.created_at, r.updated_at, r.due_at_after_days,
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
    r.id, r.user_id, r.name, r.resulting_punishment_type_id, r.penalty_type_id, r.threshold, r.mode, r.is_active, r.created_at, r.updated_at, r.due_at_after_days,
    pt.name AS penalty_type_name,
    put.name AS resulting_punishment_type_name
FROM rules r
JOIN penalty_types pt ON pt.id = r.penalty_type_id
JOIN punishment_types put ON put.id = r.resulting_punishment_type_id
WHERE r.user_id = sqlc.arg(user_id)
ORDER BY r.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListActiveRulesByUserAndPenaltyType :many
SELECT id, user_id, name, resulting_punishment_type_id, penalty_type_id, threshold, mode, is_active, created_at, updated_at, due_at_after_days
FROM rules
WHERE user_id = sqlc.arg(user_id)
  AND penalty_type_id = sqlc.arg(penalty_type_id)
  AND is_active = TRUE
ORDER BY created_at DESC;

-- name: UpdateRuleByUser :one
UPDATE rules
SET
    name = COALESCE(sqlc.narg(name), name),
    resulting_punishment_type_id = COALESCE(sqlc.narg(resulting_punishment_type_id), resulting_punishment_type_id),
    penalty_type_id = COALESCE(sqlc.narg(penalty_type_id), penalty_type_id),
    threshold = COALESCE(sqlc.narg(threshold), threshold),
    mode = COALESCE(sqlc.narg(mode), mode),
    is_active = COALESCE(sqlc.narg(is_active)::boolean, is_active),
    due_at_after_days = COALESCE(sqlc.narg(due_at_after_days), due_at_after_days),
    updated_at = NOW()
WHERE rules.id = sqlc.arg(id) AND rules.user_id = sqlc.arg(user_id)
RETURNING
    rules.id, rules.user_id, rules.name, rules.resulting_punishment_type_id, rules.penalty_type_id, rules.threshold, rules.mode, rules.is_active, rules.created_at, rules.updated_at, rules.due_at_after_days,
    (SELECT name FROM penalty_types WHERE penalty_types.id = penalty_type_id) AS penalty_type_name,
    (SELECT name FROM punishment_types WHERE punishment_types.id = resulting_punishment_type_id) AS resulting_punishment_type_name;

-- name: DeleteRuleByUser :execrows
DELETE FROM rules
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- ==================== Punishment ====================

-- name: CreatePunishment :one
WITH created_punishment AS (
    INSERT INTO punishments (
        user_id, student_id, punishment_type_id, due_at
    ) VALUES (
        sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(punishment_type_id), sqlc.arg(due_at)
    )
    RETURNING id, user_id, student_id, punishment_type_id, triggering_rule_id, created_at, due_at, resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
        user_id, student_id, punishment_type_id, triggering_rule_id, due_at
    ) VALUES (
        sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(punishment_type_id), sqlc.arg(triggering_rule_id), sqlc.arg(due_at)
    )
    RETURNING id, user_id, student_id, punishment_type_id, triggering_rule_id, created_at, due_at, resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
    RETURNING punishments.id, punishments.user_id, punishments.student_id, punishments.punishment_type_id, punishments.triggering_rule_id, punishments.created_at, punishments.due_at, punishments.resolved_at
)
SELECT
    p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at,
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
