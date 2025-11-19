package rules

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/graph"
	"github.com/rautio/react-analyzer/internal/parser"
)

// Issue represents a single rule violation
type Issue struct {
	Rule     string `json:"rule"`     // Rule identifier (e.g., "no-object-deps")
	Message  string `json:"message"`  // Human-readable message
	FilePath string `json:"filePath"` // Path to the file
	Line     uint32 `json:"line"`     // Line number (1-indexed)
	Column   uint32 `json:"column"`   // Column number (0-indexed)
}

// Rule defines the interface for all analysis rules
type Rule interface {
	// Name returns the rule identifier
	Name() string

	// Check analyzes an AST and returns any issues found
	// The resolver parameter enables cross-file analysis (can be nil for single-file rules)
	Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue
}

// GraphRule defines the interface for rules that operate on the dependency graph
// These rules require the full component graph to be built before analysis
type GraphRule interface {
	Rule // Extends Rule interface

	// CheckGraph analyzes the dependency graph and returns any issues found
	// This is called after all modules have been parsed and the graph is constructed
	CheckGraph(g *graph.Graph) []Issue
}
