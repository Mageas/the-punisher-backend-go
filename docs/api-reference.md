# API Reference - The Punisher Backend (guide technique frontend IA)

## 1) Conventions globales

### 1.1 Base URL
- Local: `http://localhost:8080`
- Prefix API: toutes les routes metier sont sous `/v1`

### 1.2 Authentification
- Routes publiques:
  - `GET /v1/health`
  - `GET /v1/auth/register/status`
  - `POST /v1/auth/register`
  - `POST /v1/auth/login`
  - `POST /v1/auth/refresh`
  - `POST /v1/auth/logout`
- Routes protegees: toutes les autres routes `/v1/**` exigent `Authorization: Bearer <access_token>`.

Flow recommande frontend:
1. `POST /v1/auth/login` avec email/password
2. Recuperer `access_token` dans le body
3. Stocker `access_token` cote client (memoire de preference)
4. Utiliser `Authorization: Bearer <token>`
5. Quand 401/expiration: `POST /v1/auth/refresh` (cookie `refresh_token` automatiquement envoye si `credentials: include`)

Notes cookie refresh:
- Nom: `refresh_token`
- `HttpOnly`, `SameSite=Strict`
- Path cookie: `/v1/auth`
- Le frontend doit utiliser `credentials: 'include'` pour `login`, `refresh`, `logout`.

