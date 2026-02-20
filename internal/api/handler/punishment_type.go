package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PunishmentTypeHandler struct {
	managed managedTypeHandler[dto.RequestPunishmentTypeDto, dto.UpdatePunishmentTypeDto, dto.ReturnPunishmentTypeDto]
}

func NewPunishmentTypeHandler(service service.PunishmentTypeService) *PunishmentTypeHandler {
	return &PunishmentTypeHandler{
		managed: managedTypeHandler[dto.RequestPunishmentTypeDto, dto.UpdatePunishmentTypeDto, dto.ReturnPunishmentTypeDto]{
			idPathParam: "punishment_type_id",
			create:      service.CreatePunishmentType,
			list:        service.ListPunishmentTypes,
			get:         service.GetPunishmentType,
			update:      service.UpdatePunishmentType,
			delete:      service.DeletePunishmentType,
		},
	}
}

func (h *PunishmentTypeHandler) CreatePunishmentType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleCreate(w, r)
}

func (h *PunishmentTypeHandler) ListPunishmentTypes(w http.ResponseWriter, r *http.Request) {
	h.managed.handleList(w, r)
}

func (h *PunishmentTypeHandler) GetPunishmentType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleGet(w, r)
}

func (h *PunishmentTypeHandler) UpdatePunishmentType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleUpdate(w, r)
}

func (h *PunishmentTypeHandler) DeletePunishmentType(w http.ResponseWriter, r *http.Request) {
	h.managed.handleDelete(w, r)
}
