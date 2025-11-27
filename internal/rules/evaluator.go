package rules

// Evaluator evaluates compiled rules against a match context.
type Evaluator struct {
	// registry contains all compiled rules.
	registry *Registry

	// stopOnFirstMatch controls whether to stop after the first matching rule.
	stopOnFirstMatch bool

	// defaultAction is the action to take when no rules match.
	defaultAction ActionType
}

// EvaluatorOption configures an Evaluator.
type EvaluatorOption func(*Evaluator)

// WithStopOnFirstMatch configures the evaluator to stop after the first match.
func WithStopOnFirstMatch(stop bool) EvaluatorOption {
	return func(e *Evaluator) {
		e.stopOnFirstMatch = stop
	}
}

// WithDefaultAction sets the default action when no rules match.
func WithDefaultAction(action ActionType) EvaluatorOption {
	return func(e *Evaluator) {
		e.defaultAction = action
	}
}

// NewEvaluator creates a new rule evaluator.
func NewEvaluator(registry *Registry, opts ...EvaluatorOption) *Evaluator {
	e := &Evaluator{
		registry:         registry,
		stopOnFirstMatch: true,
		defaultAction:    ActionAllow,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Evaluate evaluates all enabled rules against the given context.
// Returns the result of the first matching rule (if stopOnFirstMatch is true)
// or the highest priority matching rule.
func (e *Evaluator) Evaluate(ctx *MatchContext) *RuleResult {
	if e.registry == nil {
		return &RuleResult{
			Matched: false,
			Action:  e.defaultAction,
		}
	}

	rules := e.registry.GetEnabled()
	if len(rules) == 0 {
		return &RuleResult{
			Matched: false,
			Action:  e.defaultAction,
		}
	}

	// Rules are already sorted by priority (highest first).
	for _, compiled := range rules {
		if compiled.Matcher.Match(ctx) {
			return &RuleResult{
				Matched:   true,
				Rule:      compiled.Rule,
				Action:    compiled.Rule.Action.Type,
				Message:   compiled.Rule.Action.Message,
				Reference: compiled.Rule.Action.Reference,
			}
		}
	}

	// No rules matched.
	return &RuleResult{
		Matched: false,
		Action:  e.defaultAction,
	}
}

// EvaluateAll evaluates all enabled rules and returns all matching results.
// Results are ordered by priority (highest first).
func (e *Evaluator) EvaluateAll(ctx *MatchContext) []*RuleResult {
	if e.registry == nil {
		return nil
	}

	rules := e.registry.GetEnabled()
	if len(rules) == 0 {
		return nil
	}

	var results []*RuleResult

	for _, compiled := range rules {
		if compiled.Matcher.Match(ctx) {
			results = append(results, &RuleResult{
				Matched:   true,
				Rule:      compiled.Rule,
				Action:    compiled.Rule.Action.Type,
				Message:   compiled.Rule.Action.Message,
				Reference: compiled.Rule.Action.Reference,
			})
		}
	}

	return results
}

// FindMatchingRules returns all rules that match the given context.
// Useful for debugging and rule inspection.
func (e *Evaluator) FindMatchingRules(ctx *MatchContext) []*Rule {
	if e.registry == nil {
		return nil
	}

	rules := e.registry.GetEnabled()
	if len(rules) == 0 {
		return nil
	}

	var matching []*Rule

	for _, compiled := range rules {
		if compiled.Matcher.Match(ctx) {
			matching = append(matching, compiled.Rule)
		}
	}

	return matching
}

// FilterByValidator returns rules that apply to the given validator type.
// Includes rules with matching validator_type or wildcard patterns.
func (e *Evaluator) FilterByValidator(validatorType ValidatorType) []*CompiledRule {
	if e.registry == nil {
		return nil
	}

	rules := e.registry.GetEnabled()
	if len(rules) == 0 {
		return nil
	}

	// Create a minimal context for validator type matching.
	ctx := &MatchContext{
		ValidatorType: validatorType,
	}

	var filtered []*CompiledRule

	for _, compiled := range rules {
		// Check if rule has a validator type condition.
		if compiled.Rule.Match == nil || compiled.Rule.Match.ValidatorType == "" {
			// No validator type filter, include the rule.
			filtered = append(filtered, compiled)
			continue
		}

		// Check if the validator type matches.
		matcher := NewValidatorTypeMatcher(compiled.Rule.Match.ValidatorType)
		if matcher.Match(ctx) {
			filtered = append(filtered, compiled)
		}
	}

	return filtered
}
