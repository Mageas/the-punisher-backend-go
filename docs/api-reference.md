# API Reference - The Punisher Backend (guide technique frontend IA)

## 1) Conventions globales

### 1.1 Base URL
- Local: `http://localhost:8080`
- Préfixe API: toutes les routes métier sont sous `/v1`

### 1.2 Authentification
- Routes publiques:
  - `GET /v1/health`
  - `GET /v1/auth/register/status`
  - `POST /v1/auth/register`
  - `GET /v1/auth/confirm-email`
  - `POST /v1/auth/confirm-email/resend`
  - `POST /v1/auth/forgot-password`
  - `POST /v1/auth/reset-password`
  - `POST /v1/auth/login`
  - `POST /v1/auth/refresh`
  - `POST /v1/auth/logout`
- Routes protégées: toutes les autres routes `/v1/**` exigent `Authorization: Bearer <access_token>` (dont `POST /v1/auth/change-password` et `DELETE /v1/auth/refresh-tokens`).

Flux frontend recommandé:
1. `POST /v1/auth/login` avec email/password
2. Récupérer `access_token` dans le body
3. Stocker `access_token` côté client (mémoire de préférence)
4. Utiliser `Authorization: Bearer <token>`
5. Quand 401/expiration: `POST /v1/auth/refresh` (cookie `refresh_token` automatiquement envoyé si `credentials: include`)

Notes cookie refresh:
- Nom: `refresh_token`
- `HttpOnly`, `SameSite=Strict`
- Path cookie: `/v1/auth`
- Le frontend doit utiliser `credentials: 'include'` pour `login`, `refresh`, `logout`, `change-password`.

