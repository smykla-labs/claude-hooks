package dispatcher

// PipelineBlockMarker interface for pipeline block marker management.
type PipelineBlockMarker interface {
	// SetBlockMarker creates a marker indicating pipeline should be blocked.
	SetBlockMarker(errorCode, validator, message string) error

	// CheckBlockMarker checks if a valid block marker exists.
	// Returns the marker and true if valid, nil and false otherwise.
	CheckBlockMarker() (marker interface{}, valid bool)

	// IsEnabled returns whether pipeline block is enabled.
	IsEnabled() bool
}
