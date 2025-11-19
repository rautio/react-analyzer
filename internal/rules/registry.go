package rules

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/config"
	"github.com/rautio/react-analyzer/internal/graph"
	"github.com/rautio/react-analyzer/internal/parser"
)

// Registry holds all available rules
type Registry struct {
	rules  []Rule
	config *config.Config
}

// NewRegistry creates a new rule registry with all available rules
func NewRegistry(cfg *config.Config) *Registry {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &Registry{
		rules: []Rule{
			&NoObjectDeps{},
			&UnstablePropsToMemo{}, // Detects unstable props to React.memo, useMemo, useCallback
			&NoDerivedState{},      // Detects useState mirroring props via useEffect
			&NoStaleState{},        // Detects state updates without functional form
			&NoInlineProps{},       // Detects inline objects/arrays/functions in JSX props
			&DeepPropDrilling{},    // Detects props drilled through 3+ component levels
			&PlaceholderRule{},     // Demonstration of multiple rules
			// Add new rules here as they're implemented
		},
		config: cfg,
	}
}

// GetRules returns all enabled rules
func (r *Registry) GetRules() []Rule {
	var enabledRules []Rule
	for _, rule := range r.rules {
		if r.isRuleEnabled(rule) {
			enabledRules = append(enabledRules, rule)
		}
	}
	return enabledRules
}

// GetRule returns a specific rule by name
func (r *Registry) GetRule(name string) (Rule, bool) {
	for _, rule := range r.rules {
		if rule.Name() == name {
			return rule, true
		}
	}
	return nil, false
}

// isRuleEnabled checks if a rule is enabled in the configuration
func (r *Registry) isRuleEnabled(rule Rule) bool {
	ruleConfig := r.config.GetRuleConfig(rule.Name())
	return ruleConfig.Enabled
}

// RunAll runs all enabled rules on an AST and returns aggregated issues
func (r *Registry) RunAll(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var allIssues []Issue

	for _, rule := range r.rules {
		// Skip disabled rules
		if !r.isRuleEnabled(rule) {
			continue
		}

		issues := rule.Check(ast, resolver)
		allIssues = append(allIssues, issues...)
	}

	return allIssues
}

// Count returns the number of enabled rules
func (r *Registry) Count() int {
	count := 0
	for _, rule := range r.rules {
		if r.isRuleEnabled(rule) {
			count++
		}
	}
	return count
}

// RunGraph runs all enabled graph-based rules on the dependency graph
func (r *Registry) RunGraph(g *graph.Graph) []Issue {
	var allIssues []Issue

	for _, rule := range r.rules {
		// Skip disabled rules
		if !r.isRuleEnabled(rule) {
			continue
		}

		// Check if rule implements GraphRule interface
		if graphRule, ok := rule.(GraphRule); ok {
			// Special handling for configurable rules
			if rule.Name() == "deep-prop-drilling" {
				// Pass config to deep-prop-drilling rule
				if dpdRule, ok := graphRule.(*DeepPropDrilling); ok {
					issues := dpdRule.CheckGraphWithConfig(g, r.config)
					allIssues = append(allIssues, issues...)
					continue
				}
			}

			// Default behavior for other graph rules
			issues := graphRule.CheckGraph(g)
			allIssues = append(allIssues, issues...)
		}
	}

	return allIssues
}
