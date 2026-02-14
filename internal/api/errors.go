package api

import "net/http"

type ErrorResponse struct {
	Error        string        `json:"error"`
	ErrorDetails []ErrorDetail `json:"error_details,omitempty"`
	ErrorCode    int           `json:"error_code"`
}

type ErrorDetail struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type APIError struct {
	Message    string
	StatusCode int
	Details    []ErrorDetail
}

func (e *APIError) Error() string { return e.Message }

func NewAPIError(statusCode int, message string, details ...ErrorDetail) *APIError {
	return &APIError{
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

var (
	ErrInternalError      = NewAPIError(http.StatusInternalServerError, "internal_error")
	ErrInvalidRequestBody = NewAPIError(http.StatusBadRequest, "invalid_request_body")

	ErrUnauthorized       = NewAPIError(http.StatusUnauthorized, "unauthorized")
	ErrRegisterNotAllowed = NewAPIError(http.StatusUnauthorized, "register_not_allowed")

	ErrMalformedParameter = NewAPIError(http.StatusBadRequest, "malformed_parameter")
	ErrValidationFailed   = NewAPIError(http.StatusBadRequest, "validation_failed")

	ErrEmailAlreadyExists                  = NewAPIError(http.StatusConflict, "conflict", ErrorDetail{Field: "email", Error: KeyValidationEmailAlreadyExists})
	ErrInvalidCredentialsOrUserDoesntExist = NewAPIError(http.StatusUnauthorized, "invalid_credentials_or_user_doesnt_exist")

	ErrJWTInvalidSigningMethod = NewAPIError(http.StatusUnauthorized, "jwt_invalid_signing_method")
	ErrJWTInvalidToken         = NewAPIError(http.StatusUnauthorized, "jwt_invalid_token")
	ErrJWTExpired              = NewAPIError(http.StatusUnauthorized, "jwt_expired")

	ErrStudentNotFound           = NewAPIError(http.StatusNotFound, "student_not_found")
	ErrBonusTypeNotFound         = NewAPIError(http.StatusNotFound, "bonus_type_not_found")
	ErrPenaltyTypeNotFound       = NewAPIError(http.StatusNotFound, "penalty_type_not_found")
	ErrBonusNotFound             = NewAPIError(http.StatusNotFound, "bonus_not_found")
	ErrPenaltyNotFound           = NewAPIError(http.StatusNotFound, "penalty_not_found")
	ErrPunishmentTypeNotFound    = NewAPIError(http.StatusNotFound, "punishment_type_not_found")
	ErrPunishmentNotFound        = NewAPIError(http.StatusNotFound, "punishment_not_found")
	ErrPunishmentAlreadyResolved = NewAPIError(http.StatusConflict, "punishment_already_resolved")
	ErrBonusAlreadyUsed          = NewAPIError(http.StatusConflict, "bonus_already_used")

	ErrClassroomNotFound              = NewAPIError(http.StatusNotFound, "classroom_not_found")
	ErrStudentClassroomRelationExists = NewAPIError(http.StatusConflict, "student_classroom_relation_exists")
	ErrStudentOrClassroomNotFound     = NewAPIError(http.StatusNotFound, "student_or_classroom_not_found")
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
