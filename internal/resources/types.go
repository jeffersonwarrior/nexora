package resources

import (
	"fmt"
	"time"
)

// ResourceSnapshot represents resource usage at a point in time.
type ResourceSnapshot struct {
	Timestamp time.Time

	// CPU
	CPUPercent float64 // CPU usage percentage (0-100)

	// Memory
	MemUsed    uint64  // Memory used in bytes
	MemTotal   uint64  // Total memory in bytes
	MemPercent float64 // Memory usage percentage (0-100)

	// Disk
	DiskUsed    uint64  // Disk used in bytes
	DiskTotal   uint64  // Total disk in bytes
	DiskFree    uint64  // Free disk in bytes
	DiskPercent float64 // Disk usage percentage (0-100)
}

// ViolationType represents the type of resource violation.
type ViolationType string

const (
	ViolationTypeCPU  ViolationType = "cpu"
	ViolationTypeMem  ViolationType = "memory"
	ViolationTypeDisk ViolationType = "disk"
)

// Violation represents a resource threshold violation.
type Violation struct {
	Type      ViolationType
	Timestamp time.Time
	Value     float64 // Actual value
	Threshold float64 // Configured threshold
	Message   string
}

// String returns a string representation of the violation.
func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s: %.2f > %.2f at %s",
		v.Type, v.Message, v.Value, v.Threshold, v.Timestamp.Format(time.RFC3339))
}

// Summary returns formatted resource usage summary.
func (s ResourceSnapshot) Summary() string {
	return fmt.Sprintf("CPU: %.1f%% | Memory: %.1f%% (%.2fGB/%.2fGB) | Disk: %.1f%% (%.2fGB free)",
		s.CPUPercent,
		s.MemPercent,
		float64(s.MemUsed)/(1024*1024*1024),
		float64(s.MemTotal)/(1024*1024*1024),
		s.DiskPercent,
		float64(s.DiskFree)/(1024*1024*1024),
	)
}
