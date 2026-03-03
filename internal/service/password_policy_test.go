package service

import "testing"

func TestIsPasswordComplexEnough(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "valid", password: "StrongPassword123!", want: true},
		{name: "too short", password: "Aa1!short", want: false},
		{name: "missing uppercase", password: "strongpassword123!", want: false},
		{name: "missing lowercase", password: "STRONGPASSWORD123!", want: false},
		{name: "missing digit", password: "StrongPassword!!!", want: false},
		{name: "missing special", password: "StrongPassword123", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPasswordComplexEnough(tt.password)
			if got != tt.want {
				t.Fatalf("isPasswordComplexEnough(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}
