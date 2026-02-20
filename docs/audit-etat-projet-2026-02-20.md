# Audit de l'etat actuel du projet (20 fevrier 2026)

## Perimetre et methode
Audit realise sur:
- code API (`cmd/api`, `internal/api`, `internal/service`, `internal/dto`, `internal/platform`, `internal/repository`)
- SQL source SQLC (`db/sqlc/queries.sql`, `sqlc.yaml`)
- tests (`internal/api/handler`, `internal/service`, `internal/testutil`)
- documentation existante (`docs/*.md`)

Verification executee:
- `go test ./...` -> OK
- `go test -race ./...` -> OK
- `go vet ./...` -> OK
- `go test ./... -cover` -> `internal/api/handler: 92.1%`, `internal/service: 6.8%`

## Resume executif
Points forts:
- architecture en couches claire (Handler -> Service -> Repository SQLC)
- isolation multi-tenant bien appliquee (`user_id` dans les requetes SQL)
- contrats d'erreur API homogenes (`internal/api/errors.go`, `internal/platform/web/response.go`)
- tests handlers tres complets et deterministes via repository in-memory

Points faibles prioritaires:
- `P0` securite cookie refresh (`Secure: false`) dans `internal/api/handler/auth.go:60`
- `P1` coherence DTO/retours: couche `dto` fortement couplee a `repository` et `pgtype` (`internal/dto/bonus.go:7`, `internal/dto/classroom.go:7`, `internal/dto/student_kpis_history.go:7`)
- `P1` coherence API: `history_page` legacy + retour non-pagine pour l'historique (`internal/api/handler/student.go:95`)
- `P1` coherence routes: singular/plural et naming params non uniformes (`cmd/api/routes.go:53`, `cmd/api/routes.go:102`)
- `P1` coherence tests: couverture faible de la logique service et duplication importante dans certains tests volumineux

---

## 1. Coherence des DTOs et surtout des retours
### Constats
- Convention de nommage globalement stable (`Request*Dto`, `Update*Dto`, `Return*Dto`), mais exception auth avec `LoginResponseDto` et `RefreshResponseDto` (`internal/dto/auth.go:9`, `internal/dto/auth.go:14`).
- Les DTOs de sortie sont relies directement aux types SQLC (mapping depuis `repository.*Row`) dans la couche `dto` (`internal/dto/student.go:57`, `internal/dto/rule.go:75`).
- Plusieurs DTOs dependent de `pgtype` (`internal/dto/punishment.go:7`, `internal/dto/bonus.go:7`), ce qui diffuse les details PostgreSQL hors repository.
- L'historique eleve repose sur des valeurs sentinelles (UUID nil, timestamp `1970-01-01`) converties ensuite en `nil` (`internal/dto/student_kpis_history.go:11`, `db/sqlc/queries.sql:600`, `db/sqlc/queries.sql:605`).
- Incoherence de forme de retour pour la lecture collection:
- les `List*` classiques renvoient un envelope pagine (`web.NewPaginatedResponse`)
- `GET /v1/students/{id}/history` renvoie un tableau brut (pas de metadata de pagination) (`internal/api/handler/student.go:107`)

### Recommandations
1. Unifier les conventions de noms:
- soit `*ResponseDto` partout
- soit `Return*Dto` partout

2. Sortir le mapping SQLC de `internal/dto`:
- creer une couche d'adapter de mapping (`internal/repository/mapper` ou `internal/adapter/persistence`)
- garder `dto` strictement HTTP (sans import `repository`/`pgtype`)

3. Supprimer les sentinelles de l'historique:
- faire remonter des `NULL` SQL explicites
- mapper via pointeurs/nullable cote adapter

4. Standardiser les retours listes:
- soit paginer aussi `history`
- soit assumer officiellement qu'il s'agit d'un "flux" non pagine (et retirer les params de pagination)

---

## 2. Coherence des routes
### Constats
- Base `/v1` coherente, routes metier principales claires (`cmd/api/routes.go`).
- Incoherence singular/plural:
- `GET /v1/user/me` (`cmd/api/routes.go:53`) alors que le reste est majoritairement pluriel (`/students`, `/classrooms`, etc.).
- Incoherence naming des params:
- path param `studentId` en camelCase (`cmd/api/routes.go:102`, `internal/api/handler/classroom.go:169`)
- query params en snake_case (`classroom_id`, etc. dans `internal/api/handler/dashboard.go:27`)
- Incoherence pagination historique:
- `page` est parse
- mais `history_page` est encore supporte (`internal/api/handler/student.go:95`)

### Recommandations
1. Definir une convention unique de nommage URI:
- ressources en pluriel
- params en snake_case (`{student_id}`) ou tout en `id` imbrique de facon uniforme

2. Deprecier `history_page` avec fenetre de transition claire:
- conserver temporairement mais ajouter avertissement de deprecation
- retirer a date fixe

3. Extraire un helper commun de parsing `UUID` path/query:
- pour uniformiser les messages d'erreur et les `error_details`

---

