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
	"github.com/mageas/the-punisher-backend/internal/domain_errors"
)

type Error struct {
	ErrorMessage string        `json:"error_message"`
	ErrorDetails []ErrorDetail `json:"error_details,omitempty"`
	ErrorCode    int           `json:"error_code"`
}

type ErrorDetail struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

var (
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidRequestBody = errors.New("invalid request body")
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

func WriteError(w http.ResponseWriter, status int, message error, details []ErrorDetail) {
	if domainErr, ok := message.(domain_errors.DomainError); ok {
		WriteJSON(
			w,
			status,
			&Error{ErrorMessage: domainErr.GetMessage(), ErrorCode: status, ErrorDetails: []ErrorDetail{
				{Key: domainErr.GetDetailKey(), Message: domainErr.GetDetailMessage()},
			}},
			nil,
		)
		return
	}

	WriteJSON(
		w,
		status,
		&Error{ErrorMessage: message.Error(), ErrorCode: status, ErrorDetails: details},
		nil,
	)
}

func WriteServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteError(w, http.StatusInternalServerError, errors.New("the server encountered a problem and could not process your request"), nil)
}

func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	if after, ok := strings.CutPrefix(err.Error(), "json: unknown field"); ok {
		fieldName := strings.Trim(after, " \"")

		WriteError(w, http.StatusBadRequest, ErrInvalidRequestBody, []ErrorDetail{
			{Key: fieldName, Message: "Unknown field"},
		})
		return
	}

	WriteError(w, http.StatusBadRequest, ErrInvalidRequestBody, []ErrorDetail{
		{Key: "", Message: err.Error()},
	})
}

func WriteValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := []ErrorDetail{}
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				details = append(details, ErrorDetail{Key: e.Field(), Message: "This field is required"})
			case "email":
				details = append(details, ErrorDetail{Key: e.Field(), Message: "Invalid email format"})
			case "min":
				details = append(details, ErrorDetail{Key: e.Field(), Message: fmt.Sprintf("Must be at least %s characters", e.Param())})
			default:
				details = append(details, ErrorDetail{Key: e.Field(), Message: fmt.Sprintf("Validation failed on '%s'", e.Tag())})
			}
		}

		WriteError(w, http.StatusBadRequest, ErrValidationFailed, details)
		return
	}

	WriteError(w, http.StatusBadRequest, err, nil)
}
