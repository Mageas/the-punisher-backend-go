package web

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON(r *http.Request, data any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1_048_576) // 1MB max

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
