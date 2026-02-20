package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type BonusTypeHandler struct {
	managed managedTypeHandler[dto.RequestBonusTypeDto, dto.UpdateBonusTypeDto, dto.ReturnBonusTypeDto]
}

func NewBonusTypeHandler(service service.BonusTypeService) *BonusTypeHandler {
	return &BonusTypeHandler{
		managed: managedTypeHandler[dto.RequestBonusTypeDto, dto.UpdateBonusTypeDto, dto.ReturnBonusTypeDto]{
			idPathParam: "bonus_type_id",
			create:      service.CreateBonusType,
			list:        service.ListBonusTypes,
			get:         service.GetBonusType,
			update:      service.UpdateBonusType,
			delete:      service.DeleteBonusType,
		},
	}
}

func (h *BonusTypeHandler) CreateBonusType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleCreate(w, r)
}

func (h *BonusTypeHandler) ListBonusTypes(w http.ResponseWriter, r *http.Request) {
	h.managed.handleList(w, r)
}

func (h *BonusTypeHandler) GetBonusType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleGet(w, r)
}

func (h *BonusTypeHandler) UpdateBonusType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleUpdate(w, r)
}

func (h *BonusTypeHandler) DeleteBonusType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleDelete(w, r)
}
