// Package file provides validators for file operations
package file

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/smykla-labs/claude-hooks/internal/validator"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

const (
	// maxTruncateLength is the maximum length for truncating lines in warning messages
	maxTruncateLength = 60
)

// MarkdownValidator validates Markdown formatting rules
type MarkdownValidator struct {
	validator.BaseValidator
}

// NewMarkdownValidator creates a new MarkdownValidator
func NewMarkdownValidator(log logger.Logger) *MarkdownValidator {
	return &MarkdownValidator{
		BaseValidator: *validator.NewBaseValidator("validate-markdown", log),
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

	warnings := v.analyzeMarkdown(content)

	if len(warnings) > 0 {
		message := "Markdown formatting warnings"
		details := map[string]string{
			"warnings": strings.Join(warnings, "\n"),
		}
		return validator.WarnWithDetails(message, details)
	}

	return validator.Pass()
}

// getContent extracts markdown content from context
func (v *MarkdownValidator) getContent(ctx *hook.Context) (string, error) {
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

// analyzeMarkdown performs line-by-line analysis
func (v *MarkdownValidator) analyzeMarkdown(content string) []string {
	var warnings []string

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	prevLine := ""
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip first line
		if lineNum == 1 {
			prevLine = line
			continue
		}

		// Check for code block markers
		inCodeBlock = v.checkCodeBlock(line, prevLine, lineNum, inCodeBlock, &warnings)

		// Skip list checks inside code blocks
		if inCodeBlock {
			prevLine = line
			continue
		}

		// Check for first list item (transition from non-list to list)
		v.checkListItem(line, prevLine, lineNum, &warnings)

		// Check for content immediately after headers
		v.checkHeader(line, prevLine, lineNum, &warnings)

		prevLine = line
	}

	return warnings
}

// checkCodeBlock checks for code block markers and validates spacing
func (v *MarkdownValidator) checkCodeBlock(line, prevLine string, lineNum int, inCodeBlock bool, warnings *[]string) bool {
	if !v.isCodeBlockMarker(line) {
		return inCodeBlock
	}

	if !inCodeBlock {
		// Opening code block
		if !v.isEmptyLine(prevLine) && prevLine != "" {
			*warnings = append(*warnings,
				fmt.Sprintf("⚠️  Line %d: Code block should have empty line before it", lineNum),
				fmt.Sprintf("   Previous line: '%s'", v.truncate(prevLine)),
			)
		}
		return true
	}

	// Closing code block
	return false
}

// checkListItem validates list item spacing
func (v *MarkdownValidator) checkListItem(line, prevLine string, lineNum int, warnings *[]string) {
	if !v.isListItem(line) {
		return
	}

	if v.shouldWarnAboutListSpacing(prevLine) {
		*warnings = append(*warnings,
			fmt.Sprintf("⚠️  Line %d: First list item should have empty line before it", lineNum),
			fmt.Sprintf("   Previous line: '%s'", v.truncate(prevLine)),
		)
	}
}

// shouldWarnAboutListSpacing determines if a list item needs spacing before it
func (v *MarkdownValidator) shouldWarnAboutListSpacing(prevLine string) bool {
	return !v.isEmptyLine(prevLine) &&
		prevLine != "" &&
		!v.isListItem(prevLine) &&
		!v.isHeader(prevLine)
}

// checkHeader validates header spacing
func (v *MarkdownValidator) checkHeader(line, prevLine string, lineNum int, warnings *[]string) {
	if !v.isHeader(prevLine) {
		return
	}

	// Lists are allowed directly after headers
	if v.shouldWarnAboutHeaderSpacing(line) {
		*warnings = append(*warnings,
			fmt.Sprintf("⚠️  Line %d: Header should have empty line after it", lineNum-1),
			fmt.Sprintf("   Header: '%s'", v.truncate(prevLine)),
			fmt.Sprintf("   Next line: '%s'", v.truncate(line)),
		)
	}
}

// shouldWarnAboutHeaderSpacing determines if content after a header needs spacing
func (v *MarkdownValidator) shouldWarnAboutHeaderSpacing(line string) bool {
	return line != "" &&
		!v.isEmptyLine(line) &&
		!v.isHeader(line) &&
		!v.isComment(line) &&
		!v.isListItem(line)
}

var (
	codeBlockRegex = regexp.MustCompile(`^` + "```")
	listItemRegex  = regexp.MustCompile(`^[[:space:]]*[-*+][[:space:]]|^[[:space:]]*[0-9]+\.[[:space:]]`)
	headerRegex    = regexp.MustCompile(`^#+[[:space:]]`)
	commentRegex   = regexp.MustCompile(`^<!--`)
	emptyLineRegex = regexp.MustCompile(`^[[:space:]]*$`)
)

// isCodeBlockMarker checks if line starts a code block
func (v *MarkdownValidator) isCodeBlockMarker(line string) bool {
	return codeBlockRegex.MatchString(line)
}

// isListItem checks if line is a list item
func (v *MarkdownValidator) isListItem(line string) bool {
	return listItemRegex.MatchString(line)
}

// isHeader checks if line is a header
func (v *MarkdownValidator) isHeader(line string) bool {
	return headerRegex.MatchString(line)
}

// isComment checks if line is an HTML comment
func (v *MarkdownValidator) isComment(line string) bool {
	return commentRegex.MatchString(line)
}

// isEmptyLine checks if line is empty or whitespace-only
func (v *MarkdownValidator) isEmptyLine(line string) bool {
	return emptyLineRegex.MatchString(line)
}

// truncate truncates string to maxTruncateLength
func (v *MarkdownValidator) truncate(s string) string {
	if len(s) <= maxTruncateLength {
		return s
	}
	return s[:maxTruncateLength]
}