### 1.3 Types de donnees
- UUID: format canonique (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
- Date liste: `YYYY-MM-DD` (ex: `2026-02-28`)
- DateTime body (punishments): `RFC3339` (ex: `2026-03-15T18:00:00Z`)
- Bool query: `true` ou `false`

### 1.4 Pagination
- Parametre: `page` (int > 0)
- Taille fixe backend: `20`
- Si `page` invalide/absent: fallback page `1` (pas d erreur)

Format reponse paginee:
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

### 1.5 Format d erreur (global)
```json
{
  "error": "malformed_parameter",
  "error_details": [
    {
      "field": "created_from",
      "error": "validation_malformed_parameter:expected_yyyy-mm-dd"
    }
  ],
  "error_code": 400
}
```

## 2) Contrats JSON (schemas frontend)

```ts
// Pagination
interface PaginatedResponse<T> {
  page: number;
  item_per_page: number;
  total_count: number;
  previous_page: number | null;
  next_page: number | null;
  data: T[];
}

interface ErrorDetail {
  row?: number;
  field: string;
  error: string;
  value?: string;
}

interface ErrorResponse {
  error: string;
  error_code: number;
  error_details?: ErrorDetail[];
}

// Auth
interface LoginRequestDto {
  email: string;
  password: string;
}

interface LoginResponseDto {
  access_token: string;
}

interface RefreshResponseDto {
  access_token: string;
}

interface RegisterStatusResponseDto {
  register_allowed: boolean;
}

// User
interface ReturnUserDto {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
  updated_at: string;
}

// Students
interface StudentClassroomDto {
  id: string;
  name: string;
}

interface ReturnStudentDto {
  id: string;
  first_name: string;
  last_name: string;
  classrooms: StudentClassroomDto[];
  available_bonus_points: number;
  penalty_count: number;
  created_at: string;
  updated_at: string;
}

interface StudentKpisDto {
  available_bonus_points: number;
  total_bonus_points: number;
  active_bonus_count: number;
  penalty_count: number;
  total_penalty_count: number;
  total_punishment_count: number;
  overdue_punishment_count: number;
  pending_punishment_count: number;
}

interface StudentHistoryItemDto {
  type: string; // bonus | penalty | punishment
  id: string;
  penalty_type_id?: string;
  penalty_type_name?: string;
  bonus_type_id?: string;
  bonus_type_name?: string;
  points?: number;
  used_at?: string;
  punishment_type_id?: string;
  punishment_type_name?: string;
  triggering_rule_id?: string;
  triggering_rule_name?: string;
  automated?: boolean;
  due_at?: string;
  resolved_at?: string;
  created_at: string;
}

// Classrooms
interface ClassroomStudentPreviewDto {
  id: string;
  first_name: string;
  last_name: string;
}

interface ReturnClassroomDto {
  id: string;
  name: string;
  year: string | null;
  main_teacher: string | null;
  student_count: number;
  students_preview: ClassroomStudentPreviewDto[];
  created_at: string;
  updated_at: string;
}

// Bonus / Penalty / Punishment
interface ReturnBonusTypeDto {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface ReturnPenaltyTypeDto {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface ReturnPunishmentTypeDto {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface ReturnBonusDto {
  id: string;
  student_id: string;
  student_first_name: string;
  student_last_name: string;
  bonus_type_id: string;
  bonus_type_name: string;
  points: number;
  created_at: string;
  used_at: string | null;
}

interface ReturnPenaltyDto {
  id: string;
  student_id: string;
  student_first_name: string;
  student_last_name: string;
  penalty_type_id: string;
  penalty_type_name: string;
  created_at: string;
}

interface ReturnPunishmentDto {
  id: string;
  student_id: string;
  student_first_name: string;
  student_last_name: string;
  punishment_type_id: string;
  punishment_type_name: string;
  triggering_rule_id: string | null;
  triggering_rule_name: string | null;
  automated: boolean;
  created_at: string;
  due_at: string;
  resolved_at: string | null;
}

// Rules
interface ReturnRuleDto {
  id: string;
  name: string;
  resulting_punishment_type_id: string;
  resulting_punishment_type_name: string;
  penalty_type_id: string;
  penalty_type_name: string;
  threshold: number;
  due_at_after_days: number;
  mode: "after" | "at" | "every";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Dashboard
interface DashboardKpisDto {
  student_count: number;
  available_bonus_points: number;
  total_bonus_points: number;
  unused_bonus_count: number;
  penalty_count: number;
  total_punishment_count: number;
  overdue_punishment_count: number;
  pending_punishment_count: number;
}

interface ReturnDashboardDto {
  kpis: DashboardKpisDto;
  recent_penalties: ReturnPenaltyDto[];
  recent_bonuses: ReturnBonusDto[];
  pending_punishments: ReturnPunishmentDto[];
}

// Health
interface HealthCheckDto {
  status: "healthy" | "unhealthy";
  environment: string;
  version: string;
  services: Record<string, string>;
}

// Import
interface StudentImportRowErrorDto {
  row: number;
  field: string;
  message: string;
  value?: string;
}

interface StudentImportSummaryDto {
  rows_total: number;
  rows_processed: number;
  classrooms_created: number;
  classrooms_existing: number;
  students_created: number;
  students_existing: number;
  links_created: number;
  links_existing: number;
  rows_failed: number;
}

interface StudentImportResultDto {
  summary: StudentImportSummaryDto;
  errors: StudentImportRowErrorDto[];
}
```

## 3) Reference complete des endpoints

## 3.1 Health

### GET `/v1/health`
- Auth: non
- URL params: aucun
- Query params: aucun
- Body: aucun
- 200: `HealthCheckDto`
- 503: `HealthCheckDto` si DB down (`status="unhealthy"`)

Exemple 200:
```json
{
  "status": "healthy",
  "environment": "development",
  "version": "dev",
  "services": {
    "database": "healthy"
  }
}
```

## 3.2 Auth

### GET `/v1/auth/register/status`
- Auth: non
- Body: aucun
- 200: `RegisterStatusResponseDto`

Exemple:
```json
{
  "register_allowed": true
}
```

### POST `/v1/auth/register`
- Auth: non
- Body:
```json
{
  "email": "teacher@school.test",
  "first_name": "Ada",
  "last_name": "Lovelace",
  "password": "super-secret-123"
}
```
- 201: `ReturnUserDto`
- Erreurs: `register_not_allowed`, `validation_failed`, `invalid_request_body`, `conflict`

### POST `/v1/auth/login`
- Auth: non
- Body:
```json
{
  "email": "teacher@school.test",
  "password": "super-secret-123"
}
```
- 200: `LoginResponseDto`
- Side effect: set-cookie `refresh_token`
- Erreurs: `validation_failed`, `invalid_request_body`, `invalid_credentials_or_user_doesnt_exist`

Exemple 200:
```json
{
  "access_token": "<jwt_access_token>"
}
```

### POST `/v1/auth/refresh`
- Auth: non (mais cookie refresh obligatoire)
- Body: aucun
- 200: `RefreshResponseDto`
- Side effect: rotation du cookie refresh
- Erreurs: `unauthorized`, `jwt_invalid_token`, `jwt_expired`, `jwt_invalid_signing_method`

### POST `/v1/auth/logout`
- Auth: non
- Body: aucun
- 204: no content
- Side effect: suppression cookie refresh

### DELETE `/v1/auth/refresh-tokens`
- Auth: oui (`Bearer`)
- Body: aucun
- 204: no content
- Side effect: revoke tous les refresh tokens de l utilisateur
- Erreurs: `unauthorized`, `jwt_invalid_token`, `jwt_expired`, `jwt_invalid_signing_method`

## 3.3 User

### GET `/v1/user/me`
- Auth: oui
- Body: aucun
- 200: `ReturnUserDto`
- Erreurs: `unauthorized`

## 3.4 Students

### POST `/v1/students/`
- Auth: oui
- Body:
```json
{
  "first_name": "Jean",
  "last_name": "Dupont"
}
```
- 201: `ReturnStudentDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### POST `/v1/students/import`
- Auth: oui
- Body: `multipart/form-data` avec champ fichier `file` (`.csv` ou `.xlsx`)
- 200: `StudentImportResultDto`
- Erreurs: `import_file_missing`, `import_file_invalid`, `import_template_invalid`, `import_validation_failed`, `unauthorized`

Exemple curl:
```bash
curl -X POST "http://localhost:8080/v1/students/import" \
  -H "Authorization: Bearer <token>" \
  -F "file=@students.xlsx"
```

### GET `/v1/students/`
- Auth: oui
- Query params:
  - `page` (optionnel)
  - `search` (optionnel, recherche sur `first_name + last_name`)
- Exemple URL:
  - `/v1/students/?search=jean%20dupont&page=2`
- 200: `PaginatedResponse<ReturnStudentDto>`
- Erreurs: `unauthorized`

### DELETE `/v1/students/`
- Auth: oui
- Body: aucun
- 204: no content
- Erreurs: `unauthorized`

### GET `/v1/students/{student_id}`
- Auth: oui
- Path params:
  - `student_id` (uuid)
- 200: `ReturnStudentDto`
- Erreurs: `student_not_found`, `not_found` (uuid path invalide), `unauthorized`

### GET `/v1/students/{student_id}/kpis`
- Auth: oui
- 200: `StudentKpisDto`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/history`
- Auth: oui
- Query params:
  - `page` (optionnel)
- Exemple URL:
  - `/v1/students/11111111-1111-1111-1111-111111111111/history?page=1`
- 200: `PaginatedResponse<StudentHistoryItemDto>`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### PUT `/v1/students/{student_id}`
- Auth: oui
- Body (partiel):
```json
{
  "first_name": "Jeanne",
  "last_name": "Durand"
}
```
- 200: `ReturnStudentDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/students/{student_id}`
- Auth: oui
- 204: no content
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/classrooms`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnClassroomDto>`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/bonuses`
- Auth: oui
- Query params:
  - `page`
  - `state` optionnel: `used|unused`
- Exemple URL:
  - `/v1/students/{student_id}/bonuses?state=unused&page=1`
- 200: `PaginatedResponse<ReturnBonusDto>`
- Erreurs: `student_not_found`, `malformed_parameter`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/penalties`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnPenaltyDto>`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/punishments`
- Auth: oui
- Query params:
  - `page`
  - `state` optionnel: `pending|resolved`
- Exemple URL:
  - `/v1/students/{student_id}/punishments?state=pending&page=1`
- 200: `PaginatedResponse<ReturnPunishmentDto>`
- Erreurs: `student_not_found`, `malformed_parameter`, `not_found`, `unauthorized`

## 3.5 Classrooms

### POST `/v1/classrooms/`
- Auth: oui
- Body:
```json
{
  "name": "6A",
  "year": "2025-2026",
  "main_teacher": "Mme Martin"
}
```
- 201: `ReturnClassroomDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### GET `/v1/classrooms/`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnClassroomDto>`
- Erreurs: `unauthorized`

### DELETE `/v1/classrooms/`
- Auth: oui
- 204: no content
- Erreurs: `unauthorized`

### GET `/v1/classrooms/{classroom_id}`
- Auth: oui
- 200: `ReturnClassroomDto`
- Erreurs: `classroom_not_found`, `not_found`, `unauthorized`

### GET `/v1/classrooms/{classroom_id}/kpis`
- Auth: oui
- 200: `DashboardKpisDto`
- Erreurs: `classroom_not_found`, `not_found`, `unauthorized`

### PUT `/v1/classrooms/{classroom_id}`
- Auth: oui
- Body partiel:
```json
{
  "name": "6B",
  "year": "2026-2027",
  "main_teacher": "M. Bernard"
}
```
- 200: `ReturnClassroomDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `classroom_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/classrooms/{classroom_id}`
- Auth: oui
- 204: no content
- Erreurs: `classroom_not_found`, `not_found`, `unauthorized`

### POST `/v1/classrooms/{classroom_id}/students`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111"
}
```
- 204: no content
- Erreurs: `invalid_request_body`, `validation_failed`, `student_classroom_relation_exists`, `student_or_classroom_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/classrooms/{classroom_id}/students/{student_id}`
- Auth: oui
- 204: no content
- Erreurs: `student_or_classroom_not_found`, `not_found`, `unauthorized`

### GET `/v1/classrooms/{classroom_id}/students`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnStudentDto>`
- Erreurs: `classroom_not_found`, `not_found`, `unauthorized`

## 3.6 Bonus types

### POST `/v1/bonus-types/`
- Auth: oui
- Body:
```json
{
  "name": "Participation"
}
```
- 201: `ReturnBonusTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### GET `/v1/bonus-types/`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnBonusTypeDto>`
- Erreurs: `unauthorized`

### GET `/v1/bonus-types/{bonus_type_id}`
- Auth: oui
- 200: `ReturnBonusTypeDto`
- Erreurs: `bonus_type_not_found`, `not_found`, `unauthorized`

### PUT `/v1/bonus-types/{bonus_type_id}`
- Auth: oui
- Body partiel:
```json
{
  "name": "Aide aux autres"
}
```
- 200: `ReturnBonusTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `bonus_type_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/bonus-types/{bonus_type_id}`
- Auth: oui
- 204: no content
- Erreurs: `bonus_type_not_found`, `not_found`, `unauthorized`

## 3.7 Bonuses

### POST `/v1/bonuses/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "bonus_type_id": "22222222-2222-2222-2222-222222222222",
  "points": 2.5
}
```
- 201: `ReturnBonusDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `bonus_type_not_found`, `unauthorized`

