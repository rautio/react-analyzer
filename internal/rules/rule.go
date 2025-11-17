package rules

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// Issue represents a single rule violation
type Issue struct {
	Rule     string // Rule identifier (e.g., "no-object-deps")
	Message  string // Human-readable message
	FilePath string // Path to the file
	Line     uint32 // Line number (1-indexed)
	Column   uint32 // Column number (0-indexed)
}

// Rule defines the interface for all analysis rules
type Rule interface {
	// Name returns the rule identifier
	Name() string

	// Check analyzes an AST and returns any issues found
	// The resolver parameter enables cross-file analysis (can be nil for single-file rules)
	Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue
}
