package domain

import "errors"

// Error郡
var ErrNameRequired = errors.New("name is required")
var ErrEndTimeAfterStartTime = errors.New("end time must be after start time")
var ErrContentRequired = errors.New("content is required")
var ErrInvalidDirection = errors.New("direction must be up or down")
var ErrInvalidFloor = errors.New("floor must be 0 or greater")
var ErrInvalidCongestion = errors.New("congestion level must be between 0 and 3")
var ErrNotClassBooth = errors.New("likes are only for class booths")
var ErrInvalidRange = errors.New("from must be before to")
var ErrLoginIDTaken = errors.New("login id is already taken")
var ErrLoginIDRequired = errors.New("login id is required")
var ErrCannotDeleteSelf = errors.New("you cannot delete your own account")
