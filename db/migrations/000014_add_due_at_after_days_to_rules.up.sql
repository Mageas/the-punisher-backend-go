ALTER TABLE rules
ADD COLUMN due_at_after_days INTEGER NOT NULL DEFAULT 0 CHECK (due_at_after_days >= 0);
