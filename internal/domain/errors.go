package domain

import "errors"

// Error郡
var ErrNameRequired = errors.New("name is required")
var ErrEndTimeAfterStartTime = errors.New("end time must be after start time")
var ErrContentRequired = errors.New("content is required")
var ErrInvalidDirection = errors.New("direction must be up or down")
