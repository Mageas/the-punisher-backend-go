package web

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
