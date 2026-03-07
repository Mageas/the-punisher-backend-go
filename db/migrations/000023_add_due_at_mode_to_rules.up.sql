ALTER TABLE rules
ADD COLUMN due_at_mode TEXT NOT NULL DEFAULT 'days',
ADD COLUMN due_at_after_lessons INTEGER;

ALTER TABLE rules
ADD CONSTRAINT rules_due_at_mode_check
CHECK (due_at_mode IN ('days', 'next_lessons'));

ALTER TABLE rules
ADD CONSTRAINT rules_due_at_configuration_check
CHECK (
    (due_at_mode = 'days' AND due_at_after_lessons IS NULL)
    OR
    (due_at_mode = 'next_lessons' AND due_at_after_days = 0 AND due_at_after_lessons BETWEEN 1 AND 5)
);
