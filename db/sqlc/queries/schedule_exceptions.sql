-- ==================== ScheduleException ====================

-- name: CreateScheduleException :one
INSERT INTO schedule_exceptions (
    user_id, exception_type, start_date, end_date
) VALUES (
    sqlc.arg(user_id), sqlc.arg(exception_type), sqlc.arg(start_date), sqlc.arg(end_date)
)
RETURNING
    id,
    user_id,
    exception_type AS type,
    start_date,
    end_date,
    created_at,
    updated_at;

-- name: GetScheduleExceptionByUser :one
SELECT
    id,
    user_id,
    exception_type AS type,
    start_date,
    end_date,
    created_at,
    updated_at
FROM schedule_exceptions
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListScheduleExceptionsByUser :many
SELECT
    id,
    user_id,
    exception_type AS type,
    start_date,
    end_date,
    created_at,
    updated_at
FROM schedule_exceptions
WHERE user_id = sqlc.arg(user_id)
ORDER BY start_date ASC, end_date ASC, id ASC;

-- name: UpdateScheduleExceptionByUser :one
UPDATE schedule_exceptions
SET
    exception_type = COALESCE(sqlc.narg(exception_type), exception_type),
    start_date = COALESCE(sqlc.narg(start_date), start_date),
    end_date = COALESCE(sqlc.narg(end_date), end_date),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING
    id,
    user_id,
    exception_type AS type,
    start_date,
    end_date,
    created_at,
    updated_at;

-- name: DeleteScheduleExceptionByUser :execrows
DELETE FROM schedule_exceptions
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: CountScheduleExceptionOverlaps :one
SELECT COUNT(*)
FROM schedule_exceptions se
WHERE se.user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(excluded_id)::uuid IS NULL
    OR se.id <> sqlc.narg(excluded_id)::uuid
  )
  AND daterange(se.start_date, se.end_date, '[]')
      && daterange(sqlc.arg(start_date)::date, sqlc.arg(end_date)::date, '[]');
