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
		// Create issue at origin (where state is defined)
		issues = append(issues, Issue{
			Rule:     r.Name(),
			FilePath: v.Origin.FilePath,
			Line:     v.Origin.Line,
			Column:   v.Origin.Column,
			Message:  r.formatOriginMessage(v),
		})

		// Create issues at each passthrough component
		for i, comp := range v.PassthroughComponents {
			issues = append(issues, Issue{
				Rule:     r.Name(),
				FilePath: comp.FilePath,
				Line:     comp.Line,
				Column:   0, // Start of line
				Message:  r.formatPassthroughMessage(v, i),
			})
		}
	}

	return issues
}

// formatOriginMessage creates message for the state origin location
func (r *DeepPropDrilling) formatOriginMessage(v graph.PropDrillingViolation) string {
	// Build passthrough path string
	path := ""
	for i, comp := range v.PassthroughComponents {
		if i > 0 {
			path += " → "
		}
		path += comp.Name
	}

	return fmt.Sprintf(
		"State '%s' is drilled through %d component levels (%s). %s",
		v.PropName,
		v.Depth,
		path,
		v.Recommendation,
	)
}

// formatPassthroughMessage creates message for passthrough component locations
func (r *DeepPropDrilling) formatPassthroughMessage(v graph.PropDrillingViolation, componentIndex int) string {
	comp := v.PassthroughComponents[componentIndex]

	// Show position in the chain
	position := fmt.Sprintf("passthrough %d of %d", componentIndex+1, len(v.PassthroughComponents))

	return fmt.Sprintf(
		"Component '%s' passes prop '%s' without using it (%s). This prop originates from %s and is drilled through %d levels. %s",
		comp.Name,
		v.PropName,
		position,
		v.Origin.Component,
		v.Depth,
		v.Recommendation,
	)
}
