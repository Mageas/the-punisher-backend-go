package validator

import (
	"testing"
)

type validatorSample struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2"`
}

func TestValidateStructSuccess(t *testing.T) {
	in := validatorSample{Email: "test@example.com", Name: "John"}
	if err := ValidateStruct(in); err != nil {
		t.Fatalf("ValidateStruct returned error: %v", err)
	}
}

func TestValidateStructFailure(t *testing.T) {
	in := validatorSample{Email: "not-an-email", Name: ""}
	if err := ValidateStruct(in); err == nil {
		t.Fatalf("expected validation error")
	}
}