### 1.3 Types de données
- UUID: format canonique (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
- Date liste: `YYYY-MM-DD` (ex: `2026-02-28`)
- Date body (`start_date`, `end_date`, `date`): `YYYY-MM-DD`
- DateTime body (`occurred_at`, `due_at`): `RFC3339` (ex: `2026-03-15T18:00:00Z`)
- DateTime response (`*_at`): `RFC3339` en `UTC` (suffixe `Z`), normalise a la precision microseconde
- Heure body/response (`start_time`, `end_time`): `HH:MM`
- Weekday body/response: `monday|tuesday|wednesday|thursday|friday|saturday|sunday`
- Week pattern body/response: `every_week|even_weeks|odd_weeks`
- Bool query: `true` ou `false`

### 1.4 Pagination
- Parametres:
  - `page` (int > 0)
  - `item_per_page` (optionnel, int)
- Taille par defaut backend: `20`
- Bornes `item_per_page`: min `5`, max `50`
- Si `page` invalide/absent: fallback page `1` (pas d erreur)
- Si `item_per_page` invalide/absent: fallback `20` (pas d erreur)
- Si `item_per_page` < 5: force a `5`
- Si `item_per_page` > 50: force a `50`

Format réponse paginée:
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

## 2) Contrats JSON (schémas frontend)

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

interface ChangePasswordRequestDto {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

interface ChangePasswordResponseDto {
  status: "password_changed";
}

interface RegisterStatusResponseDto {
  register_allowed: boolean;
}

interface ConfirmEmailResponseDto {
  status: "email_confirmed";
}

interface ResendConfirmEmailRequestDto {
  email: string;
}

interface ResendConfirmEmailResponseDto {
  status: "confirmation_email_sent_if_needed";
}

interface ForgotPasswordRequestDto {
  email: string;
}

interface ForgotPasswordResponseDto {
  status: "password_reset_email_sent_if_needed";
}

interface ResetPasswordRequestDto {
  token: string;
  new_password: string;
  confirm_password: string;
}

interface ResetPasswordResponseDto {
  status: "password_reset";
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
  created_at: string;
  occurred_at: string;
  evaluation_label: string;
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

interface ScheduleSlotClassroomDto {
  id: string;
  name: string;
}

interface ReturnScheduleSlotDto {
  id: string;
  weekday: string;
  start_time: string;
  end_time: string;
  week_pattern: "every_week" | "even_weeks" | "odd_weeks";
  classrooms: ScheduleSlotClassroomDto[];
  created_at: string;
  updated_at: string;
}

interface ReturnScheduleExceptionDto {
  id: string;
  type: "vacation" | "public_holiday";
  start_date: string;
  end_date: string;
  created_at: string;
  updated_at: string;
}

interface NextLessonDto {
  date: string;
  start_time: string;
  end_time: string;
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

interface RequestBonusDto {
  student_id: string;
  bonus_type_id: string;
  points: number;
  occurred_at?: string; // RFC3339
  evaluation_label?: string;
}

interface UpdateBonusDto {
  points?: number; // > 0
  occurred_at?: string; // RFC3339
  evaluation_label?: string; // empty string clears the label
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
  occurred_at: string;
  evaluation_label: string;
  used_at: string | null;
}

interface RequestPenaltyDto {
  student_id: string;
  penalty_type_id: string;
  occurred_at?: string; // RFC3339
  evaluation_label?: string;
}

interface UpdatePenaltyDto {
  occurred_at?: string; // RFC3339
  evaluation_label?: string; // empty string clears the label
}

interface ReturnPenaltyDto {
  id: string;
  student_id: string;
  student_first_name: string;
  student_last_name: string;
  penalty_type_id: string;
  penalty_type_name: string;
  created_at: string;
  occurred_at: string;
  evaluation_label: string;
}

interface RequestPunishmentDto {
  student_id: string;
  punishment_type_id: string;
  due_at: string; // RFC3339
  occurred_at?: string; // RFC3339
  evaluation_label?: string;
}

interface UpdatePunishmentDto {
  occurred_at?: string; // RFC3339
  evaluation_label?: string; // empty string clears the label
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
  occurred_at: string;
  evaluation_label: string;
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
  key: string;
  value?: string;
  error_details?: string[];
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

## 3) Référence complète des endpoints

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
- Side effect: envoi d'un email de confirmation avec un lien vers `GET /v1/auth/confirm-email?token=...`
- Erreurs: `register_not_allowed`, `validation_failed`, `invalid_request_body`, `conflict`

### GET `/v1/auth/confirm-email`
- Auth: non
- Query params:
  - `token` (obligatoire, JWT de confirmation)
- 200: `ConfirmEmailResponseDto`
- Erreurs: `email_confirmation_token_missing`, `email_confirmation_token_invalid`, `email_confirmation_token_expired`, `email_confirmation_token_already_used`, `email_already_verified`, `email_confirmation_user_not_found`

Exemple 200:
```json
{
  "status": "email_confirmed"
}
```

### POST `/v1/auth/confirm-email/resend`
- Auth: non
- Body:
```json
{
  "email": "teacher@school.test"
}
```
- 200: `ResendConfirmEmailResponseDto`
- Comportement: réponse neutre. Si l'utilisateur n'existe pas ou a déjà confirmé son email, la réponse reste 200.
- Erreurs: `validation_failed`, `invalid_request_body`

Exemple 200:
```json
{
  "status": "confirmation_email_sent_if_needed"
}
```

### POST `/v1/auth/forgot-password`
- Auth: non
- Body:
```json
{
  "email": "teacher@school.test"
}
```
- 200: `ForgotPasswordResponseDto`
- Comportement: réponse neutre. Si l'utilisateur n'existe pas, la réponse reste 200.
- Side effects (si utilisateur trouvé):
  - invalide les anciens tokens de reset encore actifs
  - génère un nouveau token signé temporaire
  - envoie un email avec un lien de reset contenant `?token=...`
- Erreurs: `validation_failed`, `invalid_request_body`

Exemple 200:
```json
{
  "status": "password_reset_email_sent_if_needed"
}
```

### POST `/v1/auth/reset-password`
- Auth: non
- Body:
```json
{
  "token": "<jwt_password_reset_token>",
  "new_password": "NewSecurePass2@",
  "confirm_password": "NewSecurePass2@"
}
```
- 200: `ResetPasswordResponseDto`
- Politique mot de passe: même validation que l'inscription (`required,min=8`).
- Side effects:
  - met à jour `password_hash`, `password_changed_at`, `updated_at`
  - invalide tous les refresh tokens de l'utilisateur
  - marque le token de reset utilisé (et invalide les autres tokens actifs restants)
- Erreurs: `validation_failed`, `invalid_request_body`, `password_reset_token_missing`, `password_reset_token_invalid`, `password_reset_token_expired`, `password_reset_token_already_used`, `password_reset_user_not_found`

Exemple 200:
```json
{
  "status": "password_reset"
}
```

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
- Erreurs: `validation_failed`, `invalid_request_body`, `invalid_credentials_or_user_doesnt_exist`, `email_not_verified`

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

### POST `/v1/auth/change-password`
- Auth: oui (`Bearer`)
- Body:
```json
{
  "current_password": "CurrentPass1!",
  "new_password": "NewSecurePass2@",
  "confirm_password": "NewSecurePass2@"
}
```
- 200: `ChangePasswordResponseDto`
- Politique mot de passe: même validation que l'inscription (`required,min=8`).
- Side effects:
  - met à jour `password_hash`, `password_changed_at`, `updated_at`
  - invalide tous les refresh tokens de l'utilisateur (session courante incluse)
  - supprime le cookie refresh (`refresh_token`)
- Erreurs: `unauthorized`, `invalid_current_password`, `validation_failed`, `invalid_request_body`, `jwt_invalid_token`, `jwt_expired`, `jwt_invalid_signing_method`

Exemple 200:
```json
{
  "status": "password_changed"
}
```

### DELETE `/v1/auth/refresh-tokens`
- Auth: oui (`Bearer`)
- Body: aucun
- 204: no content
- Side effect: révoque tous les refresh tokens de l'utilisateur
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
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `first_name + last_name`)
- Exemple URL:
  - `/v1/students/?search=jean%20dupont&page=2&item_per_page=10`
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
  - `item_per_page` (optionnel, min 5, max 50)
- Tri: `occurred_at` décroissant (puis `id` décroissant)
- Exemple URL:
  - `/v1/students/11111111-1111-1111-1111-111111111111/history?page=1&item_per_page=25`
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
- Query params: `page`, `item_per_page`
- 200: `PaginatedResponse<ReturnClassroomDto>`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/bonuses`
- Auth: oui
- Query params:
  - `page`
  - `item_per_page` (optionnel, min 5, max 50)
  - `state` optionnel: `used|unused`
- Exemple URL:
  - `/v1/students/{student_id}/bonuses?state=unused&page=1&item_per_page=10`
- 200: `PaginatedResponse<ReturnBonusDto>`
- Erreurs: `student_not_found`, `malformed_parameter`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/penalties`
- Auth: oui
- Query params: `page`, `item_per_page`
- 200: `PaginatedResponse<ReturnPenaltyDto>`
- Erreurs: `student_not_found`, `not_found`, `unauthorized`

### GET `/v1/students/{student_id}/punishments`
- Auth: oui
- Query params:
  - `page`
  - `item_per_page` (optionnel, min 5, max 50)
  - `state` optionnel: `pending|resolved`
- Exemple URL:
  - `/v1/students/{student_id}/punishments?state=pending&page=1&item_per_page=10`
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
- Query params:
  - `page` (optionnel)
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `name`)
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

### GET `/v1/classrooms/{classroom_id}/next-lessons`
- Auth: oui
- 200: `NextLessonDto[]` (max `5` items)
- Regles:
  - depart au prochain jour complet
  - jour courant toujours exclu
  - prise en compte de la parite ISO (`even_weeks` / `odd_weeks`)
  - prise en compte des exceptions globales utilisateur
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
- Query params:
  - `page` (optionnel)
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `first_name + last_name`)
- 200: `PaginatedResponse<ReturnStudentDto>`
- Erreurs: `classroom_not_found`, `not_found`, `unauthorized`

## 3.6 Schedule

### POST `/v1/schedule/slots/`
- Auth: oui
- Body:
```json
{
  "weekday": "monday",
  "start_time": "08:30",
  "end_time": "09:25",
  "week_pattern": "every_week",
  "classroom_ids": [
    "11111111-1111-1111-1111-111111111111",
    "22222222-2222-2222-2222-222222222222"
  ]
}
```
- 201: `ReturnScheduleSlotDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `classroom_not_found`, `schedule_slot_conflict`, `unauthorized`

### GET `/v1/schedule/slots/`
- Auth: oui
- 200: `ReturnScheduleSlotDto[]`
- Erreurs: `unauthorized`

### GET `/v1/schedule/slots/{schedule_slot_id}`
- Auth: oui
- 200: `ReturnScheduleSlotDto`
- Erreurs: `schedule_slot_not_found`, `not_found`, `unauthorized`

### PUT `/v1/schedule/slots/{schedule_slot_id}`
- Auth: oui
- Body partiel:
```json
{
  "weekday": "tuesday",
  "start_time": "10:00",
  "end_time": "11:00",
  "week_pattern": "odd_weeks",
  "classroom_ids": [
    "11111111-1111-1111-1111-111111111111"
  ]
}
```
- 200: `ReturnScheduleSlotDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `classroom_not_found`, `schedule_slot_not_found`, `schedule_slot_conflict`, `not_found`, `unauthorized`

### DELETE `/v1/schedule/slots/{schedule_slot_id}`
- Auth: oui
- 204: no content
- Erreurs: `schedule_slot_not_found`, `not_found`, `unauthorized`

### POST `/v1/schedule/exceptions/`
- Auth: oui
- Body:
```json
{
  "type": "vacation",
  "start_date": "2026-04-06",
  "end_date": "2026-04-10"
}
```
- 201: `ReturnScheduleExceptionDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `schedule_exception_overlap`, `unauthorized`

### GET `/v1/schedule/exceptions/`
- Auth: oui
- 200: `ReturnScheduleExceptionDto[]`
- Erreurs: `unauthorized`

### GET `/v1/schedule/exceptions/{schedule_exception_id}`
- Auth: oui
- 200: `ReturnScheduleExceptionDto`
- Erreurs: `schedule_exception_not_found`, `not_found`, `unauthorized`

### PUT `/v1/schedule/exceptions/{schedule_exception_id}`
- Auth: oui
- Body partiel:
```json
{
  "type": "public_holiday",
  "start_date": "2026-05-01",
  "end_date": "2026-05-01"
}
```
- 200: `ReturnScheduleExceptionDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `schedule_exception_not_found`, `schedule_exception_overlap`, `not_found`, `unauthorized`

### DELETE `/v1/schedule/exceptions/{schedule_exception_id}`
- Auth: oui
- 204: no content
- Erreurs: `schedule_exception_not_found`, `not_found`, `unauthorized`

## 3.7 Bonus types

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
- Query params:
  - `page` (optionnel)
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `name`)
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

## 3.8 Bonuses

### POST `/v1/bonuses/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "bonus_type_id": "22222222-2222-2222-2222-222222222222",
  "points": 2.5,
  "occurred_at": "2026-02-10T09:00:00Z",
  "evaluation_label": "Participation"
}
```
- 201: `ReturnBonusDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `bonus_type_not_found`, `unauthorized`

