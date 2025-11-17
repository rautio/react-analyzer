package rules

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
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
			&MemoizedComponentUnstableProps{}, // Cross-file rule
			&PlaceholderRule{},                // Demonstration of multiple rules
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
