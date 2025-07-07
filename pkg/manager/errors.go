package manager

import "errors"

var (
	ErrMethoNotAllowed    = errors.New("method not allowed")
	ErrInvalidEventObject = errors.New("invalid event object")
)
