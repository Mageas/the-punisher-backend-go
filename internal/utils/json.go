package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidRequestBody = errors.New("invalid request body")
)

// DecodeJSON decodes the JSON request body into the given data structure.
func DecodeJSON(w http.ResponseWriter, r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// WriteJSON sends a JSON response with the given status code and data.
func WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// WriteError sends a JSON error response with the given status code and message.
func WriteError(w http.ResponseWriter, status int, message string, details ...string) {
	type envelopes struct {
		ErrorMessage string   `json:"error_message"`
		ErrorDetails []string `json:"error_details,omitempty"`
		ErrorCode    int      `json:"error_code"`
	}

	WriteJSON(w, status, &envelopes{ErrorMessage: message, ErrorCode: status, ErrorDetails: details}, nil)
}

// WriteServerError logs the error and sends a 500 Internal Server Error response.
func WriteServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteError(w, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

// WriteJSONDecodeError sends a 400 Bad Request response with the error message.
func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error(), err.Error())
}

// WriteValidationError sends a 400 Bad Request response with the error message.
func WriteValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errorMessages := make([]string, len(validationErrors))
		for i, e := range validationErrors {
			errorMessages[i] = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag())
		}

		WriteError(w, http.StatusBadRequest, ErrValidationFailed.Error(), errorMessages...)
		return
	}

	WriteError(w, http.StatusBadRequest, err.Error())
}

// WriteBadRequest sends a 400 Bad Request response with the error message.
func WriteBadRequest(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, err.Error())
}
