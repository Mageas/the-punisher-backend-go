# Playbook Ajout de Feature (BDD canonique)

Ce guide sert de procédure d'exécution pour une IA ou un dev afin d'ajouter une feature cohérente avec le modèle métier complet.

## 1. Ordre d'implémentation obligatoire

1. Migration SQL.
2. Queries SQLC.
3. Génération sqlc.
4. Erreurs API.
5. DTOs.
6. Service.
7. Handler.
8. Routes.
9. Documentation.

## 2. Patterns de référence

Références structurantes:
- `Students`: CRUD tenant-scopé + pagination.
- `Classrooms`: CRUD + relation N-N (`StudentClassrooms`).

Nouveaux domaines à calquer sur ces patterns:
- `PenaltyTypes`, `PunishmentTypes`
- `Bonuses`, `Penalties`, `Rules`, `Punishments`

## 3. Design SQL attendu par domaine

### 3.1 Types (bonus/penalty/punishment)

CRUD standard tenant-scopé:
- `Create<Type>`
- `Get<Type>ByUser`
- `Count<Types>ByUser`
- `List<Types>ByUser`
- `Update<Type>ByUser`
- `Delete<Type>ByUser`

### 3.2 Events (bonuses/penalties)

- `CreateBonus`, `CreatePenalty`
- `Get*ByUser`
- `List*ByUser`
- filtres par `student_id` utiles

Pour bonus:
- query atomique `UseBonus` (`SET used_at = NOW() WHERE id=? AND used_at IS NULL`).

### 3.3 Rules

- `CreateRule`
- `GetRuleByUser`
- `ListRulesByUser`
- `UpdateRuleByUser`
- `DeleteRuleByUser`

Validation de cohérence:
- `resulting_punishment_type_id` doit appartenir au même `user_id`.

### 3.4 Punishments

- `CreatePunishment`
- `CreatePunishmentFromRule` (optionnel, même query avec `triggering_rule_id`)
- `ListPunishmentsByUser`
- `ListPendingPunishmentsByUser`
- `ListResolvedPunishmentsByUser`
- `ResolvePunishment` (`resolved_at = NOW()` conditionnel)

Note:
- `triggering_rule_id` est réservé pour l'intégration des `Rules` (FK à ajouter quand la table existera).

## 4. Rule Engine - Implémentation recommandée

## 4.1 Entrées

- `user_id`
- `student_id`
- timestamp de l'événement courant

## 4.2 Étapes

1. Charger les règles du user.
2. Pour chaque règle, parser `conditions`.
3. Évaluer récursivement `AND`/`OR`.
4. Pour `penalty_count`, calculer le count `Penalties` pour (`student_id`, `penalty_type_id`, `user_id`).
5. Si vrai, créer une `Punishment` avec `triggering_rule_id`.

## 4.3 Protection anti-doublon (fortement recommandé)

Ajouter une règle d'idempotence pour éviter de créer en boucle la même punition lors d'événements répétés.

Options:
- contrainte unique logique (selon période),
- ou table d'historique de déclenchements.

## 5. DTO patterns

Créer dans `internal/dto/<entity>.go`:
- `Request<Entity>Dto`
- `Update<Entity>Dto`
- `Return<Entity>Dto`
- mapping repository -> DTO.

Règle:
- update partiel via pointeurs.

Exemple:

```go
type UpdateRuleDto struct {
    Name                      *string          `json:"name" validate:"omitempty,min=2,max=120"`
    ResultingPunishmentTypeID *string          `json:"resulting_punishment_type_id" validate:"omitempty,uuid"`
    Conditions                *json.RawMessage `json:"conditions" validate:"omitempty"`
}
```

## 6. Service patterns

Obligatoire:
- mapper `pgx.ErrNoRows` vers erreur métier `404`.
- mapper `rowsAffected == 0` vers `404` ou `409` selon cas.
- wrap interne: `fmt.Errorf("failed to ...: %w", err)`.

Transactions recommandées:
- `CreatePenalty + evaluateRules + createPunishments`.
- `UseBonus` atomique.
- `ResolvePunishment` atomique.

## 7. Handler patterns

Toujours:
- `userID := auth.MustUserIDFromContext(r.Context())`
- `web.DecodeJSON`
- `validator.ValidateStruct`
- `web.WriteJSON` / `web.WriteFromError`

HTTP status:
- `201` create
- `200` get/list/update/actions non destructives
- `204` delete

## 8. Erreurs métier à ajouter

Catalogue minimal:
- `ErrPenaltyTypeNotFound`
- `ErrPunishmentTypeNotFound`
- `ErrRuleNotFound`
- `ErrBonusNotFound`
- `ErrPenaltyNotFound`
- `ErrPunishmentNotFound`
- `ErrBonusAlreadyUsed`
- `ErrPunishmentAlreadyResolved`
- `ErrRuleConditionInvalid`

## 9. Tests recommandés (même si non présents)

1. Unit services:
- règles de mapping erreurs,
- transitions d'état (`used_at`, `resolved_at`),
- évaluation de conditions complexes.

2. Integration handlers:
- validation payload,
- pagination,
- isolation tenant,
- endpoints d'action (`/use`, `/resolve`).

3. End-to-end métier:
- créer pénalités successives,
- vérifier création auto de punitions,
- vérifier lien `triggering_rule_id`.

## 10. Definition of Done

Une feature est terminée si:
1. migrations et sqlc sont propres,
2. API compile,
3. comportement métier principal est validé,
4. erreurs et statuts HTTP sont cohérents,
5. docs sont à jour (`projet`, `architecture`, `api-reference`, `feature-playbook`).
