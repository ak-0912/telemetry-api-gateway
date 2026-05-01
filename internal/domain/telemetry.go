package domain

import "time"

// Telemetry is a telemetry reading for a GPU (entity).
type Telemetry struct {
	ID                  int64
	MetricName          string
	GPUID               string
	Device              string
	UUID                string
	ModelName           string
	HostName            string
	Value               float64
	LabelsRaw           string
	ProcessedAtUnixNano int64
	CreatedAt           time.Time
}
