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

	ErrUnauthorized       = errors.New("unauthorized")
	ErrRegisterNotAllowed = errors.New("register_not_allowed")

	ErrMalformedParameter = errors.New("malformed_parameter")
	ErrValidationFailed   = errors.New("validation_failed")
	ErrConflict           = errors.New("conflict")

	ErrEmailAlreadyExists                  = errors.New("email_already_exists")
	ErrInvalidCredentialsOrUserDoesntExist = errors.New("invalid_credentials_or_user_doesnt_exist")

	ErrJWTInvalidSigningMethod = errors.New("jwt_invalid_signing_method")
	ErrJWTInvalidToken         = errors.New("jwt_invalid_token")
	ErrJWTExpired              = errors.New("jwt_expired")

	ErrStudentNotFound = errors.New("student_not_found")
)

const (
	KeyValidationError              = "validation_error:%s"
	KeyValidationFieldRequired      = "validation_field_required"
	KeyValidationInvalidEmail       = "validation_invalid_email"
	KeyValidationMinLength          = "validation_min_length:%s"
	KeyValidationMaxLength          = "validation_max_length:%s"
	KeyValidationEmailAlreadyExists = "validation_email_already_exists"

	KeyValidationUnknownField       = "validation_unknown_field"
	KeyValidationMalformedParameter = "validation_malformed_parameter:expected_%s"
)
