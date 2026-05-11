package domain

import "errors"

// ErrInvalidTimeWindow is returned when start_time is strictly after end_time.
var ErrInvalidTimeWindow = errors.New("invalid time window: start_time must be before or equal to end_time")
