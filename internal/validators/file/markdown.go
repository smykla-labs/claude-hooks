// Package file provides validators for file operations
package file

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/smykla-labs/claude-hooks/internal/linters"
	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

const (
	// markdownTimeout is the timeout for markdown linting
	markdownTimeout = 10 * time.Second
)

var (
	errFileValidationNotImpl = errors.New("file-based validation not implemented")
	errNoContent             = errors.New("no content found")
)

// MarkdownValidator validates Markdown formatting rules
type MarkdownValidator struct {
	validator.BaseValidator
	linter linters.MarkdownLinter
}

// NewMarkdownValidator creates a new MarkdownValidator
func NewMarkdownValidator(linter linters.MarkdownLinter, log logger.Logger) *MarkdownValidator {
	return &MarkdownValidator{
		BaseValidator: *validator.NewBaseValidator("validate-markdown", log),
		linter:        linter,
	}
}

// Validate checks Markdown formatting rules
func (v *MarkdownValidator) Validate(ctx *hook.Context) *validator.Result {
	log := v.Logger()

	content, err := v.getContent(ctx)
	if err != nil {
		log.Debug("skipping markdown validation", "error", err)
		return validator.Pass()
	}

	if content == "" {
		return validator.Pass()
	}

	lintCtx, cancel := context.WithTimeout(context.Background(), markdownTimeout)
	defer cancel()

	result := v.linter.Lint(lintCtx, content)

	if !result.Success {
		message := "Markdown formatting errors"
		details := map[string]string{
			"errors": strings.TrimSpace(result.RawOut),
		}

		return validator.FailWithDetails(message, details)
	}

	return validator.Pass()
}

// getContent extracts markdown content from context
//
//nolint:dupl // Same pattern used across validators, extraction would add complexity
func (v *MarkdownValidator) getContent(ctx *hook.Context) (string, error) {
	log := v.Logger()

	// Try to get content from tool input (Write operation)
	if ctx.ToolInput.Content != "" {
		return ctx.ToolInput.Content, nil
	}

	// For Edit operations in PreToolUse, read file and apply edit
	if ctx.EventType == hook.PreToolUse && ctx.ToolName == hook.Edit {
		filePath := ctx.GetFilePath()
		if filePath == "" {
			return "", errNoContent
		}

		// Read original file content
		//nolint:gosec // filePath is from Claude Code tool context, not user input
		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Debug("failed to read file for edit validation", "file", filePath, "error", err)
			return "", err
		}

		// Apply the edit (replace old_string with new_string)
		oldStr := ctx.ToolInput.OldString
		newStr := ctx.ToolInput.NewString

		if oldStr == "" {
			log.Debug("no old_string in edit operation, cannot validate")
			return "", errNoContent
		}

		// Replace first occurrence (Edit tool replaces first match)
		editedContent := strings.Replace(string(originalContent), oldStr, newStr, 1)

		return editedContent, nil
	}

	// Try to get from file path (Edit or PostToolUse)
	filePath := ctx.GetFilePath()
	if filePath != "" {
		// In PostToolUse, we could read the file, but for now skip
		// as the Bash version doesn't handle this case well either
		return "", errFileValidationNotImpl
	}

	return "", errNoContent
}
