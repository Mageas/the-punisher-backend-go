# Tests des Handlers

Ce document décrit comment fonctionnent les tests backend, ce qui est couvert, et comment en ajouter de nouveaux.

## Objectif

Les tests de `internal/api/handler` valident le comportement HTTP de chaque handler:

- routes et statuts HTTP
- parsing JSON et erreurs de décodage
- validations DTO
- mapping des erreurs métier (`not_found`, `conflict`, etc.)
- erreurs internes (`500`)
- flux métier transverse `Rule -> Penalty -> Punishment`

## Organisation des fichiers

### Tests handlers

- `internal/api/handler/auth_test.go`
- `internal/api/handler/user_test.go`
- `internal/api/handler/health_test.go`
- `internal/api/handler/student_test.go`
- `internal/api/handler/classroom_test.go`
- `internal/api/handler/rule_test.go`
- `internal/api/handler/bonus_test.go`
- `internal/api/handler/penalty_test.go`
- `internal/api/handler/punishment_test.go`
- `internal/api/handler/bonus_type_test.go`
- `internal/api/handler/penalty_type_test.go`
- `internal/api/handler/punishment_type_test.go`

### Helpers de tests

- `internal/testutil/handlertest/auth.go`
  - génération de requêtes authentifiées JWT.
- `internal/testutil/httpx/http.go`
  - helpers JSON request/response.
- `internal/testutil/shared/type_handler_shared.go`
  - suites partagées pour les handlers `*type`.
- `internal/testutil/inmemory/*`
  - repository en mémoire pour exécuter les services sans DB réelle.

## Repository inmemory

Le package `internal/testutil/inmemory` simule les opérations SQLC nécessaires aux services:

- users/auth/refresh tokens
- students/classrooms + relation `student_classrooms`
- rules
- bonus/penalty/punishment (+ types)

Chaque table de test expose:

- des méthodes `Seed*` pour préparer l’état
- des constantes d’opération `Op*`
- le support `repo.SetError(Op..., err)` pour forcer les chemins d’erreur interne (`500`)

Exemple:

- `repo.SetError(inmemory.OpCreateStudent, errors.New("db down"))`
- `repo.ClearError(inmemory.OpCreateStudent)`

## Stratégie de couverture appliquée

Chaque handler est testé au minimum sur:

1. succès nominal (CRUD / actions)
2. validation DTO (`validation_failed`)
3. erreurs de decode JSON:
   - champ inconnu
   - type JSON invalide
4. erreurs de paramètres d’URL/query:
   - UUID malformé (`not_found`)
   - paramètre optionnel invalide (`malformed_parameter`)
5. erreurs métier:
   - `*_not_found`
   - `*_already_used`
   - `*_already_resolved`
   - etc.
6. erreurs internes (`internal_error`) via `SetError`

## Flux métier Rule -> Penalty -> Punishment

Couvert dans:

- `internal/api/handler/penalty_test.go` (`TestPenaltyHandlerCreateTriggersPunishmentFromRule`)

Le test vérifie que:

- une `Rule` active liée à un `PenaltyType` est seedée
- la création de penalties incrémente le compteur
- au seuil (`mode=at`, `threshold=2`) un punishment est créé automatiquement
- `triggering_rule_id`, `punishment_type_id`, `student_id` et `due_at` sont cohérents

Pour faciliter ce test en environnement in-memory:

- `service.NewPenaltyService(...)` est utilisé avec le repository in-memory
- la logique métier reste la même (évaluation des rules et création du punishment)

## Exécuter les tests

```bash
go test ./internal/api/handler
```

## Ajouter un nouveau handler

Checklist recommandée:

1. créer `internal/api/handler/<handler>_test.go`
2. créer un router test minimal avec middleware auth si nécessaire
3. couvrir les 6 catégories ci-dessus
4. ajouter/étendre `inmemory` uniquement pour les méthodes réellement utilisées
5. lancer `gofmt -w ...` puis `go test ./...`

## Notes

- Les tests sont volontairement orientés HTTP + service (pas juste unitaires purs).
- Le repository in-memory permet des tests rapides et déterministes.
- Les helpers partagés évitent la duplication et facilitent l’ajout de nouvelles suites.
