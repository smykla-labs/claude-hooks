# Rule Engine Implementation

## Overview

The rule engine provides dynamic validation configuration without modifying code.
Rules allow users to define custom validation behavior via TOML configuration.

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

**Regex indicators**: `^`, `$`, `(?`, `\\d`, `\\w`, `[`, `]`, `(`, `)`, `|`, `+`

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

## Configuration Schema (Phase 2)

Rules are configured in TOML files.

### Config Files

- Global: `~/.klaudiush/config.toml`
- Project: `.klaudiush/config.toml` or `klaudiush.toml`

### TOML Schema

```toml
[rules]
enabled = true
stop_on_first_match = true

[[rules.rules]]
name = "block-origin-push"
description = "Block pushes to origin remote"
priority = 100
enabled = true

[rules.rules.match]
validator_type = "git.push"
repo_pattern = "**/myorg/**"
remote = "origin"

[rules.rules.action]
type = "block"
message = "Don't push to origin"
reference = "GIT019"
```

### Config Precedence

1. CLI Flags (highest)
2. Environment Variables (`KLAUDIUSH_*`)
3. Project Config
4. Global Config
5. Defaults (lowest)

### Rule Merge Semantics

- Rules with same name: project overrides global
- Rules with different names: combined

### Config Types

- `pkg/config/rules.go`: RulesConfig, RuleConfig, RuleMatchConfig
- `internal/config/factory/rules_factory.go`: Creates RuleEngine
- `internal/config/koanf.go`: extractRules, mergeRules functions

## Test Coverage

148 tests including Phase 1 (126) and Phase 2 (22).

## Dependencies

- `github.com/gobwas/glob v0.2.3` - Glob pattern matching
- `github.com/knadh/koanf/v2` - Configuration loading
- `github.com/pkg/errors` - Error wrapping

## Related Files

- Investigation docs: `tmp/investigations/dynamic-validation-config/`
- Main plan: `.claude/plans/compiled-bouncing-river.md`

## Implementation Status

### Phase 1: Core Rule Engine ✅ Complete

Implemented in Session 2 (2025-11-27):

- Pattern system with auto-detection
- All matchers (9 types + composites)
- Registry with priority sorting
- Evaluator with first-match semantics
- RuleEngine main interface
- ValidatorAdapter for integration
- 126 tests, 90% coverage

### Phase 2: Configuration Schema ✅ Complete

Verified complete in Session 3 (2025-11-27):

- `pkg/config/rules.go` schema created
- Root config extended with Rules field
- Validator config extended with RulesEnabled field
- TOML loading with deep merge implemented
- Factory creates RuleEngine via RulesFactory
- 22 additional tests for config loading
- All acceptance criteria met

**Kong/Kuma Default Rules**: Deferred to Phase 3 to be added alongside PushValidator migration.

## Validator Integration Pattern

When adding rule support to validators, the pattern is:

1. **Validator struct**: Add optional `ruleAdapter *rules.RuleValidatorAdapter` field
2. **Constructor**: Accept `ruleAdapter` as last parameter (can be nil for backward compatibility)
3. **Validate()**: Check rules first, return early if matched, otherwise continue with built-in logic
4. **Factory**: Create adapter only if `f.ruleEngine != nil`, pass correct `ValidatorType` constant

**Key insight**: Nil ruleAdapter is valid - allows gradual migration without breaking existing code. Tests pass nil, production code creates adapter when rule engine is configured.
