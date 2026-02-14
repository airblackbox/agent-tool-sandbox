// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package audit records sandbox execution events as structured JSON
// suitable for OTel log export and SIEM ingestion.
package audit

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/nostalgicskinco/agent-tool-sandbox/pkg/sandbox"
)

// EventType classifies sandbox events.
type EventType string

const (
	EventExecutionStart    EventType = "sandbox.execution.start"
	EventExecutionComplete EventType = "sandbox.execution.complete"
	EventPolicyViolation   EventType = "sandbox.policy.violation"
	EventTimeout           EventType = "sandbox.execution.timeout"
)

// Event is a single sandbox audit entry.
type Event struct {
	Timestamp  time.Time       `json:"timestamp"`
	EventType  EventType       `json:"event_type"`
	ToolName   string          `json:"tool_name"`
	Command    string          `json:"command"`
	Profile    string          `json:"profile,omitempty"`
	ExitCode   int             `json:"exit_code,omitempty"`
	DurationMs float64         `json:"duration_ms,omitempty"`
	Violations []string        `json:"violations,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// Logger writes sandbox audit events.
type Logger struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewLogger creates a logger writing to w.
func NewLogger(w io.Writer) *Logger {
	return &Logger{enc: json.NewEncoder(w)}
}

// LogStart records a sandbox execution starting.
func (l *Logger) LogStart(toolName, command, profile string) {
	l.log(Event{
		EventType: EventExecutionStart,
		ToolName:  toolName,
		Command:   command,
		Profile:   profile,
	})
}

// LogResult records a completed sandbox execution.
func (l *Logger) LogResult(toolName, command, profile string, result *sandbox.Result) {
	evt := Event{
		EventType:  EventExecutionComplete,
		ToolName:   toolName,
		Command:    command,
		Profile:    profile,
		ExitCode:   result.ExitCode,
		DurationMs: float64(result.Duration.Milliseconds()),
	}
	if result.Killed {
		evt.EventType = EventTimeout
	}
	if len(result.Violations) > 0 {
		evt.EventType = EventPolicyViolation
		evt.Violations = result.Violations
	}
	if result.Error != "" {
		evt.Error = result.Error
	}
	l.log(evt)
}

func (l *Logger) log(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enc.Encode(e)
}
