ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_due_at_configuration_check;

UPDATE rules
SET due_at_after_days = 0
WHERE due_at_after_days IS NULL;

ALTER TABLE rules
ALTER COLUMN due_at_after_days SET DEFAULT 0,
ALTER COLUMN due_at_after_days SET NOT NULL;

ALTER TABLE rules
ADD CONSTRAINT rules_due_at_configuration_check
CHECK (
    (due_at_mode = 'days' AND due_at_after_lessons IS NULL)
    OR
    (due_at_mode = 'next_lessons' AND due_at_after_days = 0 AND due_at_after_lessons BETWEEN 1 AND 5)
);
