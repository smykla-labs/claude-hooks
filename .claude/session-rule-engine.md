# Rule Engine Implementation

## Overview

The rule engine provides dynamic validation configuration without modifying code. Rules allow users to define custom validation behavior via TOML configuration.

## Package Structure

```text
internal/rules/
├── types.go      - Core types, interfaces, enums
├── pattern.go    - Pattern compilation, cache (gobwas/glob)
├── matcher.go    - All matcher implementations
├── registry.go   - Rule compilation, sorting, merge
├── evaluator.go  - Rule evaluation logic
├── engine.go     - RuleEngine implementation
└── adapter.go    - Validator integration
```

## Core Types

### Rule

```go
type Rule struct {
    Name        string
    Description string
    Enabled     bool
    Priority    int         // Higher = evaluated first
    Match       *RuleMatch  // Conditions (AND logic)
    Action      *RuleAction // Block, Warn, or Allow
}
```

### RuleMatch Conditions

- `ValidatorType`: "git.push", "git.*", "*"
- `RepoPattern`: Glob or regex for repository path
- `Remote`: Exact remote name match
- `BranchPattern`: Branch name pattern
- `FilePattern`: File path pattern
- `ContentPattern`: File content regex
- `CommandPattern`: Bash command pattern
- `ToolType`: "Bash", "Write", "Edit", etc.
- `EventType`: "PreToolUse", "PostToolUse"

### Actions

- `ActionBlock`: Stops operation with error (exit 2)
- `ActionWarn`: Logs warning, allows operation
- `ActionAllow`: Explicitly allows operation

## Pattern System

Auto-detects pattern type based on syntax:

**Regex indicators**: `^`, `$`, `(?`, `\\d`, `\\w`, `[`, `]`, `(`, `)`, `|`, `+`, `.*`, `.+`

**Glob patterns** (via `gobwas/glob`):

- `*` - Single path component
- `**` - Multiple path components
- `{a,b}` - Brace expansion
- `?` - Single character

**Important**: Use `**` for multi-directory matching:

- `**/myorg/**` matches `/home/user/myorg/project`
- `*/myorg/*` only matches `/myorg/project`

## Matchers

### Simple Matchers

- `RemoteMatcher`: Exact string match
- `ValidatorTypeMatcher`: Supports wildcards (`git.*`)
- `ToolTypeMatcher`: Case-insensitive match
- `EventTypeMatcher`: Case-insensitive match

### Pattern Matchers

- `RepoPatternMatcher`: Repository root path
- `BranchPatternMatcher`: Branch name
- `FilePatternMatcher`: File path (falls back to HookContext)
- `ContentPatternMatcher`: File content (always regex)
- `CommandPatternMatcher`: Bash command

### Composite Matchers

- `CompositeMatcher(AND)`: All conditions must match
- `CompositeMatcher(OR)`: Any condition matches
- `CompositeMatcher(NOT)`: Inverts result

## Registry

- Stores compiled rules sorted by priority
- Merge semantics: same name = override, different name = combine
- Thread-safe with RWMutex

## Evaluator

- Evaluates enabled rules in priority order
- Stops on first match (configurable)
- Returns `RuleResult` with action and message

## Engine

Main entry point:

```go
engine, err := NewRuleEngine(rules,
    WithLogger(log),
    WithEngineDefaultAction(ActionAllow),
)

result := engine.Evaluate(ctx, &MatchContext{
    ValidatorType: ValidatorGitPush,
    GitContext: &GitContext{
        RepoRoot: "/path/to/repo",
        Remote:   "origin",
    },
})
```

## ValidatorAdapter

Bridges engine with validators:

```go
adapter := NewRuleValidatorAdapter(
    engine,
    ValidatorGitPush,
    WithGitContextProvider(func() *GitContext {
        return &GitContext{
            RepoRoot: gitRunner.GetRepoRoot(),
            Remote:   extractedRemote,
        }
    }),
)

// In validator.Validate():
if result := adapter.CheckRules(ctx, hookCtx); result != nil {
    return result  // Rule matched, use rule result
}
// Continue with built-in validation logic...
```

## Test Coverage

126 tests, 90% coverage.

## Dependencies

- `github.com/gobwas/glob v0.2.3` - Glob pattern matching
- `github.com/pkg/errors` - Error wrapping

## Related Files

- Investigation docs: `tmp/investigations/dynamic-validation-config/`
- Main plan: `.claude/plans/compiled-bouncing-river.md`
