// Package pipelineblock provides pipeline block marker functionality.
package pipelineblock

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cockroachdb/errors"
)

// Marker represents a pipeline block marker with expiration.
type Marker struct {
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	ErrorCode  string    `json:"error_code"`
	Validator  string    `json:"validator"`
	Message    string    `json:"message"`
	TTLSeconds int       `json:"ttl_seconds"`
}

// Manager handles pipeline block markers.
type Manager struct {
	markerPath string
	ttl        time.Duration
	enabled    bool
}

// NewManager creates a new pipeline block marker manager.
func NewManager(markerPath string, ttl time.Duration, enabled bool) *Manager {
	return &Manager{
		markerPath: markerPath,
		ttl:        ttl,
		enabled:    enabled,
	}
}

// SetBlockMarker creates a marker file indicating pipeline should be blocked.
func (m *Manager) SetBlockMarker(errorCode, validator, message string) error {
	if !m.enabled {
		return nil
	}

	now := time.Now()
	marker := Marker{
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.ttl),
		ErrorCode:  errorCode,
		Validator:  validator,
		Message:    message,
		TTLSeconds: int(m.ttl.Seconds()),
	}

	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal marker")
	}

	if err := os.WriteFile(m.markerPath, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write marker file")
	}

	return nil
}

// CheckBlockMarker checks if a valid block marker exists.
// Returns the marker and true if valid, nil and false otherwise.
func (m *Manager) CheckBlockMarker() (*Marker, bool) {
	if !m.enabled {
		return nil, false
	}

	// Check for manual clear via env var
	if os.Getenv("KLAUDIUSH_CLEAR_PIPELINE_BLOCK") != "" {
		_ = m.ClearBlockMarker() // Best effort
		return nil, false
	}

	data, err := os.ReadFile(m.markerPath)
	if err != nil {
		// No marker file exists (or can't read it)
		return nil, false
	}

	var marker Marker
	if err := json.Unmarshal(data, &marker); err != nil {
		// Invalid marker file, clear it
		_ = m.ClearBlockMarker() // Best effort
		return nil, false
	}

	// Check if marker has expired
	if time.Now().After(marker.ExpiresAt) {
		_ = m.ClearBlockMarker() // Best effort
		return nil, false
	}

	return &marker, true
}

// ClearBlockMarker removes the marker file.
func (m *Manager) ClearBlockMarker() error {
	if !m.enabled {
		return nil
	}

	err := os.Remove(m.markerPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove marker file")
	}

	return nil
}

// IsEnabled returns whether pipeline block is enabled.
func (m *Manager) IsEnabled() bool {
	return m.enabled
}
