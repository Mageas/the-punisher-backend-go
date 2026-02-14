ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_resulting_punishment_type_id_fkey;

ALTER TABLE rules
ADD CONSTRAINT rules_resulting_punishment_type_id_fkey
FOREIGN KEY (resulting_punishment_type_id)
REFERENCES punishment_types(id)
ON DELETE RESTRICT;
