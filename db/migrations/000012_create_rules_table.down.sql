DROP INDEX IF EXISTS idx_punishments_triggering_rule_id;

ALTER TABLE punishments
DROP CONSTRAINT IF EXISTS fk_punishments_triggering_rule_id;

DROP TABLE IF EXISTS rules;
