# Fonctionnalités à Implémenter

Ce document liste les fonctionnalités définies dans le projet mais non encore implémentées.

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Entités manquantes](#entités-manquantes)
- [Priorité d'implémentation](#priorité-dimplémentation)
- [Détails des fonctionnalités](#détails-des-fonctionnalités)

---

## Vue d'ensemble

### Fonctionnalités implémentées ✅

- **Users** : Authentification et création de compte (partiellement)
- **Students** : CRUD complet
- **Classrooms** : CRUD complet avec relations many-to-many
- **BonusTypes** : CRUD complet

### Fonctionnalités manquantes ❌

- **PenaltyTypes** : Types de pénalités (ex: "Bavardage", "Retard", "Oubli de matériel")
- **PunishmentTypes** : Types de punitions (ex: "Retenue", "Mot aux parents")
- **Rules** : Règles automatiques de déclenchement des punitions
- **Bonuses** : Instances de bonus attribuées aux étudiants
- **Penalties** : Instances de pénalités enregistrées
- **Punishments** : Instances de punitions assignées
- **Users** : Endpoints complets (update, delete, liste)

---

## Entités manquantes

### 1. PenaltyTypes (Types de Pénalités)

**Description :**
Les types de pénalités définissent les catégories d'incidents négatifs qu'un professeur peut enregistrer.

**Structure de la table :**
```sql
CREATE TABLE penalty_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Exemples de données :**
- "Bavardage"
- "Retard"
- "Oubli de matériel"
- "Insolence"
- "Devoirs non faits"

**Endpoints à implémenter :**
- `POST /v1/penalty-types` : Créer un type de pénalité
- `GET /v1/penalty-types` : Liste paginée
- `GET /v1/penalty-types/{id}` : Récupérer un type
- `PUT /v1/penalty-types/{id}` : Mettre à jour un type
- `DELETE /v1/penalty-types/{id}` : Supprimer un type

**Référence :**
Suivre le même pattern que `BonusType`.

---

### 2. PunishmentTypes (Types de Punitions)

**Description :**
Les types de punitions définissent les sanctions qu'un professeur peut appliquer.

**Structure de la table :**
```sql
CREATE TABLE punishment_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Exemples de données :**
- "Retenue (1h)"
- "Retenue (2h)"
- "Mot aux parents"
- "Exclusion de cours"
- "Copie de lignes (50)"
- "Copie de lignes (100)"

**Endpoints à implémenter :**
- `POST /v1/punishment-types` : Créer un type de punition
- `GET /v1/punishment-types` : Liste paginée
- `GET /v1/punishment-types/{id}` : Récupérer un type
- `PUT /v1/punishment-types/{id}` : Mettre à jour un type
- `DELETE /v1/punishment-types/{id}` : Supprimer un type

**Référence :**
Suivre le même pattern que `BonusType`.

---

### 3. Bonuses (Instances de Bonus)

**Description :**
Les bonus sont des points attribués aux étudiants pour récompenser un comportement positif.

**Structure de la table :**
```sql
CREATE TABLE bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    bonus_type_id UUID NOT NULL REFERENCES bonus_types(id) ON DELETE CASCADE,
    points INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP NULL,  -- NULL = non consommé, sinon date de consommation
    
    CONSTRAINT fk_bonuses_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_bonuses_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_bonuses_bonus_type FOREIGN KEY (bonus_type_id) REFERENCES bonus_types(id) ON DELETE CASCADE
);

CREATE INDEX idx_bonuses_user_id ON bonuses(user_id);
CREATE INDEX idx_bonuses_student_id ON bonuses(student_id);
CREATE INDEX idx_bonuses_created_at ON bonuses(created_at DESC);
CREATE INDEX idx_bonuses_used_at ON bonuses(used_at);
```

**Cycle de vie :**
1. **Création** (`created_at`) : Le bonus est attribué à un étudiant
2. **Utilisation** (`used_at`) : Le bonus est consommé (ex: pour +0.5 sur une note)

**DTOs :**
```go
type RequestBonusDto struct {
    StudentID   string `json:"student_id" validate:"required,uuid"`
    BonusTypeID string `json:"bonus_type_id" validate:"required,uuid"`
    Points      int    `json:"points" validate:"required,min=1,max=100"`
}

type UpdateBonusDto struct {
    BonusTypeID *string `json:"bonus_type_id" validate:"omitempty,uuid"`
    Points      *int    `json:"points" validate:"omitempty,min=1,max=100"`
}

type ReturnBonusDto struct {
    ID          uuid.UUID  `json:"id"`
    StudentID   uuid.UUID  `json:"student_id"`
    BonusTypeID uuid.UUID  `json:"bonus_type_id"`
    Points      int        `json:"points"`
    CreatedAt   time.Time  `json:"created_at"`
    UsedAt      *time.Time `json:"used_at"`
}
```

**Endpoints à implémenter :**
- `POST /v1/bonuses` : Attribuer un bonus
- `GET /v1/bonuses` : Liste paginée des bonus de l'utilisateur
- `GET /v1/bonuses/{id}` : Récupérer un bonus
- `PUT /v1/bonuses/{id}` : Mettre à jour un bonus
- `DELETE /v1/bonuses/{id}` : Supprimer un bonus
- `PATCH /v1/bonuses/{id}/use` : Marquer un bonus comme utilisé
- `GET /v1/students/{id}/bonuses` : Liste paginée des bonus d'un étudiant
- `GET /v1/students/{id}/bonuses/available` : Liste des bonus non utilisés d'un étudiant
- `GET /v1/students/{id}/bonuses/summary` : Résumé des bonus (total points, points disponibles)

**Logique métier spécifique :**
```go
// Marquer un bonus comme utilisé
func (s *bonusService) UseBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error) {
    // Vérifier que le bonus existe et n'est pas déjà utilisé
    bonus, err := s.repo.GetBonus(ctx, repository.GetBonusParams{
        ID:     bonusID,
        UserID: userID,
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, api.ErrBonusNotFound
        }
        return nil, fmt.Errorf("failed to get bonus: %w", err)
    }

    if bonus.UsedAt.Valid {
        return nil, api.ErrBonusAlreadyUsed
    }

    // Marquer comme utilisé
    updatedBonus, err := s.repo.MarkBonusAsUsed(ctx, bonusID)
    if err != nil {
        return nil, fmt.Errorf("failed to mark bonus as used: %w", err)
    }

    slog.Info("bonus used", "bonus_id", bonusID, "student_id", bonus.StudentID, "user_id", userID)

    return dto.BonusFromRepository(&updatedBonus), nil
}

// Résumé des bonus d'un étudiant
func (s *bonusService) GetBonusSummary(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.BonusSummaryDto, error) {
    // Total points
    totalPoints, err := s.repo.SumBonusPointsByStudent(ctx, repository.SumBonusPointsByStudentParams{
        StudentID: studentID,
        UserID:    userID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to sum bonus points: %w", err)
    }

    // Points disponibles (non utilisés)
    availablePoints, err := s.repo.SumAvailableBonusPointsByStudent(ctx, repository.SumAvailableBonusPointsByStudentParams{
        StudentID: studentID,
        UserID:    userID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to sum available bonus points: %w", err)
    }

    return &dto.BonusSummaryDto{
        StudentID:        studentID,
        TotalPoints:      totalPoints,
        AvailablePoints:  availablePoints,
        UsedPoints:       totalPoints - availablePoints,
    }, nil
}
```

**Requêtes SQL supplémentaires nécessaires :**
```sql
-- name: MarkBonusAsUsed :one
UPDATE bonuses
SET used_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND used_at IS NULL
RETURNING id, user_id, student_id, bonus_type_id, points, created_at, used_at;

-- name: SumBonusPointsByStudent :one
SELECT COALESCE(SUM(points), 0) as total_points
FROM bonuses b
JOIN students s ON s.id = b.student_id
WHERE b.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id);

-- name: SumAvailableBonusPointsByStudent :one
SELECT COALESCE(SUM(points), 0) as available_points
FROM bonuses b
JOIN students s ON s.id = b.student_id
WHERE b.student_id = sqlc.arg(student_id) 
  AND s.user_id = sqlc.arg(user_id) 
  AND b.used_at IS NULL;

-- name: ListAvailableBonusesByStudent :many
SELECT b.id, b.user_id, b.student_id, b.bonus_type_id, b.points, b.created_at, b.used_at
FROM bonuses b
JOIN students s ON s.id = b.student_id
WHERE b.student_id = sqlc.arg(student_id) 
  AND s.user_id = sqlc.arg(user_id) 
  AND b.used_at IS NULL
ORDER BY b.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);
```

---

### 4. Penalties (Instances de Pénalités)

**Description :**
Les pénalités sont des incidents négatifs enregistrés pour un étudiant.

**Structure de la table :**
```sql
CREATE TABLE penalties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    penalty_type_id UUID NOT NULL REFERENCES penalty_types(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_penalties_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_penalties_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_penalties_penalty_type FOREIGN KEY (penalty_type_id) REFERENCES penalty_types(id) ON DELETE CASCADE
);

CREATE INDEX idx_penalties_user_id ON penalties(user_id);
CREATE INDEX idx_penalties_student_id ON penalties(student_id);
CREATE INDEX idx_penalties_penalty_type_id ON penalties(penalty_type_id);
CREATE INDEX idx_penalties_created_at ON penalties(created_at DESC);
```

**DTOs :**
```go
type RequestPenaltyDto struct {
    StudentID     string `json:"student_id" validate:"required,uuid"`
    PenaltyTypeID string `json:"penalty_type_id" validate:"required,uuid"`
}

type UpdatePenaltyDto struct {
    PenaltyTypeID *string `json:"penalty_type_id" validate:"omitempty,uuid"`
}

type ReturnPenaltyDto struct {
    ID            uuid.UUID `json:"id"`
    StudentID     uuid.UUID `json:"student_id"`
    PenaltyTypeID uuid.UUID `json:"penalty_type_id"`
    CreatedAt     time.Time `json:"created_at"`
}
```

**Endpoints à implémenter :**
- `POST /v1/penalties` : Enregistrer une pénalité
- `GET /v1/penalties` : Liste paginée des pénalités de l'utilisateur
- `GET /v1/penalties/{id}` : Récupérer une pénalité
- `PUT /v1/penalties/{id}` : Mettre à jour une pénalité
- `DELETE /v1/penalties/{id}` : Supprimer une pénalité
- `GET /v1/students/{id}/penalties` : Liste paginée des pénalités d'un étudiant
- `GET /v1/students/{id}/penalties/count` : Nombre total de pénalités par type
- `GET /v1/students/{id}/penalties/count/{penaltyTypeId}` : Nombre de pénalités d'un type spécifique

**Requêtes SQL supplémentaires :**
```sql
-- name: CountPenaltiesByStudentAndType :one
SELECT COUNT(*)
FROM penalties p
JOIN students s ON s.id = p.student_id
WHERE p.student_id = sqlc.arg(student_id) 
  AND p.penalty_type_id = sqlc.arg(penalty_type_id)
  AND s.user_id = sqlc.arg(user_id);

-- name: CountPenaltiesByStudentGroupedByType :many
SELECT 
    pt.id as penalty_type_id,
    pt.name as penalty_type_name,
    COUNT(p.id) as count
FROM penalty_types pt
LEFT JOIN penalties p ON p.penalty_type_id = pt.id AND p.student_id = sqlc.arg(student_id)
JOIN students s ON s.id = sqlc.arg(student_id)
WHERE pt.user_id = sqlc.arg(user_id) AND s.user_id = sqlc.arg(user_id)
GROUP BY pt.id, pt.name
ORDER BY count DESC, pt.name;
```

**Note importante :**
Les compteurs de pénalités sont utilisés par les règles automatiques pour déclencher des punitions.

---

### 5. Punishments (Instances de Punitions)

**Description :**
Les punitions sont des sanctions assignées aux étudiants, soit manuellement soit automatiquement via les règles.

**Structure de la table :**
```sql
CREATE TABLE punishments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    punishment_type_id UUID NOT NULL REFERENCES punishment_types(id) ON DELETE CASCADE,
    triggering_rule_id UUID NULL REFERENCES rules(id) ON DELETE SET NULL,  -- NULL si manuel
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    due_at TIMESTAMP NULL,  -- Date limite pour effectuer la punition
    resolved_at TIMESTAMP NULL,  -- NULL = en attente, sinon date de résolution
    
    CONSTRAINT fk_punishments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_punishments_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_punishments_punishment_type FOREIGN KEY (punishment_type_id) REFERENCES punishment_types(id) ON DELETE CASCADE,
    CONSTRAINT fk_punishments_triggering_rule FOREIGN KEY (triggering_rule_id) REFERENCES rules(id) ON DELETE SET NULL
);

CREATE INDEX idx_punishments_user_id ON punishments(user_id);
CREATE INDEX idx_punishments_student_id ON punishments(student_id);
CREATE INDEX idx_punishments_created_at ON punishments(created_at DESC);
CREATE INDEX idx_punishments_resolved_at ON punishments(resolved_at);
CREATE INDEX idx_punishments_triggering_rule_id ON punishments(triggering_rule_id);
```

**DTOs :**
```go
type RequestPunishmentDto struct {
    StudentID         string     `json:"student_id" validate:"required,uuid"`
    PunishmentTypeID  string     `json:"punishment_type_id" validate:"required,uuid"`
    DueAt             *time.Time `json:"due_at" validate:"omitempty"`
}

type UpdatePunishmentDto struct {
    PunishmentTypeID *string    `json:"punishment_type_id" validate:"omitempty,uuid"`
    DueAt            *time.Time `json:"due_at" validate:"omitempty"`
}

type ReturnPunishmentDto struct {
    ID                 uuid.UUID  `json:"id"`
    StudentID          uuid.UUID  `json:"student_id"`
    PunishmentTypeID   uuid.UUID  `json:"punishment_type_id"`
    TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
    CreatedAt          time.Time  `json:"created_at"`
    DueAt              *time.Time `json:"due_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    Status             string     `json:"status"`  // "pending" ou "resolved"
}
```

**Endpoints à implémenter :**
- `POST /v1/punishments` : Créer une punition manuelle
- `GET /v1/punishments` : Liste paginée de toutes les punitions
- `GET /v1/punishments?status=pending` : Liste des punitions en attente
- `GET /v1/punishments?status=resolved` : Liste des punitions résolues
- `GET /v1/punishments/{id}` : Récupérer une punition
- `PUT /v1/punishments/{id}` : Mettre à jour une punition
- `DELETE /v1/punishments/{id}` : Supprimer une punition
- `PATCH /v1/punishments/{id}/resolve` : Marquer une punition comme résolue
- `GET /v1/students/{id}/punishments` : Liste paginée des punitions d'un étudiant
- `GET /v1/students/{id}/punishments/pending` : Liste des punitions en attente

**Logique métier spécifique :**
```go
// Marquer une punition comme résolue
func (s *punishmentService) ResolvePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error) {
    punishment, err := s.repo.GetPunishment(ctx, repository.GetPunishmentParams{
        ID:     punishmentID,
        UserID: userID,
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, api.ErrPunishmentNotFound
        }
        return nil, fmt.Errorf("failed to get punishment: %w", err)
    }

    if punishment.ResolvedAt.Valid {
        return nil, api.ErrPunishmentAlreadyResolved
    }

    updatedPunishment, err := s.repo.MarkPunishmentAsResolved(ctx, punishmentID)
    if err != nil {
        return nil, fmt.Errorf("failed to mark punishment as resolved: %w", err)
    }

    slog.Info("punishment resolved", "punishment_id", punishmentID, "student_id", punishment.StudentID, "user_id", userID)

    return dto.PunishmentFromRepository(&updatedPunishment), nil
}
```

**Requêtes SQL supplémentaires :**
```sql
-- name: MarkPunishmentAsResolved :one
UPDATE punishments
SET resolved_at = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND resolved_at IS NULL
RETURNING id, user_id, student_id, punishment_type_id, triggering_rule_id, created_at, due_at, resolved_at;

-- name: ListPendingPunishmentsByUser :many
SELECT id, user_id, student_id, punishment_type_id, triggering_rule_id, created_at, due_at, resolved_at
FROM punishments
WHERE user_id = sqlc.arg(user_id) AND resolved_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountPendingPunishmentsByUser :one
SELECT COUNT(*) FROM punishments WHERE user_id = sqlc.arg(user_id) AND resolved_at IS NULL;

-- name: ListPendingPunishmentsByStudent :many
SELECT p.id, p.user_id, p.student_id, p.punishment_type_id, p.triggering_rule_id, p.created_at, p.due_at, p.resolved_at
FROM punishments p
JOIN students s ON s.id = p.student_id
WHERE p.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id) AND p.resolved_at IS NULL
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);
```

---

### 6. Rules (Règles Automatiques)

**Description :**
Les règles permettent de déclencher automatiquement des punitions lorsque certaines conditions sont remplies.

**Structure de la table :**
```sql
CREATE TABLE rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    resulting_punishment_type_id UUID NOT NULL REFERENCES punishment_types(id) ON DELETE CASCADE,
    conditions JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_rules_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_rules_punishment_type FOREIGN KEY (resulting_punishment_type_id) REFERENCES punishment_types(id) ON DELETE CASCADE
);

