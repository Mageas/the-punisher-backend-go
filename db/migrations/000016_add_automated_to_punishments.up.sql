ALTER TABLE punishments
ADD COLUMN automated BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE punishments
SET automated = TRUE
WHERE triggering_rule_id IS NOT NULL;
