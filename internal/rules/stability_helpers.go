package rules

import (
	"github.com/rautio/react-analyzer/internal/parser"
)

// IsUnstableValue checks if a value is an inline object, array, or function
// These values are created fresh on every render and break memoization
func IsUnstableValue(node *parser.Node) bool {
	if node == nil {
		return false
	}

	nodeType := node.Type()

	// Inline objects: { key: value }
	if nodeType == "object" {
		return true
	}

	// Inline arrays: [1, 2, 3]
	if nodeType == "array" {
		return true
	}

	// Inline functions: () => {}, function() {}
	if nodeType == "arrow_function" || nodeType == "function" || nodeType == "function_expression" {
		return true
	}

	return false
}

// IsStabilizedValue checks if a value is wrapped in useMemo, useCallback, or is a constant
// These values maintain referential equality across renders
func IsStabilizedValue(node *parser.Node) bool {
	if node == nil {
		return false
	}

	// Check if it's a call to useMemo or useCallback
	if node.Type() == "call_expression" {
		callee := node.ChildByFieldName("function")
		if callee != nil {
			calleeName := callee.Text()
			if calleeName == "useMemo" || calleeName == "useCallback" {
				return true
			}
		}
	}

	// Check if it's a reference to an identifier (could be const, prop, etc.)
	// This is conservative - we assume identifiers might be stable
	// The caller should do deeper analysis if needed
	if node.Type() == "identifier" {
		return false // Caller should analyze further
	}

	return false
}

// GetPropName extracts the prop name from a JSX attribute node
func GetPropName(attrNode *parser.Node) string {
	if attrNode == nil || attrNode.Type() != "jsx_attribute" {
		return ""
	}

	nameNode := attrNode.ChildByFieldName("name")
	if nameNode != nil {
		return nameNode.Text()
	}

	return ""
}

// GetPropValue extracts the value node from a JSX attribute
func GetPropValue(attrNode *parser.Node) *parser.Node {
	if attrNode == nil || attrNode.Type() != "jsx_attribute" {
		return nil
	}

	// Check for jsx_expression value: prop={value}
	valueNode := attrNode.ChildByFieldName("value")
	if valueNode != nil && valueNode.Type() == "jsx_expression" {
		// Get the expression inside the braces
		for _, child := range valueNode.NamedChildren() {
			return child
		}
	}

	return valueNode
}

// IsPropIdentifier checks if a dependency is a prop parameter
// Returns true if the identifier matches a function parameter name
func IsPropIdentifier(depName string, funcNode *parser.Node) bool {
	if funcNode == nil {
		return false
	}

	// Get function parameters
	var paramsNode *parser.Node

	// Handle different function types
	switch funcNode.Type() {
	case "function_declaration", "function":
		paramsNode = funcNode.ChildByFieldName("parameters")
	case "arrow_function":
		paramsNode = funcNode.ChildByFieldName("parameters")
		// Arrow functions might have a single parameter without parens
		if paramsNode == nil {
			param := funcNode.ChildByFieldName("parameter")
			if param != nil && param.Text() == depName {
				return true
			}
		}
	default:
		return false
	}

	if paramsNode == nil {
		return false
	}

	// Check if depName matches any parameter
	// Handle both regular params and destructured params
	for _, param := range paramsNode.NamedChildren() {
		switch param.Type() {
		case "identifier":
			if param.Text() == depName {
				return true
			}
		case "object_pattern":
			// Destructured props: { config, onUpdate }
			for _, prop := range param.NamedChildren() {
				if prop.Type() == "shorthand_property_identifier_pattern" {
					if prop.Text() == depName {
						return true
					}
				} else if prop.Type() == "pair_pattern" {
					// { config: configProp } format
					valueNode := prop.ChildByFieldName("value")
					if valueNode != nil && valueNode.Text() == depName {
						return true
					}
				}
			}
		}
	}

	return false
}

// GetComponentName extracts the component name from a function node
// Note: For arrow functions, this requires the parent context to be passed separately
// since tree-sitter nodes don't have a Parent() method in our wrapper
func GetComponentName(funcNode *parser.Node) string {
	if funcNode == nil {
		return ""
	}

	// function_declaration has a name field
	if funcNode.Type() == "function_declaration" {
		nameNode := funcNode.ChildByFieldName("name")
		if nameNode != nil {
			return nameNode.Text()
		}
	}

	// For arrow functions assigned to variables (const Component = () => {}),
	// we'd need the parent node context. Since we don't have access to parent
	// in the current Node wrapper, callers should handle this case themselves
	// or we can enhance this later with a Walk-based approach

	return ""
}
