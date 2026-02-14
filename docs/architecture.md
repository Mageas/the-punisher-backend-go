# Architecture & Conventions

Ce document décrit l'architecture technique cible alignée sur la BDD canonique du projet.

## 1. Architecture globale

Flux principal:

`HTTP Route -> Handler -> Service -> Repository (sqlc) -> PostgreSQL`

Responsabilités:
- `Handler`: parsing HTTP, validation, format réponse.
- `Service`: logique métier et orchestration.
- `Repository`: accès SQL typé (sqlc).
- `Rule Engine` (dans service): évaluation des `rules.conditions`.

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
- source: table `rules.conditions` (JSONB).
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

## 6. Contrat JSON des règles

`rules.conditions` suit un arbre récursif:

Noeud logique:

```json
{
  "operator": "AND|OR",
  "triggers": ["node|leaf", "node|leaf"]
}
```

Feuille (seule règle supportée pour l'instant):

```json
{
  "type": "penalty_count",
  "penalty_type_ids": ["uuid", "uuid"],
  "threshold": 3,
  "mode": "at|every|after"
}
```

Sémantique:
- `type` permet d'ouvrir à d'autres triggers plus tard. Pour l'instant, seul `penalty_count` existe.
- `penalty_type_ids` est optionnel:
  - absent => tous les types de pénalités
  - présent => filtre "IN" (type A ou B)
- `threshold` = X
- `mode`:
  - `at`: déclenche une fois quand `count == X`
  - `every`: déclenche à chaque multiple (`count % X == 0`)
  - `after`: déclenche à chaque nouvel événement si `count > X`

Validation minimale à imposer:
- `operator` obligatoire sur noeuds.
- `triggers` non vide.
- `threshold >= 1`.
- `mode` ∈ {`at`, `every`, `after`}.
- `penalty_type_ids` appartient au même user (si fourni).

Point d'attention:
- lors de l'implémentation des Rules, veiller à supprimer les "enfants" contenus dans `conditions` qui ne sont plus référencés (ex: update/suppression de règles). Si `conditions` est stocké en JSONB, remplacer le document entier pour éviter des résidus.

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
