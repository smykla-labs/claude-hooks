// Package config provides configuration schema types for klaudiush validators.
package config

// FileConfig groups all file-related validator configurations.
type FileConfig struct {
	// Markdown validator configuration
	Markdown *MarkdownValidatorConfig `json:"markdown,omitempty" toml:"markdown"`

	// ShellScript validator configuration
	ShellScript *ShellScriptValidatorConfig `json:"shellscript,omitempty" toml:"shellscript"`

	// Terraform validator configuration
	Terraform *TerraformValidatorConfig `json:"terraform,omitempty" toml:"terraform"`

	// Workflow validator configuration (GitHub Actions)
	Workflow *WorkflowValidatorConfig `json:"workflow,omitempty" toml:"workflow"`
}

// MarkdownValidatorConfig configures the Markdown file validator.
type MarkdownValidatorConfig struct {
	ValidatorConfig

	// Timeout is the maximum time allowed for markdown linting operations.
	// Default: "10s"
	Timeout Duration `json:"timeout,omitempty" toml:"timeout"`

	// ContextLines is the number of lines before/after an edit to include for validation.
	// This allows validating edited fragments without forcing fixes for all existing issues.
	// Default: 2
	ContextLines *int `json:"context_lines,omitempty" toml:"context_lines"`

	// HeadingSpacing enforces blank lines around headings (custom rule).
	// Default: true
	HeadingSpacing *bool `json:"heading_spacing,omitempty" toml:"heading_spacing"`

	// CodeBlockFormatting enforces proper code block formatting (custom rule).
	// Default: true
	CodeBlockFormatting *bool `json:"code_block_formatting,omitempty" toml:"code_block_formatting"`

	// ListFormatting enforces proper list item formatting and spacing (custom rule).
	// Default: true
	ListFormatting *bool `json:"list_formatting,omitempty" toml:"list_formatting"`

	// UseMarkdownlint enables markdownlint-cli integration if available.
	// Default: true
	UseMarkdownlint *bool `json:"use_markdownlint,omitempty" toml:"use_markdownlint"`

	// MarkdownlintPath is the path to the markdownlint binary.
	// Default: "" (use PATH)
	MarkdownlintPath string `json:"markdownlint_path,omitempty" toml:"markdownlint_path"`

	// MarkdownlintRules configures specific markdownlint-cli rules.
	// Map of rule name (e.g., "MD022") to enabled status.
	// When not specified, all markdownlint default rules are enabled.
	// Example: {"MD022": true, "MD041": false}
	MarkdownlintRules map[string]bool `json:"markdownlint_rules,omitempty" toml:"markdownlint_rules"`

	// MarkdownlintConfig is the path to a markdownlint configuration file.
	// If specified, this file takes precedence over MarkdownlintRules.
	// Default: "" (use MarkdownlintRules or markdownlint defaults)
	MarkdownlintConfig string `json:"markdownlint_config,omitempty" toml:"markdownlint_config"`
}

// ShellScriptValidatorConfig configures the shell script validator.
type ShellScriptValidatorConfig struct {
	ValidatorConfig

	// Timeout is the maximum time allowed for shellcheck operations.
	// Default: "10s"
	Timeout Duration `json:"timeout,omitempty" toml:"timeout"`

	// ContextLines is the number of lines before/after an edit to include for validation.
	// Default: 2
	ContextLines *int `json:"context_lines,omitempty" toml:"context_lines"`

	// UseShellcheck enables shellcheck integration if available.
	// Default: true
	UseShellcheck *bool `json:"use_shellcheck,omitempty" toml:"use_shellcheck"`

	// ShellcheckSeverity is the minimum severity level for shellcheck findings.
	// Options: "error", "warning", "info", "style"
	// Default: "warning"
	ShellcheckSeverity string `json:"shellcheck_severity,omitempty" toml:"shellcheck_severity"`

	// ExcludeRules is a list of shellcheck rules to exclude (e.g., ["SC2086", "SC2154"]).
	// Default: []
	ExcludeRules []string `json:"exclude_rules,omitempty" toml:"exclude_rules"`

	// ShellcheckPath is the path to the shellcheck binary.
	// Default: "" (use PATH)
	ShellcheckPath string `json:"shellcheck_path,omitempty" toml:"shellcheck_path"`
}

// TerraformValidatorConfig configures the Terraform/OpenTofu validator.
type TerraformValidatorConfig struct {
	ValidatorConfig

	// Timeout is the maximum time allowed for terraform/tofu operations.
	// Default: "10s"
	Timeout Duration `json:"timeout,omitempty" toml:"timeout"`

	// ContextLines is the number of lines before/after an edit to include for validation.
	// Default: 2
	ContextLines *int `json:"context_lines,omitempty" toml:"context_lines"`

	// ToolPreference specifies which tool to use when both are available.
	// Options: "tofu", "terraform", "auto" (prefers tofu)
	// Default: "auto"
	ToolPreference string `json:"tool_preference,omitempty" toml:"tool_preference"`

	// CheckFormat enables terraform/tofu format checking.
	// Default: true
	CheckFormat *bool `json:"check_format,omitempty" toml:"check_format"`

	// UseTflint enables tflint integration if available.
	// Default: true
	UseTflint *bool `json:"use_tflint,omitempty" toml:"use_tflint"`

	// TerraformPath is the path to the terraform binary.
	// Default: "" (use PATH)
	TerraformPath string `json:"terraform_path,omitempty" toml:"terraform_path"`

	// TofuPath is the path to the tofu binary.
	// Default: "" (use PATH)
	TofuPath string `json:"tofu_path,omitempty" toml:"tofu_path"`

	// TflintPath is the path to the tflint binary.
	// Default: "" (use PATH)
	TflintPath string `json:"tflint_path,omitempty" toml:"tflint_path"`
}

// WorkflowValidatorConfig configures the GitHub Actions workflow validator.
type WorkflowValidatorConfig struct {
	ValidatorConfig

	// Timeout is the maximum time allowed for actionlint operations.
	// Default: "10s"
	Timeout Duration `json:"timeout,omitempty" toml:"timeout"`

	// GHAPITimeout is the maximum time allowed for GitHub API calls.
	// Default: "5s"
	GHAPITimeout Duration `json:"gh_api_timeout,omitempty" toml:"gh_api_timeout"`

	// EnforceDigestPinning requires actions to be pinned by digest.
	// Default: true
	EnforceDigestPinning *bool `json:"enforce_digest_pinning,omitempty" toml:"enforce_digest_pinning"`

	// RequireVersionComment requires a version comment when using digest pinning.
	// Format: uses: actions/checkout@sha256... # v4.1.7
	// Default: true
	RequireVersionComment *bool `json:"require_version_comment,omitempty" toml:"require_version_comment"`

	// CheckLatestVersion checks if the version comment matches the latest release.
	// Default: true
	CheckLatestVersion *bool `json:"check_latest_version,omitempty" toml:"check_latest_version"`

	// UseActionlint enables actionlint integration if available.
	// Default: true
	UseActionlint *bool `json:"use_actionlint,omitempty" toml:"use_actionlint"`

	// ActionlintPath is the path to the actionlint binary.
	// Default: "" (use PATH)
	ActionlintPath string `json:"actionlint_path,omitempty" toml:"actionlint_path"`
}
