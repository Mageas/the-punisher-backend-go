ALTER TABLE bonuses
ADD COLUMN occurred_at TIMESTAMPTZ,
ADD COLUMN evaluation_label TEXT;

UPDATE bonuses
SET occurred_at = created_at;

ALTER TABLE bonuses
ALTER COLUMN occurred_at SET NOT NULL,
ALTER COLUMN occurred_at SET DEFAULT NOW();

ALTER TABLE penalties
ADD COLUMN occurred_at TIMESTAMPTZ,
ADD COLUMN evaluation_label TEXT;

UPDATE penalties
SET occurred_at = created_at;

ALTER TABLE penalties
ALTER COLUMN occurred_at SET NOT NULL,
ALTER COLUMN occurred_at SET DEFAULT NOW();

ALTER TABLE punishments
ADD COLUMN occurred_at TIMESTAMPTZ,
ADD COLUMN evaluation_label TEXT;

UPDATE punishments
SET occurred_at = created_at;

ALTER TABLE punishments
ALTER COLUMN occurred_at SET NOT NULL,
ALTER COLUMN occurred_at SET DEFAULT NOW();
