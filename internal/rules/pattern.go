package rules

import (
	"regexp"
	"strings"
	"sync"

	"github.com/gobwas/glob"
)

// PatternType indicates whether a pattern is a glob or regex.
type PatternType int

const (
	// PatternTypeGlob indicates a glob pattern (e.g., "*/kong/*").
	PatternTypeGlob PatternType = iota

	// PatternTypeRegex indicates a regex pattern (e.g., "^.*/kong/.*$").
	PatternTypeRegex
)

// regexIndicators are strings that indicate a pattern is regex rather than glob.
var regexIndicators = []string{
	"^",   // Start anchor
	"$",   // End anchor
	"(?",  // Non-capturing group or flags
	"\\d", // Digit class
	"\\w", // Word class
	"\\s", // Whitespace class
	"\\b", // Word boundary
	"[",   // Character class start
	"]",   // Character class end
	"(",   // Capturing group start
	")",   // Capturing group end
	"|",   // Alternation
	"+",   // One or more quantifier
	".*",  // Wildcard in regex
	".+",  // One or more any
	"\\.", // Escaped dot
}

// DetectPatternType determines whether a pattern is a glob or regex.
// Returns PatternTypeRegex if the pattern contains regex-specific syntax,
// otherwise returns PatternTypeGlob.
func DetectPatternType(pattern string) PatternType {
	for _, indicator := range regexIndicators {
		if strings.Contains(pattern, indicator) {
			return PatternTypeRegex
		}
	}

	return PatternTypeGlob
}

// GlobPattern wraps a compiled glob pattern.
type GlobPattern struct {
	pattern  string
	compiled glob.Glob
}

// NewGlobPattern creates a new GlobPattern from the given pattern string.
func NewGlobPattern(pattern string) (*GlobPattern, error) {
	compiled, err := glob.Compile(pattern, '/')
	if err != nil {
		return nil, err
	}

	return &GlobPattern{
		pattern:  pattern,
		compiled: compiled,
	}, nil
}

// Match returns true if the string matches the glob pattern.
func (p *GlobPattern) Match(s string) bool {
	return p.compiled.Match(s)
}

// String returns the original pattern string.
func (p *GlobPattern) String() string {
	return p.pattern
}

// RegexPattern wraps a compiled regular expression.
type RegexPattern struct {
	pattern  string
	compiled *regexp.Regexp
}

// NewRegexPattern creates a new RegexPattern from the given pattern string.
func NewRegexPattern(pattern string) (*RegexPattern, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &RegexPattern{
		pattern:  pattern,
		compiled: compiled,
	}, nil
}

// Match returns true if the string matches the regex pattern.
func (p *RegexPattern) Match(s string) bool {
	return p.compiled.MatchString(s)
}

// String returns the original pattern string.
func (p *RegexPattern) String() string {
	return p.pattern
}

// CompilePattern compiles a pattern string, auto-detecting the pattern type.
// Returns the compiled Pattern or an error if compilation fails.
//
//nolint:ireturn // interface for polymorphism
func CompilePattern(pattern string) (Pattern, error) {
	patternType := DetectPatternType(pattern)

	switch patternType {
	case PatternTypeRegex:
		return NewRegexPattern(pattern)
	default:
		return NewGlobPattern(pattern)
	}
}

// PatternCache provides thread-safe caching of compiled patterns.
type PatternCache struct {
	mu       sync.RWMutex
	patterns map[string]Pattern
	errors   map[string]error
}

// NewPatternCache creates a new PatternCache.
func NewPatternCache() *PatternCache {
	return &PatternCache{
		patterns: make(map[string]Pattern),
		errors:   make(map[string]error),
	}
}

// Get returns a compiled pattern, compiling and caching it if necessary.
// Returns the cached error if the pattern previously failed to compile.
//
//nolint:ireturn // interface for polymorphism
func (c *PatternCache) Get(patternStr string) (Pattern, error) {
	// Fast path: check if already cached.
	c.mu.RLock()

	if p, ok := c.patterns[patternStr]; ok {
		c.mu.RUnlock()
		return p, nil
	}

	if err, ok := c.errors[patternStr]; ok {
		c.mu.RUnlock()
		return nil, err
	}

	c.mu.RUnlock()

	// Slow path: compile and cache.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock.
	if p, ok := c.patterns[patternStr]; ok {
		return p, nil
	}

	if err, ok := c.errors[patternStr]; ok {
		return nil, err
	}

	pattern, err := CompilePattern(patternStr)
	if err != nil {
		c.errors[patternStr] = err
		return nil, err
	}

	c.patterns[patternStr] = pattern

	return pattern, nil
}

// Clear removes all cached patterns.
func (c *PatternCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.patterns = make(map[string]Pattern)
	c.errors = make(map[string]error)
}

// Size returns the number of cached patterns.
func (c *PatternCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.patterns)
}

// defaultCache is the global pattern cache.
var defaultCache = NewPatternCache()

// GetCachedPattern returns a compiled pattern from the default cache.
//
//nolint:ireturn // interface for polymorphism
func GetCachedPattern(pattern string) (Pattern, error) {
	return defaultCache.Get(pattern)
}

// ClearPatternCache clears the default pattern cache.
func ClearPatternCache() {
	defaultCache.Clear()
}
