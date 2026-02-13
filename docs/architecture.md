# Architecture & Bonnes Pratiques

Ce document décrit l'architecture du projet `the-punisher-go` et les conventions à respecter pour tout développement.

## Sommaire

- [Architecture \& Bonnes Pratiques](#architecture--bonnes-pratiques)
  - [Sommaire](#sommaire)
  - [1. Architecture Globale](#1-architecture-globale)
    - [Flux de données](#flux-de-données)
  - [2. Sécurité \& Authentification](#2-sécurité--authentification)
    - [Stratégie de Tokens](#stratégie-de-tokens)
    - [Middleware](#middleware)
  - [3. Gestion des Erreurs](#3-gestion-des-erreurs)
    - [Type `APIError`](#type-apierror)
    - [Pattern](#pattern)
    - [Erreurs Fréquentes](#erreurs-fréquentes)
  - [4. Couche API (Handlers \& Routes)](#4-couche-api-handlers--routes)
    - [Routes (`cmd/api/routes.go`)](#routes-cmdapiroutesgo)
    - [Handlers (`internal/api/handler`)](#handlers-internalapihandler)
    - [Pagination](#pagination)
  - [5. Services \& Logique Métier](#5-services--logique-métier)
    - [Interfaces (`internal/service`)](#interfaces-internalservice)
    - [Responsabilités](#responsabilités)
  - [6. DTOs (Data Transfer Objects)](#6-dtos-data-transfer-objects)
    - [Nomenclature](#nomenclature)
  - [7. Base de Données (Repository)](#7-base-de-données-repository)
    - [Bonnes pratiques SQLC](#bonnes-pratiques-sqlc)
  - [8. Logging](#8-logging)
  - [9. Configuration](#9-configuration)
  - [10. Tests (Stratégie)](#10-tests-stratégie)
  - [11. Workflow de Développement](#11-workflow-de-développement)
    - [Commandes utiles (Makefile)](#commandes-utiles-makefile)

---

## 1. Architecture Globale

Le projet suit une architecture en couches (Layered Architecture) classique en Go, favorisant la séparation des responsabilités et la testabilité.

```
the-punisher-go/
├── cmd/
│   └── api/            # Point d'entrée de l'application
│       ├── main.go     # Lifecycle du serveur et connexion DB
│       └── routes.go   # Wiring des dépendances et définition des routes
├── internal/
│   ├── api/            # Couche HTTP
│   │   ├── handler/    # Contrôleurs HTTP (traitement requête/réponse)
│   │   └── middleware/ # Middlewares (Auth, Logging, etc.)
│   ├── service/        # Logique métier (Business Logic)
│   ├── repository/     # Accès aux données (généré par SQLC)
│   ├── dto/            # Objets de transfert de données (Data Transfer Objects)
│   └── platform/       # Code technique transverse
│       ├── auth/       # Middleware d'authentification
│       ├── config/     # Chargement de la configuration
│       ├── hash/       # Hashing de mots de passe (Argon2/Bcrypt)
│       ├── jwt/        # Gestion des tokens JWT
│       ├── validator/  # Validation des structs
│       └── web/        # Helpers HTTP (Réponse JSON, Erreurs, Pagination)
└── db/
    ├── migrations/     # Migrations SQL
    └── sqlc/           # Requêtes SQL brutes pour génération
```

### Flux de données
`Request` -> `Router (Chi)` -> `Handler` -> `Service` -> `Repository` -> `Database`

---

## 2. Sécurité & Authentification

L'authentification repose sur un système hybride **JWT (Access Token) + Cookie (Refresh Token)** pour allier sécurité et expérience utilisateur.

### Stratégie de Tokens
1. **Access Token (JWT)** :
   - **Durée** : Courte (ex: 15 minutes).
   - **Transport** : Header HTTP `Authorization: Bearer <token>`.
   - **Contenu** : Claims standards + `user_id`.
   - **Usage** : Authentifie chaque requête API stateless.

2. **Refresh Token (Opaque)** :
   - **Durée** : Longue (ex: 7 jours).
   - **Stockage** : Base de données (`refresh_tokens` table) pour permettre la révocation.
   - **Transport** : **Cookie HTTP-Only, Secure, SameSite=Strict**.
   - **Usage** : Utilisé uniquement sur l'endpoint `/v1/auth/refresh` pour obtenir une nouvelle paire de tokens.
   - **Sécurité** : Protégé contre le vol XSS (grâce au cookie HttpOnly) et permet la déconnexion forcée (révocation en base).

### Middleware
Le middleware `auth.AuthMiddleware` :
1. Intercepte les requêtes protégées.
2. Extrait et valide le JWT du header `Authorization`.
3. Injecte le `user_id` dans le contexte de la requête (`auth.UserIDFromContext(ctx)`).
4. Renvoie une erreur 401 si le token est invalide ou absent.

---

## 3. Gestion des Erreurs

La gestion des erreurs est centralisée et typée pour garantir des réponses API cohérentes.

### Type `APIError`
Toutes les erreurs métier sont définies dans `internal/api/errors.go` et sont du type `*api.APIError`.
Ce type encapsule :
- Un **Message** (clé de l'erreur, ex: `student_not_found`)
- Un **StatusCode** (HTTP status, ex: 404)
- Des **Details** optionnels (champs en erreur)

### Pattern
1. **Service** : Retourne une erreur sentinelle prédéfinie (`api.ErrStudentNotFound`) ou une erreur wrappée (`fmt.Errorf("failed to...: %w", err)`).
   ```go
   // internal/service/student.go
   if rowsAffected == 0 {
       return api.ErrStudentNotFound // 404 automatique
   }
   return fmt.Errorf("failed to delete student: %w", err) // 500 interne
   ```

2. **Handler** : Utilise `web.WriteFromError(w, err)` qui gère automatiquement le mapping.
   - Si l'erreur est (ou contient via `errors.As`) un `APIError`, la réponse aura le bon status code + le message JSON.
   - Sinon, c'est considéré comme une erreur serveur (500) et loggué comme telle.

### Erreurs Fréquentes
- `api.ErrInternalError` (500)
- `api.ErrUnauthorized` (401)
- `api.ErrValidationFailed` (400) - Pour les erreurs de validation `go-playground/validator`.
- `api.ErrInvalidRequestBody` (400) - Pour les JSON malformés.

---

## 4. Couche API (Handlers & Routes)

### Routes (`cmd/api/routes.go`)
- Les routes sont définies dans `routes.go` pour ne pas polluer le `main.go`.
- On utilise `chi` comme routeur.
- La nomenclature des routes est **RESTful**, pluriel, en `kebab-case` (minuscules).

**Exemples :**
- `POST /v1/auth/register`, `/v1/auth/login` (Auth)
- `POST /v1/students` (Création étudiant)
- `GET /v1/students?page=1` (Liste paginée)
- `GET /v1/classrooms/{id}/students` (Liste étudiants d'une classe)

### Handlers (`internal/api/handler`)
- **Responsabilité** : Parsing requête (body, params), validation (DTO), appel service, réponse.
- **Décodage JSON** : Toujours utiliser `web.DecodeJSON(r, &dto)`.
  - Limite stricte de **1MB** (`http.MaxBytesReader`).
  - Refuse les champs inconnus (`DisallowUnknownFields`).
- **Validation** : Toujours valider les inputs avec `validator.ValidateStruct(&req)`.
- **Réponse** : Utiliser `web.WriteJSON` (succès) ou `web.WriteFromError` (échec).

### Pagination
Les endpoints de liste (ex: `ListStudents`, `ListClassrooms`) supportent la pagination via le query param `page`.
- **Défaut** : Page 1, 20 items par page.
- **Format de Réponse** :
  ```json
  {
      "page": 1,
      "item_per_page": 20,
      "total_count": 50,
      "previous_page": null,
      "next_page": 2,
      "data": [ ... ]
  }
  ```
- **Helper** : Utiliser `web.ParsePagination(r)` et `web.NewPaginatedResponse(...)`.

---

## 5. Services & Logique Métier

### Interfaces (`internal/service`)
- Chaque service expose une **interface** (ex: `StudentService`, `ClassroomService`, `AuthService`).
- L'implémentation concrète (ex: `studentService`) est privée/non-exportée.
- Le constructeur retourne l'interface : `func NewStudentService(...) StudentService`.
- **Pourquoi ?** Facilite le mocking pour les tests et découple le handler de l'implémentation.

### Responsabilités
- Appliquer les règles métier.
- Coordonner les appels au Repository.
- Gérer les transactions (si nécessaire).
- Vérifier l'existence des données :
  - **DELETE / UPDATE** : Vérifier `rowsAffected` avec `:execrows` dans SQLC. Si 0 ligne modifiée => Retourner une erreur Not Found (`api.ErrStudentNotFound`).
- **Suppression** : Actuellement, la suppression est **physique** (Hard Delete). Les données sont définitivement supprimées de la base.
- Convertir les modèles DB en DTOs de retour via les fonctions de mapping.

---

## 6. DTOs (Data Transfer Objects)

Les DTOs (`internal/dto`) séparent le contrat d'API du modèle de base de données.

### Nomenclature
- `Request[Entity]Dto` : Pour les payloads de requête (POST/PUT).
- `Return[Entity]Dto` : Pour les payloads de réponse.
- `[Entity]FromRepository` : Fonction de mapping pure (DB Model -> DTO).
- `[Entity]ListFromRepository` : Fonction de mapping pour les listes.

```go
// internal/dto/user.go
func UserFromRepository(u *repository.User) *ReturnUserDto {
    return &ReturnUserDto{
        ID:    u.ID,
        Email: u.Email,
        // ...
    }
}
```

---

## 7. Base de Données (Repository)

- **Outil** : `sqlc` génère le code Go type-safe à partir de requêtes SQL.
- **Fichier SQL** : `db/sqlc/queries.sql`.
- **Génération** : `sqlc generate` (ou `make sqlc` si configuré).

### Bonnes pratiques SQLC
- Utiliser `sqlc.arg(param_name)` pour nommer les bind params.
- Utiliser `COALESCE` pour les mises à jour partielles (PATCH/PUT).
- Utiliser `:execrows` pour les `DELETE`, `UPDATE` et `INSERT` conditionnels afin de savoir si l'opération a réussi et retourner les bonnes erreurs métier.

---

## 8. Logging

Utilisation de `log/slog` (Go standard library).

- **Où ?** Dans la couche **Service** principalement.
- **Quand ?** Pour les opérations de mutation (création, suppresion, échecs critiques).
- **Format ?** Logs structurés JSON (par défaut en prod) ou Text (en dev).
- **Contexte ?** Toujours inclure IDs, UserID, etc.

```go
slog.Info("student created", "student_id", student.ID, "user_id", userID)
```

---

## 9. Configuration

- La configuration est chargée depuis les variables d'environnement (`.env`) via `internal/platform/config`.
- Utilisation de `joho/godotenv` pour le chargement local.
- Structure `Config` centralisée passée à l'application.

---

## 10. Tests (Stratégie)

*(À implémenter)*

- **Unit Tests** : Mocker les interfaces Repository (`Querier`) pour tester la logique des Services.
- **Integration Tests** : Tester les Handlers avec une fausse DB (ou dockerizée) et de vraies requêtes HTTP via `httptest`.
- **E2E Tests** : Scénarios complets sur l'API running.

---

## 11. Workflow de Développement

### Commandes utiles (Makefile)

- `make dev` : Lance le serveur en mode développement avec hot-reload (`air`).
- `make build` : Compile le binaire de production.
- `sqlc generate` : Régénère le code Go après modification de `queries.sql`.
- `migrate create -ext sql -dir db/migrations -seq [nom]` : Crée une nouvelle migration.
