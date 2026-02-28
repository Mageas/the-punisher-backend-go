package api

import "testing"

func TestNewAPIError(t *testing.T) {
	details := []ErrorDetail{{Field: "email", Error: "invalid"}}
	err := NewAPIError(418, "teapot", details...)

	if err.StatusCode != 418 {
		t.Fatalf("unexpected status code: %d", err.StatusCode)
	}
	if err.Message != "teapot" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
	if len(err.Details) != 1 || err.Details[0].Field != "email" {
		t.Fatalf("unexpected details: %+v", err.Details)
	}
	if err.Error() != "teapot" {
		t.Fatalf("unexpected Error() string: %s", err.Error())
	}
}
