package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/graph"
	"github.com/rautio/react-analyzer/internal/parser"
)

// DeepPropDrilling detects when props are drilled through 3+ component levels
// where intermediate components don't use the props - they only pass them down.
// This is an anti-pattern that can be solved using Context API.
type DeepPropDrilling struct{}

// Name returns the rule identifier
func (r *DeepPropDrilling) Name() string {
	return "deep-prop-drilling"
}

// Check analyzes the codebase for prop drilling violations
// Note: This rule requires graph-based analysis, so it returns nil for individual ASTs
// The actual detection happens in CheckGraph (called after all files are parsed)
func (r *DeepPropDrilling) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	// This rule operates on the graph, not individual ASTs
	// Individual file checks are not applicable
	return nil
}

// CheckGraph performs graph-based prop drilling detection
// This is called after all modules have been parsed and the graph is built
func (r *DeepPropDrilling) CheckGraph(g *graph.Graph) []Issue {
	violations := graph.DetectPropDrilling(g)

	var issues []Issue
	for _, v := range violations {
		issues = append(issues, Issue{
			Rule:     r.Name(),
			FilePath: v.Origin.FilePath,
			Line:     v.Origin.Line,
			Column:   v.Origin.Column,
			Message:  r.formatMessage(v),
		})
	}

	return issues
}

// formatMessage creates a user-friendly error message
func (r *DeepPropDrilling) formatMessage(v graph.PropDrillingViolation) string {
	// Build passthrough path string
	path := ""
	for i, comp := range v.PassthroughComponents {
		if i > 0 {
			path += " → "
		}
		path += comp.Name
	}

	if path == "" {
		// No passthrough components (shouldn't happen for depth >= 3, but safety check)
		return fmt.Sprintf(
			"Prop '%s' is drilled through %d component levels. %s",
			v.PropName,
			v.Depth,
			v.Recommendation,
		)
	}

	return fmt.Sprintf(
		"Prop '%s' is drilled through %d component levels (%s). %s",
		v.PropName,
		v.Depth,
		path,
		v.Recommendation,
	)
}
