CREATE TABLE punishments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    punishment_type_id UUID NOT NULL REFERENCES punishment_types(id) ON DELETE RESTRICT,
    triggering_rule_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_punishments_user_id ON punishments(user_id);
CREATE INDEX idx_punishments_student_id ON punishments(student_id);
CREATE INDEX idx_punishments_punishment_type_id ON punishments(punishment_type_id);
CREATE INDEX idx_punishments_resolved_at ON punishments(resolved_at);
