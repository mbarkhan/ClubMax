package domain

import "errors"

var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource already exists")
	ErrForbidden  = errors.New("access denied")
	ErrInvalid    = errors.New("invalid input")
	ErrInvalidOTP = errors.New("invalid or expired verification code")
	ErrGymLimit   = errors.New("coach gym limit reached")
)
