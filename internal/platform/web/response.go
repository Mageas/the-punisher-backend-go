package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mageas/the-punisher-backend/internal/api"
)

func WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message error, details []api.ErrorDetail) {
	WriteJSON(
		w,
		status,
		&api.ErrorResponse{Error: message.Error(), ErrorCode: status, ErrorDetails: details},
		nil,
	)
}

func WriteImportError(w http.ResponseWriter, status int, message error, details []api.ImportErrorDetail) {
	WriteJSON(
		w,
		status,
		&api.ImportErrorResponse{Error: message.Error(), ErrorCode: status, ErrorDetails: details},
		nil,
	)
}

func WriteAPIError(w http.ResponseWriter, apiErr *api.APIError, details []api.ErrorDetail) {
	if details == nil {
		details = apiErr.Details
	}

	WriteError(w, apiErr.StatusCode, apiErr, details)
}

func WriteServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteAPIError(w, api.ErrInternalError, nil)
}

func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		WriteAPIError(w, api.ErrInvalidRequestBody, []api.ErrorDetail{
			{Field: unmarshalTypeErr.Field, Error: fmt.Sprintf(api.KeyValidationMalformedParameter, unmarshalTypeErr.Type.String())},
		})
		return
	}

	if after, ok := strings.CutPrefix(err.Error(), "json: unknown field"); ok {
		fieldName := strings.Trim(after, " \"")

		WriteAPIError(w, api.ErrInvalidRequestBody, []api.ErrorDetail{
			{Field: fieldName, Error: api.KeyValidationUnknownField},
		})
		return
	}

	WriteAPIError(w, api.ErrInvalidRequestBody, []api.ErrorDetail{
		{Field: "", Error: err.Error()},
	})
}

func WriteValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := []api.ErrorDetail{}
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: api.KeyValidationFieldRequired})
			case "email":
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: api.KeyValidationInvalidEmail})
			case "min":
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(api.KeyValidationMinLength, e.Param())})
			case "max":
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(api.KeyValidationMaxLength, e.Param())})
			case "oneof":
				details = append(details, api.ErrorDetail{
					Field: e.Field(),
					Error: fmt.Sprintf(api.KeyValidationOneOf, strings.Join(strings.Fields(e.Param()), "|")),
				})
			default:
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(api.KeyValidationError, e.Tag())})
			}
		}

		WriteAPIError(w, api.ErrValidationFailed, details)
		return
	}

	WriteError(w, http.StatusBadRequest, err, nil)
}

func WriteFromError(w http.ResponseWriter, err error) {
	var importErr *api.ImportValidationError
	if errors.As(err, &importErr) {
		WriteImportError(w, importErr.StatusCode, importErr, importErr.Details)
		return
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		WriteAPIError(w, apiErr, nil)
		return
	}

	WriteServerError(w, err)
}
