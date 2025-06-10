package model

import "errors"

// ErrInvalidPageSize is returned when a page size parameter is not a positive number.
// This error indicates that the requested page size is invalid for pagination operations.
var ErrInvalidPageSize = errors.New("invalid page size")
