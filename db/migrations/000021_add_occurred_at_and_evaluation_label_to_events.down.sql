ALTER TABLE bonuses
DROP COLUMN IF EXISTS evaluation_label,
DROP COLUMN IF EXISTS occurred_at;

ALTER TABLE penalties
DROP COLUMN IF EXISTS evaluation_label,
DROP COLUMN IF EXISTS occurred_at;

ALTER TABLE punishments
DROP COLUMN IF EXISTS evaluation_label,
DROP COLUMN IF EXISTS occurred_at;
