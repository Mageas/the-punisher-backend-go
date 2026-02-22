package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

func parseBodyRFC3339(w http.ResponseWriter, rawValue, field string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(rawValue))
	if err != nil {
		web.WriteAPIError(w, api.ErrInvalidRequestBody, []api.ErrorDetail{
			{Field: field, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "rfc3339_datetime")},
		})
		return time.Time{}, false
	}

	return parsed, true
}
