CREATE TABLE bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    bonus_type_id UUID NOT NULL REFERENCES bonus_types(id) ON DELETE CASCADE,
    points DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

CREATE INDEX idx_bonuses_user_id ON bonuses(user_id);
CREATE INDEX idx_bonuses_student_id ON bonuses(student_id);
CREATE INDEX idx_bonuses_bonus_type_id ON bonuses(bonus_type_id);
