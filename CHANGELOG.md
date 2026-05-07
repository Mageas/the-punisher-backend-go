# Changelog

Tous les changements notables de ce projet seront documentés dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/),
et ce projet adhère au [Versionnage Sémantique](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Ajouté
- Documentation d'observabilité Grafana / Loki (`docs/grafana/`, `docs/observability-grafana-loki.md`)

### Modifié
- Mises à jour de `README.md` et `AGENTS.md`

## [2026-03-13]

### Modifié
- Simplification de la réponse `/dashboard` en ne conservant que les KPIs (#112)

## [2026-03-12]

### Ajouté
- Application du fuseau horaire utilisateur aux plannings, règles et filtres de dates (#110)
- Endpoint bulk bonuses par classe et suppression de la route bulk students inutilisée (#108)

## [2026-03-11]

### Ajouté
- Routes de création en masse scopées par classe pour pénalités, punitions et étudiants (#106)

### Corrigé
- Arrondi à deux décimales des floats dans les réponses liées aux bonus (#104)

## [2026-03-07]

### Ajouté
- Dates d'échéance « prochain cours » depuis le contexte de classe de la pénalité (#98)
- Support des champs `due_at` nullable sur les règles selon le mode (#102)

### Corrigé
- Isolation par utilisateur appliquée à toutes les requêtes SQL métier (#100)

## [2026-03-06]

### Ajouté
- Emplois du temps par classe et gestion des exceptions (#95)
- Champs `occurred_at` et `evaluation_label` sur bonus, pénalités et punitions (#94)
- Formatage unifié des dates en RFC3339 UTC
- Champ `points` dans la route de mise à jour d'un bonus

### Corrigé
- `evaluation_label` non-nullable avec chaîne vide par défaut

## [2026-03-04]

### Ajouté
- Paramètre `item_per_page` configurable sur toutes les routes paginées (#91)

## [2026-03-03]

### Ajouté
- Flow de mot de passe oublié (#84)
- Route de changement de mot de passe pour utilisateur connecté (#85)

## [2026-03-02]

### Ajouté
- Confirmation d'email à l'inscription (#83)
- Codes d'erreur structurés pour l'import de students et classrooms (#81)

### Modifié
- Templates d'issues GitHub et fichier `AGENTS.md`

## [2026-03-01]

### Ajouté
- Recherche sur les classes (#78)
- Recherche des étudiants dans une classe (#76)
- Recherche sur les endpoints de types (bonus, pénalité, punition) (#74)
- Cibles Makefile pour le coverage (#71)

## [2026-02-28]

### Ajouté
- Suite de tests d'intégration (#69)
- Batterie de tests unitaires couvrant platform, services et middlewares

### Modifié
- Centralisation des filtres de recherche pour pénalités, punitions et bonus (#68)
- Nettoyage massif des anciens tests avant réintroduction

## [2026-02-27]

### Ajouté
- Route d'import CSV/JSON d'étudiants et de classes (#64)
- Routes de suppression en masse pour étudiants et classes (#65)

## [2026-02-25]

### Ajouté
- Route `POST /auth/logout-all` invalidant tous les refresh tokens de l'utilisateur (#58)

## [2026-02-24]

### Ajouté
- Route `POST /auth/logout` révoquant le refresh token courant (#61)
- Webhook de déploiement Dokploy sur push `main` (#59)

## [2026-02-23]

### Ajouté
- Endpoint public indiquant si l'inscription est ouverte
- Dockerfile multi-stage et pipeline GitHub Actions de build d'image
- Refonte des réponses KPIs et nouvelle route `GET /classrooms/:id/kpis`

## [2026-02-22]

### Ajouté
- Helper standardisé `WriteAPIError`

### Corrigé
- Renvoi d'un 404 sur UUID invalide au lieu d'un 400

## [2026-02-21]

### Modifié
- Erreurs de repository dédiées pour supprimer les imports `pgx` dans les services
- Suppression du concept « managed types » et tests sur `internal/platform/*`
- Usage direct des DTOs dans les tests au lieu de structs ad-hoc
- Handler 404 global et fusion des routes students + classrooms
- Cascade `ON DELETE` sur les clés étrangères des `*_types`

## [2026-02-20]

### Ajouté
- Rotation systématique du refresh token à chaque refresh
- Champ `automated` pour marquer les punitions générées par une règle

### Modifié
- Modularisation des requêtes, routes et tests, introduction de la suite d'intégration
- Remaniement des signatures et conventions handlers + services
- Suppression des anciennes routes `:id` et création d'un groupe `auth`
- Découplage des DTOs des types `pgx/pgtype` au profit des types Go natifs

### Corrigé
- Cookie `Secure` + `SameSite` et validation UUID renforcée dans les handlers

## [2026-02-18]

### Ajouté
- Route `GET /me` pour le profil de l'utilisateur courant
- Endpoint `GET /dashboard`
- Endpoint profil étudiant
- Enrichissement des classes (stats, students count)
- Enrichissement des étudiants, punitions, bonus, pénalités et règles
- Recherche sur étudiants, bonus et punitions

### Modifié
- Split du profil étudiant en deux réponses : KPIs et historique
- Remplacement de `student_profile` par `student_kpis_history`

### Corrigé
- Configuration CORS entièrement pilotable via `.env`

## [2026-02-17]

### Modifié
- Documentation du plan d'enrichissement de l'API

## [2026-02-15]

### Ajouté
- Middleware CORS initial

## [2026-02-14]

### Ajouté
- Fonctionnalité « règles automatiques » : CRUD + déclenchement de punitions
- CRUD punitions et catégories de punitions
- CRUD pénalités et catégories de pénalités
- Paramètre fonctionnel supplémentaire pour les bonus
- Workflow CI `.github/workflows/tests.yml`
- Première couverture de tests des handlers

### Modifié
- README initial
- Mise à jour et simplification de la documentation des règles

### Corrigé
- Création atomique pénalité + punition associée (transaction)
- Correction des triggers SQL de pénalité

## [2026-02-13]

### Ajouté
- CRUD bonus et catégories de bonus
- CRUD des classes

### Modifié
- Passe de lisibilité globale sur le code
- Série de mises à jour documentaires (nomenclature, responses enrichies)

### Corrigé
- Renommage du champ `bonus.name`

## [2026-02-11]

### Ajouté
- CRUD des élèves

## [2026-02-10]

### Ajouté
- Variable d'environnement pour activer ou désactiver l'inscription

## [2025-12-21]

### Corrigé
- Amélioration de la gestion d'erreurs HTTP/domain

## [2025-12-20]

### Corrigé
- Amélioration de la gestion d'erreurs HTTP/domain

## [2025-12-16]

### Ajouté
- Stockage de l'IP associée à chaque refresh token (audit)

## [2025-12-15]

### Ajouté
- Persistance des refresh tokens en base
- Authentification JWT (access token) sans persistance

## [2025-12-14]

### Ajouté
- Seeder (`internal/seeder/`)
- Type `APIError`
- Premiers tests d'erreurs
- Validator et décodage JSON centralisé

### Modifié
- Mise en place de l'architecture en couches (handler / service / repository / dto / platform)
- Retour à la stack `repository` + `sqlc` après expérimentations

## [2025-12-13]

### Ajouté
- Première intégration JWT
- Tests initiaux
- Couches `domain` et `repository` de base

## [2025-12-12]

### Ajouté
- Bootstrap du projet
