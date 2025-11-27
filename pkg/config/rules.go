// Package config provides configuration schema types for klaudiush validators.
package config

// RulesConfig contains the dynamic rule configuration.
type RulesConfig struct {
	// Enabled controls whether the rule engine is active.
	// Default: true
	Enabled *bool `json:"enabled,omitempty" koanf:"enabled" toml:"enabled"`

	// StopOnFirstMatch controls whether to stop after the first matching rule.
	// Default: true
	StopOnFirstMatch *bool `json:"stop_on_first_match,omitempty" koanf:"stop_on_first_match" toml:"stop_on_first_match"`

	// Rules is the list of validation rules.
	Rules []RuleConfig `json:"rules,omitempty" koanf:"rules" toml:"rules"`
}

// RuleConfig represents a single validation rule configuration.
type RuleConfig struct {
	// Name uniquely identifies this rule. Used for override precedence.
	Name string `json:"name,omitempty" koanf:"name" toml:"name"`

	// Description provides human-readable explanation of the rule.
	Description string `json:"description,omitempty" koanf:"description" toml:"description"`

	// Enabled controls whether this rule is active.
	// Default: true
	Enabled *bool `json:"enabled,omitempty" koanf:"enabled" toml:"enabled"`

	// Priority determines evaluation order (higher = evaluated first).
	// Default: 0
	Priority int `json:"priority,omitempty" koanf:"priority" toml:"priority"`

	// Match contains the conditions that must be satisfied.
	Match *RuleMatchConfig `json:"match,omitempty" koanf:"match" toml:"match"`

	// Action specifies what happens when the rule matches.
	Action *RuleActionConfig `json:"action,omitempty" koanf:"action" toml:"action"`
}

// RuleMatchConfig contains all conditions for a rule to match.
// All non-empty conditions must be satisfied (AND logic).
type RuleMatchConfig struct {
	// ValidatorType filters by validator type (supports wildcards).
	// Examples: "git.push", "git.*", "*"
	ValidatorType string `json:"validator_type,omitempty" koanf:"validator_type" toml:"validator_type"`

	// RepoPattern matches against the repository root path.
	// Supports glob patterns (e.g., "**/myorg/**") and regex.
	RepoPattern string `json:"repo_pattern,omitempty" koanf:"repo_pattern" toml:"repo_pattern"`

	// Remote matches against git remote name (exact match).
	Remote string `json:"remote,omitempty" koanf:"remote" toml:"remote"`

	// BranchPattern matches against branch name.
	// Supports glob patterns (e.g., "feat/*") and regex.
	BranchPattern string `json:"branch_pattern,omitempty" koanf:"branch_pattern" toml:"branch_pattern"`

	// FilePattern matches against file path.
	// Supports glob patterns (e.g., "**/*.md") and regex.
	FilePattern string `json:"file_pattern,omitempty" koanf:"file_pattern" toml:"file_pattern"`

	// ContentPattern matches against file content.
	// Always treated as regex.
	ContentPattern string `json:"content_pattern,omitempty" koanf:"content_pattern" toml:"content_pattern"`

	// CommandPattern matches against bash command.
	// Supports glob patterns and regex.
	CommandPattern string `json:"command_pattern,omitempty" koanf:"command_pattern" toml:"command_pattern"`

	// ToolType matches against the hook tool type.
	// Examples: "Bash", "Write", "Edit"
	ToolType string `json:"tool_type,omitempty" koanf:"tool_type" toml:"tool_type"`

	// EventType matches against the hook event type.
	// Examples: "PreToolUse", "PostToolUse"
	EventType string `json:"event_type,omitempty" koanf:"event_type" toml:"event_type"`
}

// RuleActionConfig specifies what happens when a rule matches.
type RuleActionConfig struct {
	// Type is the action to take (block, warn, allow).
	// Default: "block"
	Type string `json:"type,omitempty" koanf:"type" toml:"type"`

	// Message is the human-readable message to display.
	Message string `json:"message,omitempty" koanf:"message" toml:"message"`

	// Reference is an optional error reference code (e.g., "GIT019").
	Reference string `json:"reference,omitempty" koanf:"reference" toml:"reference"`
}

// IsEnabled returns true if the rules engine is enabled.
// Returns true if Enabled is nil (default behavior).
func (r *RulesConfig) IsEnabled() bool {
	if r == nil || r.Enabled == nil {
		return true
	}

	return *r.Enabled
}

// ShouldStopOnFirstMatch returns true if evaluation should stop on first match.
// Returns true if StopOnFirstMatch is nil (default behavior).
func (r *RulesConfig) ShouldStopOnFirstMatch() bool {
	if r == nil || r.StopOnFirstMatch == nil {
		return true
	}

	return *r.StopOnFirstMatch
}

// IsRuleEnabled returns true if the rule is enabled.
// Returns true if Enabled is nil (default behavior).
func (r *RuleConfig) IsRuleEnabled() bool {
	if r.Enabled == nil {
		return true
	}

	return *r.Enabled
}

// GetActionType returns the action type, defaulting to "block" if not set.
func (a *RuleActionConfig) GetActionType() string {
	if a == nil || a.Type == "" {
		return "block"
	}

	return a.Type
}
