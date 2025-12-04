// Package crashdump provides crash dump functionality for klaudiush.
package crashdump

import (
	"time"
)

// CrashInfo contains all diagnostic information about a crash.
type CrashInfo struct {
	// ID is the unique identifier for this crash dump.
	ID string `json:"id"`

	// Timestamp is when the crash occurred.
	Timestamp time.Time `json:"timestamp"`

	// PanicValue is the value passed to panic().
	PanicValue string `json:"panic_value"`

	// StackTrace is the stack trace of the panicking goroutine.
	StackTrace string `json:"stack_trace"`

	// Runtime contains runtime information at the time of crash.
	Runtime RuntimeInfo `json:"runtime"`

	// Context contains the hook context if available.
	Context *ContextInfo `json:"context,omitempty"`

	// Config contains a sanitized config snapshot.
	Config map[string]any `json:"config,omitempty"`

	// Metadata contains additional diagnostic information.
	Metadata DumpMetadata `json:"metadata"`
}

// RuntimeInfo contains Go runtime information at the time of crash.
type RuntimeInfo struct {
	// GOOS is the operating system (e.g., "darwin", "linux").
	GOOS string `json:"goos"`

	// GOARCH is the architecture (e.g., "amd64", "arm64").
	GOARCH string `json:"goarch"`

	// GoVersion is the Go version used to build the binary.
	GoVersion string `json:"go_version"`

	// NumGoroutine is the number of goroutines at crash time.
	NumGoroutine int `json:"num_goroutine"`

	// NumCPU is the number of CPUs available.
	NumCPU int `json:"num_cpu"`
}

// ContextInfo contains the hook context at the time of crash.
type ContextInfo struct {
	// EventType is the type of hook event (PreToolUse, PostToolUse, Notification).
	EventType string `json:"event_type"`

	// ToolName is the name of the tool being invoked.
	ToolName string `json:"tool_name"`

	// Command is the command being executed (for Bash tool).
	Command string `json:"command,omitempty"`

	// FilePath is the file path (for file operations).
	FilePath string `json:"file_path,omitempty"`
}

// DumpMetadata contains additional context about the crash dump.
type DumpMetadata struct {
	// Version is the klaudiush version.
	Version string `json:"version"`

	// User is the username who triggered the crash.
	User string `json:"user,omitempty"`

	// Hostname is the machine hostname.
	Hostname string `json:"hostname,omitempty"`

	// WorkingDir is the current working directory.
	WorkingDir string `json:"working_dir,omitempty"`
}

// DumpSummary provides a short summary for listing crash dumps.
type DumpSummary struct {
	// ID is the unique identifier for this crash dump.
	ID string `json:"id"`

	// Timestamp is when the crash occurred.
	Timestamp time.Time `json:"timestamp"`

	// PanicValue is a truncated version of the panic value.
	PanicValue string `json:"panic_value"`

	// FilePath is the path to the dump file.
	FilePath string `json:"file_path"`

	// Size is the size of the dump file in bytes.
	Size int64 `json:"size"`
}
