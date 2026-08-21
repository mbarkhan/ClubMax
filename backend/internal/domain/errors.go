package domain

import "errors"

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string { return "invalid field: " + e.Field }
func (e *ValidationError) Unwrap() error { return ErrInvalid }

func InvalidField(field string) error { return &ValidationError{Field: field} }

var (
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource already exists")
	ErrForbidden          = errors.New("access denied")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrProtectedAdmin     = errors.New("default administrator cannot be deleted")
	ErrInvalid            = errors.New("invalid input")
	ErrInvalidOTP         = errors.New("invalid or expired verification code")
	ErrGymLimit           = errors.New("coach gym limit reached")
	ErrGymImageLimit      = errors.New("gym image limit reached")
	ErrGymFull            = errors.New("gym capacity reached")
	ErrInsufficientFunds  = errors.New("insufficient wallet funds")
	ErrWithdrawalLimit    = errors.New("withdrawal amount is outside policy limits")
	ErrInvalidState       = errors.New("invalid state transition")
	ErrMessageRecipient   = errors.New("message recipient is not allowed")
)
