package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

type managedTypeHandler[TCreateReq any, TUpdateReq any, TReturn any] struct {
	idPathParam string

	create func(ctx context.Context, userID uuid.UUID, req TCreateReq) (*TReturn, error)
	list   func(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*TReturn, int64, error)
	get    func(ctx context.Context, userID, resourceID uuid.UUID) (*TReturn, error)
	update func(ctx context.Context, userID, resourceID uuid.UUID, req TUpdateReq) (*TReturn, error)
	delete func(ctx context.Context, userID, resourceID uuid.UUID) error
}

func (h managedTypeHandler[TCreateReq, TUpdateReq, TReturn]) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req TCreateReq
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.create(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, res, nil)
}

func (h managedTypeHandler[TCreateReq, TUpdateReq, TReturn]) handleList(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	items, totalCount, err := h.list(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(items, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h managedTypeHandler[TCreateReq, TUpdateReq, TReturn]) handleGet(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	resourceID, ok := parsePathUUID(w, r, h.idPathParam, h.idPathParam, "id")
	if !ok {
		return
	}

	res, err := h.get(r.Context(), userID, resourceID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h managedTypeHandler[TCreateReq, TUpdateReq, TReturn]) handleUpdate(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	resourceID, ok := parsePathUUID(w, r, h.idPathParam, h.idPathParam, "id")
	if !ok {
		return
	}

	var req TUpdateReq
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.update(r.Context(), userID, resourceID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h managedTypeHandler[TCreateReq, TUpdateReq, TReturn]) handleDelete(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	resourceID, ok := parsePathUUID(w, r, h.idPathParam, h.idPathParam, "id")
	if !ok {
		return
	}

	if err := h.delete(r.Context(), userID, resourceID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
