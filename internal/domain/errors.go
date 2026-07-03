package domain

import "errors"

// Error郡
var ErrNameRequired = errors.New("name is required")
var ErrEndTimeAfterStartTime = errors.New("end time must be after start time")
var ErrSessionNumberLessThanOne = errors.New("session number must be at least 1")
var ErrInvalidLiveStatus = errors.New("invalid live status value")
var ErrContentRequired = errors.New("content is required")

