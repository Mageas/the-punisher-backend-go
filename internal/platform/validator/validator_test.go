package validator

import (
	"testing"
)

type TestStruct struct {
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"min=18"`
	Email string `json:"email" validate:"email"`
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name      string
		input     TestStruct
		expectErr bool
	}{
		{
			name: "valid_struct",
			input: TestStruct{
				Name:  "Jean Dupont",
				Age:   25,
				Email: "jean.dupont@example.com",
			},
			expectErr: false,
		},
		{
			name: "missing_required_field",
			input: TestStruct{
				Name:  "",
				Age:   25,
				Email: "jean.dupont@example.com",
			},
			expectErr: true,
		},
		{
			name: "invalid_min_value",
			input: TestStruct{
				Name:  "Jean Dupont",
				Age:   17,
				Email: "jean.dupont@example.com",
			},
			expectErr: true,
		},
		{
			name: "invalid_email_format",
			input: TestStruct{
				Name:  "Jean Dupont",
				Age:   25,
				Email: "not-an-email",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStruct(tc.input)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}
