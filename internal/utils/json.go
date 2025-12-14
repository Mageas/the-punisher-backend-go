package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mageas/the-punisher-backend/internal/apierr"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message error, details []apierr.ErrorDetail) {
	WriteJSON(
		w,
		status,
		&apierr.Error{Error: message.Error(), ErrorCode: status, ErrorDetails: details},
		nil,
	)
}

func WriteServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteError(w, http.StatusInternalServerError, apierr.ErrInternalError, nil)
}

func WriteConflictError(w http.ResponseWriter, field string, errorKey string) {
	details := []apierr.ErrorDetail{
		{Field: field, Error: errorKey},
	}
	WriteError(w, http.StatusConflict, apierr.ErrConflict, details)
}

func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	if after, ok := strings.CutPrefix(err.Error(), "json: unknown field"); ok {
		fieldName := strings.Trim(after, " \"")

		WriteError(w, http.StatusBadRequest, apierr.ErrInvalidRequestBody, []apierr.ErrorDetail{
			{Field: fieldName, Error: apierr.KeyValidationUnknownField},
		})
		return
	}

	WriteError(w, http.StatusBadRequest, apierr.ErrInvalidRequestBody, []apierr.ErrorDetail{
		{Field: "", Error: err.Error()},
	})
}

func WriteValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := []apierr.ErrorDetail{}
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				details = append(details, apierr.ErrorDetail{Field: e.Field(), Error: apierr.KeyValidationFieldRequired})
			case "email":
				details = append(details, apierr.ErrorDetail{Field: e.Field(), Error: apierr.KeyValidationInvalidEmail})
			case "min":
				details = append(details, apierr.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(apierr.KeyValidationMinLength, e.Param())})
			case "max":
				details = append(details, apierr.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(apierr.KeyValidationMaxLength, e.Param())})
			default:
				details = append(details, apierr.ErrorDetail{Field: e.Field(), Error: fmt.Sprintf(apierr.KeyValidationError, e.Tag())})
			}
		}

		WriteError(w, http.StatusBadRequest, apierr.ErrValidationFailed, details)
		return
	}

	WriteError(w, http.StatusBadRequest, err, nil)
}
