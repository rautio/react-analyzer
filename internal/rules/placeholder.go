package rules

import "github.com/oskari/react-analyzer/internal/parser"

// PlaceholderRule is a demonstration rule that doesn't report any issues
// This shows the registry can handle multiple rules
type PlaceholderRule struct{}

// Name returns the rule identifier
func (r *PlaceholderRule) Name() string {
	return "placeholder"
}

// Check analyzes an AST (currently does nothing)
func (r *PlaceholderRule) Check(ast *parser.AST) []Issue {
	// This is a placeholder - real rules will implement actual checks
	return nil
}
