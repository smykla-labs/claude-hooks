// Package fixers provides auto-fix functionality for doctor checks.
package fixers

import (
	"context"
	"fmt"
	"strings"

	internalconfig "github.com/smykla-labs/klaudiush/internal/config"
	"github.com/smykla-labs/klaudiush/internal/doctor"
	ruleschecker "github.com/smykla-labs/klaudiush/internal/doctor/checkers/rules"
	"github.com/smykla-labs/klaudiush/internal/prompt"
	"github.com/smykla-labs/klaudiush/pkg/config"
)

const (
	// disabledNote is the full description for rules with no existing description.
	disabledNote = "DISABLED BY DOCTOR: fix the rule configuration and re-enable"

	// disabledSuffix is appended to existing descriptions.
	disabledSuffix = "[DISABLED BY DOCTOR: fix and re-enable]"
)

// RulesFixer fixes invalid rules by disabling them.
type RulesFixer struct {
	prompter     prompt.Prompter
	writer       *internalconfig.Writer
	rulesChecker *ruleschecker.RulesChecker
}

// NewRulesFixer creates a new RulesFixer.
func NewRulesFixer(prompter prompt.Prompter) *RulesFixer {
	return &RulesFixer{
		prompter:     prompter,
		writer:       internalconfig.NewWriter(),
		rulesChecker: ruleschecker.NewRulesChecker(),
	}
}

// ID returns the fixer identifier.
func (*RulesFixer) ID() string {
	return "fix_invalid_rules"
}

// Description returns a human-readable description.
func (*RulesFixer) Description() string {
	return "Disable invalid rules in configuration (rules can be re-enabled after manual fix)"
}

// CanFix checks if this fixer can fix the given result.
func (*RulesFixer) CanFix(result doctor.CheckResult) bool {
	return result.FixID == "fix_invalid_rules" && result.Status == doctor.StatusFail
}

// Fix disables invalid rules in the configuration.
func (f *RulesFixer) Fix(ctx context.Context, interactive bool) error {
	// Run the rules checker to get current issues
	f.rulesChecker.Check(ctx)
	issues := f.rulesChecker.GetIssues()

	if len(issues) == 0 {
		return nil
	}

	cfg, err := f.loadConfig()
	if err != nil {
		return err
	}

	if cfg.Rules == nil || len(cfg.Rules.Rules) == 0 {
		return nil
	}

	// Collect indices of rules to disable
	rulesToDisable := f.collectFixableRules(issues)
	if len(rulesToDisable) == 0 {
		return nil
	}

	// Confirm with user if interactive
	if interactive {
		if !f.confirmFix(len(rulesToDisable)) {
			return nil
		}
	}

	// Disable the invalid rules
	f.disableRules(cfg, rulesToDisable)

	// Write back to project config
	if err := f.writer.WriteProject(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// loadConfig loads the configuration without validation.
// This allows fixing invalid configurations.
func (*RulesFixer) loadConfig() (*config.Config, error) {
	loader, err := internalconfig.NewKoanfLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	// Use LoadWithoutValidation to allow loading invalid configs
	cfg, err := loader.LoadWithoutValidation(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// collectFixableRules collects indices of rules that can be fixed.
func (*RulesFixer) collectFixableRules(issues []ruleschecker.RuleIssue) map[int]bool {
	rulesToDisable := make(map[int]bool)

	for _, issue := range issues {
		if issue.Fixable {
			rulesToDisable[issue.RuleIndex] = true
		}
	}

	return rulesToDisable
}

// confirmFix prompts the user for confirmation.
func (f *RulesFixer) confirmFix(count int) bool {
	msg := fmt.Sprintf(
		"Disable %d invalid rule(s)? (They can be re-enabled after manual fix)",
		count,
	)

	confirmed, err := f.prompter.Confirm(msg, true)
	if err != nil {
		return false
	}

	return confirmed
}

// disableRules marks the specified rules as disabled.
func (*RulesFixer) disableRules(cfg *config.Config, rulesToDisable map[int]bool) {
	for idx := range rulesToDisable {
		if idx < len(cfg.Rules.Rules) {
			disabled := false
			cfg.Rules.Rules[idx].Enabled = &disabled

			// Add a description note if not present
			desc := cfg.Rules.Rules[idx].Description
			if desc == "" {
				cfg.Rules.Rules[idx].Description = disabledNote
			} else if !containsDisabledNote(desc) {
				cfg.Rules.Rules[idx].Description = desc + " " + disabledSuffix
			}
		}
	}
}

// containsDisabledNote checks if description already has the disabled note.
func containsDisabledNote(desc string) bool {
	if desc == "" {
		return false
	}

	return desc == disabledNote || strings.HasSuffix(desc, disabledSuffix)
}
