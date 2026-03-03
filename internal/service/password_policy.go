package service

import (
	"unicode"

	"github.com/mageas/the-punisher-backend/internal/api"
)

func validatePasswordComplexity(password string, field string) error {
	if !isPasswordComplexEnough(password) {
		return api.NewAPIError(
			api.ErrPasswordPolicyViolation.StatusCode,
			api.ErrPasswordPolicyViolation.Message,
			api.ErrorDetail{
				Field: field,
				Error: api.KeyValidationPasswordPolicy,
			},
		)
	}

	return nil
}

func isPasswordComplexEnough(password string) bool {
	if len(password) < 12 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}