## 3. Coherence du placement fichiers/dossiers
### Constats
- Architecture technique saine et lisible (`docs/architecture.md` est alignee au code global).
- Les dossiers `handler`, `service`, `dto` sont plats et commencent a grossir.
- Plusieurs fichiers deviennent volumineux:
- `internal/service/classroom.go:276 lignes`
- `internal/service/penalty.go:247 lignes`
- `internal/api/handler/classroom.go:223 lignes`
- Beaucoup de duplication entre `bonus_type`, `penalty_type`, `punishment_type` (handler + service + tests).

### Recommandations
1. Fragmenter par domaine fonctionnel, par exemple:
- `internal/api/handler/classroom/` (`crud.go`, `membership.go`)
- `internal/service/classroom/` (`crud.go`, `membership.go`, `enrichment.go`)

2. Factoriser les patterns repetitifs des "types" (`bonus_type`, `penalty_type`, `punishment_type`):
- helper de CRUD commun
- helpers de validation/mapping communs

3. Garder `internal/repository` strictement generation SQLC + petits wrappers d'infra, sans faire remonter ces types dans les couches HTTP.

---

## 4. Coherence des tests
### Constats
- Tres bon niveau de tests handlers (statuts, validations, decode, erreurs metier, erreurs internes).
- Coverage package `internal/api/handler` elevee (`92.1%`).
- Coverage package `internal/service` faible (`6.8%`) alors que la logique sensible (transactions, rule engine) est principalement en service.
- Peu/pas de tests directs pour `internal/platform/auth`, `internal/platform/jwt`, `internal/platform/web`, `internal/dto`.
- Fichiers de tests tres gros et parfois repetitifs:
- `internal/api/handler/classroom_test.go:670`
- `internal/api/handler/punishment_test.go:530`
- `internal/api/handler/student_test.go:516`

### Recommandations
1. Renforcer les tests unitaires services (priorite):
- transactions `CreatePenalty` (`internal/service/penalty.go:40`)
- comportements limites du rule engine (`shouldTriggerRule`, `dueAt`)
- erreurs repository (`ErrNoRows`, erreurs techniques)

2. Ajouter des tests unitaires de securite:
- cookie auth (flags `Secure`, `HttpOnly`, `SameSite`)
- verification JWT (issuer/audience/expiration)

3. Continuer la factorisation des tests handlers:
- etendre la logique de `internal/testutil/shared/type_handler_shared.go` aux autres domaines repetitifs

---

## 5. Fragmentation possible en helpers/fonctions
Candidats prioritaires:
1. Parsing UUID + erreurs API uniformes:
- code duplique dans quasiment tous les handlers (`internal/api/handler/*.go`).

2. Parsing options de liste (`page`, `state`, `search`):
- deja partiellement factorise via `internal/platform/web/query_params.go`, a etendre.

3. Pattern CRUD des "types":
- `internal/api/handler/bonus_type.go`, `internal/api/handler/penalty_type.go`, `internal/api/handler/punishment_type.go`
- `internal/service/bonus_type.go`, `internal/service/penalty_type.go`, `internal/service/punishment_type.go`

4. Mapping DTO <-> rows SQLC:
- extraire vers une couche mapper dediee pour sortir la dependance `repository/pgtype` des DTO HTTP.

---

## 6. Focus securite et clarte (priorite)
### Risques securite
1. `P0` cookie refresh non securise en transport:
- `Secure: false` (`internal/api/handler/auth.go:60`).

2. `P1` refresh token non rotate au refresh:
- le token est valide jusqu'a expiration sans rotation/revocation systematique (`internal/service/auth.go:90`).

3. `P1` panic possible si middleware absent:
- `MustUserIDFromContext` panique (`internal/platform/auth/middleware.go:62`).

### Risques clarte/maintenabilite
1. Couplage SQLC/PG dans DTO et service (difficulte de lecture + evolution).
2. Sentinelles temporelles/UUID pour l'historique (contrat implicite fragile).
3. Melange conventions route/query (`studentId`, `history_page`, pluralisation).

---

## Plan d'action recommande (pragmatique)
### Sprint 1 (court, securite + coherence API)
1. Activer `Secure` sur cookie refresh en prod (configurable par env).
2. Uniformiser parsing d'ID et erreurs (`error_details` explicites).
3. Deprecier `history_page` et standardiser `page` uniquement.
4. Normaliser naming route params (`student_id` ou convention unique).

### Sprint 2 (clarte structurelle)
1. Refactoriser handlers/services volumineux en sous-fichiers par responsabilite.
2. Factoriser CRUD des "types" (handlers + services + tests).
3. Etendre les tests unitaires services sur flux critiques (transactions/rules).

### Sprint 3 (decouplage durable)
1. Isoler mapping SQLC dans des adapters dedies.
2. Nettoyer DTOs HTTP de toute dependance `repository/pgtype`.
3. Revoir l'historique pour supprimer les sentinelles SQL.

---

## Conclusion
Le projet est globalement solide en architecture et en comportement HTTP. Les priorites pour atteindre ton objectif "securite + clarte" sont:
1. corriger les points securite auth,
2. unifier les conventions DTO/routes/retours,
3. reduire le couplage SQLC/PostgreSQL hors couche repository,
4. reequilibrer les tests vers la logique service.
