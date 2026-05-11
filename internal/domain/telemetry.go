// Package domain contains core entities and domain-level errors.
// It has no dependencies on infrastructure or frameworks.
package domain

import "time"

// Telemetry represents a single GPU telemetry reading (domain entity).
type Telemetry struct {
	ID                  int64     // Database primary key.
	MetricName          string    // e.g. "gpu_utilization", "gpu_memory_used".
	GPUID               string    // Unique GPU identifier (used for routing/querying).
	Device              string    // Device path, e.g. "cuda:0".
	UUID                string    // Hardware UUID of the GPU.
	ModelName           string    // GPU model, e.g. "A100", "H100".
	HostName            string    // Host machine name.
	Value               float64   // Metric value (unitless; interpretation depends on MetricName).
	LabelsRaw           string    // Arbitrary JSON labels attached by the collector.
	ProcessedAtUnixNano int64     // Collector-side timestamp in nanoseconds since epoch (used for time filtering).
	CreatedAt           time.Time // Database insertion timestamp.
}
