package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

func parsePathUUID(w http.ResponseWriter, r *http.Request, field string, paramNames ...string) (uuid.UUID, bool) {
	if len(paramNames) == 0 {
		paramNames = []string{field}
	}

	for _, paramName := range paramNames {
		rawValue := strings.TrimSpace(chi.URLParam(r, paramName))
		if rawValue == "" {
			continue
		}

		parsed, err := uuid.Parse(rawValue)
		if err == nil {
			return parsed, true
		}
	}

	writeUUIDParseError(w, http.StatusBadRequest, api.ErrMalformedParameter, field)
	return uuid.Nil, false
}

func parseBodyUUID(w http.ResponseWriter, rawValue string, field string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(strings.TrimSpace(rawValue))
	if err != nil {
		writeUUIDParseError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, field)
		return uuid.Nil, false
	}

	return parsed, true
}

func parseOptionalQueryUUID(w http.ResponseWriter, r *http.Request, name string) (*uuid.UUID, bool) {
	rawValue := strings.TrimSpace(r.URL.Query().Get(name))
	if rawValue == "" {
		return nil, true
	}

	parsed, err := uuid.Parse(rawValue)
	if err != nil {
		writeUUIDParseError(w, http.StatusBadRequest, api.ErrMalformedParameter, name)
		return nil, false
	}

	return &parsed, true
}

func writeUUIDParseError(w http.ResponseWriter, status int, apiErr error, field string) {
	web.WriteError(w, status, apiErr, []api.ErrorDetail{
		{Field: field, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "uuid")},
	})
}
