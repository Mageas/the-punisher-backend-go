package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mageas/the-punisher-backend/internal/api"
	platformvalidator "github.com/mageas/the-punisher-backend/internal/platform/validator"
)

func decodeErrorResponse(t *testing.T, rr *httptest.ResponseRecorder) api.ErrorResponse {
	t.Helper()

	var body api.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	h := http.Header{}
	h.Set("X-Test", "1")

	err := WriteJSON(rr, http.StatusCreated, map[string]string{"ok": "true"}, h)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content-type: %s", got)
	}
	if got := rr.Header().Get("X-Test"); got != "1" {
		t.Fatalf("missing copied header")
	}
}

func TestWriteJSONDecodeErrorUnknownField(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSONDecodeError(rr, errors.New(`json: unknown field "foo"`))

	body := decodeErrorResponse(t, rr)
	if body.Error != api.ErrInvalidRequestBody.Message {
		t.Fatalf("unexpected error: %s", body.Error)
	}
	if len(body.ErrorDetails) != 1 || body.ErrorDetails[0].Field != "foo" {
		t.Fatalf("unexpected details: %+v", body.ErrorDetails)
	}
}

func TestWriteJSONDecodeErrorType(t *testing.T) {
	type payload struct {
		Age int `json:"age"`
	}

	var p payload
	err := json.Unmarshal([]byte(`{"age":"bad"}`), &p)
	if err == nil {
		t.Fatalf("expected unmarshal error")
	}

	rr := httptest.NewRecorder()
	WriteJSONDecodeError(rr, err)

	body := decodeErrorResponse(t, rr)
	if body.Error != api.ErrInvalidRequestBody.Message {
		t.Fatalf("unexpected error: %s", body.Error)
	}
	if len(body.ErrorDetails) != 1 {
		t.Fatalf("expected one detail")
	}
	if body.ErrorDetails[0].Field != "age" {
		t.Fatalf("expected field age, got %+v", body.ErrorDetails)
	}
}

func TestWriteJSONDecodeErrorFallback(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSONDecodeError(rr, errors.New("decode failed"))

	body := decodeErrorResponse(t, rr)
	if len(body.ErrorDetails) != 1 {
		t.Fatalf("expected one detail")
	}
	if body.ErrorDetails[0].Field != "" || body.ErrorDetails[0].Error != "decode failed" {
		t.Fatalf("unexpected details: %+v", body.ErrorDetails)
	}
}

func TestWriteValidationError(t *testing.T) {
	type payload struct {
		RequiredField string `json:"required_field" validate:"required"`
		Email         string `json:"email" validate:"email"`
		MinField      string `json:"min_field" validate:"min=3"`
		MaxField      string `json:"max_field" validate:"max=2"`
		AlphaField    string `json:"alpha_field" validate:"alpha"`
	}

	err := platformvalidator.ValidateStruct(payload{
		RequiredField: "",
		Email:         "bad",
		MinField:      "ab",
		MaxField:      "abcd",
		AlphaField:    "123",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	rr := httptest.NewRecorder()
	WriteValidationError(rr, err)

	body := decodeErrorResponse(t, rr)
	if body.Error != api.ErrValidationFailed.Message {
		t.Fatalf("unexpected error: %s", body.Error)
	}
	if len(body.ErrorDetails) == 0 {
		t.Fatalf("expected validation details")
	}

	joined := ""
	for _, d := range body.ErrorDetails {
		joined += d.Field + ":" + d.Error + "|"
	}
	if !strings.Contains(joined, "required_field:validation_field_required") {
		t.Fatalf("expected required detail, got %s", joined)
	}
	if !strings.Contains(joined, "email:validation_invalid_email") {
		t.Fatalf("expected email detail, got %s", joined)
	}
	if !strings.Contains(joined, "min_field:validation_min_length:3") {
		t.Fatalf("expected min detail, got %s", joined)
	}
	if !strings.Contains(joined, "max_field:validation_max_length:2") {
		t.Fatalf("expected max detail, got %s", joined)
	}
	if !strings.Contains(joined, "alpha_field:validation_error:alpha") {
		t.Fatalf("expected default-tag detail, got %s", joined)
	}
}

func TestWriteValidationErrorNonValidationError(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteValidationError(rr, errors.New("bad request"))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := decodeErrorResponse(t, rr)
	if body.Error != "bad request" {
		t.Fatalf("unexpected body error: %s", body.Error)
	}
}

func TestWriteFromError(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteFromError(rr, api.ErrUnauthorized)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	WriteFromError(rr2, errors.New("boom"))
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr2.Code)
	}
}
