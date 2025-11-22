// Package parser provides Bash command parsing capabilities using mvdan.cc/sh
package parser

import (
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// CmdType represents the type of command.
type CmdType int

const (
	// CmdTypeSimple represents a simple command (e.g., "git add file.txt").
	CmdTypeSimple CmdType = iota
	// CmdTypePipe represents a pipeline (e.g., "ls | grep foo").
	CmdTypePipe
	// CmdTypeSubshell represents a subshell (e.g., "(cd dir && git commit)").
	CmdTypeSubshell
	// CmdTypeCmdSubst represents command substitution (e.g., "$(git log)").
	CmdTypeCmdSubst
	// CmdTypeChain represents chained commands (&&, ||, ;).
	CmdTypeChain
)

// String returns string representation of CmdType.
func (t CmdType) String() string {
	switch t {
	case CmdTypeSimple:
		return "Simple"
	case CmdTypePipe:
		return "Pipe"
	case CmdTypeSubshell:
		return "Subshell"
	case CmdTypeCmdSubst:
		return "CmdSubst"
	case CmdTypeChain:
		return "Chain"
	default:
		return "Unknown"
	}
}

// Location represents position in source code.
type Location struct {
	Line   uint
	Column uint
}

// Command represents a parsed command with metadata.
type Command struct {
	Name     string   // Command name (e.g., "git")
	Args     []string // Command arguments
	Location Location // Position in source
	Type     CmdType  // Command type
	Raw      string   // Raw command string
}

// String returns a string representation of the command.
func (c *Command) String() string {
	if len(c.Args) == 0 {
		return c.Name
	}

	return fmt.Sprintf("%s %s", c.Name, strings.Join(c.Args, " "))
}

// FullCommand returns the complete command as a string slice.
func (c *Command) FullCommand() []string {
	result := []string{c.Name}
	result = append(result, c.Args...)

	return result
}

// wordToString converts syntax.Word to string, handling quotes and expansions.
func wordToString(word *syntax.Word) string {
	if word == nil {
		return ""
	}

	var result strings.Builder

	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			result.WriteString(p.Value)
		case *syntax.SglQuoted:
			result.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, dqPart := range p.Parts {
				if lit, ok := dqPart.(*syntax.Lit); ok {
					result.WriteString(lit.Value)
				}
			}
		}
	}

	return result.String()
}

// wordsToStrings converts a slice of syntax.Word to string slice.
func wordsToStrings(words []*syntax.Word) []string {
	result := make([]string, 0, len(words))

	for _, word := range words {
		if s := wordToString(word); s != "" {
			result = append(result, s)
		}
	}

	return result
}
