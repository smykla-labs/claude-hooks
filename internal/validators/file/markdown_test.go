package file_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/claude-hooks/internal/validators/file"
	"github.com/smykla-labs/claude-hooks/pkg/hook"
	"github.com/smykla-labs/claude-hooks/pkg/logger"
)

var _ = Describe("MarkdownValidator", func() {
	var (
		v   *file.MarkdownValidator
		ctx *hook.Context
	)

	BeforeEach(func() {
		v = file.NewMarkdownValidator(logger.NewNoOpLogger())
		ctx = &hook.Context{
			EventType: hook.PreToolUse,
			ToolName:  hook.Write,
		}
	})

	Describe("Name", func() {
		It("returns correct validator name", func() {
			Expect(v.Name()).To(Equal("validate-markdown"))
		})
	})

	Describe("Validate", func() {
		Context("with valid markdown", func() {
			It("passes for empty content", func() {
				ctx.ToolInput.Content = ""
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("passes for markdown with proper spacing", func() {
				content := `# Header

Some text here.

- List item 1
- List item 2

` + "```" + `bash
code here
` + "```" + `

More text.
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("passes for consecutive list items", func() {
				content := `- Item 1
- Item 2
- Item 3
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("passes for list after header", func() {
				content := `## Features
- Feature 1
- Feature 2
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("passes for consecutive headers", func() {
				content := `# Title
## Subtitle
### Section
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("passes for header followed by comment", func() {
				content := `# Header
<!-- Comment -->
Text
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})
		})

		Context("code block validation", func() {
			It("warns when code block has no empty line before", func() {
				content := `Some text
` + "```" + `bash
code
` + "```"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 2: Code block should have empty line before it"))
			})

			It("passes when code block has empty line before", func() {
				content := `Some text

` + "```" + `bash
code
` + "```"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})

			It("ignores list markers inside code blocks", func() {
				content := `
` + "```" + `bash
- this is not a list
* also not a list
1. still not a list
` + "```"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})

			It("handles multiple code blocks", func() {
				content := `Text

` + "```" + `
code1
` + "```" + `

More text
` + "```" + `
code2
` + "```"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 8: Code block should have empty line before it"))
			})
		})

		Context("list item validation", func() {
			It("warns when first list item has no empty line before", func() {
				content := `Some text
- List item
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 2: First list item should have empty line before it"))
			})

			It("passes when first list item has empty line before", func() {
				content := `Some text

- List item
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})

			It("handles different list markers", func() {
				content := `Text
- Dash item
Text
* Star item
Text
+ Plus item
Text
1. Ordered item
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 2: First list item"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 4: First list item"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 6: First list item"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 8: First list item"))
			})

			It("handles indented list items", func() {
				content := `Text
  - Indented item
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 2: First list item"))
			})

			It("does not warn for consecutive list items", func() {
				content := `
- Item 1
- Item 2
- Item 3
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})
		})

		Context("header validation", func() {
			It("warns when header has no empty line after", func() {
				content := `# Header
Text immediately after
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 1: Header should have empty line after it"))
			})

			It("passes when header has empty line after", func() {
				content := `# Header

Text after empty line
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})

			It("handles different header levels", func() {
				content := `# H1
Text
## H2
Text
### H3
Text
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 1: Header should have empty line after it"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 3: Header should have empty line after it"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 5: Header should have empty line after it"))
			})
		})

		Context("edge cases", func() {
			It("skips validation for Edit operations", func() {
				ctx.ToolName = hook.Edit
				ctx.ToolInput.FilePath = "/path/to/file.md"
				ctx.ToolInput.Content = ""
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("skips validation when no content available", func() {
				ctx.ToolInput.Content = ""
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
			})

			It("handles truncation of long lines in warnings", func() {
				longLine := strings.Repeat("x", 100)
				content := longLine + "\n- List item\n"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Details["warnings"]).To(ContainSubstring("Previous line: '" + strings.Repeat("x", 60)))
				Expect(result.Details["warnings"]).NotTo(ContainSubstring(strings.Repeat("x", 70)))
			})

			It("handles empty lines properly", func() {
				content := `
` + "```" + `
code
` + "```" + `
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})
		})

		Context("complex scenarios", func() {
			It("handles mixed formatting issues", func() {
				content := `# Title
Immediate text
- List without space
` + "```" + `
code
` + "```"
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeFalse())
				Expect(result.ShouldBlock).To(BeFalse())
				Expect(result.Message).To(Equal("Markdown formatting warnings"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 1: Header should have empty line after it"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 3: First list item should have empty line before it"))
				Expect(result.Details["warnings"]).To(ContainSubstring("Line 4: Code block should have empty line before it"))
			})

			It("handles real-world markdown example", func() {
				content := `# Project Title

## Overview

This is a description.

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

` + "```" + `bash
npm install
` + "```" + `

## Usage

1. Step one
2. Step two

Done!
`
				ctx.ToolInput.Content = content
				result := v.Validate(ctx)
				Expect(result.Passed).To(BeTrue())
				Expect(result.Message).To(BeEmpty())
			})
		})
	})
})
