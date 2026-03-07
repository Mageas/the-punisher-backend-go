-- ==================== StudentClassroom ====================

-- name: AddStudentToClassroom :execrows
INSERT INTO student_classrooms (user_id, student_id, classroom_id)
SELECT sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(classroom_id)
WHERE EXISTS (
    SELECT 1 FROM students st WHERE st.id = sqlc.arg(student_id) AND st.user_id = sqlc.arg(user_id)
) AND EXISTS (
    SELECT 1 FROM classrooms cl WHERE cl.id = sqlc.arg(classroom_id) AND cl.user_id = sqlc.arg(user_id)
);

-- name: RemoveStudentFromClassroom :execrows
DELETE FROM student_classrooms
WHERE user_id = sqlc.arg(user_id)
  AND student_id = sqlc.arg(student_id)
  AND classroom_id = sqlc.arg(classroom_id);

-- name: CountStudentsByClassroom :one
SELECT COUNT(*)
FROM student_classrooms sc
JOIN students s ON s.id = sc.student_id AND s.user_id = sc.user_id
JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
WHERE sc.user_id = sqlc.arg(user_id)
  AND sc.classroom_id = sqlc.arg(classroom_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  );

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
JOIN student_classrooms sc ON sc.student_id = s.id AND sc.user_id = s.user_id
JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
WHERE sc.user_id = sqlc.arg(user_id)
  AND sc.classroom_id = sqlc.arg(classroom_id)
  AND (
    sqlc.narg(search)::text IS NULL
    OR CONCAT_WS(' ', s.first_name, s.last_name) ILIKE '%' || sqlc.narg(search)::text || '%'
  )
ORDER BY s.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListClassroomRefsByStudentIDs :many
SELECT
    sc.student_id,
    c.id AS classroom_id,
    c.name AS classroom_name
FROM student_classrooms sc
JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
JOIN students s ON s.id = sc.student_id AND s.user_id = sc.user_id
WHERE sc.user_id = sqlc.arg(user_id)
  AND sc.student_id = ANY(sqlc.arg(student_ids)::uuid[])
ORDER BY c.created_at DESC;

-- name: CountClassroomsByStudent :one
SELECT COUNT(*)
FROM student_classrooms sc
JOIN students s ON s.id = sc.student_id AND s.user_id = sc.user_id
JOIN classrooms c ON c.id = sc.classroom_id AND c.user_id = sc.user_id
WHERE sc.user_id = sqlc.arg(user_id)
  AND sc.student_id = sqlc.arg(student_id);

-- name: ListClassroomsByStudent :many
SELECT
    c.id, c.user_id, c.name, c.year, c.main_teacher, c.created_at, c.updated_at,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc2
        WHERE sc2.classroom_id = c.id
          AND sc2.user_id = c.user_id
    ), 0)::bigint AS student_count,
    COALESCE((
        SELECT SUM(b.points)
        FROM student_classrooms sc2
        JOIN students s2 ON s2.id = sc2.student_id AND s2.user_id = sc2.user_id
        JOIN bonuses b ON b.student_id = s2.id AND b.user_id = s2.user_id
        WHERE sc2.classroom_id = c.id
          AND sc2.user_id = c.user_id
          AND b.user_id = c.user_id
          AND b.used_at IS NULL
    ), 0)::double precision AS total_bonus_points,
    COALESCE((
        SELECT COUNT(*)
        FROM student_classrooms sc2
        JOIN students s2 ON s2.id = sc2.student_id AND s2.user_id = sc2.user_id
        JOIN penalties p ON p.student_id = s2.id AND p.user_id = s2.user_id
        WHERE sc2.classroom_id = c.id
          AND sc2.user_id = c.user_id
          AND p.user_id = c.user_id
    ), 0)::bigint AS total_penalty_count
FROM classrooms c
JOIN student_classrooms sc ON sc.classroom_id = c.id AND sc.user_id = c.user_id
JOIN students s ON s.id = sc.student_id AND s.user_id = sc.user_id
WHERE sc.user_id = sqlc.arg(user_id)
  AND sc.student_id = sqlc.arg(student_id)
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
    JOIN students s ON s.id = sc.student_id AND s.user_id = sc.user_id
    WHERE sc.user_id = c.user_id
      AND sc.classroom_id = c.id
    ORDER BY s.created_at DESC
    LIMIT sqlc.arg(preview_limit)
) preview ON TRUE
WHERE c.user_id = sqlc.arg(user_id)
  AND c.id = ANY(sqlc.arg(classroom_ids)::uuid[])
ORDER BY c.id;