CREATE INDEX idx_rules_user_id ON rules(user_id);
CREATE INDEX idx_rules_is_active ON rules(is_active);
CREATE INDEX idx_rules_conditions ON rules USING GIN (conditions);
```

**Structure du JSON `conditions` :**
```json
{
  "operator": "OR",
  "triggers": [
    {
      "type": "penalty_count",
      "penalty_type_id": "uuid-oubli-materiel",
      "threshold": 3
    },
    {
      "operator": "AND",
      "triggers": [
        {
          "type": "penalty_count",
          "penalty_type_id": "uuid-bavardage",
          "threshold": 2
        },
        {
          "type": "penalty_count",
          "penalty_type_id": "uuid-insolence",
          "threshold": 1
        }
      ]
    }
  ]
}
```

**Logique :**
- `operator` : "AND" ou "OR"
- `triggers` : Liste de conditions
- Chaque condition peut être :
  - Simple : `{ "type": "penalty_count", "penalty_type_id": "...", "threshold": N }`
  - Composée : `{ "operator": "AND/OR", "triggers": [...] }`

**DTOs :**
```go
type RequestRuleDto struct {
    Name                       string          `json:"name" validate:"required,min=2,max=200"`
    ResultingPunishmentTypeID  string          `json:"resulting_punishment_type_id" validate:"required,uuid"`
    Conditions                 json.RawMessage `json:"conditions" validate:"required"`
    IsActive                   bool            `json:"is_active"`
}

