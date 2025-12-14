package domain_errors

import "errors"

type DomainError interface {
	error
	GetMessage() string
	GetDetailKey() string
	GetDetailMessage() string

	WithKey(key string) DomainError
	Is(target error) bool
}

type domainError struct {
	message       string
	detailKey     string
	detailMessage string
}

func (e *domainError) WithKey(key string) DomainError {
	newErr := *e
	newErr.detailKey = key
	return &newErr
}

func (e *domainError) Is(target error) bool {
	t, ok := target.(*domainError)
	if !ok {
		return false
	}
	return e.message == t.message
}

func (e *domainError) Error() string {
	return e.message
}
func (e *domainError) GetMessage() string {
	return e.message
}

func (e *domainError) GetDetailKey() string {
	return e.detailKey
}

func (e *domainError) GetDetailMessage() string {
	return e.detailMessage
}

func NewDomainError(msg, key, detailMsg string) DomainError {
	return &domainError{
		message:       msg,
		detailKey:     key,
		detailMessage: detailMsg,
	}
}

var (
	ErrUserNotFound       = errors.New("user_not_found")
	ErrEmailAlreadyExists = NewDomainError("email_already_exists", "email", "This email is already registered")
	ErrInvalidPassword    = NewDomainError("invalid_password", "password", "Password is too weak")
)
