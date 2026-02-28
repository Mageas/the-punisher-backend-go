package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type decodePayload struct {
	Name string `json:"name"`
}

func TestDecodeJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john"}`))
	var data decodePayload

	if err := DecodeJSON(r, &data); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}
	if data.Name != "john" {
		t.Fatalf("unexpected decoded value: %+v", data)
	}
}

func TestDecodeJSONUnknownField(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john","extra":1}`))
	var data decodePayload

	if err := DecodeJSON(r, &data); err == nil {
		t.Fatalf("expected unknown field error")
	}
}