type UpdateRuleDto struct {
    Name                       *string         `json:"name" validate:"omitempty,min=2,max=200"`
    ResultingPunishmentTypeID  *string         `json:"resulting_punishment_type_id" validate:"omitempty,uuid"`
    Conditions                 json.RawMessage `json:"conditions" validate:"omitempty"`
    IsActive                   *bool           `json:"is_active"`
}

type ReturnRuleDto struct {
    ID                        uuid.UUID       `json:"id"`
    Name                      string          `json:"name"`
    ResultingPunishmentTypeID uuid.UUID       `json:"resulting_punishment_type_id"`
    Conditions                json.RawMessage `json:"conditions"`
    IsActive                  bool            `json:"is_active"`
    CreatedAt                 time.Time       `json:"created_at"`
    UpdatedAt                 time.Time       `json:"updated_at"`
}
```

**Endpoints à implémenter :**
- `POST /v1/rules` : Créer une règle
- `GET /v1/rules` : Liste paginée des règles
- `GET /v1/rules/{id}` : Récupérer une règle
- `PUT /v1/rules/{id}` : Mettre à jour une règle
- `DELETE /v1/rules/{id}` : Supprimer une règle
- `PATCH /v1/rules/{id}/activate` : Activer une règle
- `PATCH /v1/rules/{id}/deactivate` : Désactiver une règle
- `POST /v1/rules/{id}/evaluate/{studentId}` : Évaluer une règle pour un étudiant spécifique
- `POST /v1/penalties/{id}/evaluate-rules` : Évaluer toutes les règles après ajout d'une pénalité

**Logique d'évaluation (complexe) :**

```go
// Évaluer une règle pour un étudiant
func (s *ruleService) EvaluateRule(ctx context.Context, userID uuid.UUID, ruleID uuid.UUID, studentID uuid.UUID) (bool, error) {
    rule, err := s.repo.GetRule(ctx, repository.GetRuleParams{
        ID:     ruleID,
        UserID: userID,
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return false, api.ErrRuleNotFound
        }
        return false, fmt.Errorf("failed to get rule: %w", err)
    }

    if !rule.IsActive {
        return false, nil
    }

    // Parser les conditions JSON
    var conditions RuleConditions
    if err := json.Unmarshal(rule.Conditions, &conditions); err != nil {
        return false, fmt.Errorf("failed to parse rule conditions: %w", err)
    }

    // Évaluer récursivement
    triggered, err := s.evaluateConditions(ctx, userID, studentID, &conditions)
    if err != nil {
        return false, err
    }

    return triggered, nil
}

