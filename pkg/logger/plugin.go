package logger

import "time"

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source,omitempty"`
	Line      int            `json:"line,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// LogFilter defines criteria for filtering log entries
type LogFilter struct {
	Levels     []string          `json:"levels,omitempty"`      // Filter by log levels (INFO, DEBUG, etc)
	Sources    []string          `json:"sources,omitempty"`     // Filter by source files
	Contains   []string          `json:"contains,omitempty"`    // Messages must contain these strings
	StartTime  *time.Time        `json:"start_time,omitempty"`  // Only entries after this time
	EndTime    *time.Time        `json:"end_time,omitempty"`    // Only entries before this time
	FieldMatch map[string]string `json:"field_match,omitempty"` // Match specific field values
}

// LogPlugin defines the interface for log handlers
type LogPlugin interface {
	// Handle processes a log entry
	Handle(entry LogEntry) error
	// ShouldHandle determines if this plugin should handle the entry
	ShouldHandle(entry LogEntry) bool
	// Initialize sets up the plugin
	Initialize() error
	// Close cleans up any resources
	Close() error
}
