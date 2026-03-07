ALTER TABLE students
ADD CONSTRAINT students_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE classrooms
ADD CONSTRAINT classrooms_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE bonus_types
ADD CONSTRAINT bonus_types_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE penalty_types
ADD CONSTRAINT penalty_types_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE punishment_types
ADD CONSTRAINT punishment_types_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE rules
ADD CONSTRAINT rules_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE schedule_slots
ADD CONSTRAINT schedule_slots_user_id_id_key UNIQUE (user_id, id);

ALTER TABLE student_classrooms
ADD COLUMN user_id UUID;

UPDATE student_classrooms sc
SET user_id = s.user_id
FROM students s
WHERE s.id = sc.student_id;

ALTER TABLE student_classrooms
ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE student_classrooms
DROP CONSTRAINT student_classrooms_pkey;

ALTER TABLE student_classrooms
ADD CONSTRAINT student_classrooms_pkey PRIMARY KEY (user_id, student_id, classroom_id);

CREATE INDEX idx_student_classrooms_user_student_id
    ON student_classrooms (user_id, student_id);

CREATE INDEX idx_student_classrooms_user_classroom_id
    ON student_classrooms (user_id, classroom_id);

ALTER TABLE schedule_slot_classrooms
ADD COLUMN user_id UUID;

UPDATE schedule_slot_classrooms ssc
SET user_id = s.user_id
FROM schedule_slots s
WHERE s.id = ssc.schedule_slot_id;

ALTER TABLE schedule_slot_classrooms
ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE schedule_slot_classrooms
DROP CONSTRAINT schedule_slot_classrooms_pkey;

ALTER TABLE schedule_slot_classrooms
ADD CONSTRAINT schedule_slot_classrooms_pkey PRIMARY KEY (user_id, schedule_slot_id, classroom_id);

CREATE INDEX idx_schedule_slot_classrooms_user_slot_id
    ON schedule_slot_classrooms (user_id, schedule_slot_id);

CREATE INDEX idx_schedule_slot_classrooms_user_classroom_id
    ON schedule_slot_classrooms (user_id, classroom_id);

ALTER TABLE bonuses
DROP CONSTRAINT IF EXISTS bonuses_student_id_fkey,
DROP CONSTRAINT IF EXISTS bonuses_bonus_type_id_fkey;

ALTER TABLE bonuses
ADD CONSTRAINT bonuses_student_user_id_fkey
FOREIGN KEY (user_id, student_id)
REFERENCES students(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT bonuses_bonus_type_user_id_fkey
FOREIGN KEY (user_id, bonus_type_id)
REFERENCES bonus_types(user_id, id)
ON DELETE CASCADE;

CREATE INDEX idx_bonuses_user_student_id
    ON bonuses (user_id, student_id);

CREATE INDEX idx_bonuses_user_bonus_type_id
    ON bonuses (user_id, bonus_type_id);

ALTER TABLE penalties
DROP CONSTRAINT IF EXISTS penalties_student_id_fkey,
DROP CONSTRAINT IF EXISTS penalties_penalty_type_id_fkey;

ALTER TABLE penalties
ADD CONSTRAINT penalties_student_user_id_fkey
FOREIGN KEY (user_id, student_id)
REFERENCES students(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT penalties_penalty_type_user_id_fkey
FOREIGN KEY (user_id, penalty_type_id)
REFERENCES penalty_types(user_id, id)
ON DELETE CASCADE;

CREATE INDEX idx_penalties_user_student_id
    ON penalties (user_id, student_id);

CREATE INDEX idx_penalties_user_penalty_type_id
    ON penalties (user_id, penalty_type_id);

ALTER TABLE rules
DROP CONSTRAINT IF EXISTS rules_resulting_punishment_type_id_fkey,
DROP CONSTRAINT IF EXISTS rules_penalty_type_id_fkey;

ALTER TABLE rules
ADD CONSTRAINT rules_resulting_punishment_type_user_id_fkey
FOREIGN KEY (user_id, resulting_punishment_type_id)
REFERENCES punishment_types(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT rules_penalty_type_user_id_fkey
FOREIGN KEY (user_id, penalty_type_id)
REFERENCES penalty_types(user_id, id)
ON DELETE CASCADE;

CREATE INDEX idx_rules_user_resulting_punishment_type_id
    ON rules (user_id, resulting_punishment_type_id);

ALTER TABLE punishments
DROP CONSTRAINT IF EXISTS punishments_student_id_fkey,
DROP CONSTRAINT IF EXISTS punishments_punishment_type_id_fkey,
DROP CONSTRAINT IF EXISTS fk_punishments_triggering_rule_id;

ALTER TABLE punishments
ADD CONSTRAINT punishments_student_user_id_fkey
FOREIGN KEY (user_id, student_id)
REFERENCES students(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT punishments_punishment_type_user_id_fkey
FOREIGN KEY (user_id, punishment_type_id)
REFERENCES punishment_types(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT punishments_triggering_rule_user_id_fkey
FOREIGN KEY (user_id, triggering_rule_id)
REFERENCES rules(user_id, id)
ON DELETE SET NULL (triggering_rule_id);

CREATE INDEX idx_punishments_user_student_id
    ON punishments (user_id, student_id);

CREATE INDEX idx_punishments_user_punishment_type_id
    ON punishments (user_id, punishment_type_id);

CREATE INDEX idx_punishments_user_triggering_rule_id
    ON punishments (user_id, triggering_rule_id);

ALTER TABLE student_classrooms
DROP CONSTRAINT IF EXISTS student_classrooms_student_id_fkey,
DROP CONSTRAINT IF EXISTS student_classrooms_classroom_id_fkey;

ALTER TABLE student_classrooms
ADD CONSTRAINT student_classrooms_student_user_id_fkey
FOREIGN KEY (user_id, student_id)
REFERENCES students(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT student_classrooms_classroom_user_id_fkey
FOREIGN KEY (user_id, classroom_id)
REFERENCES classrooms(user_id, id)
ON DELETE CASCADE;

ALTER TABLE schedule_slot_classrooms
DROP CONSTRAINT IF EXISTS schedule_slot_classrooms_schedule_slot_id_fkey,
DROP CONSTRAINT IF EXISTS schedule_slot_classrooms_classroom_id_fkey;

ALTER TABLE schedule_slot_classrooms
ADD CONSTRAINT schedule_slot_classrooms_slot_user_id_fkey
FOREIGN KEY (user_id, schedule_slot_id)
REFERENCES schedule_slots(user_id, id)
ON DELETE CASCADE,
ADD CONSTRAINT schedule_slot_classrooms_classroom_user_id_fkey
FOREIGN KEY (user_id, classroom_id)
REFERENCES classrooms(user_id, id)
ON DELETE CASCADE;
