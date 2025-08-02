package logger

import (
	"encoding/json"
	"io"
	"strings"
	"time"
)

// Entry represents a structured log entry.
type Entry struct {
	Timestamp     time.Time `json:"timestamp"`
	Level         string    `json:"level"`
	CorrelationID string    `json:"correlation_id"`
	TenantID      string    `json:"tenant_id,omitempty"`
	Subject       string    `json:"subject,omitempty"`
	Action        string    `json:"action,omitempty"`
	Resource      string    `json:"resource,omitempty"`
	Decision      string    `json:"decision,omitempty"`
	PolicyID      string    `json:"policy_id,omitempty"`
	Reason        string    `json:"reason,omitempty"`
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// ParseLevel converts string to Level.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger writes structured logs to the provided writer.
type Logger struct {
	w        io.Writer
	minLevel Level
}

// New creates a new Logger.
func New(w io.Writer, min Level) *Logger {
	return &Logger{w: w, minLevel: min}
}

// Log writes the entry as JSON to the writer.
func (l *Logger) Log(e Entry) {
	if ParseLevel(e.Level) < l.minLevel {
		return
	}
	e.Timestamp = time.Now().UTC()
	enc := json.NewEncoder(l.w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(e)
}
