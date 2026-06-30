package domain

import "errors"

// Sentinel domain errors. Services return these; the httputil error mapper
// translates them into HTTP status codes and the API error envelope.
var (
	// ErrNotFound indicates a requested resource does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict indicates a uniqueness or state conflict (e.g. duplicate email).
	ErrConflict = errors.New("resource conflict")
	// ErrForbidden indicates the acting user lacks permission for the action.
	ErrForbidden = errors.New("forbidden")
	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrValidation indicates a business-rule validation failure.
	ErrValidation = errors.New("validation failed")
	// ErrInvalidCredentials indicates a failed login attempt.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