### GET `/v1/bonuses/`
- Auth: oui
- Query params (metier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `bonus_type_id` (uuid)
  - `state` (`used|unused`)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - `page`
- Exemple URL:
  - `/v1/bonuses/?state=unused&classroom_id=33333333-3333-3333-3333-333333333333&created_from=2026-02-01&created_to=2026-02-28&page=1`
- 200: `PaginatedResponse<ReturnBonusDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/bonuses/{bonus_id}`
- Auth: oui
- 200: `ReturnBonusDto`
- Erreurs: `bonus_not_found`, `not_found`, `unauthorized`

### POST `/v1/bonuses/{bonus_id}/use`
- Auth: oui
- Body: aucun
- 200: `ReturnBonusDto` (avec `used_at` non null)
- Erreurs: `bonus_not_found`, `bonus_already_used`, `not_found`, `unauthorized`

### DELETE `/v1/bonuses/{bonus_id}`
- Auth: oui
- 204: no content
- Erreurs: `bonus_not_found`, `not_found`, `unauthorized`

## 3.8 Penalty types

### POST `/v1/penalty-types/`
- Auth: oui
- Body:
```json
{
  "name": "Bavardage"
}
```
- 201: `ReturnPenaltyTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### GET `/v1/penalty-types/`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnPenaltyTypeDto>`
- Erreurs: `unauthorized`

### GET `/v1/penalty-types/{penalty_type_id}`
- Auth: oui
- 200: `ReturnPenaltyTypeDto`
- Erreurs: `penalty_type_not_found`, `not_found`, `unauthorized`

### PUT `/v1/penalty-types/{penalty_type_id}`
- Auth: oui
- Body partiel:
```json
{
  "name": "Retard"
}
```
- 200: `ReturnPenaltyTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `penalty_type_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/penalty-types/{penalty_type_id}`
- Auth: oui
- 204: no content
- Erreurs: `penalty_type_not_found`, `not_found`, `unauthorized`

## 3.9 Penalties

### POST `/v1/penalties/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "penalty_type_id": "44444444-4444-4444-4444-444444444444"
}
```
- 201: `ReturnPenaltyDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `penalty_type_not_found`, `unauthorized`

