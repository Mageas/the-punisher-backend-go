package api

import "errors"

type Error struct {
	Error        string        `json:"error"`
	ErrorDetails []ErrorDetail `json:"error_details,omitempty"`
	ErrorCode    int           `json:"error_code"`
}

type ErrorDetail struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

var (
	ErrInternalError      = errors.New("internal_error")
	ErrInvalidRequestBody = errors.New("invalid_request_body")

	ErrValidationFailed = errors.New("validation_failed")
	ErrConflict         = errors.New("conflict")

	ErrUserNotFound                        = errors.New("user_not_found")
	ErrEmailAlreadyExists                  = errors.New("email_already_exists")
	ErrInvalidCredentialsOrUserDoesntExist = errors.New("invalid_credentials_or_user_doesnt_exist")
)

const (
	KeyValidationError              = "validation_error:%s"
	KeyValidationFieldRequired      = "validation_field_required"
	KeyValidationInvalidEmail       = "validation_invalid_email"
	KeyValidationMinLength          = "validation_min_length:%s"
	KeyValidationMaxLength          = "validation_max_length:%s"
	KeyValidationEmailAlreadyExists = "validation_email_already_exists"

	KeyValidationUnknownField = "validation_unknown_field"
)
