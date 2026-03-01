# The Punisher Backend (Go)

API backend pour la gestion disciplinaire en classe:
- bonus (`bonuses`)
- pénalités (`penalties`)
- punitions (`punishments`)
- règles automatiques (`rules`)

Stack principal: Go, PostgreSQL, sqlc, Chi.

## Prérequis

- Go `1.24.10` (cf. `go.mod`)
- Docker + Docker Compose (pour PostgreSQL)
- `make`

Outils utiles selon les commandes:
- `air` (pour `make dev`)
- `migrate` (pour les migrations)
- `sqlc` (pour régénérer le repository)

## Installation rapide

1. Cloner le repo
```bash
git clone https://github.com/Mageas/the-punisher-backend-go.git
cd the-punisher-backend-go
```

2. Configurer l'environnement
```bash
cp .env.dist .env
```

3. Démarrer PostgreSQL
```bash
docker compose up -d
```

4. Appliquer les migrations
```bash
make migrate-up
```

5. (Optionnel) Seed des données
```bash
make seed
```

Seed d'une instance distante via API
```bash
go run ./cmd/seed-api --base-url http://localhost:8080
# ou
make seed-api SEED_API_URL=http://localhost:8080
```

Options utiles:
- `--email` (default `admin@test.fr`)
- `--password` (default `admin@test.fr`)
- `--class-count`, `--students-per-class`
- `--bonus-chance`, `--penalty-chance`, `--punishment-chance`
- `--max-bonuses`, `--max-penalties`, `--max-punishments`

Variables d'environnement possibles:
- `SEED_API_URL`
- `SEED_API_EMAIL`
- `SEED_API_PASSWORD`

## Lancer l'application

Mode dev (hot reload via Air):
```bash
make dev
```

Sans Air:
```bash
go run ./cmd/api
```

API par défaut: `http://localhost:8080`

## Configuration CORS

Variables disponibles (dans `.env`) :
- `CORS_ALLOWED_ORIGINS` (CSV), ex: `http://localhost:3000,http://localhost:5173`
- `CORS_ALLOWED_METHODS` (CSV), ex: `GET,POST,PUT,DELETE,OPTIONS`
- `CORS_ALLOWED_HEADERS` (CSV), ex: `Accept,Authorization,Content-Type,X-CSRF-Token`
- `CORS_EXPOSED_HEADERS` (CSV), ex: `Link`
- `CORS_ALLOW_CREDENTIALS` (`true|false`)
- `CORS_MAX_AGE` (secondes)
- `JWT_REFRESH_COOKIE_SECURE` (`true|false`) pour forcer le flag `Secure` du cookie refresh

Note sécurité : si `CORS_ALLOW_CREDENTIALS=true`, les origins contenant `*` sont refusées au démarrage.
Par défaut, `JWT_REFRESH_COOKIE_SECURE=true` quand `APP_ENV=production`.

## Build

```bash
make build
./bin/main
```

## Tests

Tests unitaires (par défaut):
```bash
go test ./...
# ou
make test
```

Tests d'intégration (Testcontainers + Docker requis):
```bash
go test -tags=integration ./tests/integration/...
# ou
make test-integration
```

Package spécifique (unitaires):
```bash
go test ./internal/service
go test ./internal/api/handler
```

## Commandes utiles

```bash
make sqlc            # régénère le code sqlc
make migrate-create <name>
make migrate-up
make migrate-down
make seed
make seed-api SEED_API_URL=http://localhost:8080
make reset-seed
```

## CI

Une GitHub Action exécute les tests automatiquement sur chaque PR:
- `.github/workflows/tests.yml`

## Documentation

- `docs/api-reference.md`
- `docs/architecture.md`
- `docs/projet.md`
- `docs/testing-handlers.md`
