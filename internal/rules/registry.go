package rules

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/graph"
	"github.com/rautio/react-analyzer/internal/parser"
)

// Registry holds all available rules
type Registry struct {
	rules []Rule
}

// NewRegistry creates a new rule registry with all available rules
func NewRegistry() *Registry {
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
	}
}

// GetRules returns all registered rules
func (r *Registry) GetRules() []Rule {
	return r.rules
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

// RunAll runs all registered rules on an AST and returns aggregated issues
func (r *Registry) RunAll(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var allIssues []Issue

	for _, rule := range r.rules {
		issues := rule.Check(ast, resolver)
		allIssues = append(allIssues, issues...)
	}

	return allIssues
}

// Count returns the number of registered rules
func (r *Registry) Count() int {
	return len(r.rules)
}

// RunGraph runs all graph-based rules on the dependency graph
func (r *Registry) RunGraph(g *graph.Graph) []Issue {
	var allIssues []Issue

	for _, rule := range r.rules {
		// Check if rule implements GraphRule interface
		if graphRule, ok := rule.(GraphRule); ok {
			issues := graphRule.CheckGraph(g)
			allIssues = append(allIssues, issues...)
		}
	}

	return allIssues
}
