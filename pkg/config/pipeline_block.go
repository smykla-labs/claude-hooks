// Package config provides configuration schema types for klaudiush validators.
package config

// PipelineBlockConfig configures the pipeline block marker system.
type PipelineBlockConfig struct {
	// Enabled controls whether pipeline block marking is active.
	// When enabled, blocking validation errors create a marker file that
	// causes subsequent hook invocations to fail immediately.
	// Default: false
	Enabled *bool `json:"enabled,omitempty" koanf:"enabled" toml:"enabled"`

	// TTLSeconds is the time-to-live for block markers in seconds.
	// Markers expire after this duration, allowing normal validation to resume.
	// This prevents indefinite blocking after a pipeline completes.
	// Default: 5
	TTLSeconds *int `json:"ttl_seconds,omitempty" koanf:"ttl_seconds" toml:"ttl_seconds"`

	// MarkerFile is the path to the marker file (relative to working directory).
	// The marker file stores pipeline block state as JSON with expiration timestamp.
	// Default: ".klaudiush-pipeline-block"
	MarkerFile string `json:"marker_file,omitempty" koanf:"marker_file" toml:"marker_file"`
}

// IsEnabled returns whether pipeline block is enabled.
func (p *PipelineBlockConfig) IsEnabled() bool {
	if p == nil || p.Enabled == nil {
		return false
	}

	return *p.Enabled
}

// GetTTLSeconds returns the TTL with default fallback.
func (p *PipelineBlockConfig) GetTTLSeconds() int {
	if p == nil || p.TTLSeconds == nil {
		return 5 // Default: 5 seconds
	}

	return *p.TTLSeconds
}

// GetMarkerFile returns the marker file path with default fallback.
func (p *PipelineBlockConfig) GetMarkerFile() string {
	if p == nil || p.MarkerFile == "" {
		return ".klaudiush-pipeline-block"
	}

	return p.MarkerFile
}
