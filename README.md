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

## Build

```bash
make build
./bin/main
```

## Tests

Tous les tests:
```bash
go test ./...
```

Package spécifique:
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
