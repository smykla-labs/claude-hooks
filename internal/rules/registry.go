package rules

import (
	"cmp"
	"slices"
	"sync"

	"github.com/pkg/errors"
)

// CompiledRule represents a rule with its pre-compiled matcher.
type CompiledRule struct {
	// Rule is the original rule configuration.
	Rule *Rule

	// Matcher is the compiled matcher for this rule.
	Matcher Matcher
}

// Registry stores compiled rules sorted by priority.
type Registry struct {
	mu    sync.RWMutex
	rules []*CompiledRule
}

// NewRegistry creates a new empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make([]*CompiledRule, 0),
	}
}

// Add compiles and adds a rule to the registry.
// Returns an error if the rule's matcher cannot be compiled.
func (r *Registry) Add(rule *Rule) error {
	if rule == nil {
		return errors.New("rule cannot be nil")
	}

	if rule.Name == "" {
		return errors.New("rule name cannot be empty")
	}

	if rule.Action == nil {
		return errors.New("rule action cannot be nil")
	}

	// Compile the matcher.
	matcher, err := BuildMatcher(rule.Match)
	if err != nil {
		return errors.Wrap(err, "failed to compile rule matcher")
	}

	// If no conditions specified, match everything.
	if matcher == nil {
		matcher = &AlwaysMatcher{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate name and update if exists.
	for i, existing := range r.rules {
		if existing.Rule.Name == rule.Name {
			r.rules[i] = &CompiledRule{
				Rule:    rule,
				Matcher: matcher,
			}

			r.sortRulesLocked()

			return nil
		}
	}

	// Add new rule.
	r.rules = append(r.rules, &CompiledRule{
		Rule:    rule,
		Matcher: matcher,
	})

	r.sortRulesLocked()

	return nil
}

// AddAll compiles and adds multiple rules to the registry.
// Returns the first error encountered, if any.
func (r *Registry) AddAll(rules []*Rule) error {
	for _, rule := range rules {
		if err := r.Add(rule); err != nil {
			return err
		}
	}

	return nil
}

// Remove removes a rule by name.
// Returns true if the rule was found and removed.
func (r *Registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules {
		if rule.Rule.Name == name {
			r.rules = slices.Delete(r.rules, i, i+1)
			return true
		}
	}

	return false
}

// Get returns a rule by name.
// Returns nil if not found.
func (r *Registry) Get(name string) *CompiledRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if rule.Rule.Name == name {
			return rule
		}
	}

	return nil
}

// GetAll returns all compiled rules sorted by priority.
func (r *Registry) GetAll() []*CompiledRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*CompiledRule, len(r.rules))
	copy(result, r.rules)

	return result
}

// GetEnabled returns all enabled rules sorted by priority.
func (r *Registry) GetEnabled() []*CompiledRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*CompiledRule, 0, len(r.rules))

	for _, rule := range r.rules {
		if rule.Rule.Enabled {
			result = append(result, rule)
		}
	}

	return result
}

// Size returns the number of rules in the registry.
func (r *Registry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.rules)
}

// Clear removes all rules from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make([]*CompiledRule, 0)
}

// sortRulesLocked sorts rules by priority (descending) then by name (ascending).
// Must be called with write lock held.
func (r *Registry) sortRulesLocked() {
	slices.SortFunc(r.rules, func(a, b *CompiledRule) int {
		// Higher priority first.
		if result := cmp.Compare(b.Rule.Priority, a.Rule.Priority); result != 0 {
			return result
		}

		// Then by name (alphabetical).
		return cmp.Compare(a.Rule.Name, b.Rule.Name)
	})
}

// Merge combines rules from another registry into this one.
// Rules with the same name will be overwritten (source takes precedence).
func (r *Registry) Merge(source *Registry) {
	if source == nil {
		return
	}

	source.mu.RLock()
	sourceRules := make([]*CompiledRule, len(source.rules))
	copy(sourceRules, source.rules)
	source.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Build a map of existing rules for quick lookup.
	existing := make(map[string]int, len(r.rules))
	for i, rule := range r.rules {
		existing[rule.Rule.Name] = i
	}

	// Merge source rules.
	for _, srcRule := range sourceRules {
		if idx, ok := existing[srcRule.Rule.Name]; ok {
			// Override existing rule.
			r.rules[idx] = srcRule
		} else {
			// Add new rule.
			r.rules = append(r.rules, srcRule)
			existing[srcRule.Rule.Name] = len(r.rules) - 1
		}
	}

	r.sortRulesLocked()
}

// MergeRules combines two rule slices with override semantics.
// Rules with the same name from the override slice take precedence.
// Returns a new sorted slice.
func MergeRules(base, override []*Rule) []*Rule {
	merged := make(map[string]*Rule, len(base)+len(override))

	for _, rule := range base {
		merged[rule.Name] = rule
	}

	for _, rule := range override {
		merged[rule.Name] = rule
	}

	result := make([]*Rule, 0, len(merged))
	for _, rule := range merged {
		result = append(result, rule)
	}

	// Sort by priority (descending) then by name (ascending).
	slices.SortFunc(result, func(a, b *Rule) int {
		if result := cmp.Compare(b.Priority, a.Priority); result != 0 {
			return result
		}

		return cmp.Compare(a.Name, b.Name)
	})

	return result
}
