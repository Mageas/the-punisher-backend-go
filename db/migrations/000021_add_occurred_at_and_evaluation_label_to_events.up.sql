ALTER TABLE bonuses
ADD COLUMN occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN evaluation_label TEXT NOT NULL DEFAULT '';

UPDATE bonuses
SET occurred_at = created_at;

ALTER TABLE penalties
ADD COLUMN occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN evaluation_label TEXT NOT NULL DEFAULT '';

UPDATE penalties
SET occurred_at = created_at;

ALTER TABLE punishments
ADD COLUMN occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
ADD COLUMN evaluation_label TEXT NOT NULL DEFAULT '';

UPDATE punishments
SET occurred_at = created_at;
