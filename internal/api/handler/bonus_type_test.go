package handler_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/dto"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

func TestBonusTypeHandlerCRUDSuccess(t *testing.T) {
	shared.RunTypeHandlerCRUDSuccess[dto.ReturnBonusTypeDto](t, bonusTypeSuite())
}

func TestBonusTypeHandlerDTOValidations(t *testing.T) {
	shared.RunTypeHandlerDTOValidations(t, bonusTypeSuite())
}

func TestBonusTypeHandlerDecodeAndIDErrors(t *testing.T) {
	shared.RunTypeHandlerDecodeAndIDErrors(t, bonusTypeSuite())
}

func TestBonusTypeHandlerNotFoundAndInternalErrors(t *testing.T) {
	shared.RunTypeHandlerNotFoundAndInternalErrors(t, bonusTypeSuite())
}

func bonusTypeSuite() shared.ManagedTypeSuite {
	return shared.ManagedTypeSuite{
		BasePath:      "/v1/bonus-types",
		NotFoundError: api.ErrBonusTypeNotFound.Error(),
		OpCreate:      inmemory.OpCreateBonusType,
		OpList:        inmemory.OpListBonusTypesByUser,
		Seed: func(repo *inmemory.Repository, seed shared.TypeSeed) {
			repo.SeedBonusType(repository.BonusType{
				ID:     seed.ID,
				UserID: seed.UserID,
				Name:   seed.Name,
			})
		},
		NewRouter: func(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
			svc := service.NewBonusTypeService(repo)
			h := handler.NewBonusTypeHandler(svc)

			r := chi.NewRouter()
			r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
			r.Route("/v1/bonus-types", func(r chi.Router) {
				r.Post("/", h.CreateBonusType)
				r.Get("/", h.ListBonusTypes)
				r.Get("/{bonus_type_id}", h.GetBonusType)
				r.Put("/{bonus_type_id}", h.UpdateBonusType)
				r.Delete("/{bonus_type_id}", h.DeleteBonusType)
			})

			return r
		},
	}
}
