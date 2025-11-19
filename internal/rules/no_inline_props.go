package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// NoInlineProps detects when JSX props are set to inline objects, arrays, or functions
// that are created on every render, causing unnecessary re-renders and breaking memoization.
//
// Bad:
//
//	<Component config={{ theme: 'dark' }} />  // New object every render
//	<List items={[1, 2, 3]} />                // New array every render
//	<Button onClick={() => click()} />        // New function every render
//
// Good:
//
//	const CONFIG = { theme: 'dark' };
//	<Component config={CONFIG} />             // Same reference every render
type NoInlineProps struct{}

// Name returns the rule identifier
func (r *NoInlineProps) Name() string {
	return "no-inline-props"
}

// Check analyzes the AST for inline prop values
func (r *NoInlineProps) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Acquire global tree-sitter lock before walking ASTs
	// Tree-sitter C library is not thread-safe - must serialize all AST operations
	if resolver != nil {
		resolver.LockTreeSitter()
		defer resolver.UnlockTreeSitter()
	}

	// Single-pass walk looking for jsx_attribute nodes
	ast.Root.Walk(func(node *parser.Node) bool {
		// Only process JSX attributes
		if node.Type() != "jsx_attribute" {
			return true
		}

		// Get the prop name using shared helper
		propName := GetPropName(node)
		if propName == "" {
			return true
		}

		// Get the prop value using shared helper
		propValue := GetPropValue(node)
		if propValue == nil {
			return true // Boolean prop like <Input disabled />
		}

		// Check if the value is an inline object/array/function
		if !IsUnstableValue(propValue) {
			return true
		}

		// Found a violation - determine type name for message
		valueType := "value"
		switch propValue.Type() {
		case "object":
			valueType = "object"
		case "array":
			valueType = "array"
		case "arrow_function", "function", "function_expression":
			valueType = "function"
		}

		suggestion := r.getSuggestion(valueType)
		line, col := node.StartPoint()

		issues = append(issues, Issue{
			Rule: r.Name(),
			Message: fmt.Sprintf(
				"Prop '%s' receives an inline %s, creating a new reference every render. Extract to a constant or use %s.",
				propName,
				valueType,
				suggestion,
			),
			FilePath: ast.FilePath,
			Line:     line + 1,
			Column:   col + 1,
		})

		return true
	})

	return issues
}

// getSuggestion returns the appropriate fix suggestion based on value type
func (r *NoInlineProps) getSuggestion(valueType string) string {
	switch valueType {
	case "function":
		return "useCallback"
	case "object", "array":
		return "useMemo"
	default:
		return "useMemo"
	}
}
