ALTER TABLE schedule_slot_classrooms
DROP CONSTRAINT IF EXISTS schedule_slot_classrooms_slot_user_id_fkey,
DROP CONSTRAINT IF EXISTS schedule_slot_classrooms_classroom_user_id_fkey;

ALTER TABLE schedule_slot_classrooms
ADD CONSTRAINT schedule_slot_classrooms_schedule_slot_id_fkey
FOREIGN KEY (schedule_slot_id)
REFERENCES schedule_slots(id)
ON DELETE CASCADE,
ADD CONSTRAINT schedule_slot_classrooms_classroom_id_fkey
FOREIGN KEY (classroom_id)
REFERENCES classrooms(id)
ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_schedule_slot_classrooms_user_slot_id;
DROP INDEX IF EXISTS idx_schedule_slot_classrooms_user_classroom_id;

ALTER TABLE schedule_slot_classrooms
DROP CONSTRAINT IF EXISTS schedule_slot_classrooms_pkey;

ALTER TABLE schedule_slot_classrooms
ADD CONSTRAINT schedule_slot_classrooms_pkey PRIMARY KEY (schedule_slot_id, classroom_id);

ALTER TABLE schedule_slot_classrooms
DROP COLUMN IF EXISTS user_id;

ALTER TABLE student_classrooms
DROP CONSTRAINT IF EXISTS student_classrooms_student_user_id_fkey,
DROP CONSTRAINT IF EXISTS student_classrooms_classroom_user_id_fkey;

ALTER TABLE student_classrooms
ADD CONSTRAINT student_classrooms_student_id_fkey
FOREIGN KEY (student_id)
REFERENCES students(id)
ON DELETE CASCADE,
ADD CONSTRAINT student_classrooms_classroom_id_fkey
FOREIGN KEY (classroom_id)
REFERENCES classrooms(id)
ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_student_classrooms_user_student_id;
DROP INDEX IF EXISTS idx_student_classrooms_user_classroom_id;

ALTER TABLE student_classrooms
DROP CONSTRAINT IF EXISTS student_classrooms_pkey;

ALTER TABLE student_classrooms
ADD CONSTRAINT student_classrooms_pkey PRIMARY KEY (student_id, classroom_id);

ALTER TABLE student_classrooms
DROP COLUMN IF EXISTS user_id;

ALTER TABLE punishments
DROP CONSTRAINT IF EXISTS punishments_student_user_id_fkey,
DROP CONSTRAINT IF EXISTS punishments_punishment_type_user_id_fkey,
DROP CONSTRAINT IF EXISTS punishments_triggering_rule_user_id_fkey;

DROP INDEX IF EXISTS idx_punishments_user_student_id;
DROP INDEX IF EXISTS idx_punishments_user_punishment_type_id;
DROP INDEX IF EXISTS idx_punishments_user_triggering_rule_id;

ALTER TABLE punishments
ADD CONSTRAINT punishments_student_id_fkey
FOREIGN KEY (student_id)
REFERENCES students(id)
ON DELETE CASCADE,
ADD CONSTRAINT punishments_punishment_type_id_fkey
FOREIGN KEY (punishment_type_id)
REFERENCES punishment_types(id)
ON DELETE CASCADE,
ADD CONSTRAINT fk_punishments_triggering_rule_id
FOREIGN KEY (triggering_rule_id)
REFERENCES rules(id)
ON DELETE SET NULL;

ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_resulting_punishment_type_user_id_fkey,
DROP CONSTRAINT IF EXISTS rules_penalty_type_user_id_fkey;

DROP INDEX IF EXISTS idx_rules_user_resulting_punishment_type_id;

ALTER TABLE rules
ADD CONSTRAINT rules_resulting_punishment_type_id_fkey
FOREIGN KEY (resulting_punishment_type_id)
REFERENCES punishment_types(id)
ON DELETE CASCADE,
ADD CONSTRAINT rules_penalty_type_id_fkey
FOREIGN KEY (penalty_type_id)
REFERENCES penalty_types(id)
ON DELETE CASCADE;

ALTER TABLE penalties
DROP CONSTRAINT IF EXISTS penalties_student_user_id_fkey,
DROP CONSTRAINT IF EXISTS penalties_penalty_type_user_id_fkey;

DROP INDEX IF EXISTS idx_penalties_user_student_id;
DROP INDEX IF EXISTS idx_penalties_user_penalty_type_id;

ALTER TABLE penalties
ADD CONSTRAINT penalties_student_id_fkey
FOREIGN KEY (student_id)
REFERENCES students(id)
ON DELETE CASCADE,
ADD CONSTRAINT penalties_penalty_type_id_fkey
FOREIGN KEY (penalty_type_id)
REFERENCES penalty_types(id)
ON DELETE CASCADE;

ALTER TABLE bonuses
DROP CONSTRAINT IF EXISTS bonuses_student_user_id_fkey,
DROP CONSTRAINT IF EXISTS bonuses_bonus_type_user_id_fkey;

DROP INDEX IF EXISTS idx_bonuses_user_student_id;
DROP INDEX IF EXISTS idx_bonuses_user_bonus_type_id;

ALTER TABLE bonuses
ADD CONSTRAINT bonuses_student_id_fkey
FOREIGN KEY (student_id)
REFERENCES students(id)
ON DELETE CASCADE,
ADD CONSTRAINT bonuses_bonus_type_id_fkey
FOREIGN KEY (bonus_type_id)
REFERENCES bonus_types(id)
ON DELETE CASCADE;

ALTER TABLE schedule_slots
DROP CONSTRAINT IF EXISTS schedule_slots_user_id_id_key;

ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_user_id_id_key;

ALTER TABLE punishment_types
DROP CONSTRAINT IF EXISTS punishment_types_user_id_id_key;

ALTER TABLE penalty_types
DROP CONSTRAINT IF EXISTS penalty_types_user_id_id_key;

ALTER TABLE bonus_types
DROP CONSTRAINT IF EXISTS bonus_types_user_id_id_key;

ALTER TABLE classrooms
DROP CONSTRAINT IF EXISTS classrooms_user_id_id_key;

ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_user_id_id_key;