### GET `/v1/bonuses/`
- Auth: oui
- Query params (métier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `bonus_type_id` (uuid)
  - `state` (`used|unused`)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - Note: `created_from` / `created_to` filtrent `occurred_at` (date métier)
  - `page`
  - `item_per_page` (optionnel, min 5, max 50)
- Tri: `occurred_at` décroissant (puis `id` décroissant)
- Exemple URL:
  - `/v1/bonuses/?state=unused&classroom_id=33333333-3333-3333-3333-333333333333&created_from=2026-02-01&created_to=2026-02-28&page=1&item_per_page=10`
- 200: `PaginatedResponse<ReturnBonusDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/bonuses/{bonus_id}`
- Auth: oui
- 200: `ReturnBonusDto`
- Erreurs: `bonus_not_found`, `not_found`, `unauthorized`

### PUT `/v1/bonuses/{bonus_id}`
- Auth: oui
- Body partiel:
```json
{
  "points": 3.5,
  "occurred_at": "2026-02-09T08:00:00Z",
  "evaluation_label": ""
}
```
- `points` met a jour la valeur des points (doit etre `> 0`).
- `evaluation_label: ""` supprime explicitement le libellé.
- 200: `ReturnBonusDto`
- Erreurs: `invalid_request_body`, `bonus_not_found`, `not_found`, `unauthorized`

### POST `/v1/bonuses/{bonus_id}/use`
- Auth: oui
- Body: aucun
- 200: `ReturnBonusDto` (avec `used_at` non null)
- Erreurs: `bonus_not_found`, `bonus_already_used`, `not_found`, `unauthorized`

### DELETE `/v1/bonuses/{bonus_id}`
- Auth: oui
- 204: no content
- Erreurs: `bonus_not_found`, `not_found`, `unauthorized`

## 3.9 Penalty types

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
- Query params:
  - `page` (optionnel)
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `name`)
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

