package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

func TestPenaltyTypeHandlerCRUDSuccess(t *testing.T) {
	shared.RunTypeHandlerCRUDSuccess(t, penaltyTypeSuite())
}

func TestPenaltyTypeHandlerDTOValidations(t *testing.T) {
	shared.RunTypeHandlerDTOValidations(t, penaltyTypeSuite())
}

func TestPenaltyTypeHandlerDecodeAndIDErrors(t *testing.T) {
	shared.RunTypeHandlerDecodeAndIDErrors(t, penaltyTypeSuite())
}

func TestPenaltyTypeHandlerNotFoundAndInternalErrors(t *testing.T) {
	shared.RunTypeHandlerNotFoundAndInternalErrors(t, penaltyTypeSuite())
}

func penaltyTypeSuite() shared.ManagedTypeSuite {
	return shared.ManagedTypeSuite{
		BasePath:      "/v1/penalty-types",
		NotFoundError: api.ErrPenaltyTypeNotFound.Error(),
		OpCreate:      inmemory.OpCreatePenaltyType,
		OpList:        inmemory.OpListPenaltyTypesByUser,
		Seed: func(repo *inmemory.Repository, seed shared.TypeSeed) {
			repo.SeedPenaltyType(repository.PenaltyType{
				ID:     seed.ID,
				UserID: seed.UserID,
				Name:   seed.Name,
			})
		},
		NewRouter: func(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
			svc := service.NewPenaltyTypeService(repo)
			h := handler.NewPenaltyTypeHandler(svc)

			r := chi.NewRouter()
			r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
			r.Route("/v1/penalty-types", func(r chi.Router) {
				r.Post("/", h.CreatePenaltyType)
				r.Get("/", h.ListPenaltyTypes)
				r.Get("/{penalty_type_id}", h.GetPenaltyType)
				r.Put("/{penalty_type_id}", h.UpdatePenaltyType)
				r.Delete("/{penalty_type_id}", h.DeletePenaltyType)
			})

			return r
		},
	}
}
