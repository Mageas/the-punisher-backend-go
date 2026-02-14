# API Reference - Cible Métier

Base path: `/v1`

Ce document décrit le contrat API cible aligné sur la BDD canonique.

## 1. Conventions globales

- Auth Bearer requise hors `/health` et `/auth/*`.
- JSON strict (unknown fields interdits).
- Pagination via `?page=` (taille fixe 20).
- Erreur standard:

```json
{
  "error": "validation_failed",
  "error_code": 400,
  "error_details": [
    { "field": "name", "error": "validation_field_required" }
  ]
}
```

## 2. Health

### GET `/health`

- `200` healthy
- `503` unhealthy

## 3. Auth

### POST `/auth/register`
### POST `/auth/login`
### POST `/auth/refresh`

Principes:
- access token dans le body.
- refresh token en cookie `HttpOnly`.

## 4. Students

### POST `/students`
### GET `/students?page=1`
### GET `/students/{id}`
### PUT `/students/{id}`
### DELETE `/students/{id}`

Note:
- les réponses `students` restent plates (pas d'IDs de relations embarqués).
- pour les relations, utiliser les endpoints dédiés ci-dessous.

### GET `/students/{id}/classrooms?page=1`
### GET `/students/{id}/bonuses?page=1`
Paramètre optionnel:
- `state`: `used|unused`. Exemple: `GET /students/{id}/bonuses?page=1&state=used`.
### GET `/students/{id}/penalties?page=1`
### GET `/students/{id}/punishments?page=1`
Paramètre optionnel:
- `state`: `pending|resolved`. Exemple: `GET /students/{id}/punishments?page=1&state=resolved`.

## 5. Classrooms

### POST `/classrooms`
### GET `/classrooms?page=1`
### GET `/classrooms/{id}`
### PUT `/classrooms/{id}`
### DELETE `/classrooms/{id}`
### POST `/classrooms/{id}/students`

Body:

```json
{ "student_id": "uuid" }
```

### DELETE `/classrooms/{id}/students/{studentId}`
### GET `/classrooms/{id}/students?page=1`

## 6. Types catalogues

### Bonus Types
- `POST /bonus-types`
- `GET /bonus-types`
- `GET /bonus-types/{id}`
- `PUT /bonus-types/{id}`
- `DELETE /bonus-types/{id}`

### Penalty Types
- `POST /penalty-types`
- `GET /penalty-types`
- `GET /penalty-types/{id}`
- `PUT /penalty-types/{id}`
- `DELETE /penalty-types/{id}`

### Punishment Types
- `POST /punishment-types`
- `GET /punishment-types`
- `GET /punishment-types/{id}`
- `PUT /punishment-types/{id}`
- `DELETE /punishment-types/{id}`

## 7. Bonuses

### POST `/bonuses`

Body:

```json
{
  "student_id": "uuid",
  "bonus_type_id": "uuid",
  "points": 1.5
}
```

### GET `/bonuses?page=1`
Paramètre optionnel:
- `state`: `used|unused`. Exemple: `GET /bonuses?page=1&state=used`.
### GET `/bonuses/{id}`
### DELETE `/bonuses/{id}`
### POST `/bonuses/{id}/use`

Effet:
- passe `used_at` de `NULL` à timestamp.
- si déjà utilisé: conflit.

## 8. Penalties

### POST `/penalties`

Body:

```json
{
  "student_id": "uuid",
  "penalty_type_id": "uuid"
}
```

Effet métier:
- enregistre une pénalité,
- déclenche l'évaluation des règles actives (`is_active = true`).

### GET `/penalties?page=1`
### GET `/penalties/{id}`
### DELETE `/penalties/{id}`

## 9. Rules

### POST `/rules`

Body:

```json
{
  "name": "3 oublis matériel => retenue",
  "resulting_punishment_type_id": "uuid",
  "penalty_type_id": "uuid-oubli-materiel",
  "threshold": 3,
  "due_at_after_days": 7,
  "mode": "every",
  "is_active": true
}
```

Contraintes:
- `mode`: `after|at|every`
- `threshold >= 1`
- `due_at_after_days >= 0`
- `is_active` booléen (`true` par défaut).
- `penalty_type_id` doit appartenir au même user.
- `resulting_punishment_type_id` doit appartenir au même user.
- seules les règles avec `is_active = true` peuvent déclencher automatiquement une `Punishment`.
- si un `penalty_type` est supprimé, ses règles associées sont supprimées en cascade.
- lors d'un déclenchement automatique, `punishments.due_at = now + due_at_after_days`.
- `punishments.resolved_at` reste `NULL` à la création.

Autres exemples:

```json
{
  "name": "3 retards => retenue (déclenchement au seuil)",
  "resulting_punishment_type_id": "uuid",
  "penalty_type_id": "uuid-retard",
  "threshold": 3,
  "due_at_after_days": 3,
  "mode": "at",
  "is_active": true
}
```

```json
{
  "name": "Tous les 5 bavardages => mot aux parents",
  "resulting_punishment_type_id": "uuid",
  "penalty_type_id": "uuid-bavardage",
  "threshold": 5,
  "due_at_after_days": 14,
  "mode": "every",
  "is_active": false
}
```

```json
{
  "name": "Après 2 bavardages, chaque nouveau bavardage => mot aux parents",
  "resulting_punishment_type_id": "uuid",
  "penalty_type_id": "uuid-bavardage",
  "threshold": 2,
  "due_at_after_days": 10,
  "mode": "after",
  "is_active": true
}
```

### GET `/rules?page=1`
### GET `/rules/{id}`
### PUT `/rules/{id}`
### DELETE `/rules/{id}`

## 10. Punishments

### POST `/punishments`

Création manuelle.

Body:

```json
{
  "student_id": "uuid",
  "punishment_type_id": "uuid",
  "due_at": "2026-02-20T17:00:00Z"
}
```

### GET `/punishments?page=1`
Paramètre optionnel:
- `state`: `pending|resolved`. Exemple: `GET /punishments?page=1&state=pending`.
### GET `/punishments/{id}`
### DELETE `/punishments/{id}`
### POST `/punishments/{id}/resolve`

Effet:
- passe `resolved_at` de `NULL` à timestamp.
- si déjà résolue: conflit.

Format de réponse `punishments`:
- `triggering_rule_id` est toujours présent dans le JSON.
- valeur `null` pour une punition manuelle, UUID pour une punition automatique.

## 11. Codes d'erreur recommandés

- `400 validation_failed`
- `400 invalid_request_body`
- `400 malformed_parameter`
- `401 unauthorized`
- `404 student_not_found`
- `404 classroom_not_found`
- `404 bonus_type_not_found`
- `404 penalty_type_not_found`
- `404 punishment_type_not_found`
- `404 rule_not_found`
- `404 bonus_not_found`
- `404 penalty_not_found`
- `404 punishment_not_found`
- `409 student_classroom_relation_exists`
- `409 bonus_already_used`
- `409 punishment_already_resolved`