### GET `/v1/penalties/`
- Auth: oui
- Query params (metier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `penalty_type_id` (uuid)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - `page`
- Exemple URL:
  - `/v1/penalties/?penalty_type_id=44444444-4444-4444-4444-444444444444&classroom_id=33333333-3333-3333-3333-333333333333&created_from=2026-02-01&created_to=2026-02-28&page=1`
- 200: `PaginatedResponse<ReturnPenaltyDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/penalties/{penalty_id}`
- Auth: oui
- 200: `ReturnPenaltyDto`
- Erreurs: `penalty_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/penalties/{penalty_id}`
- Auth: oui
- 204: no content
- Erreurs: `penalty_not_found`, `not_found`, `unauthorized`

## 3.10 Punishment types

### POST `/v1/punishment-types/`
- Auth: oui
- Body:
```json
{
  "name": "Exercice supplementaire"
}
```
- 201: `ReturnPunishmentTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### GET `/v1/punishment-types/`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnPunishmentTypeDto>`
- Erreurs: `unauthorized`

### GET `/v1/punishment-types/{punishment_type_id}`
- Auth: oui
- 200: `ReturnPunishmentTypeDto`
- Erreurs: `punishment_type_not_found`, `not_found`, `unauthorized`

### PUT `/v1/punishment-types/{punishment_type_id}`
- Auth: oui
- Body partiel:
```json
{
  "name": "Presentation orale"
}
```
- 200: `ReturnPunishmentTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `punishment_type_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/punishment-types/{punishment_type_id}`
- Auth: oui
- 204: no content
- Erreurs: `punishment_type_not_found`, `not_found`, `unauthorized`

## 3.11 Punishments

### POST `/v1/punishments/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "punishment_type_id": "55555555-5555-5555-5555-555555555555",
  "due_at": "2026-03-15T18:00:00Z"
}
```
- 201: `ReturnPunishmentDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `punishment_type_not_found`, `unauthorized`

