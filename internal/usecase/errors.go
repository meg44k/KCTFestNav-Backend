package usecase

import "errors"

var ErrForbidden = errors.New("permission denied")
var ErrUnauthorized = errors.New("unauthorized")
var ErrTooMany = errors.New("too many requests")
