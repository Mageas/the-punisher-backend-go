# The Punisher Backend (Go)

API backend pour la gestion disciplinaire en classe:
- bonus (`bonuses`)
- pénalités (`penalties`)
- punitions (`punishments`)
- règles automatiques (`rules`)
- emploi du temps (`schedule`)

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

## Pagination API

Toutes les routes paginées utilisent les query params:
- `page` (optionnel, int > 0, défaut: `1`)
- `item_per_page` (optionnel, int, défaut: `20`)

Règles sur `item_per_page`:
- min: `5`
- max: `50`
- valeur invalide/non numérique: fallback `20`
- valeur < 5: forcée à `5`
- valeur > 50: forcée à `50`

## Créations batch par classe

Le backend expose aussi des routes de création multiple scoppées par classe:
- `POST /v1/classrooms/{classroom_id}/bonuses/bulk`
- `POST /v1/classrooms/{classroom_id}/penalties/bulk`
- `POST /v1/classrooms/{classroom_id}/punishments/bulk`

Elles permettent de cibler plusieurs élèves d'une même classe en une seule requête, avec rollback complet si la validation métier échoue en cours de traitement.

## Emploi du temps

Le backend expose désormais un sous-domaine `schedule` pour:
- gérer des créneaux hebdomadaires multi-classes avec `weekday`, `start_time`, `end_time` et `week_pattern`
- gérer des interruptions globales utilisateur (`vacation`, `public_holiday`) sur journées entières
- calculer les `5` prochains cours d'une classe via `GET /v1/classrooms/{classroom_id}/next-lessons`
- servir de base aux règles automatiques avec une échéance en `days` ou sur les `5` prochains cours (`next_lessons`, échéance positionnée au début du cours) en utilisant la classe fournie à la création d'une pénalité, ou la classe unique de l'élève si elle peut être déduite

Ces calculs calendrier/horaires sont interprétés dans le fuseau horaire de l'utilisateur.

Formats métier associés:
- `weekday`: texte anglais minuscule (`monday` à `sunday`)
- `start_time` / `end_time`: `HH:MM`
- `week_pattern`: `every_week`, `even_weeks`, `odd_weeks`
- `start_date` / `end_date`: `YYYY-MM-DD`

## Isolation des données métier

Toutes les requêtes SQL métier sont scoppées par `user_id` pour les domaines:
- `students`, `classrooms`, `student_classrooms`
- `bonus_types`, `bonuses`
- `penalty_types`, `penalties`
- `punishment_types`, `punishments`, `rules`
- `schedule_slots`, `schedule_slot_classrooms`, `schedule_exceptions`
- agrégats métier (`dashboard`, historique élève)

Le schéma PostgreSQL empêche désormais aussi les relations cross-user via des contraintes composites basées sur `(user_id, id)`.

Hors périmètre de cette règle:
- les flux d'authentification
- les tables de tokens
- la gestion des cookies refresh

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

## Emails transactionnels (SMTP)

Lors d'une inscription (`POST /v1/auth/register`), le backend:
- crée un token signé de confirmation email avec expiration
- enregistre l'utilisateur avec le fuseau horaire par défaut `Europe/Paris` (exposé en lecture, non modifiable via l'API pour l'instant)
- envoie automatiquement un email via SMTP après validation de la transaction en base
- valide l'email via `GET /v1/auth/confirm-email?token=<token>`
- permet de renvoyer un nouveau lien via `POST /v1/auth/confirm-email/resend`
- bloque le login tant que l'email n'est pas confirmé

Variables d'environnement associées:
- `EMAIL_CONFIRMATION_SECRET`
- `EMAIL_CONFIRMATION_EXPIRATION_IN_HOURS`
- `EMAIL_CONFIRMATION_BASE_URL` (URL absolue valide requise, ex: `https://api.example.com/v1/auth/confirm-email`)
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM_EMAIL`
- `SMTP_FROM_NAME`

## Réinitialisation de mot de passe (lien email)

Endpoints publics:
- `POST /v1/auth/forgot-password`
- `POST /v1/auth/reset-password`

Body `forgot-password`:
```json
{
  "email": "teacher@school.test"
}
```

Body `reset-password`:
```json
{
  "token": "<jwt_password_reset_token>",
  "new_password": "NewSecurePass2@",
  "confirm_password": "NewSecurePass2@"
}
```

Regles appliquees:
- `forgot-password` repond de maniere neutre (200) meme si l'utilisateur n'existe pas
- si l'utilisateur existe, generation d'un token signe temporaire + envoi d'un email contenant un lien unique
- `reset-password` verifie la validite du token (invalide/expire/deja utilise)
- verification `new_password == confirm_password`
- mise a jour du mot de passe (`password_hash`, `password_changed_at`)
- invalidation des refresh tokens utilisateur
- invalidation du token de reset apres usage

Variables d'environnement associees:
- `PASSWORD_RESET_SECRET`
- `PASSWORD_RESET_EXPIRATION_IN_HOURS`
- `PASSWORD_RESET_BASE_URL` (URL absolue valide requise, ex: `https://app.example.com/reset-password`)

## Changement de mot de passe (authentifie)

Endpoint protege:
- `POST /v1/auth/change-password`

Body JSON:
```json
{
  "current_password": "CurrentPass1!",
  "new_password": "NewSecurePass2@",
  "confirm_password": "NewSecurePass2@"
}
```

Regles appliquees:
- verification du mot de passe actuel
- verification `new_password == confirm_password`
- validation du nouveau mot de passe identique a l inscription (`required,min=8`)

Effets de bord:
- mise a jour de `password_hash` + `password_changed_at`
- invalidation de tous les refresh tokens utilisateur (y compris session courante)
- suppression du cookie `refresh_token`

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
