# API Endpoints Documentation

Ce document décrit tous les endpoints API disponibles dans l'application The Punisher.

## Table des matières

- [Authentification](#authentification)
- [Utilisateurs](#utilisateurs)
- [Étudiants](#étudiants)
- [Classes](#classes)
- [Types de Bonus](#types-de-bonus)
- [Santé](#santé)

---

## Authentification

### POST /v1/auth/register

Création d'un nouveau compte utilisateur (professeur).

**Authentification requise :** Non

**Corps de requête :**
```json
{
  "email": "prof@example.com",
  "first_name": "Jean",
  "last_name": "Dupont",
  "password": "motdepasse123"
}
```

**Validation :**
- `email` : requis, format email valide
- `first_name` : requis, minimum 2 caractères, maximum 70 caractères
- `last_name` : requis, minimum 2 caractères, maximum 70 caractères
- `password` : requis, minimum 8 caractères

**Réponse (201 Created) :**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "prof@example.com",
  "first_name": "Jean",
  "last_name": "Dupont",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
  ```json
  {
    "error": "validation_failed",
    "error_code": 400,
    "error_details": [
      {
        "field": "email",
        "error": "validation_invalid_email"
      }
    ]
  }
  ```
- `409 Conflict` : Email déjà utilisé
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

---

### POST /v1/auth/login

Connexion d'un utilisateur existant. Retourne un access token (JWT) et définit un refresh token dans un cookie HTTP-Only.

**Authentification requise :** Non

**Corps de requête :**
```json
{
  "email": "prof@example.com",
  "password": "motdepasse123"
}
```

**Validation :**
- `email` : requis, format email valide
- `password` : requis, minimum 8 caractères

**Réponse (200 OK) :**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Cookies définis :**
- `refresh_token` : Token opaque, HTTP-Only, Secure, SameSite=Strict, expire après 7 jours

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `401 Unauthorized` : Identifiants incorrects
  ```json
  {
    "error": "invalid_credentials_or_user_doesnt_exist",
    "error_code": 401
  }
  ```

**Utilisation de l'access token :**

Pour les requêtes authentifiées, ajouter le header :
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### POST /v1/auth/refresh

Renouvellement de l'access token à l'aide du refresh token stocké dans le cookie.

**Authentification requise :** Oui (via cookie refresh_token)

**Corps de requête :** Aucun

**Réponse (200 OK) :**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Cookies définis :**
- `refresh_token` : Nouveau refresh token, HTTP-Only, Secure, SameSite=Strict

**Erreurs possibles :**
- `401 Unauthorized` : Refresh token invalide ou expiré

**Notes :**
- Le refresh token est automatiquement révoqué et un nouveau est créé
- L'ancien refresh token ne peut plus être utilisé après cette opération

---

## Utilisateurs

Les utilisateurs représentent les professeurs qui utilisent l'application.

**Note :** Pour l'instant, l'API ne propose que l'endpoint de création (via `/v1/auth/register`). Les endpoints de mise à jour, suppression ou liste ne sont pas encore implémentés.

---

## Étudiants

Les étudiants appartiennent à un utilisateur (professeur) et peuvent être membres de plusieurs classes.

### POST /v1/students

Création d'un nouvel étudiant.

**Authentification requise :** Oui

**Corps de requête :**
```json
{
  "first_name": "Alice",
  "last_name": "Martin"
}
```

**Validation :**
- `first_name` : requis, minimum 2 caractères, maximum 70 caractères
- `last_name` : requis, minimum 2 caractères, maximum 70 caractères

**Réponse (201 Created) :**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "first_name": "Alice",
  "last_name": "Martin",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `401 Unauthorized` : Non authentifié

---

### GET /v1/students

Liste paginée des étudiants de l'utilisateur authentifié.

**Authentification requise :** Oui

**Query Parameters :**
- `page` (optionnel) : Numéro de page (défaut: 1)

**Réponse (200 OK) :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 45,
  "previous_page": null,
  "next_page": 2,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174001",
      "first_name": "Alice",
      "last_name": "Martin",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174002",
      "first_name": "Bob",
      "last_name": "Durand",
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**Notes sur la pagination :**
- Par défaut, 20 éléments par page
- `previous_page` : null si on est sur la première page, sinon numéro de la page précédente
- `next_page` : null si on est sur la dernière page, sinon numéro de la page suivante
- Les étudiants sont triés par date de création (plus récents en premier)

---

### GET /v1/students/{id}

Récupération d'un étudiant spécifique.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de l'étudiant

**Réponse (200 OK) :**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "first_name": "Alice",
  "last_name": "Martin",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
  ```json
  {
    "error": "malformed_parameter",
    "error_code": 400
  }
  ```
- `404 Not Found` : Étudiant non trouvé
  ```json
  {
    "error": "student_not_found",
    "error_code": 404
  }
  ```

---

### PUT /v1/students/{id}

Mise à jour d'un étudiant. Tous les champs sont optionnels.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de l'étudiant

**Corps de requête :**
```json
{
  "first_name": "Alice",
  "last_name": "Martin-Dupont"
}
```

**Validation :**
- `first_name` (optionnel) : minimum 2 caractères, maximum 70 caractères
- `last_name` (optionnel) : minimum 2 caractères, maximum 70 caractères

**Réponse (200 OK) :**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "first_name": "Alice",
  "last_name": "Martin-Dupont",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T14:20:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `404 Not Found` : Étudiant non trouvé

**Note :**
- Utilise `COALESCE` en base de données : seuls les champs fournis sont mis à jour
- Les champs non fournis conservent leur valeur actuelle

---

### DELETE /v1/students/{id}

Suppression d'un étudiant. **Attention : Suppression physique (hard delete) définitive.**

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de l'étudiant

**Réponse (204 No Content) :**

Aucun corps de réponse.

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Étudiant non trouvé

**Note importante :**
- La suppression est **définitive** (hard delete)
- Si l'étudiant est lié à des classes, ces relations sont supprimées en cascade (grâce aux contraintes ON DELETE CASCADE en base)
- Si l'étudiant a des bonus, pénalités ou punitions, ces données sont également supprimées en cascade

---

### GET /v1/students/{id}/classrooms

Liste paginée des classes auxquelles appartient l'étudiant.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de l'étudiant

**Query Parameters :**
- `page` (optionnel) : Numéro de page (défaut: 1)

**Réponse (200 OK) :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 3,
  "previous_page": null,
  "next_page": null,
  "data": [
    {
      "id": "223e4567-e89b-12d3-a456-426614174001",
      "name": "3ème A",
      "year": "2023-2024",
      "main_teacher": "M. Dupont",
      "created_at": "2024-01-10T09:00:00Z",
      "updated_at": "2024-01-10T09:00:00Z"
    }
  ]
}
```

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Étudiant non trouvé

---

## Classes

Les classes (classrooms) appartiennent à un utilisateur (professeur) et peuvent contenir plusieurs étudiants.

### POST /v1/classrooms

Création d'une nouvelle classe.

**Authentification requise :** Oui

**Corps de requête :**
```json
{
  "name": "3ème A",
  "year": "2023-2024",
  "main_teacher": "M. Dupont"
}
```

**Validation :**
- `name` : requis, minimum 2 caractères, maximum 100 caractères
- `year` (optionnel) : maximum 20 caractères
- `main_teacher` (optionnel) : maximum 100 caractères

**Réponse (201 Created) :**
```json
{
  "id": "223e4567-e89b-12d3-a456-426614174001",
  "name": "3ème A",
  "year": "2023-2024",
  "main_teacher": "M. Dupont",
  "created_at": "2024-01-10T09:00:00Z",
  "updated_at": "2024-01-10T09:00:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `401 Unauthorized` : Non authentifié

---

### GET /v1/classrooms

Liste paginée des classes de l'utilisateur authentifié.

**Authentification requise :** Oui

**Query Parameters :**
- `page` (optionnel) : Numéro de page (défaut: 1)

**Réponse (200 OK) :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 12,
  "previous_page": null,
  "next_page": null,
  "data": [
    {
      "id": "223e4567-e89b-12d3-a456-426614174001",
      "name": "3ème A",
      "year": "2023-2024",
      "main_teacher": "M. Dupont",
      "created_at": "2024-01-10T09:00:00Z",
      "updated_at": "2024-01-10T09:00:00Z"
    },
    {
      "id": "223e4567-e89b-12d3-a456-426614174002",
      "name": "4ème B",
      "year": "2023-2024",
      "main_teacher": null,
      "created_at": "2024-01-10T09:15:00Z",
      "updated_at": "2024-01-10T09:15:00Z"
    }
  ]
}
```

**Notes :**
- Les classes sont triées par date de création (plus récentes en premier)
- Les champs `year` et `main_teacher` peuvent être `null`

---

### GET /v1/classrooms/{id}

Récupération d'une classe spécifique.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe

**Réponse (200 OK) :**
```json
{
  "id": "223e4567-e89b-12d3-a456-426614174001",
  "name": "3ème A",
  "year": "2023-2024",
  "main_teacher": "M. Dupont",
  "created_at": "2024-01-10T09:00:00Z",
  "updated_at": "2024-01-10T09:00:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Classe non trouvée
  ```json
  {
    "error": "classroom_not_found",
    "error_code": 404
  }
  ```

---

### PUT /v1/classrooms/{id}

Mise à jour d'une classe. Tous les champs sont optionnels.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe

**Corps de requête :**
```json
{
  "name": "3ème A - Section Européenne",
  "year": "2024-2025"
}
```

**Validation :**
- `name` (optionnel) : minimum 2 caractères, maximum 100 caractères
- `year` (optionnel) : maximum 20 caractères
- `main_teacher` (optionnel) : maximum 100 caractères

**Réponse (200 OK) :**
```json
{
  "id": "223e4567-e89b-12d3-a456-426614174001",
  "name": "3ème A - Section Européenne",
  "year": "2024-2025",
  "main_teacher": "M. Dupont",
  "created_at": "2024-01-10T09:00:00Z",
  "updated_at": "2024-01-15T16:30:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `404 Not Found` : Classe non trouvée

---

### DELETE /v1/classrooms/{id}

Suppression d'une classe. **Attention : Suppression physique (hard delete) définitive.**

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe

**Réponse (204 No Content) :**

Aucun corps de réponse.

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Classe non trouvée

**Note importante :**
- La suppression est **définitive** (hard delete)
- Les relations avec les étudiants sont supprimées en cascade (via ON DELETE CASCADE)

---

### POST /v1/classrooms/{id}/students

Ajout d'un étudiant à une classe.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe

**Corps de requête :**
```json
{
  "student_id": "123e4567-e89b-12d3-a456-426614174001"
}
```

**Validation :**
- `student_id` : requis, format UUID valide

**Réponse (204 No Content) :**

Aucun corps de réponse.

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `404 Not Found` : Étudiant ou classe non trouvé(e)
  ```json
  {
    "error": "student_or_classroom_not_found",
    "error_code": 404
  }
  ```
- `409 Conflict` : L'étudiant est déjà dans cette classe
  ```json
  {
    "error": "student_classroom_relation_exists",
    "error_code": 409
  }
  ```

**Note :**
- Vérifie que l'étudiant et la classe appartiennent bien à l'utilisateur authentifié
- Un étudiant peut appartenir à plusieurs classes

---

### DELETE /v1/classrooms/{id}/students/{studentId}

Retrait d'un étudiant d'une classe.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe
- `studentId` : UUID de l'étudiant

**Réponse (204 No Content) :**

Aucun corps de réponse.

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Relation non trouvée ou classe n'appartient pas à l'utilisateur

**Note :**
- Ne supprime pas l'étudiant, mais seulement sa relation avec la classe
- Si l'étudiant n'est pas dans la classe, retourne 404

---

### GET /v1/classrooms/{id}/students

Liste paginée des étudiants d'une classe.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID de la classe

**Query Parameters :**
- `page` (optionnel) : Numéro de page (défaut: 1)

**Réponse (200 OK) :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 28,
  "previous_page": null,
  "next_page": 2,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174001",
      "first_name": "Alice",
      "last_name": "Martin",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174002",
      "first_name": "Bob",
      "last_name": "Durand",
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Classe non trouvée

---

## Types de Bonus

Les types de bonus définissent les catégories de récompenses qu'un professeur peut attribuer (ex: "Participation active", "Devoir rendu en avance").

### POST /v1/bonus-types

Création d'un nouveau type de bonus.

**Authentification requise :** Oui

**Corps de requête :**
```json
{
  "name": "Participation active"
}
```

**Validation :**
- `name` : requis, minimum 2 caractères, maximum 100 caractères

**Réponse (201 Created) :**
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174001",
  "name": "Participation active",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `401 Unauthorized` : Non authentifié

---

### GET /v1/bonus-types

Liste paginée des types de bonus de l'utilisateur authentifié.

**Authentification requise :** Oui

**Query Parameters :**
- `page` (optionnel) : Numéro de page (défaut: 1)

**Réponse (200 OK) :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 5,
  "previous_page": null,
  "next_page": null,
  "data": [
    {
      "id": "323e4567-e89b-12d3-a456-426614174001",
      "name": "Participation active",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    },
    {
      "id": "323e4567-e89b-12d3-a456-426614174002",
      "name": "Devoir rendu en avance",
      "created_at": "2024-01-15T10:05:00Z",
      "updated_at": "2024-01-15T10:05:00Z"
    }
  ]
}
```

---

### GET /v1/bonus-types/{id}

Récupération d'un type de bonus spécifique.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID du type de bonus

**Réponse (200 OK) :**
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174001",
  "name": "Participation active",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Type de bonus non trouvé
  ```json
  {
    "error": "bonus_type_not_found",
    "error_code": 404
  }
  ```

---

### PUT /v1/bonus-types/{id}

Mise à jour d'un type de bonus.

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID du type de bonus

**Corps de requête :**
```json
{
  "name": "Participation exceptionnelle"
}
```

**Validation :**
- `name` (optionnel) : minimum 2 caractères, maximum 100 caractères

**Réponse (200 OK) :**
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174001",
  "name": "Participation exceptionnelle",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T14:30:00Z"
}
```

**Erreurs possibles :**
- `400 Bad Request` : Données invalides
- `404 Not Found` : Type de bonus non trouvé

---

### DELETE /v1/bonus-types/{id}

Suppression d'un type de bonus. **Attention : Suppression physique (hard delete) définitive.**

**Authentification requise :** Oui

**Path Parameters :**
- `id` : UUID du type de bonus

**Réponse (204 No Content) :**

Aucun corps de réponse.

**Erreurs possibles :**
- `400 Bad Request` : ID malformé
- `404 Not Found` : Type de bonus non trouvé

**Note importante :**
- La suppression est **définitive** (hard delete)
- Si des bonus instances utilisent ce type, ils seront également supprimés en cascade (lorsque la table `bonuses` sera implémentée)

---

## Santé

### GET /v1/health

Endpoint de vérification de l'état de l'application.

**Authentification requise :** Non

**Réponse (200 OK) :**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "database": "connected"
}
```

**Utilisation :**
- Monitoring de l'application
- Health checks pour les conteneurs Docker / Kubernetes
- Vérification de la connexion à la base de données

---

## Codes d'erreur communs

### Erreurs d'authentification

- `401 Unauthorized` : Token manquant, invalide ou expiré
  ```json
  {
    "error": "unauthorized",
    "error_code": 401
  }
  ```

### Erreurs de validation

- `400 Bad Request` : Données de requête invalides
  ```json
  {
    "error": "validation_failed",
    "error_code": 400,
    "error_details": [
      {
        "field": "first_name",
        "error": "validation_min_length:2"
      }
    ]
  }
  ```

### Erreurs serveur

- `500 Internal Server Error` : Erreur serveur interne
  ```json
  {
    "error": "internal_error",
    "error_code": 500
  }
  ```

---

## Notes importantes

### Isolation des données

Toutes les ressources (étudiants, classes, types de bonus, etc.) sont isolées par utilisateur (`user_id`). Un utilisateur ne peut voir et modifier que ses propres données.

### Pagination

- Par défaut : 20 éléments par page
- Page par défaut : 1
- Le format de réponse paginée est toujours identique :
  ```json
  {
    "page": <numéro de page actuel>,
    "item_per_page": 20,
    "total_count": <nombre total d'éléments>,
    "previous_page": <numéro page précédente ou null>,
    "next_page": <numéro page suivante ou null>,
    "data": [...]
  }
  ```

### Suppressions

- Toutes les suppressions sont **physiques** (hard delete)
- Les données supprimées ne peuvent pas être récupérées
- Les relations en cascade sont gérées automatiquement par la base de données

### Timestamps

- Tous les timestamps sont au format ISO 8601 (UTC)
- Format : `2024-01-15T10:30:00Z`
- `created_at` : Date de création de la ressource
- `updated_at` : Date de dernière modification de la ressource
