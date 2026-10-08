package usecase

import "errors"

var ErrForbidden = errors.New("permission denied")
var ErrUnauthorized = errors.New("unauthorized")
