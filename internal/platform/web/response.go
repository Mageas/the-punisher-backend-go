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
		&api.Error{Error: message.Error(), ErrorCode: status, ErrorDetails: details},
		nil,
	)
}

func WriteServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteError(w, http.StatusInternalServerError, api.ErrInternalError, nil)
}

func WriteConflictError(w http.ResponseWriter, field string, errorKey string) {
	details := []api.ErrorDetail{
		{Field: field, Error: errorKey},
	}
	WriteError(w, http.StatusConflict, api.ErrConflict, details)
}

func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	if after, ok := strings.CutPrefix(err.Error(), "json: unknown field"); ok {
		fieldName := strings.Trim(after, " \"")

		WriteError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, []api.ErrorDetail{
			{Field: fieldName, Error: api.KeyValidationUnknownField},
		})
		return
	}

	WriteError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, []api.ErrorDetail{
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
			default:
				details = append(details, api.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(api.KeyValidationError, e.Tag())})
			}
		}

		WriteError(w, http.StatusBadRequest, api.ErrValidationFailed, details)
		return
	}

	WriteError(w, http.StatusBadRequest, err, nil)
}
