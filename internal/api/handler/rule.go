package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type RuleHandler struct {
	service service.RuleService
}

func NewRuleHandler(service service.RuleService) *RuleHandler {
	return &RuleHandler{
		service: service,
	}
}

func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestRuleDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	rule, err := h.service.CreateRule(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, rule, nil)
}

func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	rules, totalCount, err := h.service.ListRules(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(rules, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *RuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	ruleID, ok := parsePathUUID(w, r, "rule_id")
	if !ok {
		return
	}

	rule, err := h.service.GetRule(r.Context(), userID, ruleID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, rule, nil)
}

func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	ruleID, ok := parsePathUUID(w, r, "rule_id")
	if !ok {
		return
	}

	var req dto.UpdateRuleDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	rule, err := h.service.UpdateRule(r.Context(), userID, ruleID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, rule, nil)
}

func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	ruleID, ok := parsePathUUID(w, r, "rule_id")
	if !ok {
		return
	}

	if err := h.service.DeleteRule(r.Context(), userID, ruleID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
