CREATE TABLE rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR NOT NULL,
    resulting_punishment_type_id UUID NOT NULL REFERENCES punishment_types(id) ON DELETE RESTRICT,
    penalty_type_id UUID NOT NULL REFERENCES penalty_types(id) ON DELETE CASCADE,
    threshold INTEGER NOT NULL CHECK (threshold >= 1),
    mode VARCHAR NOT NULL CHECK (mode IN ('after', 'at', 'every')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rules_user_id ON rules(user_id);
CREATE INDEX idx_rules_penalty_type_id ON rules(penalty_type_id);
CREATE INDEX idx_rules_resulting_punishment_type_id ON rules(resulting_punishment_type_id);
CREATE INDEX idx_rules_user_penalty_type_active ON rules(user_id, penalty_type_id, is_active);

ALTER TABLE punishments
ADD CONSTRAINT fk_punishments_triggering_rule_id
FOREIGN KEY (triggering_rule_id) REFERENCES rules(id) ON DELETE SET NULL;

CREATE INDEX idx_punishments_triggering_rule_id ON punishments(triggering_rule_id);
