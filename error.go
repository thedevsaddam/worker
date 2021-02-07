package gworker

import "errors"

// available errors
var (
	ErrJobConflict      = errors.New(packageName + ": job already registered")
	ErrInvalidJob       = errors.New(packageName + ": job must be a func")
	ErrInvalidJobReturn = errors.New(packageName + ": job must return error type")
)
