# Audit SQLC et couplage PostgreSQL (20 fevrier 2026)

## Reponse courte
Oui, l'implementation actuelle est **fortement liee a PostgreSQL** dans les `DTOs` et `services`, et faiblement dans les `handlers`.

Niveau de couplage estime:
- Handlers: faible a moyen
- Services: fort
- DTOs: fort
- Tests in-memory: fort

## 1. Etat actuel du couplage

### 1.1 Couplage au niveau SQLC (normal et attendu)
- `sqlc.yaml` cible explicitement PostgreSQL (`engine: postgresql`) et `pgx/v5` (`sql_package: pgx/v5`) (`sqlc.yaml:3`, `sqlc.yaml:10`).
- Le SQL utilise des specificites PostgreSQL:
- casts `::type` (`db/sqlc/queries.sql:103`, `db/sqlc/queries.sql:418`)
- `ILIKE` (`db/sqlc/queries.sql:104`, `db/sqlc/queries.sql:1071`)
- `ANY(...::uuid[])` (`db/sqlc/queries.sql:347`, `db/sqlc/queries.sql:407`)
- `NOW()` (`db/sqlc/queries.sql:41`, `db/sqlc/queries.sql:1121`)

Conclusion: couplage DB au niveau repository/SQLC = normal.

### 1.2 Couplage qui depasse la couche repository
#### DTOs
- Import direct de `internal/repository` dans de nombreux DTOs (`internal/dto/student.go:7`, `internal/dto/rule.go:7`).
- Import direct de `pgtype` dans DTOs (`internal/dto/bonus.go:7`, `internal/dto/punishment.go:7`, `internal/dto/student_kpis_history.go:7`).
- Mapping direct depuis `repository.*Row` dans les DTOs (`internal/dto/student.go:57`, `internal/dto/rule.go:75`).

#### Services
- Usage massif de `pgtype.*` pour construire les params SQLC (`internal/service/student.go:76`, `internal/service/rule.go:130`, `internal/service/punishment.go:78`).
- Usage direct de `pgx.ErrNoRows` (`internal/service/auth.go:40`, `internal/service/classroom.go:73`, `internal/service/penalty.go:73`).
- Interface transactionnelle dependante `pgx.Tx` (`internal/service/penalty.go:30`).

#### Handlers
- Couplage faible: peu de references PostgreSQL directes.
- Mais certaines decisions de format sont influencees par le schema SQLC (ex: historique/sentinelles via service+dto).

#### Tests
- Repository in-memory clone l'API SQLC, avec types `repository.*` et `pgtype` (`internal/testutil/inmemory/repository.go:9`, `internal/testutil/inmemory/bonuses_table.go:298`).
- Le fake transaction implemente `pgx.Tx` (`internal/testutil/inmemory/tx.go:23`).

## 2. Points sensibles identifies
1. **Couche DTO non portable**
- Toute evolution de SQLC impacte directement les contrats DTO.

2. **Services dependants des types Postgres**
- La logique metier manipule des details techniques (`pgtype.Text`, `pgtype.Bool`, `pgtype.UUID`).

3. **Historique eleve fragile**
- Contrat base sur sentinelles SQL (`db/sqlc/queries.sql:600`, `db/sqlc/queries.sql:605`) puis conversion cote DTO (`internal/dto/student_kpis_history.go:125`).

4. **Effort de migration DB eleve**
- Migrer vers un autre moteur implique aujourd'hui des changements transverses (repository + services + DTO + tests).

## 3. Plan d'action recommande (si objectif = reduire le verrou PostgreSQL)

### Phase 0 - Sans rupture (court terme)
1. Documenter officiellement la frontiere:
- `repository` est la seule couche autorisee a parler SQL/pgx/pgtype.

2. Ajouter des tests de contrat d'adapter repository:
- garantir les invariants metier sans exposer `pgtype` hors adapter.

### Phase 1 - Decouplage DTO (gain rapide)
1. Creer des modeles de domaine/lecture neutres:
- ex: `internal/domain/model` ou `internal/core/model`.

2. Deplacer les mappings `repository row -> DTO` vers un adapter dedie:
- ex: `internal/adapter/persistence/sqlcmapper`.

3. Retirer `repository` et `pgtype` de `internal/dto`.

Impact attendu:
- baisse immediate du couplage DTO a SQLC/PG.

### Phase 2 - Decouplage service
1. Introduire des types d'entree metier neutres pour filtres optionnels:
- ex: `OptionalBool`, `OptionalString`, etc. (ou pointeurs simples dans des structs metier).

2. Convertir ces types en `pgtype.*` uniquement dans l'adapter repository.

3. Centraliser la traduction d'erreurs techniques:
- `pgx.ErrNoRows` -> `ErrNotFound` metier dans l'adapter.

Impact attendu:
- la logique service n'a plus besoin de `pgx`/`pgtype`.

### Phase 3 - Historique et modeles de lecture
1. Refaire `ListStudentHistory` pour renvoyer des `NULL` explicites (pas de sentinelles).
2. Mapper les `NULL` vers pointeurs/metadonnees directement dans l'adapter.
3. Simplifier `internal/dto/student_kpis_history.go` (supprimer logique sentinel).

Impact attendu:
- moins de code implicite, meilleure clarte, moins d'erreurs futures.

### Phase 4 - Option multi-DB (seulement si besoin produit)
1. Definir un port repository metier (interfaces stables cote domaine).
2. Garder SQLC/PG comme premier adapter.
3. Ajouter un second adapter (si necessaire) sans toucher handlers/services.

## 4. Priorisation concrete

Priorite immediate:
1. Phase 1 (DTO) - fort ROI en clarte et maintenance.
2. Phase 2 (service) - decouplage metier reel.
3. Phase 3 (historique) - reduit dette technique visible.

A faire seulement si besoin explicite de portabilite DB:
4. Phase 4 (multi-adapter DB).

## 5. Conclusion
Ton code est propre pour une stack assumee `Go + SQLC + PostgreSQL`, mais le couplage depasse actuellement la couche persistence.

Si ton objectif est securite + clarte + evolutivite, la trajectoire recommandee est:
1. isoler SQLC/pgx/pgtype dans un adapter,
2. garder DTO/services en types metier neutres,
3. supprimer les sentinelles SQL dans l'historique.

Cela permet de conserver SQLC (et ses gains) sans enfermer toute l'application dans des details PostgreSQL.
