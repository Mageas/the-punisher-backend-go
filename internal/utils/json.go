package utils

import (
	"encoding/json"
	"log/slog"
	"maps"
	"net/http"
)

// WriteJSON sends a JSON response with the given status code and data.
func WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// WriteError sends a JSON error response with the given status code and message.
func WriteError(w http.ResponseWriter, status int, message string) {
	type envelopes struct {
		ErrorMessage string `json:"error_message"`
		ErrorCode    int    `json:"error_code"`
	}

	WriteJSON(w, status, &envelopes{ErrorMessage: message, ErrorCode: status}, nil)
}

// ServerError logs the error and sends a 500 Internal Server Error response.
func ServerError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	WriteError(w, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

// BadRequest sends a 400 Bad Request response with the error message.
func BadRequest(w http.ResponseWriter, err error) {
	WriteError(w, http.StatusBadRequest, err.Error())
}
