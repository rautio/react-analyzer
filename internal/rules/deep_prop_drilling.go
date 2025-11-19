package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/config"
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

// CheckGraph performs graph-based prop drilling detection with default config
// This is called after all modules have been parsed and the graph is built
func (r *DeepPropDrilling) CheckGraph(g *graph.Graph) []Issue {
	// Use default threshold of 2
	return r.CheckGraphWithConfig(g, nil)
}

// CheckGraphWithConfig performs graph-based prop drilling detection with configuration
func (r *DeepPropDrilling) CheckGraphWithConfig(g *graph.Graph, cfg *config.Config) []Issue {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Get maxDepth from config and convert to minPassthroughComponents
	options := cfg.GetDeepPropDrillingOptions()
	// maxDepth of N means we allow chains up to N components
	// which means we warn when we have (N-1) or more passthrough components
	// Example: maxDepth=3 (App→Dashboard→Sidebar) allows 1 passthrough, warns on 2+
	minPassthrough := options.MaxDepth - 1

	// Detect violations with configured threshold
	violations := graph.DetectPropDrilling(g, minPassthrough)

	var issues []Issue
	for _, v := range violations {
		// Build related information for the full chain
		relatedInfo := r.buildRelatedInformation(v)

		// Create issue at origin (where state is defined)
		issues = append(issues, Issue{
			Rule:     r.Name(),
			FilePath: v.Origin.FilePath,
			Line:     v.Origin.Line,
			Column:   v.Origin.Column,
			Message:  r.formatOriginMessage(v),
			Related:  relatedInfo,
		})

		// Create issues at each passthrough component
		for i, comp := range v.PassthroughComponents {
			issues = append(issues, Issue{
				Rule:     r.Name(),
				FilePath: comp.FilePath,
				Line:     comp.Line,
				Column:   0, // Start of line
				Message:  r.formatPassthroughMessage(v, i),
				Related:  relatedInfo,
			})
		}
	}

	return issues
}

// formatOriginMessage creates message for the state origin location
func (r *DeepPropDrilling) formatOriginMessage(v graph.PropDrillingViolation) string {
	// Build complete component chain: Origin → Passthroughs → Consumer
	chain := v.Origin.Component

	for _, comp := range v.PassthroughComponents {
		chain += " → " + comp.Name
	}

	chain += " → " + v.Consumer.Component

	return fmt.Sprintf(
		"State '%s' is drilled through %d component levels (%s). %s",
		v.PropName,
		v.Depth,
		chain,
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

// buildRelatedInformation creates the full chain of related locations
func (r *DeepPropDrilling) buildRelatedInformation(v graph.PropDrillingViolation) []RelatedInformation {
	var related []RelatedInformation

	// Total depth is the number of components in the chain
	totalComponents := v.Depth

	// Add origin location (component 1)
	related = append(related, RelatedInformation{
		FilePath: v.Origin.FilePath,
		Line:     v.Origin.Line,
		Column:   v.Origin.Column,
		Message:  fmt.Sprintf("Component 1 of %d: '%s' defines state '%s'", totalComponents, v.Origin.Component, v.PropName),
	})

	// Add each passthrough component (components 2 through N-1)
	for i, comp := range v.PassthroughComponents {
		position := i + 2 // Component position in chain (1-indexed, +2 because origin is 1)
		related = append(related, RelatedInformation{
			FilePath: comp.FilePath,
			Line:     comp.Line,
			Column:   0,
			Message:  fmt.Sprintf("Component %d of %d: '%s' passes prop without using it", position, totalComponents, comp.Name),
		})
	}

	// Add consumer location (final component)
	related = append(related, RelatedInformation{
		FilePath: v.Consumer.FilePath,
		Line:     v.Consumer.Line,
		Column:   v.Consumer.Column,
		Message:  fmt.Sprintf("Component %d of %d: '%s' consumes prop '%s'", totalComponents, totalComponents, v.Consumer.Component, v.PropName),
	})

	return related
}
