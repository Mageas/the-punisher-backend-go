package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mageas/the-punisher-backend/internal/api"
)

// ParseEnumQueryParam returns the normalized value (lowercase), whether it was present, and an error on invalid value.
func ParseEnumQueryParam(r *http.Request, name string, allowed []string) (string, bool, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return "", false, nil
	}

	normalized := strings.ToLower(raw)
	for _, value := range allowed {
		if normalized == strings.ToLower(value) {
			return normalized, true, nil
		}
	}

	return "", true, fmt.Errorf("invalid %s", name)
}

func EnumExpected(values []string) string {
	if len(values) == 0 {
		return ""
	}

	lower := make([]string, 0, len(values))
	for _, value := range values {
		lower = append(lower, strings.ToLower(value))
	}

	return strings.Join(lower, "_or_")
}

// ParseEnumQueryParamToBool parses an enum query param and maps it to a bool pointer.
// It returns nil if the param is absent, and an error with details if invalid.
func ParseEnumQueryParamToBool(r *http.Request, name string, trueValue string, falseValue string) (*bool, []api.ErrorDetail, error) {
	allowed := []string{trueValue, falseValue}
	value, hasValue, err := ParseEnumQueryParam(r, name, allowed)
	if err != nil {
		return nil, []api.ErrorDetail{
			{Field: name, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, EnumExpected(allowed))},
		}, err
	}

	if !hasValue {
		return nil, nil, nil
	}

	isTrue := strings.EqualFold(value, trueValue)
	return &isTrue, nil, nil
}