// Évaluer récursivement les conditions
func (s *ruleService) evaluateConditions(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, conditions *RuleConditions) (bool, error) {
    if conditions.Type == "penalty_count" {
        // Condition simple : compter les pénalités
        count, err := s.repo.CountPenaltiesByStudentAndType(ctx, repository.CountPenaltiesByStudentAndTypeParams{
            StudentID:     studentID,
            PenaltyTypeID: conditions.PenaltyTypeID,
            UserID:        userID,
        })
        if err != nil {
            return false, fmt.Errorf("failed to count penalties: %w", err)
        }

        return count >= int64(conditions.Threshold), nil
    }

    // Condition composée (AND/OR)
    if conditions.Operator == "AND" {
        for _, trigger := range conditions.Triggers {
            result, err := s.evaluateConditions(ctx, userID, studentID, &trigger)
            if err != nil {
                return false, err
            }
            if !result {
                return false, nil  // AND : un seul false suffit
            }
        }
        return true, nil
    }

    if conditions.Operator == "OR" {
        for _, trigger := range conditions.Triggers {
            result, err := s.evaluateConditions(ctx, userID, studentID, &trigger)
            if err != nil {
                return false, err
            }
            if result {
                return true, nil  // OR : un seul true suffit
            }
        }
        return false, nil
    }

    return false, fmt.Errorf("invalid operator: %s", conditions.Operator)
}

