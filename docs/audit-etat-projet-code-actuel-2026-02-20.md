# Audit du Projet (Code Actuel)

Date: 2026-02-20
Portee: analyse du code source actuel uniquement (pas des anciens documents d'audit)

## 1. Methode

Verifications executees:
- `go test ./...` -> OK
- `go test -cover ./...` -> OK
- `go vet ./...` -> OK

Couverture observee:
- `internal/api/handler`: 92.6%
- `internal/service`: 12.7%
- autres packages: majoritairement 0%

## 2. Synthese Executif

Le projet est sain sur sa base API (handlers testes et conventions d'erreurs globalement coherentes), mais il y a des incoherences de contrat entre routes/tests/docs, des zones de duplication importantes, et plusieurs points securite/clarte a traiter en priorite.

Priorites proposees:
- P0: securite auth/session et robustesse entree utilisateur
- P1: coherence DTO/retours/routes/tests
- P2: fragmentation/factorisation pour clarte et maintenabilite

## 3. Forces Actuelles

- Architecture claire en couches: `handler -> service -> repository(sqlc)` (`cmd/api/routes.go:16`, `internal/service/*.go`, `internal/repository/*`).
- Format d'erreur uniforme via `api.ErrorResponse` + helpers `web.Write*` (`internal/api/errors.go:5`, `internal/platform/web/response.go:25`).
- Validation DTO centralisee (`internal/platform/validator/validator.go:10`).
- Requetes SQL multi-tenant bien scopees par `user_id` dans l'ensemble des queries (`db/sqlc/queries.sql`).
- Tests handlers solides et riches en cas d'erreur (`internal/api/handler/*_test.go`).

## 4. Points Faibles et Modifications Recommandees

### 4.1 Coherence DTOs et Retours

Constats:
- Incoherence majeure: l'historique eleve supporte `page` mais retourne un tableau brut, pas un objet pagine (`internal/api/handler/student.go:86`, `internal/api/handler/student_kpis_history_test.go:119`).
- Validation `due_at` de punition non harmonisee: parse manuel RFC3339, renvoie `invalid_request_body` sans `error_details` de champ (`internal/api/handler/punishment.go:47`).
- Parsing d'UUID pour les rules fait aussi cote service (doublon de responsabilite avec validation handler) (`internal/service/rule.go:36`, `internal/service/rule.go:134`).

Recommandations:
- Unifier toutes les routes `List*` sur un contrat unique:
  - Option A: `PaginatedResponse` partout (incluant `/students/{id}/history`).
  - Option B: assumer tableau brut pour `history` et retirer `page` (moins recommande).
- Ajouter une validation de format `due_at` explicite avec detail de champ (`due_at`) pour alignement DX/API.
- Garder le parsing d'UUID en handler seulement; service doit recevoir des types deja valides.

### 4.2 Coherence des Routes

Constats:
- Routes runtime utilisent des params nommes (`{student_id}`, `{bonus_id}`, etc.) (`cmd/api/routes.go:79`, `cmd/api/routes.go:142`, `cmd/api/routes.go:159`).
- Plusieurs tests utilisent `{id}` generique et rely sur fallback `parsePathUUID(..., "id")` (`internal/api/handler/bonus_test.go:497`, `internal/api/handler/penalty_test.go:457`, `internal/api/handler/rule_test.go:479`).
- `GET /v1/user/me` est monte differemment dans les tests (`/v1/user/me/`) vs runtime (`/v1/user` + `/me`) (`internal/api/handler/user_test.go:425`, `cmd/api/routes.go:53`).
- Middleware auth repete dans chaque groupe de routes, risque d'oubli lors d'ajout futur (`cmd/api/routes.go:80`, `cmd/api/routes.go:95`, `cmd/api/routes.go:116`, etc.).

Recommandations:
- Aligner strictement les routers de tests sur `cmd/api/routes.go` (noms params inclus).
- Supprimer les compatibilites legacy inutiles (`id`, `studentId`) quand la migration tests est faite (`internal/api/handler/uuid_params.go:14`, `internal/api/handler/classroom_membership.go:52`).
- Introduire un sous-router protege unique (`/v1`) avec middleware auth applique une fois, puis exceptions explicites (auth/health).

### 4.3 Coherence du Placement Fichiers/Dossiers

Constats:
- Bonne separation globale, mais certains fichiers deviennent des points de congestion:
  - `db/sqlc/queries.sql`: 1138 lignes.
  - `cmd/api/routes.go`: 183 lignes de wiring.
  - Gros fichiers de tests handlers (ex: `internal/api/handler/classroom_test.go`: 659 lignes).

Recommandations:
- Fragmenter `db/sqlc/queries.sql` par domaine (`queries_users.sql`, `queries_students.sql`, etc.) et configurer sqlc sur dossier.
- Decomposer le montage routes en modules de domaine (`cmd/api/routes_students.go`, `routes_classrooms.go`, etc.).
- Scinder les gros tests par themes (`*_crud_test.go`, `*_validation_test.go`, `*_errors_test.go`).

### 4.4 Coherence des Tests

Constats:
- Tres bonne couverture handler, mais faible couverture service (12.7%) et quasi nulle sur `platform/*`.
- Aucune suite integration reelle PostgreSQL (les tests reposent sur repository in-memory).
- Le in-memory reproduit bien beaucoup de cas, mais ne garantit pas les subtilites SQL/Postgres (casts, contraintes, collations, perf, plans).

Recommandations:
- Ajouter une couche de tests integration DB (docker postgres) sur les flux critiques:
  - auth login/refresh
  - penalty -> rule -> punishment transaction
  - bonus use (idempotence/atomicite)
- Augmenter la couverture service sur `student`, `classroom`, `dashboard`, `auth`.
- Ajouter tests unitaires pour `internal/platform/web`, `internal/platform/jwt`, `internal/platform/config`.
- Ajouter test de non-regression doc/contrat pour `history` (pagination + format).

## 5. Securite et Clarte (Priorite Haute)

### P0 - A corriger en premier

- Refresh token stocke en clair en base (`internal/service/auth.go:71`, `db/sqlc/queries.sql:26`).
  - Action: stocker un hash du refresh token (et comparer hash), pas le token brut.
- Pas de rotation du refresh token au `refresh` (`internal/service/auth.go:90`).
  - Action: rotation a usage unique (revoke ancien + emettre nouveau).
- Validation JWT incomplete (issuer/audience non verifies explicitement) (`internal/platform/jwt/token.go:36`).
  - Action: parser/verifier avec contraintes `iss`/`aud` attendues.
- Protection brute-force absente sur login (`internal/api/handler/auth.go:33`).
  - Action: rate limiting + lockout progressif + observabilite security events.

### P1 - Important

- `MustUserIDFromContext` panic si middleware absent (`internal/platform/auth/middleware.go:62`).
  - Action: version safe dans handlers (retour 401) ou garde-fou central.
- Overflow pagination possible sur pages enormes (`internal/platform/web/pagination.go:31`).
  - Action: borner `page` et calculer offset en `int64` avec clamp.
- `DecodeJSON` n'interdit pas explicitement les tokens JSON supplementaires (`internal/platform/web/decode.go:14`).
  - Action: second decode et verification EOF stricte.

## 6. Opportunites de Fragmentation en Helpers

Candidats concrets:
- Service helpers d'existence et mapping erreurs not-found:
  - `internal/service/common_existence.go` pour `ensureStudentExists`, `ensureClassroomExists`, etc.
- Handler helpers de parsing/validation champs speciaux:
  - ex `parseRFC3339Field("due_at")` au lieu de parse manuel ad hoc.
- Route registration helpers:
  - fonctions `mountStudentsRoutes`, `mountRulesRoutes`, etc.
- Test helpers:
  - factory de router de test alignant exactement les params du runtime.

## 7. Plan d'Action (Propose)

### Sprint 1 (Securite + Contrat)

- Hash refresh tokens + rotation refresh.
- Verification JWT issuer/audience.
- Correction contrat `history` (choix pagination unique).
- Validation detaillee `due_at`.

### Sprint 2 (Coherence Routes/Tests)

- Aligner tous les tests sur les vrais params de routes.
- Retirer fallback params legacy (`id`, `studentId`) apres alignement.
- Centraliser middleware auth sur un sous-router protege.

### Sprint 3 (Clarte + Dette technique)

- Split `queries.sql` par domaine.
- Split `routes.go` par domaine.
- Split gros fichiers de tests.
- Ajouter integration tests postgres pour flux critiques.

## 8. Conclusion

La base est robuste cote handlers et conventions d'erreurs. Les gains les plus importants viennent maintenant de la securisation auth/session, de l'alignement strict des contrats API (retours + routes + docs + tests), et d'une meilleure modularisation pour maintenir la clarte a mesure que le code grossit.
