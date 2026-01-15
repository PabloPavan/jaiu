package ports

import "errors"

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrConflict = errors.New("conflict")
