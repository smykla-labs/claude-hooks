package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

const (
	// terraformTimeout is the timeout for terraform/tofu commands
	terraformTimeout = 10 * time.Second
)

// TerraformValidator validates Terraform/OpenTofu file formatting
type TerraformValidator struct {
	validator.BaseValidator
}

// NewTerraformValidator creates a new TerraformValidator
func NewTerraformValidator(log logger.Logger) *TerraformValidator {
	return &TerraformValidator{
		BaseValidator: *validator.NewBaseValidator("validate-terraform", log),
	}
}

// Validate checks Terraform formatting and optionally runs tflint
func (v *TerraformValidator) Validate(ctx *hook.Context) *validator.Result {
	log := v.Logger()
	content, err := v.getContent(ctx)
	if err != nil {
		log.Debug("skipping terraform validation", "error", err)
		return validator.Pass()
	}

	if content == "" {
		return validator.Pass()
	}

	// Detect which tool to use
	tool := v.detectTool()
	log.Debug("detected terraform tool", "tool", tool)

	// Create temp file for validation
	tmpFile, err := v.createTempFile(content)
	if err != nil {
		log.Debug("failed to create temp file", "error", err)
		return validator.Pass()
	}
	defer v.cleanupTempFile(tmpFile)

	var warnings []string

	// Run format check
	if fmtWarning := v.checkFormat(tool, tmpFile); fmtWarning != "" {
		warnings = append(warnings, fmtWarning)
	}

	// Run tflint if available
	if lintWarnings := v.runTflint(tmpFile); len(lintWarnings) > 0 {
		warnings = append(warnings, lintWarnings...)
	}

	if len(warnings) > 0 {
		message := "Terraform validation warnings"
		details := map[string]string{
			"warnings": strings.Join(warnings, "\n"),
		}
		return validator.WarnWithDetails(message, details)
	}

	return validator.Pass()
}

// getContent extracts terraform content from context
func (v *TerraformValidator) getContent(ctx *hook.Context) (string, error) {
	// Try to get content from tool input (Write operation)
	if ctx.ToolInput.Content != "" {
		return ctx.ToolInput.Content, nil
	}

	// For Edit operations in PreToolUse, we can't easily get final content
	// Skip validation
	if ctx.EventType == hook.PreToolUse && ctx.ToolName == hook.Edit {
		return "", errors.New("cannot validate Edit operations in PreToolUse")
	}

	// Try to get from file path (Edit or PostToolUse)
	filePath := ctx.GetFilePath()
	if filePath != "" {
		// In PostToolUse, we could read the file, but for now skip
		// as the Bash version doesn't handle this case well either
		return "", errors.New("file-based validation not implemented")
	}

	return "", errors.New("no content found")
}

// detectTool detects whether to use tofu or terraform
func (v *TerraformValidator) detectTool() string {
	// Check for tofu first (takes precedence)
	if _, err := exec.LookPath("tofu"); err == nil {
		return "tofu"
	}

	// Fall back to terraform
	if _, err := exec.LookPath("terraform"); err == nil {
		return "terraform"
	}

	return ""
}

// createTempFile creates a temporary .tf file with the content
func (v *TerraformValidator) createTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "terraform-*.tf")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// cleanupTempFile removes the temporary file
func (v *TerraformValidator) cleanupTempFile(path string) {
	if err := os.Remove(path); err != nil {
		v.Logger().Debug("failed to remove temp file", "path", path, "error", err)
	}
}

// checkFormat runs terraform/tofu fmt -check
func (v *TerraformValidator) checkFormat(tool, filePath string) string {
	if tool == "" {
		return "⚠️  Neither 'tofu' nor 'terraform' found in PATH - skipping format check"
	}

	ctx, cancel := context.WithTimeout(context.Background(), terraformTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, tool, "fmt", "-check", "-diff", filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		// Formatting is correct
		return ""
	}

	// Format check failed
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
		diff := stdout.String()
		if diff == "" {
			diff = stderr.String()
		}
		return fmt.Sprintf("⚠️  Terraform formatting issues detected:\n%s\n   Run '%s fmt %s' to fix",
			strings.TrimSpace(diff), tool, filepath.Base(filePath))
	}

	v.Logger().Debug("fmt command failed", "error", err, "stderr", stderr.String())
	return fmt.Sprintf("⚠️  Failed to run '%s fmt -check': %v", tool, err)
}

// runTflint runs tflint on the file if available
func (v *TerraformValidator) runTflint(filePath string) []string {
	// Check if tflint is available
	if _, err := exec.LookPath("tflint"); err != nil {
		v.Logger().Debug("tflint not found in PATH, skipping")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), terraformTimeout)
	defer cancel()

	// Run tflint on the file
	cmd := exec.CommandContext(ctx, "tflint", "--format=compact", filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(stdout.String())

	if err != nil {
		// tflint returns non-zero on findings
		if output != "" {
			return []string{"⚠️  tflint findings:\n" + output}
		}
		v.Logger().Debug("tflint failed", "error", err, "stderr", stderr.String())
		return nil
	}

	// No findings
	return nil
}
