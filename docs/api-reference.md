# API Reference

Base path: `/v1`

## 1. Conventions globales

- Auth Bearer requise sur toutes les routes sauf :
  - `GET /health`
  - `GET /auth/register/status`
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
- JSON strict :
  - champs inconnus rejetés (`400 invalid_request_body`)
  - validation via tags `validate` (`400 validation_failed`)
- Pagination :
  - query param `?page=`
  - taille fixe : `20`
  - si `page` invalide ou <= 0, fallback sur page `1`
- Dates : format RFC3339.
- Réponse d'erreur standard :

```json
{
  "error": "validation_failed",
  "error_code": 400,
  "error_details": [
    { "field": "first_name", "error": "validation_field_required" }
  ]
}
```

### Réponse paginée standard

Toutes les routes `List*` renvoient :

```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 42,
  "previous_page": null,
  "next_page": 2,
  "data": []
}
```

## 2. Objets de réponse

### Student

```json
{
  "id": "uuid",
  "first_name": "Lucas",
  "last_name": "Dubois",
  "classrooms": [{ "id": "uuid", "name": "6eme A" }],
  "available_bonus_points": 3,
  "penalty_count": 5,
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### Classroom

```json
{
  "id": "uuid",
  "name": "6eme A",
  "year": "2025-2026",
  "main_teacher": "Mme Martin",
  "student_count": 12,
  "students_preview": [
    { "id": "uuid", "first_name": "Lucas", "last_name": "Dubois" }
  ],
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

`students_preview` est limité à 5 élèves max.

### Bonus

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Emma",
  "student_last_name": "Bernard",
  "bonus_type_id": "uuid",
  "bonus_type_name": "Participation",
  "points": 1,
  "created_at": "2026-02-18T10:00:00Z",
  "used_at": null
}
```

### Penalty

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Lucas",
  "student_last_name": "Dubois",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "created_at": "2026-02-18T10:00:00Z"
}
```