// Déclencher une punition si la règle est satisfaite
func (s *ruleService) TriggerPunishmentIfNeeded(ctx context.Context, userID uuid.UUID, ruleID uuid.UUID, studentID uuid.UUID) error {
    triggered, err := s.EvaluateRule(ctx, userID, ruleID, studentID)
    if err != nil {
        return err
    }

    if !triggered {
        return nil  // Règle non déclenchée
    }

    rule, err := s.repo.GetRule(ctx, repository.GetRuleParams{
        ID:     ruleID,
        UserID: userID,
    })
    if err != nil {
        return fmt.Errorf("failed to get rule: %w", err)
    }

    // Créer la punition
    punishment, err := s.repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
        UserID:             userID,
        StudentID:          studentID,
        PunishmentTypeID:   rule.ResultingPunishmentTypeID,
        TriggeringRuleID:   pgtype.UUID{Bytes: ruleID, Valid: true},
    })
    if err != nil {
        return fmt.Errorf("failed to create punishment: %w", err)
    }

    slog.Info("punishment triggered by rule", 
        "punishment_id", punishment.ID, 
        "rule_id", ruleID, 
        "student_id", studentID, 
        "user_id", userID)

    return nil
}
```

**Note importante :**
L'évaluation des règles doit être déclenchée après chaque ajout de pénalité. Cela peut être fait :
1. Dans le service `PenaltyService.CreatePenalty` : appeler `RuleService.EvaluateAllRulesForStudent`
2. Via un endpoint dédié : `POST /v1/penalties/{id}/evaluate-rules`

---

### 7. Users (Endpoints complets)

**Endpoints manquants :**
- `GET /v1/users/me` : Récupérer le profil de l'utilisateur authentifié
- `PUT /v1/users/me` : Mettre à jour le profil
- `DELETE /v1/users/me` : Supprimer le compte
- `PUT /v1/users/me/password` : Changer le mot de passe
- `GET /v1/users/me/stats` : Statistiques (nombre d'étudiants, classes, etc.)

**DTOs supplémentaires :**
```go
type UpdateUserDto struct {
    FirstName *string `json:"first_name" validate:"omitempty,min=2,max=70"`
    LastName  *string `json:"last_name" validate:"omitempty,min=2,max=70"`
}

