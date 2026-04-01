package errs

import "errors"

// Sentinel errors used across service layers. Transport layers map these to
// appropriate HTTP status codes; the service layer wraps or returns them
// directly.
var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
	ErrConflict = errors.New("conflict")
)
