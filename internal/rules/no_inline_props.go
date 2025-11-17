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

	filePath := ast.FilePath

	// Walk the AST to find JSX elements
	ast.Root.Walk(func(node *parser.Node) bool {
		// Only process JSX elements
		if node.Type() != "jsx_element" && node.Type() != "jsx_self_closing_element" {
			return true
		}

		// Get the opening element (contains attributes/props)
		openingElement := r.getOpeningElement(node)
		if openingElement == nil {
			return true
		}

		// Walk through attributes to find inline values
		openingElement.Walk(func(attrNode *parser.Node) bool {
			if attrNode.Type() != "jsx_attribute" {
				return true
			}

			// Get the prop name
			propName := r.getPropName(attrNode)
			if propName == "" {
				return true
			}

			// Get the prop value
			propValue := r.getPropValue(attrNode)
			if propValue == nil {
				return true // Boolean prop like <Input disabled />
			}

			// Check if the value is an inline object/array/function
			if IsUnstableValue(propValue) {
				valueType := r.getValueTypeName(propValue.Type())
				suggestion := r.getSuggestion(valueType)

				line, col := attrNode.StartPoint()

				issues = append(issues, Issue{
					Rule: r.Name(),
					Message: fmt.Sprintf(
						"Prop '%s' receives an inline %s, creating a new reference every render. Extract to a constant or use %s.",
						propName,
						valueType,
						suggestion,
					),
					FilePath: filePath,
					Line:     line + 1,
					Column:   col + 1,
				})
			}

			return true
		})

		return true
	})

	return issues
}

// getOpeningElement gets the jsx_opening_element from a JSX element
func (r *NoInlineProps) getOpeningElement(node *parser.Node) *parser.Node {
	nodeType := node.Type()

	// Self-closing elements are their own opening element
	if nodeType == "jsx_self_closing_element" {
		return node
	}

	// Regular elements have a jsx_opening_element child
	if nodeType == "jsx_element" {
		for _, child := range node.Children() {
			if child.Type() == "jsx_opening_element" {
				return child
			}
		}
	}

	return nil
}

// getValueTypeName returns a friendly name for the inline value type
func (r *NoInlineProps) getValueTypeName(expressionType string) string {
	switch expressionType {
	case "object":
		return "object"
	case "array":
		return "array"
	case "arrow_function", "function", "function_expression":
		return "function"
	default:
		return "value"
	}
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

// getPropName returns the name of a JSX attribute
func (r *NoInlineProps) getPropName(attrNode *parser.Node) string {
	for _, child := range attrNode.Children() {
		if child.Type() == "property_identifier" {
			return child.Text()
		}
	}
	return ""
}

// getPropValue returns the value expression of a JSX attribute
func (r *NoInlineProps) getPropValue(attrNode *parser.Node) *parser.Node {
	// Find the jsx_expression child
	for _, child := range attrNode.Children() {
		if child.Type() == "jsx_expression" {
			// Get the expression inside the braces
			namedChildren := child.NamedChildren()
			if len(namedChildren) > 0 {
				return namedChildren[0]
			}
		}
	}
	return nil
}
