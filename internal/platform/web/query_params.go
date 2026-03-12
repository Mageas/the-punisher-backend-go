package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

// ParseOptionalUUIDQueryParam parses an optional UUID query param.
func ParseOptionalUUIDQueryParam(r *http.Request, name string) (*uuid.UUID, []api.ErrorDetail, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil, nil
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil, []api.ErrorDetail{
			{Field: name, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "uuid")},
		}, err
	}

	return &parsed, nil, nil
}

// ParseOptionalBoolQueryParam parses an optional bool query param with true/false values.
func ParseOptionalBoolQueryParam(r *http.Request, name string) (*bool, []api.ErrorDetail, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil, nil
	}

	switch strings.ToLower(raw) {
	case "true":
		v := true
		return &v, nil, nil
	case "false":
		v := false
		return &v, nil, nil
	default:
		return nil, []api.ErrorDetail{
			{Field: name, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "true_or_false")},
		}, fmt.Errorf("invalid %s", name)
	}
}

// ParseOptionalDateQueryParam parses an optional YYYY-MM-DD query param into a date-only time value.
func ParseOptionalDateQueryParam(r *http.Request, name string) (*time.Time, []api.ErrorDetail, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil, nil
	}

	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, []api.ErrorDetail{
			{Field: name, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "yyyy-mm-dd")},
		}, err
	}

	parsed = parsed.UTC()
	return &parsed, nil, nil
}

// ValidateDateRange ensures from <= to when both bounds are provided.
func ValidateDateRange(from, to *time.Time, fromField, toField string) ([]api.ErrorDetail, error) {
	if from == nil || to == nil {
		return nil, nil
	}

	if from.After(*to) {
		return []api.ErrorDetail{
			{
				Field: fromField,
				Error: fmt.Sprintf(api.KeyValidationMalformedParameter, fmt.Sprintf("%s_lte_%s", fromField, toField)),
			},
		}, fmt.Errorf("invalid range %s > %s", fromField, toField)
	}

	return nil, nil
}

// ParseSearchQueryParam returns a normalized search value (single-spaced, trimmed) or nil if empty.
func ParseSearchQueryParam(r *http.Request, name string) *string {
	raw := r.URL.Query().Get(name)
	normalized := strings.Join(strings.Fields(raw), " ")
	if normalized == "" {
		return nil
	}

	return &normalized
}
