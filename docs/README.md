# Documentation The Punisher Backend

Bienvenue dans la documentation complète du backend The Punisher.

## 📚 Table des matières

### Pour démarrer
- [Projet.md](./projet.md) - Vue d'ensemble du projet et objectifs
- [Architecture.md](./architecture.md) - Architecture technique et bonnes pratiques

### Pour utiliser l'API
- [API Endpoints](./api-endpoints.md) - Documentation complète de tous les endpoints REST
- [Exemples](./examples.md) - Exemples concrets et cas d'usage

### Pour développer
- [Guide d'Implémentation](./feature-implementation-guide.md) - Guide détaillé pour ajouter des fonctionnalités
- [Fonctionnalités Manquantes](./missing-features.md) - Liste des features à implémenter

---

## 🎯 Quel document lire ?

### Je suis un utilisateur de l'API
👉 Commencez par **[API Endpoints](./api-endpoints.md)** pour découvrir tous les endpoints disponibles.  
👉 Consultez **[Exemples](./examples.md)** pour voir des cas d'usage concrets avec cURL.

### Je suis un développeur qui rejoint le projet
👉 Lisez d'abord **[Projet.md](./projet.md)** pour comprendre les objectifs.  
👉 Puis **[Architecture.md](./architecture.md)** pour comprendre la structure du code.  
👉 Enfin **[Guide d'Implémentation](./feature-implementation-guide.md)** pour savoir comment ajouter des features.

### Je suis une IA qui doit ajouter une fonctionnalité
👉 Lisez **[Guide d'Implémentation](./feature-implementation-guide.md)** - c'est conçu pour vous !  
👉 Consultez **[Fonctionnalités Manquantes](./missing-features.md)** pour voir ce qui est à faire.  
👉 Référez-vous à **[Architecture.md](./architecture.md)** pour les conventions.

### Je veux comprendre le modèle de données
👉 Consultez **[Projet.md](./projet.md)** - Section "Modèle de Données" avec le diagramme ER.

---

## 📖 Résumé des documents

### [Projet.md](./projet.md)
**Objectif :** Vue d'ensemble du projet The Punisher

**Contenu :**
- Objectifs et vision du projet
- Fonctionnalités clés
- Modèle de données complet (diagramme ER)
- Règles métier
- Système de points (bonus/malus)
- Système de règles automatiques

**À lire si :** Vous découvrez le projet ou voulez comprendre le business.

---

### [Architecture.md](./architecture.md)
**Objectif :** Architecture technique et conventions de développement

**Contenu :**
- Architecture en couches (Handlers → Services → Repository)
- Gestion de la sécurité et authentification (JWT + Refresh Token)
- Gestion des erreurs
- Pagination
- DTOs et mapping
- SQLC et base de données
- Logging avec slog
- Workflow de développement

**À lire si :** Vous voulez comprendre comment le code est organisé.

---

### [API Endpoints](./api-endpoints.md)
**Objectif :** Documentation complète de l'API REST

**Contenu :**
- **Authentification** : `/v1/auth/register`, `/v1/auth/login`, `/v1/auth/refresh`
- **Étudiants** : CRUD complet + relations avec classes
- **Classes** : CRUD complet + gestion des étudiants
- **Types de Bonus** : CRUD complet
- **Santé** : Health check

Pour chaque endpoint :
- Méthode HTTP et URL
- Authentification requise ou non
- Corps de requête avec validation
- Réponse attendue (JSON)
- Exemples de requêtes/réponses
- Erreurs possibles

**À lire si :** Vous développez un frontend ou utilisez l'API.

---

### [Exemples](./examples.md)
**Objectif :** Cas d'usage concrets avec exemples cURL

**Contenu :**
- Scénarios complets (inscription, configuration, utilisation)
- Exemples de requêtes cURL pour tous les endpoints
- Cas d'usage typiques (gestion de classe, recherche, mise à jour)
- Gestion des erreurs avec exemples
- Scripts bash utiles

**À lire si :** Vous voulez tester l'API rapidement.

---

### [Guide d'Implémentation](./feature-implementation-guide.md)
**Objectif :** Guide step-by-step pour implémenter une nouvelle fonctionnalité

**Contenu :**
- Vue d'ensemble du processus (8 étapes)
- Exemples de référence (Student, Classroom, BonusType)
- Détails de chaque étape :
  1. Migration de base de données
  2. Requêtes SQL (SQLC)
  3. Génération du code
  4. DTOs
  5. Service
  6. Handler
  7. Routes
  8. Erreurs
- Checklist complète
- Patterns et conventions
- Exemples complets de code

**À lire si :** Vous ajoutez une nouvelle entité ou fonctionnalité.

---

### [Fonctionnalités Manquantes](./missing-features.md)
**Objectif :** Liste détaillée des features à implémenter

**Contenu :**
- Vue d'ensemble (implémenté ✅ vs manquant ❌)
- Détails de chaque entité manquante :
  - **PenaltyTypes** : Types de pénalités
  - **PunishmentTypes** : Types de punitions
  - **Bonuses** : Instances de bonus
  - **Penalties** : Instances de pénalités
  - **Punishments** : Instances de punitions
  - **Rules** : Règles automatiques (complexe)
  - **Users** : Endpoints complets

Pour chaque entité :
- Description
- Structure de la table SQL
- DTOs
- Endpoints à implémenter
- Logique métier spécifique
- Requêtes SQL supplémentaires
- Exemples de code

**À lire si :** Vous voulez implémenter une fonctionnalité manquante.

---

## 🔍 Recherche rapide

### Comment faire X ?

| Je veux... | Document à consulter | Section |
|------------|---------------------|---------|
| Créer un compte utilisateur | [API Endpoints](./api-endpoints.md) | Authentification → POST /v1/auth/register |
| Me connecter | [API Endpoints](./api-endpoints.md) | Authentification → POST /v1/auth/login |
| Créer un étudiant | [API Endpoints](./api-endpoints.md) | Étudiants → POST /v1/students |
| Voir mes classes | [API Endpoints](./api-endpoints.md) | Classes → GET /v1/classrooms |
| Ajouter une nouvelle entité | [Guide d'Implémentation](./feature-implementation-guide.md) | Étapes détaillées |
| Comprendre le modèle de données | [Projet.md](./projet.md) | Section 3 : Modèle de Données |
| Comprendre l'architecture | [Architecture.md](./architecture.md) | Section 1 : Architecture Globale |
| Gérer les erreurs | [Architecture.md](./architecture.md) | Section 3 : Gestion des Erreurs |
| Tester l'API avec cURL | [Exemples](./examples.md) | Exemples de requêtes cURL |
| Implémenter les pénalités | [Fonctionnalités Manquantes](./missing-features.md) | Section 4 : Penalties |
| Implémenter les règles | [Fonctionnalités Manquantes](./missing-features.md) | Section 6 : Rules |

---

## 🚀 Démarrage rapide

### 1. Pour un développeur backend

```bash
# Lire dans cet ordre :
1. docs/projet.md                           # Comprendre le projet
2. docs/architecture.md                     # Comprendre l'architecture
3. docs/feature-implementation-guide.md     # Apprendre à développer
```

### 2. Pour un développeur frontend

```bash
# Lire dans cet ordre :
1. docs/api-endpoints.md    # Documentation de l'API
2. docs/examples.md          # Exemples d'utilisation
```

### 3. Pour une IA

```bash
# Lire dans cet ordre :
1. docs/feature-implementation-guide.md  # Guide complet pour implémenter
2. docs/missing-features.md              # Ce qui est à faire
3. docs/architecture.md                  # Conventions et patterns
```

---

## 📊 État d'avancement du projet

### Implémenté ✅
- ✅ Authentification (JWT + Refresh Token)
- ✅ Users (création uniquement)
- ✅ Students (CRUD complet)
- ✅ Classrooms (CRUD complet)
- ✅ Relations Students ↔ Classrooms (many-to-many)
- ✅ BonusTypes (CRUD complet)

### À implémenter ❌
- ❌ PenaltyTypes (types de pénalités)
- ❌ PunishmentTypes (types de punitions)
- ❌ Bonuses (instances de bonus)
- ❌ Penalties (instances de pénalités)
- ❌ Punishments (instances de punitions)
- ❌ Rules (règles automatiques)
- ❌ Users (endpoints complets : update, delete, profil)

**Voir [Fonctionnalités Manquantes](./missing-features.md) pour plus de détails.**

---

## 🛠️ Technologies utilisées

- **Langage** : Go 1.21+
- **Framework HTTP** : Chi router
- **Base de données** : PostgreSQL
- **ORM** : SQLC (génération de code type-safe)
- **Migrations** : golang-migrate
- **Authentification** : JWT (golang-jwt/jwt)
- **Hashing** : Argon2 / Bcrypt
- **Validation** : go-playground/validator
- **Logging** : log/slog (Go standard library)
- **Hot reload** : Air

---

## 📝 Conventions de code

### Nommage

| Type | Convention | Exemple |
|------|-----------|---------|
| Tables DB | pluriel, snake_case | `students`, `bonus_types` |
| Colonnes DB | singular, snake_case | `first_name`, `created_at` |
| Interfaces Go | singular, PascalCase | `StudentService` |
| Structs Go | singular, PascalCase | `studentService` |
| Méthodes Go | verbe + nom, PascalCase | `CreateStudent` |
| Fichiers Go | singular, snake_case | `student.go` |
| Routes API | pluriel, kebab-case | `/v1/bonus-types` |
| JSON fields | snake_case | `first_name`, `created_at` |

### Architecture

```
Request → Router → Handler → Service → Repository → Database
                      ↓
                     DTO
```

### Validation

- Utiliser les tags `validate` sur les DTOs
- Valider avec `validator.ValidateStruct(req)`
- Retourner des erreurs typées (`api.ErrValidationFailed`)

### Erreurs

- Service : Retourner `api.ErrXxxNotFound` ou wrap avec `fmt.Errorf`
- Handler : Utiliser `web.WriteFromError(w, err)`
- Toujours logger les erreurs importantes avec `slog`

---

## 🔗 Liens utiles

### Documentation externe
- [Go Documentation](https://go.dev/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [SQLC](https://docs.sqlc.dev/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [go-playground/validator](https://github.com/go-playground/validator)

### Repos similaires
- Pas de repos similaires référencés pour l'instant

---

## 💡 Conseils

### Pour bien commencer
1. Lisez **tous** les fichiers de documentation dans l'ordre recommandé
2. Explorez le code existant (Student, Classroom, BonusType)
3. Testez l'API avec les exemples cURL
4. Implémentez une petite fonctionnalité pour vous familiariser

### Pour implémenter une feature
1. Lisez le guide d'implémentation en entier
2. Regardez les exemples de référence
3. Suivez la checklist étape par étape
4. Testez au fur et à mesure
5. Documentez votre feature dans api-endpoints.md

### Pour debugger
1. Vérifiez les logs (slog)
2. Testez les requêtes SQL directement dans psql
3. Utilisez les erreurs typées pour identifier le problème
4. Vérifiez l'isolation multi-tenant (user_id)

---

## 📞 Support

Pour toute question sur la documentation ou le projet :
- Ouvrir une issue sur GitHub
- Consulter les exemples de code existants
- Relire les sections pertinentes de cette documentation

---

## 📄 Licence

Ce projet est privé. Tous droits réservés.

---

**Dernière mise à jour :** Février 2024
