# Architecture & Conventions

Ce document décrit l'architecture technique cible alignée sur la BDD canonique du projet.

## 1. Architecture globale

Flux principal:

`HTTP Route -> Handler -> Service -> Repository (sqlc) -> PostgreSQL`

Responsabilités:
- `Handler`: parsing HTTP, validation, format réponse.
- `Service`: logique métier et orchestration.
- `Repository`: accès SQL typé (sqlc).
- `Rule Engine` (dans service): évaluation des triggers des `rules`.

Structure du Repo:
```text
the-punisher-go/
├── cmd/
│   ├── api/
│   │   ├── main.go      # démarrage serveur, DB pool, config
│   │   └── routes.go    # wiring dépendances et routes HTTP
│   └── seed/
│       └── main.go      # exécution des seeders
├── db/
│   ├── migrations/      # schéma SQL versionné
│   └── sqlc/
│       └── queries.sql  # source sqlc
├── internal/
│   ├── api/
│   │   ├── errors.go
│   │   └── handler/
│   ├── dto/
│   ├── platform/
│   │   ├── auth/
│   │   ├── config/
│   │   ├── hash/
│   │   ├── jwt/
│   │   ├── validator/
│   │   └── web/
│   ├── repository/      # généré par sqlc
│   ├── seeder/
│   └── service/
├── docs/
└── sqlc.yaml
```

## 2. Modules métier

Le backend est organisé par domaine:
- `auth`
- `students`
- `classrooms`
- `student_classrooms`
- `bonus_types`
- `penalty_types`
- `punishment_types`
- `bonuses`
- `penalties`
- `rules`
- `punishments`

## 3. Invariants de sécurité

1. Multi-tenant obligatoire:
- toutes les queries métier incluent `user_id`.

2. Ownership relationnel:
- toute opération `student <-> classroom` valide le même `user_id`.
- même principe pour `bonus_type`, `penalty_type`, `punishment_type`.

3. Visibilité:
- une ressource valide mais hors tenant retourne `404` (pas `403`).

## 4. Règles métier structurantes

### 4.1 Events

- `Bonuses` et `Penalties` sont des événements datés.
- ils alimentent l'état métier dérivé (compteurs, disponibilité).

### 4.2 Rule Engine

- input: élève, user, nouvel événement de pénalité.
- source: table `rules` (`penalty_type_id`, `threshold`, `mode`, `is_active`).
- filtre obligatoire: évaluer uniquement les règles où `is_active = true`.
- output: zéro ou plusieurs `punishments`.

### 4.3 Punishments

- manuelle: `triggering_rule_id = NULL`.
- automatique: `triggering_rule_id` renseigné.
- statut:
  - pending: `resolved_at IS NULL`
  - resolved: `resolved_at IS NOT NULL`

### 4.4 Bonuses

- bonus disponible si `used_at IS NULL`.
- consommation atomique: `UPDATE ... WHERE used_at IS NULL`.

## 5. Transaction boundaries recommandées

1. `CreatePenalty`:
- insérer pénalité,
- évaluer règles,
- créer punition(s) automatique(s) si nécessaire,
- commit unique.

2. `UseBonus`:
- marquer bonus consommé (`used_at = NOW()`),
- vérifier `rowsAffected == 1` pour éviter double consommation.

## 6. Contrat des règles

Chaque règle stocke un trigger simple:

```json
{
  "penalty_type_id": "uuid",
  "threshold": 3,
  "mode": "at|every|after",
  "is_active": true
}
```

Sémantique:
- `penalty_type_id`: type de pénalité observé pour la règle.
- `threshold`: nombre de pénalités à atteindre.
- `mode`:
  - `at`: déclenche une fois quand `count == threshold`
  - `every`: déclenche à chaque multiple (`count % threshold == 0`)
  - `after`: déclenche à chaque nouvel événement si `count > threshold`
- `is_active`: active/désactive la règle sans suppression.

Validation minimale à imposer:
- `threshold >= 1`.
- `mode` ∈ {`at`, `every`, `after`}.
- `is_active` booléen (par défaut `true` à la création).
- `penalty_type_id` appartient au même user.

Intégrité référentielle:
- la relation `rules.penalty_type_id -> penalty_types.id` doit être en `ON DELETE CASCADE`.
- suppression d'un `penalty_type` => suppression automatique des règles associées.

## 7. SQLC conventions

Dans `db/sqlc/queries.sql`:
- créer un bloc de queries par domaine.
- utiliser `:execrows` pour les updates/deletes conditionnels.
- utiliser `COALESCE(sqlc.narg(...), ...)` pour updates partielles.
- paginer via `LIMIT/OFFSET`.

Patterns utiles:
- `Count*ByUser`
- `List*ByUser`
- `Get*ByUser`
- `Update*ByUser`
- `Delete*ByUser`

## 8. API conventions

1. Decode:
- `web.DecodeJSON` (max body 1MB, unknown fields interdits).

2. Validation:
- tags `validate` sur DTO.

3. Erreurs:
- centralisées via `api.APIError`.
- réponse uniforme:
  - `error`
  - `error_details`
  - `error_code`

4. Pagination:
- query `page`.
- taille fixe 20.
- réponse standard `PaginatedResponse`.

## 9. Checklist implémentation d'une entité

1. Migration `up/down`.
2. Queries SQLC.
3. `make sqlc`.
4. DTOs.
5. Service.
6. Handler.
7. Routes.
8. Erreurs API dédiées.
9. Documentation (`docs/projet.md`, `docs/api-reference.md`, `docs/feature-playbook.md`).

## 10. Exemple de workflow complet (penalty -> punishment)

1. `POST /penalties`.
2. Service `CreatePenalty` persiste l'événement.
3. Service charge les règles du user.
4. Rule engine calcule les règles satisfaites.
5. Service crée les `punishments` associées.
6. Réponse inclut la pénalité créée et éventuellement les punitions générées.

## 11. Commandes utiles

- `make dev`
- `make build`
- `make sqlc`
- `make migrate-create <name>`
- `make migrate-up`
- `make migrate-down`
