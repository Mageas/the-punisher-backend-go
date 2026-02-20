# Audit SQLC / Postgres - Niveau de Couplage et Plan d'Action

Date: 2026-02-20
Portee: code actuel uniquement

## 1. Reponse Directe a ta Question

Es-tu fortement liee a Postgres dans DTOs/services/handlers?

- DTOs: non, couplage faible.
- Handlers: non, couplage faible.
- Services: oui, couplage moyen a fort.
- Repository + persistence: oui, couplage tres fort (normal avec sqlc postgres).

## 2. Cartographie du Couplage

## 2.1 DTOs (faible)

- DTOs utilisent surtout `uuid.UUID`, `time.Time`, types Go standards (`internal/dto/*.go`).
- Pas de `pgx`, `pgtype`, `sql.Null*` dans les DTOs.

Conclusion: les DTOs ne sont pas lies a Postgres de facon bloquante.

## 2.2 Handlers (faible)

- Handlers manipulent HTTP + DTO + interfaces de services (`internal/api/handler/*.go`).
- Pas d'import `pgx`/`pgconn` en handlers.

Conclusion: couche HTTP decouplable sans gros impact DB.

## 2.3 Services (moyen/fort)

Points de couplage visibles:
- Detection not-found via `pgx.ErrNoRows` dans plusieurs services:
  - `internal/service/bonus.go:10`
  - `internal/service/punishment.go:11`
  - `internal/service/rule.go:10`
  - `internal/service/student.go:10`
  - `internal/service/classroom_crud.go:10`
- Dependance explicite aux erreurs PostgreSQL `pgconn.PgError` code `23505`:
  - `internal/service/classroom_membership.go:11`, `internal/service/classroom_membership.go:26`
- Transactions basees sur `pgx.Tx`:
  - `internal/service/penalty.go:11`, `internal/service/penalty.go:32`
  - `internal/repository/transactional.go:15`
- Signatures fortement liees aux types sqlc generes `repository.*Params`, `repository.*Row`.

Conclusion: la logique metier est partiellement couplee aux details d'implementation sqlc/pgx.

## 2.4 Repository / SQLC / SQL (tres fort)

- `sqlc.yaml` force `engine: postgresql` et `sql_package: pgx/v5` (`sqlc.yaml:3`, `sqlc.yaml:10`).
- Queries avec features postgres specifiques:
  - `ILIKE`, casts `::uuid`, `ANY(...::uuid[])`, `NOW()`, `COALESCE(sqlc.narg(...))` (`db/sqlc/queries.sql`).
- Interface `DBTX` et `WithTx` bases sur types `pgx` (`internal/repository/db.go:10`, `internal/repository/db.go:28`).

Conclusion: la couche persistence est volontairement et fortement postgres-dependante.

## 3. Evaluation Globale

Niveau de couplage par couche:
- DTO: 1/5
- Handler: 1/5
- Service: 3.5/5
- Repository/SQL: 5/5

Lecture pragmatique:
- Ce n'est pas un probleme si l'objectif est "Postgres-first".
- C'est un probleme si tu veux pouvoir changer de base rapidement (MySQL/SQLite/Cloud SQL variant).

## 4. Risques si Migration DB Future

- Rework massif des queries sqlc (syntaxe, fonctions, casts).
- Regeneration repository + impact cascade sur services/tests in-memory.
- Logique d'erreur metier actuellement branchee sur `pgx.ErrNoRows` / `pgconn`.
- Transactions `pgx.Tx` a abstraire.

## 5. Plan d'Action Recommande

Objectif propose: reduire le couplage dans les couches metier sans renoncer a sqlc/postgres aujourd'hui.

### Phase 1 - Frontiere metier claire (priorite haute)

- Introduire des interfaces metier par domaine (ex: `StudentRepo`, `BonusRepo`, `RuleRepo`) dans `internal/service` ou `internal/domain`.
- Les interfaces doivent exposer des types metier (pas `repository.*Row`).
- Definir des erreurs metier communes (`ErrNotFound`, `ErrConflict`, etc.) sans `pgx`.

Resultat: services non dependants de `pgx` ni de structs sqlc.

### Phase 2 - Adapter SQLC dedie (priorite haute)

- Creer `internal/adapter/persistence/sqlc` (ou similaire) qui:
  - appelle `internal/repository` (genere sqlc),
  - mappe vers types metier,
  - convertit `pgx.ErrNoRows` et `pgconn` vers erreurs metier.
- Deplacer les checks `pgx.ErrNoRows` hors des services.

Resultat: couplage Postgres concentre dans un seul adaptateur.

### Phase 3 - Transaction abstraction (priorite moyenne)

- Introduire un `TxManager` metier:
  - `WithinTx(ctx, func(repos DomainRepos) error) error`
- Implementer une version sqlc/pgx dans l'adapter.
- Retirer `pgx.Tx` des services (`penaltyService`).

Resultat: services transactionnels mais DB-agnostiques.

### Phase 4 - Tests de contrat adapter (priorite moyenne)

- Garder les tests handlers actuels.
- Ajouter tests de contrat adapter SQLC + postgres (docker) pour:
  - mapping erreurs,
  - transactions,
  - filtres multi-tenant.

Resultat: securise une evolution future de persistence.

### Phase 5 - Option migration multi-DB (si necessaire)

- Une fois frontiere stable, implementer un 2e adapter (si besoin reel).
- Sinon rester Postgres, mais avec dette technique fortement reduite.

## 6. Decision Cadre (Recommandation)

Deux strategies possibles:
- Strategie A (pragmatique recommandee): rester Postgres, mais decoupler les services maintenant.
- Strategie B (ambitieuse): preparer migration multi-DB immediate (cout eleve, ROI faible si pas besoin produit court terme).

Recommendation: Strategie A.

## 7. Checklist de Changement Minimal (ordre conseille)

1. Introduire erreurs metier persistence-agnostiques.
2. Retirer `pgx.ErrNoRows` des services.
3. Extraire un adapter SQLC par domaine.
4. Abstraire la transaction `CreatePenalty`.
5. Ajouter tests integration postgres du nouvel adapter.

## 8. Conclusion

Aujourd'hui, tu n'es pas fortement liee a Postgres dans les DTOs/handlers, mais tu l'es dans les services et tres fortement dans repository/sqlc (ce qui est attendu). Avec un refactor en adaptateur, tu peux conserver sqlc + Postgres tout en retrouvant clarte, securite de conception, et marge d'evolution future.
