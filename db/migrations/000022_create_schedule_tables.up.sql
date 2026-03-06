CREATE TABLE schedule_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weekday INTEGER NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    start_time VARCHAR(5) NOT NULL CHECK (start_time ~ '^(?:[01][0-9]|2[0-3]):[0-5][0-9]$'),
    end_time VARCHAR(5) NOT NULL CHECK (end_time ~ '^(?:[01][0-9]|2[0-3]):[0-5][0-9]$'),
    week_pattern VARCHAR NOT NULL CHECK (week_pattern IN ('every_week', 'even_weeks', 'odd_weeks')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT schedule_slots_time_range CHECK (end_time::time > start_time::time)
);

CREATE INDEX idx_schedule_slots_user_weekday
    ON schedule_slots (user_id, weekday, start_time);

CREATE TABLE schedule_slot_classrooms (
    schedule_slot_id UUID NOT NULL REFERENCES schedule_slots(id) ON DELETE CASCADE,
    classroom_id UUID NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,
    PRIMARY KEY (schedule_slot_id, classroom_id)
);

CREATE INDEX idx_schedule_slot_classrooms_classroom_id
    ON schedule_slot_classrooms (classroom_id);

CREATE TABLE schedule_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exception_type VARCHAR NOT NULL CHECK (exception_type IN ('vacation', 'public_holiday')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT schedule_exceptions_date_range CHECK (end_date >= start_date)
);

CREATE INDEX idx_schedule_exceptions_user_dates
    ON schedule_exceptions (user_id, start_date, end_date);