### Punishment

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Lucas",
  "student_last_name": "Dubois",
  "punishment_type_id": "uuid",
  "punishment_type_name": "Retenue",
  "triggering_rule_id": "uuid",
  "triggering_rule_name": "3 bavardages => retenue",
  "created_at": "2026-02-18T10:00:00Z",
  "due_at": "2026-02-25T10:00:00Z",
  "resolved_at": null
}
```

Pour une punition manuelle :
- `triggering_rule_id = null`
- `triggering_rule_name = null`

### Rule

```json
{
  "id": "uuid",
  "name": "3 bavardages => retenue",
  "resulting_punishment_type_id": "uuid",
  "resulting_punishment_type_name": "Retenue",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "threshold": 3,
  "due_at_after_days": 7,
  "mode": "every",
  "is_active": true,
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### Dashboard

```json
{
  "kpis": {
    "student_count": 34,
    "available_bonus_points": 14.5,
    "total_bonus_points": 26.5,
    "unused_bonus_count": 12,
    "penalty_count": 47,
    "total_punishment_count": 9,
    "overdue_punishment_count": 2,
    "pending_punishment_count": 5
  },
  "recent_penalties": [],
  "recent_bonuses": [],
  "pending_punishments": []
}
```

Chaque liste est limitée à 10 éléments.

### Student KPIs

```json
{
  "available_bonus_points": 3,
  "total_bonus_points": 6,
  "active_bonus_count": 2,
  "penalty_count": 5,
  "total_penalty_count": 5,
  "total_punishment_count": 3,
  "overdue_punishment_count": 1,
  "pending_punishment_count": 1
}
```

### Student History

```json
[
  {
    "type": "punishment",
    "id": "uuid",
    "punishment_type_id": "uuid",
    "punishment_type_name": "Retenue",
    "triggering_rule_id": "uuid",
    "triggering_rule_name": "3 bavardages => retenue",
    "due_at": "2026-02-25T10:00:00Z",
    "resolved_at": null,
    "created_at": "2026-02-18T10:00:00Z"
  },
  {
    "type": "penalty",
    "id": "uuid",
    "penalty_type_id": "uuid",
    "penalty_type_name": "Bavardage",
    "created_at": "2026-02-18T09:00:00Z"
  },
  {
    "type": "bonus",
    "id": "uuid",
    "bonus_type_id": "uuid",
    "bonus_type_name": "Participation",
    "points": 1,
    "used_at": null,
    "created_at": "2026-02-18T08:00:00Z"
  }
]
```

`history` est trié par `created_at` desc et paginé (taille fixe 20).

## 3. Health

### GET `/health`

Réponses :
- `200` si healthy
- `503` si unhealthy

Body :

```json
{
  "status": "healthy",
  "environment": "dev",
  "version": "dev",
  "services": {
    "database": "healthy"
  }
}
```

## 4. Auth

### GET `/auth/register/status`

Réponse `200` :

```json
{
  "register_allowed": true
}
```

### POST `/auth/register`

Body :

```json
{
  "email": "teacher@example.com",
  "first_name": "Jean",
  "last_name": "Dupont",
  "password": "password123"
}
```

Réponses :
- `201` -> `ReturnUserDto`
- `401 register_not_allowed` si `APP_ALLOW_REGISTER=false`

### POST `/auth/login`

Body :

```json
{
  "email": "teacher@example.com",
  "password": "password123"
}
```

Réponse `200` :

```json
{
  "access_token": "jwt"
}
```

Cookie `HttpOnly` posé :
- nom : `refresh_token`
- path : `/v1/auth/refresh`

### POST `/auth/refresh`

Nécessite le cookie `refresh_token`.
Le cookie `refresh_token` est roté (nouvelle valeur) à chaque appel réussi.

Réponse `200` :

```json
{
  "access_token": "jwt"
}
```

### GET `/user/me`

Auth :
- Bearer token requis

Réponses :
- `200` -> `ReturnUserDto`
- `401` -> `unauthorized`

## 5. Dashboard

### GET `/dashboard?classroom_id=uuid`

Query params :
- `classroom_id` optionnel

Réponses :
- `200` -> `ReturnDashboardDto`
- `404 not_found` si `classroom_id` n'est pas un UUID
- `404 classroom_not_found` si classe absente pour l'utilisateur

## 6. Students

### POST `/students/`

Body :

```json
{
  "first_name": "Lucas",
  "last_name": "Dubois"
}
```

Réponse :
- `201` -> `ReturnStudentDto`

### GET `/students/`

Query params :
- `page` optionnel
- `search` optionnel : recherche sur `prénom + nom` de l'élève

Réponse :
- `200` -> `PaginatedResponse<ReturnStudentDto>`

### GET `/students/{student_id}`

Réponse :
- `200` -> `ReturnStudentDto`

### PUT `/students/{student_id}`

Body partiel :

```json
{
  "first_name": "Nouveau",
  "last_name": "Nom"
}
```

Réponse :
- `200` -> `ReturnStudentDto`

### DELETE `/students/{student_id}`

Réponse :
- `204` no content

### GET `/students/{student_id}/kpis`

Query params :
- aucun

Réponse :
- `200` -> `StudentKpisDto`

### GET `/students/{student_id}/history`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<StudentHistoryItemDto>`

### GET `/students/{student_id}/classrooms`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnClassroomDto>`

### GET `/students/{student_id}/bonuses`

Query params :
- `page` optionnel
- `state` optionnel : `used|unused`

Réponse :
- `200` -> `PaginatedResponse<ReturnBonusDto>`

### GET `/students/{student_id}/penalties`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnPenaltyDto>`

### GET `/students/{student_id}/punishments`

Query params :
- `page` optionnel
- `state` optionnel : `pending|resolved`

Réponse :
- `200` -> `PaginatedResponse<ReturnPunishmentDto>`

## 7. Classrooms

### POST `/classrooms/`

Body :

```json
{
  "name": "6eme A",
  "year": "2025-2026",
  "main_teacher": "Mme Martin"
}
```

Réponse :
- `201` -> `ReturnClassroomDto`

### GET `/classrooms/`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnClassroomDto>`

### GET `/classrooms/{classroom_id}`

Réponse :
- `200` -> `ReturnClassroomDto`

### GET `/classrooms/{classroom_id}/kpis`

Réponse :
- `200` -> `DashboardKpisDto`

### PUT `/classrooms/{classroom_id}`

Body partiel :

```json
{
  "name": "6eme B",
  "year": "2026-2027",
  "main_teacher": "M. Leroy"
}
```

Réponse :
- `200` -> `ReturnClassroomDto`

### DELETE `/classrooms/{classroom_id}`

Réponse :
- `204` no content

### POST `/classrooms/{classroom_id}/students`

Body :

```json
{
  "student_id": "uuid"
}
```

Réponse :
- `204` no content

### DELETE `/classrooms/{classroom_id}/students/{student_id}`

Réponse :
- `204` no content

### GET `/classrooms/{classroom_id}/students`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnStudentDto>`

## 8. Bonus Types

### POST `/bonus-types/`
### GET `/bonus-types/`
### GET `/bonus-types/{bonus_type_id}`
### PUT `/bonus-types/{bonus_type_id}`
### DELETE `/bonus-types/{bonus_type_id}`

Payloads :
- create/update body : `{ "name": "Participation" }`
- list : `PaginatedResponse<ReturnBonusTypeDto>`
- get/create/update : `ReturnBonusTypeDto`
- delete : `204`

## 9. Penalty Types

### POST `/penalty-types/`
### GET `/penalty-types/`
### GET `/penalty-types/{penalty_type_id}`
### PUT `/penalty-types/{penalty_type_id}`
### DELETE `/penalty-types/{penalty_type_id}`

Payloads :
- create/update body : `{ "name": "Bavardage" }`
- list : `PaginatedResponse<ReturnPenaltyTypeDto>`
- get/create/update : `ReturnPenaltyTypeDto`
- delete : `204`

## 10. Punishment Types

### POST `/punishment-types/`
### GET `/punishment-types/`
### GET `/punishment-types/{punishment_type_id}`
### PUT `/punishment-types/{punishment_type_id}`
### DELETE `/punishment-types/{punishment_type_id}`

Payloads :
- create/update body : `{ "name": "Retenue" }`
- list : `PaginatedResponse<ReturnPunishmentTypeDto>`
- get/create/update : `ReturnPunishmentTypeDto`
- delete : `204`

## 11. Bonuses

### POST `/bonuses/`

Body :

```json
{
  "student_id": "uuid",
  "bonus_type_id": "uuid",
  "points": 1.5
}
```

Réponse :
- `201` -> `ReturnBonusDto`

### GET `/bonuses/`

Query params :
- `page` optionnel
- `state` optionnel : `used|unused`
- `search` optionnel : recherche sur `prénom + nom` de l'élève

Réponse :
- `200` -> `PaginatedResponse<ReturnBonusDto>`

### GET `/bonuses/{bonus_id}`

Réponse :
- `200` -> `ReturnBonusDto`

### POST `/bonuses/{bonus_id}/use`

Effet :
- passe `used_at` de `null` à `now`

Réponse :
- `200` -> `ReturnBonusDto`

### DELETE `/bonuses/{bonus_id}`

Réponse :
- `204` no content

## 12. Penalties

### POST `/penalties/`

Body :

```json
{
  "student_id": "uuid",
  "penalty_type_id": "uuid"
}
```

Effet métier :
- crée une pénalité
- évalue les règles actives
- peut créer automatiquement une punition

Réponse :
- `201` -> `ReturnPenaltyDto`

### GET `/penalties/`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnPenaltyDto>`

### GET `/penalties/{penalty_id}`

Réponse :
- `200` -> `ReturnPenaltyDto`

### DELETE `/penalties/{penalty_id}`

Réponse :
- `204` no content

## 13. Punishments

### POST `/punishments/`

Body :

```json
{
  "student_id": "uuid",
  "punishment_type_id": "uuid",
  "due_at": "2026-02-20T17:00:00Z"
}
```

Réponse :
- `201` -> `ReturnPunishmentDto`

### GET `/punishments/`

Query params :
- `page` optionnel
- `state` optionnel : `pending|resolved`
- `search` optionnel : recherche sur `prénom + nom` de l'élève

Réponse :
- `200` -> `PaginatedResponse<ReturnPunishmentDto>`

### GET `/punishments/{punishment_id}`

Réponse :
- `200` -> `ReturnPunishmentDto`

### POST `/punishments/{punishment_id}/resolve`

Effet :
- passe `resolved_at` de `null` à `now`

Réponse :
- `200` -> `ReturnPunishmentDto`

### DELETE `/punishments/{punishment_id}`

Réponse :
- `204` no content

## 14. Rules

### POST `/rules/`

Body :

```json
{
  "name": "3 bavardages => retenue",
  "resulting_punishment_type_id": "uuid",
  "penalty_type_id": "uuid",
  "threshold": 3,
  "due_at_after_days": 7,
  "mode": "every",
  "is_active": true
}
```

Contraintes :
- `mode` : `after|at|every`
- `threshold >= 1`
- `due_at_after_days >= 0`

Réponse :
- `201` -> `ReturnRuleDto`

### GET `/rules/`

Query params :
- `page` optionnel

Réponse :
- `200` -> `PaginatedResponse<ReturnRuleDto>`

### GET `/rules/{rule_id}`

Réponse :
- `200` -> `ReturnRuleDto`

### PUT `/rules/{rule_id}`

Body partiel possible :

```json
{
  "name": "4 bavardages => retenue",
  "threshold": 4,
  "mode": "at",
  "is_active": false
}
```

Réponse :
- `200` -> `ReturnRuleDto`

### DELETE `/rules/{rule_id}`

Réponse :
- `204` no content

## 15. Codes d'erreur utilisés

Erreurs globales :
- `internal_error`
- `invalid_request_body`
- `malformed_parameter`
- `validation_failed`
- `unauthorized`

Auth :
- `register_not_allowed`
- `invalid_credentials_or_user_doesnt_exist`
- `jwt_invalid_signing_method`
- `jwt_invalid_token`
- `jwt_expired`

Not found :
- `student_not_found`
- `classroom_not_found`
- `bonus_type_not_found`
- `penalty_type_not_found`
- `punishment_type_not_found`
- `rule_not_found`
- `bonus_not_found`
- `penalty_not_found`
- `punishment_not_found`
- `student_or_classroom_not_found`

Conflicts :
- `conflict` (email déjà utilisé, via `error_details`)
- `student_classroom_relation_exists`
- `bonus_already_used`
- `punishment_already_resolved`

## 16. Définitions des DTOs

### ReturnUserDto

```json
{
  "id": "uuid",
  "email": "teacher@example.com",
  "first_name": "Jean",
  "last_name": "Dupont",
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### ReturnDashboardDto

```json
{
  "kpis": {
    "student_count": 34,
    "available_bonus_points": 14.5,
    "total_bonus_points": 26.5,
    "unused_bonus_count": 12,
    "penalty_count": 47,
    "total_punishment_count": 9,
    "overdue_punishment_count": 2,
    "pending_punishment_count": 5
  },
  "recent_penalties": [],
  "recent_bonuses": [],
  "pending_punishments": []
}
```

### DashboardKpisDto

```json
{
  "student_count": 34,
  "available_bonus_points": 14.5,
  "total_bonus_points": 26.5,
  "unused_bonus_count": 12,
  "penalty_count": 47,
  "total_punishment_count": 9,
  "overdue_punishment_count": 2,
  "pending_punishment_count": 5
}
```

### ReturnStudentDto

```json
{
  "id": "uuid",
  "first_name": "Lucas",
  "last_name": "Dubois",
  "classrooms": [{ "id": "uuid", "name": "6eme A" }],
  "available_bonus_points": 3,
  "penalty_count": 5,
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### StudentKpisDto

```json
{
  "available_bonus_points": 3,
  "total_bonus_points": 6,
  "active_bonus_count": 2,
  "penalty_count": 5,
  "total_penalty_count": 5,
  "total_punishment_count": 3,
  "overdue_punishment_count": 1,
  "pending_punishment_count": 1
}
```

### StudentHistoryItemDto

```json
{
  "type": "punishment|penalty|bonus",
  "id": "uuid",
  "penalty_type_id": "uuid|null",
  "penalty_type_name": "string|null",
  "bonus_type_id": "uuid|null",
  "bonus_type_name": "string|null",
  "points": 1.5,
  "used_at": "2026-02-18T10:00:00Z|null",
  "punishment_type_id": "uuid|null",
  "punishment_type_name": "string|null",
  "triggering_rule_id": "uuid|null",
  "triggering_rule_name": "string|null",
  "due_at": "2026-02-25T10:00:00Z|null",
  "resolved_at": "2026-02-26T10:00:00Z|null",
  "created_at": "2026-02-18T10:00:00Z"
}
```

### ReturnClassroomDto

```json
{
  "id": "uuid",
  "name": "6eme A",
  "year": "2025-2026",
  "main_teacher": "Mme Martin",
  "student_count": 12,
  "students_preview": [
    { "id": "uuid", "first_name": "Lucas", "last_name": "Dubois" }
  ],
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### ReturnBonusTypeDto

```json
{
  "id": "uuid",
  "name": "Participation",
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### ReturnPenaltyTypeDto

```json
{
  "id": "uuid",
  "name": "Bavardage",
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### ReturnPunishmentTypeDto

```json
{
  "id": "uuid",
  "name": "Retenue",
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```

### ReturnBonusDto

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Emma",
  "student_last_name": "Bernard",
  "bonus_type_id": "uuid",
  "bonus_type_name": "Participation",
  "points": 1,
  "created_at": "2026-02-18T10:00:00Z",
  "used_at": "2026-02-19T10:00:00Z|null"
}
```

### ReturnPenaltyDto

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Lucas",
  "student_last_name": "Dubois",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "created_at": "2026-02-18T10:00:00Z"
}
```

### ReturnPunishmentDto

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Lucas",
  "student_last_name": "Dubois",
  "punishment_type_id": "uuid",
  "punishment_type_name": "Retenue",
  "triggering_rule_id": "uuid|null",
  "triggering_rule_name": "string|null",
  "created_at": "2026-02-18T10:00:00Z",
  "due_at": "2026-02-25T10:00:00Z",
  "resolved_at": "2026-02-26T10:00:00Z|null"
}
```

### ReturnRuleDto

```json
{
  "id": "uuid",
  "name": "3 bavardages => retenue",
  "resulting_punishment_type_id": "uuid",
  "resulting_punishment_type_name": "Retenue",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "threshold": 3,
  "due_at_after_days": 7,
  "mode": "after|at|every",
  "is_active": true,
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z"
}
```
