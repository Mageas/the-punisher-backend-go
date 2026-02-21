ALTER TABLE penalties
DROP CONSTRAINT IF EXISTS penalties_penalty_type_id_fkey;

ALTER TABLE penalties
ADD CONSTRAINT penalties_penalty_type_id_fkey
FOREIGN KEY (penalty_type_id)
REFERENCES penalty_types(id)
ON DELETE RESTRICT;

ALTER TABLE punishments
DROP CONSTRAINT IF EXISTS punishments_punishment_type_id_fkey;

ALTER TABLE punishments
ADD CONSTRAINT punishments_punishment_type_id_fkey
FOREIGN KEY (punishment_type_id)
REFERENCES punishment_types(id)
ON DELETE RESTRICT;
