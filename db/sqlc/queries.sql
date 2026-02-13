-- name: CreateUser :one
INSERT INTO users (
    email, first_name, last_name, password_hash
) VALUES (
    LOWER(sqlc.arg(email)), sqlc.arg(first_name), sqlc.arg(last_name), sqlc.arg(password_hash)
)
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: UserEmailExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1
);

-- name: GetUserCredentialsByEmailForAuth :one
SELECT id, email, password_hash FROM users WHERE email = LOWER(sqlc.arg(email)) LIMIT 1;

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

-- name: CreateStudent :one
INSERT INTO students (
    user_id, first_name, last_name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(first_name), sqlc.arg(last_name)
)
RETURNING id, user_id, first_name, last_name, created_at, updated_at;

-- name: GetStudentByUser :one
SELECT id, user_id, first_name, last_name, created_at, updated_at
FROM students
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountStudentsByUser :one
SELECT COUNT(*) FROM students WHERE user_id = sqlc.arg(user_id);

-- name: ListStudentsByUser :many
SELECT id, user_id, first_name, last_name, created_at, updated_at
FROM students
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateStudentByUser :one
UPDATE students
SET
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, first_name, last_name, created_at, updated_at;

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
RETURNING id, user_id, name, year, main_teacher, created_at, updated_at;

-- name: GetClassroomByUser :one
SELECT id, user_id, name, year, main_teacher, created_at, updated_at
FROM classrooms
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountClassroomsByUser :one
SELECT COUNT(*) FROM classrooms WHERE user_id = sqlc.arg(user_id);

-- name: ListClassroomsByUser :many
SELECT id, user_id, name, year, main_teacher, created_at, updated_at
FROM classrooms
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: UpdateClassroomByUser :one
UPDATE classrooms
SET
    name = COALESCE(sqlc.narg(name), name),
    year = COALESCE(sqlc.narg(year), year),
    main_teacher = COALESCE(sqlc.narg(main_teacher), main_teacher),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, year, main_teacher, created_at, updated_at;

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
SELECT s.id, s.user_id, s.first_name, s.last_name, s.created_at, s.updated_at
FROM students s
JOIN student_classrooms sc ON sc.student_id = s.id
JOIN classrooms c ON c.id = sc.classroom_id
WHERE sc.classroom_id = sqlc.arg(classroom_id) AND c.user_id = sqlc.arg(user_id)
ORDER BY s.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountClassroomsByStudent :one
SELECT COUNT(*)
FROM student_classrooms sc
JOIN students s ON s.id = sc.student_id
WHERE sc.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id);

-- name: ListClassroomsByStudent :many
SELECT c.id, c.user_id, c.name, c.year, c.main_teacher, c.created_at, c.updated_at
FROM classrooms c
JOIN student_classrooms sc ON sc.classroom_id = c.id
JOIN students s ON s.id = sc.student_id
WHERE sc.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id)
ORDER BY c.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);
-- ==================== BonusType ====================

-- name: CreateBonusType :one
INSERT INTO bonus_types (
    user_id, name
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name)
)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetBonusType :one
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

-- name: UpdateBonusType :one
UPDATE bonus_types
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeleteBonusType :execrows
DELETE FROM bonus_types
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
