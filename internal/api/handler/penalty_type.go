package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PenaltyTypeHandler struct {
	managed managedTypeHandler[dto.RequestPenaltyTypeDto, dto.UpdatePenaltyTypeDto, dto.ReturnPenaltyTypeDto]
}

func NewPenaltyTypeHandler(service service.PenaltyTypeService) *PenaltyTypeHandler {
	return &PenaltyTypeHandler{
		managed: managedTypeHandler[dto.RequestPenaltyTypeDto, dto.UpdatePenaltyTypeDto, dto.ReturnPenaltyTypeDto]{
			idPathParam: "penalty_type_id",
			create:      service.CreatePenaltyType,
			list:        service.ListPenaltyTypes,
			get:         service.GetPenaltyType,
			update:      service.UpdatePenaltyType,
			delete:      service.DeletePenaltyType,
		},
	}
}

func (h *PenaltyTypeHandler) CreatePenaltyType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleCreate(w, r)
}

func (h *PenaltyTypeHandler) ListPenaltyTypes(w http.ResponseWriter, r *http.Request) {
	h.managed.handleList(w, r)
}

func (h *PenaltyTypeHandler) GetPenaltyType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleGet(w, r)
}

func (h *PenaltyTypeHandler) UpdatePenaltyType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleUpdate(w, r)
}

func (h *PenaltyTypeHandler) DeletePenaltyType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleDelete(w, r)
}