type ChangePasswordDto struct {
    OldPassword string `json:"old_password" validate:"required,min=8"`
    NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UserStatsDto struct {
    StudentCount   int64 `json:"student_count"`
    ClassroomCount int64 `json:"classroom_count"`
    BonusCount     int64 `json:"bonus_count"`
    PenaltyCount   int64 `json:"penalty_count"`
}
```

---

## Priorité d'implémentation

### Priorité 1 (Fondations) 🔴
1. **PenaltyTypes** - Nécessaire pour les penalties
2. **PunishmentTypes** - Nécessaire pour les punishments
3. **Penalties** - Enregistrement des incidents
4. **Punishments** - Application des sanctions

### Priorité 2 (Fonctionnalités avancées) 🟡
5. **Bonuses** - Système de récompenses
6. **Rules** - Automatisation des punitions

### Priorité 3 (Améliorations) 🟢
7. **Users** (endpoints complets) - Gestion du profil

---

## Détails des fonctionnalités

### Workflow typique d'utilisation

1. **Configuration initiale** (par le professeur) :
   - Créer des `PenaltyTypes` (ex: Bavardage, Retard)
   - Créer des `PunishmentTypes` (ex: Retenue 1h, Mot aux parents)
   - Créer des `BonusTypes` (ex: Participation active)
   - Créer des `Students`
   - Créer des `Classrooms`
   - Associer les étudiants aux classes

2. **Utilisation quotidienne** :
   - Enregistrer des `Penalties` pour les incidents
   - Attribuer des `Bonuses` pour les comportements positifs
   - Créer manuellement des `Punishments` si nécessaire
   - Marquer les `Punishments` comme résolues

3. **Automatisation (optionnel)** :
   - Créer des `Rules` pour automatiser les punitions
   - Les règles s'évaluent automatiquement après chaque `Penalty`

4. **Consultation** :
   - Voir les statistiques d'un étudiant
   - Voir les punitions en attente
   - Voir l'historique des bonus/pénalités

---

## Dépendances entre entités

```
Users
├── Students
├── Classrooms
│   └── StudentClassrooms (relation)
├── BonusTypes
│   └── Bonuses (instances)
├── PenaltyTypes
│   └── Penalties (instances)
├── PunishmentTypes
│   └── Punishments (instances)
└── Rules
    └── déclenche → Punishments
```

**Ordre d'implémentation recommandé :**
1. PenaltyTypes (indépendant)
2. PunishmentTypes (indépendant)
3. Penalties (dépend de PenaltyTypes)
4. Punishments (dépend de PunishmentTypes)
5. Bonuses (dépend de BonusTypes - déjà implémenté)
6. Rules (dépend de PenaltyTypes, PunishmentTypes, Penalties)

---

## Notes finales

- Toutes ces entités suivent le même pattern architectural que les entités existantes
- Référez-vous au guide `feature-implementation-guide.md` pour les détails d'implémentation
- Les exemples de `Student`, `Classroom` et `BonusType` sont vos références
- N'oubliez pas de documenter chaque nouvelle fonctionnalité dans `api-endpoints.md`
