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

func TestPunishmentTypeHandlerCRUDSuccess(t *testing.T) {
	shared.RunTypeHandlerCRUDSuccess(t, punishmentTypeSuite())
}

func TestPunishmentTypeHandlerDTOValidations(t *testing.T) {
	shared.RunTypeHandlerDTOValidations(t, punishmentTypeSuite())
}

func TestPunishmentTypeHandlerDecodeAndIDErrors(t *testing.T) {
	shared.RunTypeHandlerDecodeAndIDErrors(t, punishmentTypeSuite())
}

func TestPunishmentTypeHandlerNotFoundAndInternalErrors(t *testing.T) {
	shared.RunTypeHandlerNotFoundAndInternalErrors(t, punishmentTypeSuite())
}

func punishmentTypeSuite() shared.ManagedTypeSuite {
	return shared.ManagedTypeSuite{
		BasePath:      "/v1/punishment-types",
		NotFoundError: api.ErrPunishmentTypeNotFound.Error(),
		OpCreate:      inmemory.OpCreatePunishmentType,
		OpList:        inmemory.OpListPunishmentTypesByUser,
		Seed: func(repo *inmemory.Repository, seed shared.TypeSeed) {
			repo.SeedPunishmentType(repository.PunishmentType{
				ID:     seed.ID,
				UserID: seed.UserID,
				Name:   seed.Name,
			})
		},
		NewRouter: func(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
			svc := service.NewPunishmentTypeService(repo)
			h := handler.NewPunishmentTypeHandler(svc)

			r := chi.NewRouter()
			r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
			r.Route("/v1/punishment-types", func(r chi.Router) {
				r.Post("/", h.CreatePunishmentType)
				r.Get("/", h.ListPunishmentTypes)
				r.Get("/{id}", h.GetPunishmentType)
				r.Put("/{id}", h.UpdatePunishmentType)
				r.Delete("/{id}", h.DeletePunishmentType)
			})

			return r
		},
	}
}
