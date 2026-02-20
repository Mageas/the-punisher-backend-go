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
