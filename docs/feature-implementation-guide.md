# Guide d'Implémentation de Nouvelles Fonctionnalités

Ce document fournit un guide détaillé pour implémenter de nouvelles fonctionnalités dans The Punisher Backend. Il est particulièrement conçu pour faciliter l'ajout de features par une IA.

## Table des matières

- [Vue d'ensemble du processus](#vue-densemble-du-processus)
- [Exemples de référence](#exemples-de-référence)
- [Étapes détaillées](#étapes-détaillées)
- [Checklist complète](#checklist-complète)
- [Patterns et conventions](#patterns-et-conventions)
- [Exemples complets](#exemples-complets)

---

## Vue d'ensemble du processus

Pour implémenter une nouvelle entité/ressource (ex: Penalties, Punishments, etc.), suivre ces étapes **dans cet ordre** :

1. **Migration de base de données** - Créer la table
2. **Requêtes SQL** - Définir les opérations CRUD dans `db/sqlc/queries.sql`
3. **Génération du code** - Exécuter `sqlc generate`
4. **DTO** - Créer les structures de données d'entrée/sortie
5. **Service** - Implémenter la logique métier
6. **Handler** - Créer les contrôleurs HTTP
7. **Routes** - Enregistrer les routes dans `cmd/api/routes.go`
8. **Erreurs** - Ajouter les erreurs spécifiques si nécessaire

---

## Exemples de référence

**Entités simples (CRUD complet) :**
- `Student` : `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/student.go`
- `BonusType` : `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/bonus_type.go`

**Entités avec relations :**
- `Classroom` : `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/classroom.go`
  - Relations many-to-many avec Students
  - Opérations : AddStudentToClassroom, RemoveStudentFromClassroom
  - Listes : ListStudentsByClassroom, ListClassroomsByStudent

**Authentification :**
- `Auth` : `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/auth.go`

---

## Étapes détaillées

### 1. Migration de base de données

**Commande :**
```bash
make migrate-create <nom_descriptif>
# Exemple: make migrate-create create_penalties_table
```

**Fichier up (création) :**
```sql
-- db/migrations/NNNNNN_create_penalties_table.up.sql
CREATE TABLE IF NOT EXISTS penalties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    penalty_type_id UUID NOT NULL REFERENCES penalty_types(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Index pour optimiser les requêtes fréquentes
    CONSTRAINT fk_penalties_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_penalties_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_penalties_penalty_type FOREIGN KEY (penalty_type_id) REFERENCES penalty_types(id) ON DELETE CASCADE
);

CREATE INDEX idx_penalties_user_id ON penalties(user_id);
CREATE INDEX idx_penalties_student_id ON penalties(student_id);
CREATE INDEX idx_penalties_created_at ON penalties(created_at DESC);
```

**Fichier down (rollback) :**
```sql
-- db/migrations/NNNNNN_create_penalties_table.down.sql
DROP TABLE IF EXISTS penalties;
```

**Points clés :**
- Toujours inclure `user_id` pour l'isolation multi-tenant
- Utiliser `UUID` pour les IDs (avec `gen_random_uuid()`)
- Ajouter `created_at` et `updated_at` si pertinent
- Définir `ON DELETE CASCADE` pour les clés étrangères
- Créer des index sur les colonnes fréquemment interrogées

**Appliquer la migration :**
```bash
make migrate-up
```

---

### 2. Requêtes SQL (SQLC)

Ajouter les requêtes dans `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/db/sqlc/queries.sql`

**Pattern standard pour un CRUD complet :**

```sql
-- ==================== Penalty ====================

-- name: CreatePenalty :one
INSERT INTO penalties (
    user_id, student_id, penalty_type_id
) VALUES (
    sqlc.arg(user_id), sqlc.arg(student_id), sqlc.arg(penalty_type_id)
)
RETURNING id, user_id, student_id, penalty_type_id, created_at;

-- name: GetPenalty :one
SELECT id, user_id, student_id, penalty_type_id, created_at
FROM penalties
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: CountPenaltiesByUser :one
SELECT COUNT(*) FROM penalties WHERE user_id = sqlc.arg(user_id);

-- name: ListPenaltiesByUser :many
SELECT id, user_id, student_id, penalty_type_id, created_at
FROM penalties
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: ListPenaltiesByStudent :many
SELECT p.id, p.user_id, p.student_id, p.penalty_type_id, p.created_at
FROM penalties p
JOIN students s ON s.id = p.student_id
WHERE p.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id)
ORDER BY p.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountPenaltiesByStudent :one
SELECT COUNT(*)
FROM penalties p
JOIN students s ON s.id = p.student_id
WHERE p.student_id = sqlc.arg(student_id) AND s.user_id = sqlc.arg(user_id);

-- name: UpdatePenalty :one
UPDATE penalties
SET
    penalty_type_id = COALESCE(sqlc.narg(penalty_type_id), penalty_type_id)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, student_id, penalty_type_id, created_at;

-- name: DeletePenalty :execrows
DELETE FROM penalties
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);
```

**Points clés :**
- Utiliser `sqlc.arg(nom)` pour les paramètres requis
- Utiliser `sqlc.narg(nom)` pour les paramètres optionnels (nullable)
- Toujours filtrer par `user_id` pour l'isolation
- Utiliser `:execrows` pour DELETE/UPDATE afin de vérifier les lignes affectées
- Utiliser `:one` pour retourner un seul résultat
- Utiliser `:many` pour retourner une liste
- Inclure `query_limit` et `query_offset` pour la pagination
- Trier par `created_at DESC` pour avoir les plus récents en premier

**Générer le code Go :**
```bash
make sqlc
# ou: sqlc generate
```

---

### 3. DTOs (Data Transfer Objects)

Créer le fichier `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/dto/penalty.go`

```go
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

// DTO pour la création
type RequestPenaltyDto struct {
	StudentID     string `json:"student_id" validate:"required,uuid"`
	PenaltyTypeID string `json:"penalty_type_id" validate:"required,uuid"`
}

// DTO pour la mise à jour (tous les champs optionnels)
type UpdatePenaltyDto struct {
	PenaltyTypeID *string `json:"penalty_type_id" validate:"omitempty,uuid"`
}

// DTO de retour
type ReturnPenaltyDto struct {
	ID            uuid.UUID `json:"id"`
	StudentID     uuid.UUID `json:"student_id"`
	PenaltyTypeID uuid.UUID `json:"penalty_type_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// Fonction de mapping depuis le repository
func PenaltyFromRepository(p *repository.Penalty) *ReturnPenaltyDto {
	if p == nil {
		return nil
	}

	return &ReturnPenaltyDto{
		ID:            p.ID,
		StudentID:     p.StudentID,
		PenaltyTypeID: p.PenaltyTypeID,
		CreatedAt:     p.CreatedAt,
	}
}

// Fonction de mapping pour les listes
func PenaltyListFromRepository(penalties []repository.Penalty) []*ReturnPenaltyDto {
	dtos := make([]*ReturnPenaltyDto, 0, len(penalties))

	for _, p := range penalties {
		if dto := PenaltyFromRepository(&p); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
```

**Règles de validation courantes :**
- `required` : Champ obligatoire
- `uuid` : Format UUID valide
- `email` : Format email valide
- `min=N` : Longueur minimale
- `max=N` : Longueur maximale
- `omitempty` : Optionnel, validé uniquement si fourni

**Points clés :**
- Ne jamais exposer `user_id` dans les DTOs (géré automatiquement via le contexte)
- Utiliser des pointeurs (`*string`) pour les champs optionnels dans les Update DTOs
- Préfixer les noms : `Request`, `Update`, `Return`
- Fonction de mapping : `[Entity]FromRepository`
- Fonction de mapping liste : `[Entity]ListFromRepository`

---

### 4. Service (Logique métier)

Créer le fichier `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/penalty.go`

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

// Interface du service (pour faciliter le mocking)
type PenaltyService interface {
	CreatePenalty(ctx context.Context, userID uuid.UUID, req dto.RequestPenaltyDto) (*dto.ReturnPenaltyDto, error)
	GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error)
	ListPenalties(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnPenaltyDto, int64, error)
	ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnPenaltyDto, int64, error)
	UpdatePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID, req dto.UpdatePenaltyDto) (*dto.ReturnPenaltyDto, error)
	DeletePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) error
}

// Implémentation privée
type penaltyService struct {
	repo repository.Querier
}

// Constructeur (retourne l'interface)
func NewPenaltyService(repo repository.Querier) PenaltyService {
	return &penaltyService{repo: repo}
}

// Création
func (s *penaltyService) CreatePenalty(ctx context.Context, userID uuid.UUID, req dto.RequestPenaltyDto) (*dto.ReturnPenaltyDto, error) {
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, api.ErrMalformedParameter
	}

	penaltyTypeID, err := uuid.Parse(req.PenaltyTypeID)
	if err != nil {
		return nil, api.ErrMalformedParameter
	}

	// Vérifier que l'étudiant appartient à l'utilisateur
	_, err = s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to verify student: %w", err)
	}

	// Vérifier que le penalty type appartient à l'utilisateur
	_, err = s.repo.GetPenaltyType(ctx, repository.GetPenaltyTypeParams{
		ID:     penaltyTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to verify penalty type: %w", err)
	}

	penalty, err := s.repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        userID,
		StudentID:     studentID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create penalty: %w", err)
	}

	slog.Info("penalty created", "penalty_id", penalty.ID, "student_id", studentID, "user_id", userID)

	return dto.PenaltyFromRepository(&penalty), nil
}

// Récupération
func (s *penaltyService) GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error) {
	penalty, err := s.repo.GetPenalty(ctx, repository.GetPenaltyParams{
		ID:     penaltyID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to get penalty: %w", err)
	}

	return dto.PenaltyFromRepository(&penalty), nil
}

// Liste paginée
func (s *penaltyService) ListPenalties(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnPenaltyDto, int64, error) {
	totalCount, err := s.repo.CountPenaltiesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalties: %w", err)
	}

	penalties, err := s.repo.ListPenaltiesByUser(ctx, repository.ListPenaltiesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalties: %w", err)
	}

	return dto.PenaltyListFromRepository(penalties), totalCount, nil
}

// Liste par étudiant
func (s *penaltyService) ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnPenaltyDto, int64, error) {
	// Vérifier que l'étudiant existe et appartient à l'utilisateur
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to verify student: %w", err)
	}

	totalCount, err := s.repo.CountPenaltiesByStudent(ctx, repository.CountPenaltiesByStudentParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalties by student: %w", err)
	}

	penalties, err := s.repo.ListPenaltiesByStudent(ctx, repository.ListPenaltiesByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalties by student: %w", err)
	}

	return dto.PenaltyListFromRepository(penalties), totalCount, nil
}

// Mise à jour
func (s *penaltyService) UpdatePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID, req dto.UpdatePenaltyDto) (*dto.ReturnPenaltyDto, error) {
	params := repository.UpdatePenaltyParams{
		ID:     penaltyID,
		UserID: userID,
	}

	if req.PenaltyTypeID != nil {
		penaltyTypeID, err := uuid.Parse(*req.PenaltyTypeID)
		if err != nil {
			return nil, api.ErrMalformedParameter
		}

		// Vérifier que le penalty type appartient à l'utilisateur
		_, err = s.repo.GetPenaltyType(ctx, repository.GetPenaltyTypeParams{
			ID:     penaltyTypeID,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, api.ErrPenaltyTypeNotFound
			}
			return nil, fmt.Errorf("failed to verify penalty type: %w", err)
		}

		params.PenaltyTypeID = pgtype.UUID{Bytes: penaltyTypeID, Valid: true}
	}

	penalty, err := s.repo.UpdatePenalty(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to update penalty: %w", err)
	}

	return dto.PenaltyFromRepository(&penalty), nil
}

// Suppression
func (s *penaltyService) DeletePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) error {
	rowsAffected, err := s.repo.DeletePenalty(ctx, repository.DeletePenaltyParams{
		ID:     penaltyID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete penalty: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrPenaltyNotFound
	}

	slog.Info("penalty deleted", "penalty_id", penaltyID, "user_id", userID)

	return nil
}
```

**Points clés :**
- Interface publique exportée (PenaltyService)
- Implémentation privée non-exportée (penaltyService)
- Toujours vérifier l'appartenance des ressources liées à l'utilisateur
- Utiliser `errors.Is(err, pgx.ErrNoRows)` pour détecter les ressources non trouvées
- Retourner des erreurs typées (`api.ErrXxxNotFound`)
- Logger les opérations de mutation avec `slog.Info`
- Gérer les rowsAffected pour les DELETE/UPDATE
- Wrap les erreurs internes avec `fmt.Errorf("...: %w", err)`

---

### 5. Handler (Contrôleurs HTTP)

Créer le fichier `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/api/handler/penalty.go`

```go
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PenaltyHandler struct {
	service service.PenaltyService
}

func NewPenaltyHandler(service service.PenaltyService) *PenaltyHandler {
	return &PenaltyHandler{service: service}
}

// POST /v1/penalties
func (h *PenaltyHandler) CreatePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestPenaltyDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	penalty, err := h.service.CreatePenalty(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, penalty, nil)
}

// GET /v1/penalties
func (h *PenaltyHandler) ListPenalties(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	penalties, totalCount, err := h.service.ListPenalties(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

// GET /v1/penalties/{id}
func (h *PenaltyHandler) GetPenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	penalty, err := h.service.GetPenalty(r.Context(), userID, penaltyID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, penalty, nil)
}

// PUT /v1/penalties/{id}
func (h *PenaltyHandler) UpdatePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	var req dto.UpdatePenaltyDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	penalty, err := h.service.UpdatePenalty(r.Context(), userID, penaltyID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, penalty, nil)
}

// DELETE /v1/penalties/{id}
func (h *PenaltyHandler) DeletePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	if err := h.service.DeletePenalty(r.Context(), userID, penaltyID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /v1/students/{id}/penalties
func (h *PenaltyHandler) ListPenaltiesByStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	limit, offset, page := web.ParsePagination(r)

	penalties, totalCount, err := h.service.ListPenaltiesByStudent(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}
```

**Points clés :**
- Toujours récupérer `userID` du contexte avec `auth.MustUserIDFromContext`
- Parser les UUID avec `uuid.Parse(chi.URLParam(r, "id"))`
- Toujours décoder et valider les inputs
- Utiliser `web.WriteJSON` pour les succès
- Utiliser `web.WriteFromError` pour les erreurs
- Status codes :
  - `201 Created` : Création réussie
  - `200 OK` : Succès (GET, PUT)
  - `204 No Content` : Suppression réussie
  - `400 Bad Request` : Données invalides
  - `404 Not Found` : Ressource non trouvée

---

### 6. Routes

Modifier `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/cmd/api/routes.go`

```go
// Dans la fonction mount()

penaltyService := service.NewPenaltyService(repo)
penaltyHandler := handler.NewPenaltyHandler(penaltyService)

r.Route("/v1/penalties", func(r chi.Router) {
    r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
    r.Post("/", penaltyHandler.CreatePenalty)
    r.Get("/", penaltyHandler.ListPenalties)
    r.Get("/{id}", penaltyHandler.GetPenalty)
    r.Put("/{id}", penaltyHandler.UpdatePenalty)
    r.Delete("/{id}", penaltyHandler.DeletePenalty)
})

// Ajouter aussi dans la section students si nécessaire
r.Route("/v1/students", func(r chi.Router) {
    r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
    // ... routes existantes ...
    r.Get("/{id}/penalties", penaltyHandler.ListPenaltiesByStudent)
})
```

**Points clés :**
- Utiliser `r.Route` pour grouper les routes avec un préfixe commun
- Toujours ajouter le middleware d'authentification
- Respecter la convention RESTful : pluriel en kebab-case
- Pattern :
  - `POST /v1/entities` : Création
  - `GET /v1/entities` : Liste paginée
  - `GET /v1/entities/{id}` : Récupération
  - `PUT /v1/entities/{id}` : Mise à jour
  - `DELETE /v1/entities/{id}` : Suppression

---

### 7. Erreurs

Ajouter dans `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/api/errors.go`

```go
var (
	// ... erreurs existantes ...
	
	ErrPenaltyNotFound     = NewAPIError(http.StatusNotFound, "penalty_not_found")
	ErrPenaltyTypeNotFound = NewAPIError(http.StatusNotFound, "penalty_type_not_found")
)
```

**Erreurs courantes à définir :**
- `Err[Entity]NotFound` : 404 pour ressource non trouvée
- `Err[Entity]AlreadyExists` : 409 pour conflit
- `ErrInvalid[Something]` : 400 pour données invalides

---

## Checklist complète

### Phase 1 : Base de données
- [ ] Créer la migration (up et down)
- [ ] Ajouter les index appropriés
- [ ] Définir les contraintes de clés étrangères avec ON DELETE CASCADE
- [ ] Appliquer la migration (`make migrate-up`)

### Phase 2 : Repository (SQLC)
- [ ] Ajouter les requêtes SQL dans `db/sqlc/queries.sql`
  - [ ] Create (`:one`)
  - [ ] Get (`:one`)
  - [ ] Count (`:one`)
  - [ ] List (`:many`)
  - [ ] Update (`:one`) si applicable
  - [ ] Delete (`:execrows`)
  - [ ] Requêtes de relation si applicable
- [ ] Générer le code (`make sqlc`)
- [ ] Vérifier que les types générés sont corrects

### Phase 3 : DTOs
- [ ] Créer `internal/dto/[entity].go`
- [ ] Définir `Request[Entity]Dto` avec les validations
- [ ] Définir `Update[Entity]Dto` avec les champs optionnels
- [ ] Définir `Return[Entity]Dto`
- [ ] Implémenter `[Entity]FromRepository`
- [ ] Implémenter `[Entity]ListFromRepository`

### Phase 4 : Service
- [ ] Créer `internal/service/[entity].go`
- [ ] Définir l'interface du service
- [ ] Implémenter la structure privée
- [ ] Implémenter le constructeur
- [ ] Implémenter Create
- [ ] Implémenter Get
- [ ] Implémenter List
- [ ] Implémenter Update (si applicable)
- [ ] Implémenter Delete
- [ ] Ajouter les logs appropriés (`slog.Info`)
- [ ] Vérifier l'appartenance des ressources liées

### Phase 5 : Handler
- [ ] Créer `internal/api/handler/[entity].go`
- [ ] Définir la structure du handler
- [ ] Implémenter le constructeur
- [ ] Implémenter Create
- [ ] Implémenter List
- [ ] Implémenter Get
- [ ] Implémenter Update (si applicable)
- [ ] Implémenter Delete

### Phase 6 : Routes
- [ ] Ajouter les routes dans `cmd/api/routes.go`
- [ ] Initialiser le service
- [ ] Initialiser le handler
- [ ] Définir les routes avec authentification

### Phase 7 : Erreurs
- [ ] Ajouter les erreurs spécifiques dans `internal/api/errors.go`

### Phase 8 : Documentation
- [ ] Ajouter l'entité dans `docs/api-endpoints.md`
- [ ] Documenter tous les endpoints
- [ ] Ajouter des exemples de requêtes/réponses
- [ ] Documenter les erreurs possibles

---

## Patterns et conventions

### Nommage

**Base de données :**
- Tables : pluriel, snake_case (`students`, `bonus_types`)
- Colonnes : singular, snake_case (`first_name`, `created_at`)

**Go :**
- Interfaces : singular, PascalCase (`StudentService`)
- Structs : singular, PascalCase (`studentService`)
- Méthodes : verbe + nom, PascalCase (`CreateStudent`)
- Fichiers : singular, snake_case (`student.go`)

**API :**
- Routes : pluriel, kebab-case (`/v1/bonus-types`)
- JSON fields : snake_case (`first_name`, `created_at`)

### Gestion des erreurs

**Dans le service :**
```go
// Erreur de ressource non trouvée
if errors.Is(err, pgx.ErrNoRows) {
    return nil, api.ErrStudentNotFound
}

// Erreur interne wrappée
return nil, fmt.Errorf("failed to create student: %w", err)
```

**Dans le handler :**
```go
// Toutes les erreurs sont gérées par web.WriteFromError
if err != nil {
    web.WriteFromError(w, err)
    return
}
```

### Logging

```go
// Log de création
slog.Info("student created", "student_id", student.ID, "user_id", userID)

// Log de suppression
slog.Info("student deleted", "student_id", studentID, "user_id", userID)

// Log avec relation
slog.Info("student added to classroom", "student_id", studentID, "classroom_id", classroomID, "user_id", userID)
```

### Validation

**Tags de validation :**
- `required` : Obligatoire
- `uuid` : UUID valide
- `email` : Email valide
- `min=N` : Longueur min
- `max=N` : Longueur max
- `omitempty` : Optionnel

**Exemple :**
```go
type RequestStudentDto struct {
    FirstName string `json:"first_name" validate:"required,min=2,max=70"`
    LastName  string `json:"last_name" validate:"required,min=2,max=70"`
}
```

### Pagination

**Toujours retourner :**
```go
response := web.NewPaginatedResponse(items, totalCount, page)
web.WriteJSON(w, http.StatusOK, response, nil)
```

**Format de réponse :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 45,
  "previous_page": null,
  "next_page": 2,
  "data": [...]
}
```

---

## Exemples complets

### Exemple 1 : Entité simple sans relations

Voir l'implémentation complète de `BonusType` :
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/bonus_type.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/api/handler/bonus_type.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/dto/bonus_type.go`

### Exemple 2 : Entité avec relations simples

Voir l'implémentation complète de `Student` :
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/student.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/api/handler/student.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/dto/student.go`

### Exemple 3 : Entité avec relations many-to-many

Voir l'implémentation complète de `Classroom` :
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/service/classroom.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/api/handler/classroom.go`
- `/home/runner/work/the-punisher-backend-go/the-punisher-backend-go/internal/dto/classroom.go`

**Points clés des relations many-to-many :**
- Table de jointure : `student_classrooms`
- Méthodes spécifiques : `AddStudentToClassroom`, `RemoveStudentFromClassroom`
- Listes croisées : `ListStudentsByClassroom`, `ListClassroomsByStudent`

---

## Commandes utiles

```bash
# Créer une migration
make migrate-create create_penalties_table

# Appliquer les migrations
make migrate-up

# Revenir en arrière (rollback)
make migrate-down

# Générer le code SQLC
make sqlc

# Lancer le serveur en mode développement
make dev

# Compiler le binaire
make build
```

---

## Résumé des fichiers à créer/modifier

Pour une nouvelle entité complète (ex: Penalty) :

1. `db/migrations/NNNNNN_create_penalties_table.up.sql`
2. `db/migrations/NNNNNN_create_penalties_table.down.sql`
3. `db/sqlc/queries.sql` (ajouter les requêtes)
4. `internal/dto/penalty.go`
5. `internal/service/penalty.go`
6. `internal/api/handler/penalty.go`
7. `cmd/api/routes.go` (modifier)
8. `internal/api/errors.go` (modifier si nouvelles erreurs)
9. `docs/api-endpoints.md` (documenter)

---

## Notes finales

- **Toujours** vérifier l'isolation multi-tenant (filtrer par `user_id`)
- **Toujours** valider les inputs
- **Toujours** gérer les erreurs de manière cohérente
- **Toujours** documenter les nouveaux endpoints
- Les exemples de Student, Classroom et BonusType sont vos références principales
- En cas de doute, se référer à l'implémentation existante
