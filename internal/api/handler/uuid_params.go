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

func parsePathUUID(w http.ResponseWriter, r *http.Request, field string) (uuid.UUID, bool) {
	rawValue := strings.TrimSpace(chi.URLParam(r, field))
	parsed, err := uuid.Parse(rawValue)
	if err != nil {
		writeUUIDParseError(w, api.ErrNotFound, field)
		return uuid.Nil, false
	}

	return parsed, true
}

func parseBodyUUID(w http.ResponseWriter, rawValue string, field string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(strings.TrimSpace(rawValue))
	if err != nil {
		writeUUIDParseError(w, api.ErrInvalidRequestBody, field)
		return uuid.Nil, false
	}

	return parsed, true
}

func parseOptionalBodyUUID(w http.ResponseWriter, rawValue *string, field string) (*uuid.UUID, bool) {
	if rawValue == nil {
		return nil, true
	}

	parsed, ok := parseBodyUUID(w, *rawValue, field)
	if !ok {
		return nil, false
	}

	return &parsed, true
}

func parseBodyUUIDSlice(w http.ResponseWriter, rawValues []string, field string) ([]uuid.UUID, bool) {
	seen := make(map[uuid.UUID]struct{}, len(rawValues))
	values := make([]uuid.UUID, 0, len(rawValues))

	for _, rawValue := range rawValues {
		parsed, err := uuid.Parse(strings.TrimSpace(rawValue))
		if err != nil {
			writeUUIDParseError(w, api.ErrInvalidRequestBody, field)
			return nil, false
		}

		if _, exists := seen[parsed]; exists {
			continue
		}

		seen[parsed] = struct{}{}
		values = append(values, parsed)
	}

	return values, true
}

func writeUUIDParseError(w http.ResponseWriter, apiErr *api.APIError, field string) {
	web.WriteAPIError(w, apiErr, []api.ErrorDetail{
		{Field: field, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "uuid")},
	})
}