### GET `/v1/punishments/`
- Auth: oui
- Query params (metier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `punishment_type_id` (uuid)
  - `state` (`pending|resolved`)
  - `automated` (`true|false`)
  - `overdue` (`true|false`)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - `due_from` (`YYYY-MM-DD`)
  - `due_to` (`YYYY-MM-DD`)
  - `page`
- Exemple URL:
  - `/v1/punishments/?state=pending&overdue=true&automated=false&classroom_id=33333333-3333-3333-3333-333333333333&due_from=2026-03-01&due_to=2026-03-31&page=1`
- 200: `PaginatedResponse<ReturnPunishmentDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/punishments/{punishment_id}`
- Auth: oui
- 200: `ReturnPunishmentDto`
- Erreurs: `punishment_not_found`, `not_found`, `unauthorized`

### POST `/v1/punishments/{punishment_id}/resolve`
- Auth: oui
- Body: aucun
- 200: `ReturnPunishmentDto` (avec `resolved_at` non null)
- Erreurs: `punishment_not_found`, `punishment_already_resolved`, `not_found`, `unauthorized`

### DELETE `/v1/punishments/{punishment_id}`
- Auth: oui
- 204: no content
- Erreurs: `punishment_not_found`, `not_found`, `unauthorized`

## 3.12 Rules

### POST `/v1/rules/`
- Auth: oui
- Body:
```json
{
  "name": "3 retards = punition",
  "resulting_punishment_type_id": "55555555-5555-5555-5555-555555555555",
  "penalty_type_id": "44444444-4444-4444-4444-444444444444",
  "threshold": 3,
  "due_at_after_days": 2,
  "mode": "after",
  "is_active": true
}
```
- 201: `ReturnRuleDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `punishment_type_not_found`, `penalty_type_not_found`, `unauthorized`

### GET `/v1/rules/`
- Auth: oui
- Query params: `page`
- 200: `PaginatedResponse<ReturnRuleDto>`
- Erreurs: `unauthorized`

### GET `/v1/rules/{rule_id}`
- Auth: oui
- 200: `ReturnRuleDto`
- Erreurs: `rule_not_found`, `not_found`, `unauthorized`

### PUT `/v1/rules/{rule_id}`
- Auth: oui
- Body partiel:
```json
{
  "name": "4 retards = punition",
  "threshold": 4,
  "mode": "every",
  "is_active": false
}
```
- 200: `ReturnRuleDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `rule_not_found`, `punishment_type_not_found`, `penalty_type_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/rules/{rule_id}`
- Auth: oui
- 204: no content
- Erreurs: `rule_not_found`, `not_found`, `unauthorized`

## 3.13 Dashboard

### GET `/v1/dashboard/`
- Auth: oui
- Query params:
  - `classroom_id` optionnel (uuid)
- Exemple URL:
  - `/v1/dashboard/?classroom_id=33333333-3333-3333-3333-333333333333`
- 200: `ReturnDashboardDto`
- Erreurs: `malformed_parameter`, `classroom_not_found`, `unauthorized`

Exemple 200 (tronque):
```json
{
  "kpis": {
    "student_count": 28,
    "available_bonus_points": 64.5,
    "total_bonus_points": 91,
    "unused_bonus_count": 19,
    "penalty_count": 13,
    "total_punishment_count": 5,
    "overdue_punishment_count": 1,
    "pending_punishment_count": 2
  },
  "recent_penalties": [],
  "recent_bonuses": [],
  "pending_punishments": []
}
```

## 4) Notes importantes pour le frontend IA

- Les listes `bonuses`, `penalties`, `punishments` sont maintenant basees sur filtres metier (pas de `search` texte sur l eleve).
- La recherche eleve reste sur `GET /v1/students/?search=...`.
- Pour filtrer des evenements d un eleve: utiliser `student_id`.
- Les bornes `created_to` / `due_to` sont inclusives sur la journee (backend fait `< date + 1 day`).
- `overdue=true` sur punishments signifie: `resolved_at IS NULL` ET `due_at < now()`.
- Pour les IDs path invalides (`{student_id}`, etc.), le backend renvoie `404 not_found` avec `error_details` indiquant le champ invalide.

## 5) Catalogue complet des erreurs disponibles

## 5.1 Schema unique

```json
{
  "error": "<string>",
  "error_code": 400,
  "error_details": [
    {
      "row": 12,
      "field": "created_from",
      "error": "validation_malformed_parameter:expected_yyyy-mm-dd",
      "value": "15-01-2026"
    }
  ]
}
```

## 5.2 Erreurs sans `error_details` (ou details optionnels)

### 500
- `internal_error`
```json
{ "error": "internal_error", "error_code": 500 }
```

### 404
- `not_found`
```json
{ "error": "not_found", "error_code": 404 }
```
- `student_not_found`
```json
{ "error": "student_not_found", "error_code": 404 }
```
- `bonus_type_not_found`
```json
{ "error": "bonus_type_not_found", "error_code": 404 }
```
- `penalty_type_not_found`
```json
{ "error": "penalty_type_not_found", "error_code": 404 }
```
- `rule_not_found`
```json
{ "error": "rule_not_found", "error_code": 404 }
```
- `bonus_not_found`
```json
{ "error": "bonus_not_found", "error_code": 404 }
```
- `penalty_not_found`
```json
{ "error": "penalty_not_found", "error_code": 404 }
```
- `punishment_type_not_found`
```json
{ "error": "punishment_type_not_found", "error_code": 404 }
```
- `punishment_not_found`
```json
{ "error": "punishment_not_found", "error_code": 404 }
```
- `classroom_not_found`
```json
{ "error": "classroom_not_found", "error_code": 404 }
```
- `student_or_classroom_not_found`
```json
{ "error": "student_or_classroom_not_found", "error_code": 404 }
```

### 401
- `unauthorized`
```json
{ "error": "unauthorized", "error_code": 401 }
```
- `register_not_allowed`
```json
{ "error": "register_not_allowed", "error_code": 401 }
```
- `invalid_credentials_or_user_doesnt_exist`
```json
{ "error": "invalid_credentials_or_user_doesnt_exist", "error_code": 401 }
```
- `jwt_invalid_signing_method`
```json
{ "error": "jwt_invalid_signing_method", "error_code": 401 }
```
- `jwt_invalid_token`
```json
{ "error": "jwt_invalid_token", "error_code": 401 }
```
- `jwt_expired`
```json
{ "error": "jwt_expired", "error_code": 401 }
```

### 409
- `punishment_already_resolved`
```json
{ "error": "punishment_already_resolved", "error_code": 409 }
```
- `bonus_already_used`
```json
{ "error": "bonus_already_used", "error_code": 409 }
```
- `student_classroom_relation_exists`
```json
{ "error": "student_classroom_relation_exists", "error_code": 409 }
```

## 5.3 Erreurs avec `error_details` importantes

### `conflict` (email deja pris) - 409
```json
{
  "error": "conflict",
  "error_code": 409,
  "error_details": [
    {
      "field": "email",
      "error": "validation_email_already_exists"
    }
  ]
}
```

### `invalid_request_body` - 400
Cas classiques:
- type JSON invalide (`expected_string`, `expected_number`, etc.)
- champ JSON inconnu
- UUID body invalide
- datetime RFC3339 invalide

Exemples:
```json
{
  "error": "invalid_request_body",
  "error_code": 400,
  "error_details": [
    {
      "field": "student_id",
      "error": "validation_malformed_parameter:expected_uuid"
    }
  ]
}
```

```json
{
  "error": "invalid_request_body",
  "error_code": 400,
  "error_details": [
    {
      "field": "due_at",
      "error": "validation_malformed_parameter:expected_rfc3339_datetime"
    }
  ]
}
```

```json
{
  "error": "invalid_request_body",
  "error_code": 400,
  "error_details": [
    {
      "field": "unknown_field",
      "error": "validation_unknown_field"
    }
  ]
}
```

### `validation_failed` - 400
Exemple:
```json
{
  "error": "validation_failed",
  "error_code": 400,
  "error_details": [
    {
      "field": "Password",
      "error": "validation_min_length:8"
    },
    {
      "field": "Email",
      "error": "validation_invalid_email"
    }
  ]
}
```

### `malformed_parameter` - 400
Cas classiques:
- UUID query invalide (`student_id`, `classroom_id`, etc.)
- enum invalide (`state`)
- bool invalide (`automated`, `overdue`)
- date invalide (`created_from`, `due_to`)
- plage invalide (`from > to`)

Exemples:
```json
{
  "error": "malformed_parameter",
  "error_code": 400,
  "error_details": [
    {
      "field": "state",
      "error": "validation_malformed_parameter:expected_pending_or_resolved"
    }
  ]
}
```

```json
{
  "error": "malformed_parameter",
  "error_code": 400,
  "error_details": [
    {
      "field": "created_from",
      "error": "validation_malformed_parameter:expected_created_from_lte_created_to"
    }
  ]
}
```

### `not_found` avec detail UUID path invalide - 404
Exemple:
```json
{
  "error": "not_found",
  "error_code": 404,
  "error_details": [
    {
      "field": "student_id",
      "error": "validation_malformed_parameter:expected_uuid"
    }
  ]
}
```

## 5.4 Erreurs import students

### `import_file_missing` - 400
```json
{
  "error": "import_file_missing",
  "error_code": 400,
  "error_details": [
    {
      "field": "file",
      "error": "file_field_is_required"
    }
  ]
}
```

### `import_file_invalid` - 400
Exemples `error_details` possibles:
- `expected_multipart_form_data`
- `failed_to_parse_multipart_form`
- `unsupported_file_type`
- `failed_to_read_xlsx`
- `failed_to_read_xlsx_rows`
- `failed_to_read_csv`

Exemple:
```json
{
  "error": "import_file_invalid",
  "error_code": 400,
  "error_details": [
    {
      "field": "file",
      "error": "unsupported_file_type",
      "value": ".txt"
    }
  ]
}
```

### `import_template_invalid` - 400
Exemples `error_details` possibles:
- `missing_headers_row`
- `missing_header_eleves`
- `missing_header_classes`

Exemple:
```json
{
  "error": "import_template_invalid",
  "error_code": 400,
  "error_details": [
    {
      "field": "headers",
      "error": "missing_header_eleves"
    }
  ]
}
```

### `import_validation_failed` - 400
- Erreurs ligne par ligne sur le contenu fichier
- `row` est renseigne

Exemple:
```json
{
  "error": "import_validation_failed",
  "error_code": 400,
  "error_details": [
    {
      "row": 4,
      "field": "eleves",
      "error": "student name format must be 'NOM Prenom'",
      "value": "Jean"
    },
    {
      "row": 4,
      "field": "classes",
      "error": "at least one classroom is required",
      "value": ""
    }
  ]
}
```
