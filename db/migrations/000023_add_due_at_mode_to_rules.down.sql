ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_due_at_configuration_check;

ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_due_at_mode_check;

ALTER TABLE rules
DROP COLUMN IF EXISTS due_at_after_lessons,
DROP COLUMN IF EXISTS due_at_mode;
