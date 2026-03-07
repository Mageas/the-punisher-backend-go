-- ==================== ScheduleSlot ====================

-- name: CreateScheduleSlot :one
INSERT INTO schedule_slots (
    user_id, weekday, start_time, end_time, week_pattern
) VALUES (
    sqlc.arg(user_id), sqlc.arg(weekday), sqlc.arg(start_time), sqlc.arg(end_time), sqlc.arg(week_pattern)
)
RETURNING
    id, user_id, weekday, start_time, end_time, week_pattern, created_at, updated_at;

-- name: GetScheduleSlotByUser :one
SELECT
    id, user_id, weekday, start_time, end_time, week_pattern, created_at, updated_at
FROM schedule_slots
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListScheduleSlotsByUser :many
SELECT
    id, user_id, weekday, start_time, end_time, week_pattern, created_at, updated_at
FROM schedule_slots
WHERE user_id = sqlc.arg(user_id)
ORDER BY weekday ASC, start_time ASC, id ASC;

-- name: UpdateScheduleSlotByUser :one
UPDATE schedule_slots
SET
    weekday = COALESCE(sqlc.narg(weekday), weekday),
    start_time = COALESCE(sqlc.narg(start_time), start_time),
    end_time = COALESCE(sqlc.narg(end_time), end_time),
    week_pattern = COALESCE(sqlc.narg(week_pattern), week_pattern),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING
    id, user_id, weekday, start_time, end_time, week_pattern, created_at, updated_at;

-- name: DeleteScheduleSlotByUser :execrows
DELETE FROM schedule_slots
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: CountScheduleSlotConflicts :one
SELECT COUNT(*)
FROM schedule_slots s
WHERE s.user_id = sqlc.arg(user_id)
  AND s.weekday = sqlc.arg(weekday)
  AND (
    sqlc.narg(excluded_id)::uuid IS NULL
    OR s.id <> sqlc.narg(excluded_id)::uuid
  )
  AND sqlc.arg(start_time)::time < s.end_time::time
  AND s.start_time::time < sqlc.arg(end_time)::time
  AND (
    s.week_pattern = 'every_week'
    OR sqlc.arg(week_pattern)::text = 'every_week'
    OR s.week_pattern = sqlc.arg(week_pattern)::text
  );

-- name: CreateScheduleSlotClassroomRelation :execrows
INSERT INTO schedule_slot_classrooms (
    user_id, schedule_slot_id, classroom_id
)
SELECT
    sqlc.arg(user_id), sqlc.arg(schedule_slot_id), sqlc.arg(classroom_id)
WHERE EXISTS (
    SELECT 1
    FROM schedule_slots s
    WHERE s.id = sqlc.arg(schedule_slot_id)
      AND s.user_id = sqlc.arg(user_id)
)
  AND EXISTS (
    SELECT 1
    FROM classrooms c
    WHERE c.id = sqlc.arg(classroom_id)
      AND c.user_id = sqlc.arg(user_id)
)
ON CONFLICT DO NOTHING;

-- name: DeleteScheduleSlotClassroomRelationsBySlot :execrows
DELETE FROM schedule_slot_classrooms
WHERE user_id = sqlc.arg(user_id)
  AND schedule_slot_id = sqlc.arg(schedule_slot_id);

-- name: ListScheduleSlotClassroomRefsBySlotIDs :many
SELECT
    ssc.schedule_slot_id,
    c.id AS classroom_id,
    c.name AS classroom_name
FROM schedule_slot_classrooms ssc
JOIN schedule_slots s ON s.id = ssc.schedule_slot_id AND s.user_id = ssc.user_id
JOIN classrooms c ON c.id = ssc.classroom_id AND c.user_id = ssc.user_id
WHERE ssc.user_id = sqlc.arg(user_id)
  AND ssc.schedule_slot_id = ANY(sqlc.arg(schedule_slot_ids)::uuid[])
ORDER BY ssc.schedule_slot_id ASC, c.name ASC, c.id ASC;

-- name: CountClassroomsByIDsAndUser :one
SELECT COUNT(*)
FROM classrooms
WHERE user_id = sqlc.arg(user_id)
  AND id = ANY(sqlc.arg(classroom_ids)::uuid[]);

-- name: ListScheduleSlotsByClassroom :many
SELECT
    s.id, s.user_id, s.weekday, s.start_time, s.end_time, s.week_pattern, s.created_at, s.updated_at
FROM schedule_slots s
JOIN schedule_slot_classrooms ssc ON ssc.schedule_slot_id = s.id AND ssc.user_id = s.user_id
WHERE s.user_id = sqlc.arg(user_id)
  AND ssc.classroom_id = sqlc.arg(classroom_id)
ORDER BY s.weekday ASC, s.start_time ASC, s.id ASC;

-- name: DeleteOrphanScheduleSlotsByUser :execrows
DELETE FROM schedule_slots s
WHERE s.user_id = sqlc.arg(user_id)
  AND NOT EXISTS (
    SELECT 1
    FROM schedule_slot_classrooms ssc
    WHERE ssc.schedule_slot_id = s.id
      AND ssc.user_id = s.user_id
  );
