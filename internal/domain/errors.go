package domain

import "errors"

var (
	// ErrInvalidTimeWindow indicates start is after end for a telemetry query.
	ErrInvalidTimeWindow = errors.New("invalid time window: start_time must be before or equal to end_time")
)