## 3.10 Penalties

### POST `/v1/penalties/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "penalty_type_id": "44444444-4444-4444-4444-444444444444",
  "occurred_at": "2026-02-10T09:00:00Z",
  "evaluation_label": "Retard"
}
```
- 201: `ReturnPenaltyDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `penalty_type_not_found`, `unauthorized`

### GET `/v1/penalties/`
- Auth: oui
- Query params (métier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `penalty_type_id` (uuid)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - Note: `created_from` / `created_to` filtrent `occurred_at` (date métier)
  - `page`
  - `item_per_page` (optionnel, min 5, max 50)
- Tri: `occurred_at` décroissant (puis `id` décroissant)
- Exemple URL:
  - `/v1/penalties/?penalty_type_id=44444444-4444-4444-4444-444444444444&classroom_id=33333333-3333-3333-3333-333333333333&created_from=2026-02-01&created_to=2026-02-28&page=1&item_per_page=10`
- 200: `PaginatedResponse<ReturnPenaltyDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/penalties/{penalty_id}`
- Auth: oui
- 200: `ReturnPenaltyDto`
- Erreurs: `penalty_not_found`, `not_found`, `unauthorized`

### PUT `/v1/penalties/{penalty_id}`
- Auth: oui
- Body partiel:
```json
{
  "occurred_at": "2026-02-09T08:00:00Z",
  "evaluation_label": ""
}
```
- `evaluation_label: ""` supprime explicitement le libellé.
- 200: `ReturnPenaltyDto`
- Erreurs: `invalid_request_body`, `penalty_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/penalties/{penalty_id}`
- Auth: oui
- 204: no content
- Erreurs: `penalty_not_found`, `not_found`, `unauthorized`

## 3.11 Punishment types

### POST `/v1/punishment-types/`
- Auth: oui
- Body:
```json
{
  "name": "Exercice supplémentaire"
}
```
- 201: `ReturnPunishmentTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `unauthorized`

### GET `/v1/punishment-types/`
- Auth: oui
- Query params:
  - `page` (optionnel)
  - `item_per_page` (optionnel, min 5, max 50)
  - `search` (optionnel, recherche sur `name`)
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
  "name": "Présentation orale"
}
```
- 200: `ReturnPunishmentTypeDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `punishment_type_not_found`, `not_found`, `unauthorized`

### DELETE `/v1/punishment-types/{punishment_type_id}`
- Auth: oui
- 204: no content
- Erreurs: `punishment_type_not_found`, `not_found`, `unauthorized`

## 3.12 Punishments

### POST `/v1/punishments/`
- Auth: oui
- Body:
```json
{
  "student_id": "11111111-1111-1111-1111-111111111111",
  "punishment_type_id": "55555555-5555-5555-5555-555555555555",
  "due_at": "2026-03-15T18:00:00Z",
  "occurred_at": "2026-02-10T09:00:00Z",
  "evaluation_label": "Travail non rendu"
}
```
- 201: `ReturnPunishmentDto`
- Erreurs: `validation_failed`, `invalid_request_body`, `student_not_found`, `punishment_type_not_found`, `unauthorized`

### GET `/v1/punishments/`
- Auth: oui
- Query params (métier):
  - `student_id` (uuid)
  - `classroom_id` (uuid)
  - `punishment_type_id` (uuid)
  - `state` (`pending|resolved`)
  - `automated` (`true|false`)
  - `overdue` (`true|false`)
  - `created_from` (`YYYY-MM-DD`)
  - `created_to` (`YYYY-MM-DD`)
  - Note: `created_from` / `created_to` filtrent `occurred_at` (date métier)
  - `due_from` (`YYYY-MM-DD`)
  - `due_to` (`YYYY-MM-DD`)
  - `page`
  - `item_per_page` (optionnel, min 5, max 50)
- Tri: `occurred_at` décroissant (puis `id` décroissant)
- Exemple URL:
  - `/v1/punishments/?state=pending&overdue=true&automated=false&classroom_id=33333333-3333-3333-3333-333333333333&due_from=2026-03-01&due_to=2026-03-31&page=1&item_per_page=10`
- 200: `PaginatedResponse<ReturnPunishmentDto>`
- Erreurs: `malformed_parameter`, `unauthorized`

### GET `/v1/punishments/{punishment_id}`
- Auth: oui
- 200: `ReturnPunishmentDto`
- Erreurs: `punishment_not_found`, `not_found`, `unauthorized`

### PUT `/v1/punishments/{punishment_id}`
- Auth: oui
- Body partiel:
```json
{
  "occurred_at": "2026-02-09T08:00:00Z",
  "evaluation_label": ""
}
```
- `evaluation_label: ""` supprime explicitement le libellé.
- 200: `ReturnPunishmentDto`
- Erreurs: `invalid_request_body`, `punishment_not_found`, `not_found`, `unauthorized`

### POST `/v1/punishments/{punishment_id}/resolve`
- Auth: oui
- Body: aucun
- 200: `ReturnPunishmentDto` (avec `resolved_at` non null)
- Erreurs: `punishment_not_found`, `punishment_already_resolved`, `not_found`, `unauthorized`

### DELETE `/v1/punishments/{punishment_id}`
- Auth: oui
- 204: no content
- Erreurs: `punishment_not_found`, `not_found`, `unauthorized`

## 3.13 Rules

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
- Query params: `page`, `item_per_page`
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

## 3.14 Dashboard

### GET `/v1/dashboard/`
- Auth: oui
- Query params:
  - `classroom_id` optionnel (uuid)
- Tri des listes `recent_penalties`, `recent_bonuses`, `pending_punishments`: `occurred_at` décroissant
- Exemple URL:
  - `/v1/dashboard/?classroom_id=33333333-3333-3333-3333-333333333333`
- 200: `ReturnDashboardDto`
- Erreurs: `malformed_parameter`, `classroom_not_found`, `unauthorized`

Exemple 200 (tronqué):
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

- Les listes `bonuses`, `penalties`, `punishments` sont maintenant basées sur des filtres métier (pas de `search` texte sur l'élève).
- Le tri des listes `bonuses`, `penalties`, `punishments`, de l'historique élève et du dashboard (`recent_*`, `pending_punishments`) se fait sur `occurred_at`.
- La recherche élève est disponible sur `GET /v1/students/?search=...` et `GET /v1/classrooms/{classroom_id}/students?search=...`.
- Les listes `classrooms`, `bonus-types`, `penalty-types`, `punishment-types` supportent `search` (sur `name`).
- Pour filtrer des événements d'un élève: utiliser `student_id`.
- Les bornes `created_to` / `due_to` sont inclusives sur la journée (backend fait `< date + 1 day`).
- `created_from` / `created_to` filtrent la date métier `occurred_at` (et non la date technique `created_at`).
- `created_at` reste la date technique de création; `occurred_at` est la date métier de l'événement.
- `overdue=true` sur punishments signifie: `resolved_at IS NULL` ET `due_at < now()`.
- Pour les IDs path invalides (`{student_id}`, etc.), le backend renvoie `404 not_found` avec `error_details` indiquant le champ invalide.

## 5) Catalogue complet des erreurs disponibles

## 5.1 Schéma unique

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

## 5.2 Erreurs sans `error_details` (ou détails optionnels)

### 400
- `malformed_parameter`
```json
{ "error": "malformed_parameter", "error_code": 400 }
```
- `invalid_request_body`
```json
{ "error": "invalid_request_body", "error_code": 400 }
```
- `validation_failed`
```json
{ "error": "validation_failed", "error_code": 400 }
```
- `email_confirmation_token_missing`
```json
{ "error": "email_confirmation_token_missing", "error_code": 400 }
```
- `email_confirmation_token_invalid`
```json
{ "error": "email_confirmation_token_invalid", "error_code": 400 }
```
- `email_confirmation_token_expired`
```json
{ "error": "email_confirmation_token_expired", "error_code": 400 }
```
- `password_reset_token_missing`
```json
{ "error": "password_reset_token_missing", "error_code": 400 }
```
- `password_reset_token_invalid`
```json
{ "error": "password_reset_token_invalid", "error_code": 400 }
```
- `password_reset_token_expired`
```json
{ "error": "password_reset_token_expired", "error_code": 400 }
```

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
- `schedule_slot_not_found`
```json
{ "error": "schedule_slot_not_found", "error_code": 404 }
```
- `schedule_exception_not_found`
```json
{ "error": "schedule_exception_not_found", "error_code": 404 }
```
- `student_or_classroom_not_found`
```json
{ "error": "student_or_classroom_not_found", "error_code": 404 }
```
- `email_confirmation_user_not_found`
```json
{ "error": "email_confirmation_user_not_found", "error_code": 404 }
```
- `password_reset_user_not_found`
```json
{ "error": "password_reset_user_not_found", "error_code": 404 }
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
- `invalid_current_password`
```json
{ "error": "invalid_current_password", "error_code": 401 }
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

### 403
- `email_not_verified`
```json
{ "error": "email_not_verified", "error_code": 403 }
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
- `schedule_slot_conflict`
```json
{ "error": "schedule_slot_conflict", "error_code": 409 }
```
- `schedule_exception_overlap`
```json
{ "error": "schedule_exception_overlap", "error_code": 409 }
```
- `email_confirmation_token_already_used`
```json
{ "error": "email_confirmation_token_already_used", "error_code": 409 }
```
- `email_already_verified`
```json
{ "error": "email_already_verified", "error_code": 409 }
```
- `password_reset_token_already_used`
```json
{ "error": "password_reset_token_already_used", "error_code": 409 }
```

## 5.3 Erreurs avec `error_details` importantes

### `conflict` (email déjà pris) - 409
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
- horaire `HH:MM` invalide
- date `YYYY-MM-DD` invalide

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

Pour `POST /v1/auth/change-password` et `POST /v1/auth/reset-password`, des clés possibles dans `error_details`:
- `validation_password_confirmation_mismatch`
- `validation_min_length:8`
- `validation_one_of:foo|bar|baz` pour les champs validés avec `oneof`

Exemple `oneof`:
```json
{
  "error": "validation_failed",
  "error_code": 400,
  "error_details": [
    {
      "field": "week_pattern",
      "error": "validation_one_of:every_week|even_weeks|odd_weeks"
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

### `not_found` avec détail UUID path invalide - 404
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
- `row` est renseigné
- `error` est une clé stable `import_*` (pas un message texte)
- `error_details` (liste de strings) apporte le contexte de validation (min/max/expected/field)

Exemples de clés possibles:
- `import_required_field`
- `import_invalid_format`
- `import_invalid_length`
- `import_no_data_rows`
- `import_max_rows_exceeded`

Exemple:
```json
{
  "error": "import_validation_failed",
  "error_code": 400,
  "error_details": [
    {
      "row": 4,
      "field": "eleves",
      "error": "import_invalid_format",
      "value": "Jean",
      "error_details": ["expected:uppercase_last_name_then_first_name"]
    },
    {
      "row": 4,
      "field": "classes",
      "error": "import_invalid_length",
      "value": "A",
      "error_details": ["min:2", "max:100"]
    }
  ]
}
```
