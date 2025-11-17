package rules

import (
	"fmt"

	"github.com/oskari/react-analyzer/internal/analyzer"
	"github.com/oskari/react-analyzer/internal/parser"
)

// MemoizedComponentUnstableProps detects when a parent passes unstable props to a memoized child component
// This is anti-pattern #5 from the catalog: React.memo is useless when parent passes unstable props
type MemoizedComponentUnstableProps struct{}

// Name returns the rule identifier
func (r *MemoizedComponentUnstableProps) Name() string {
	return "memoized-component-unstable-props"
}

// Check analyzes an AST for inline props passed to memoized components
func (r *MemoizedComponentUnstableProps) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Skip if resolver is not available (shouldn't happen, but safety check)
	if resolver == nil {
		return nil
	}

	// Walk the AST to find JSX elements
	ast.Root.Walk(func(node *parser.Node) bool {
		if node.Type() != "jsx_element" && node.Type() != "jsx_self_closing_element" {
			return true
		}

		// Get the JSX opening element
		openingElement := r.getOpeningElement(node)
		if openingElement == nil {
			return true
		}

		// Get the component name being rendered
		componentName := r.getComponentName(openingElement)
		if componentName == "" {
			return true
		}

		// Check if this component is memoized
		isMemoized, err := r.isComponentMemoized(componentName, ast.FilePath, resolver)
		if err != nil || !isMemoized {
			return true
		}

		// Component is memoized - check for unstable props
		unstableProps := r.findUnstableProps(openingElement)
		for _, prop := range unstableProps {
			issues = append(issues, Issue{
				Rule:     r.Name(),
				Message:  fmt.Sprintf("Passing unstable %s to memoized component '%s' breaks memoization", prop.propType, componentName),
				FilePath: ast.FilePath,
				Line:     prop.line,
				Column:   prop.column,
			})
		}

		return true
	})

	return issues
}

// unstableProp represents a prop that's unstable (inline object/array/function)
type unstableProp struct {
	propType string // "object", "array", "function"
	line     uint32
	column   uint32
}

// getOpeningElement gets the jsx_opening_element from a JSX element
func (r *MemoizedComponentUnstableProps) getOpeningElement(node *parser.Node) *parser.Node {
	nodeType := node.Type()

	if nodeType == "jsx_self_closing_element" {
		return node
	}

	if nodeType == "jsx_element" {
		for _, child := range node.Children() {
			if child.Type() == "jsx_opening_element" {
				return child
			}
		}
	}

	return nil
}

// getComponentName extracts the component name from a JSX opening element
func (r *MemoizedComponentUnstableProps) getComponentName(openingElement *parser.Node) string {
	for _, child := range openingElement.Children() {
		nodeType := child.Type()
		if nodeType == "identifier" || nodeType == "jsx_identifier" {
			return child.Text()
		}
	}
	return ""
}

// isComponentMemoized checks if a component is wrapped in React.memo
func (r *MemoizedComponentUnstableProps) isComponentMemoized(componentName string, currentFile string, resolver *analyzer.ModuleResolver) (bool, error) {
	// Load the current module to check imports
	module, err := resolver.GetModule(currentFile)
	if err != nil {
		return false, err
	}

	// First check if it's defined locally in the same file
	if symbol, exists := module.Symbols[componentName]; exists {
		return symbol.IsMemoized, nil
	}

	// Not local - check if it's imported
	var importSource string
	for _, imp := range module.Imports {
		// Check default import
		if imp.Default == componentName {
			importSource = imp.Source
			break
		}
		// Check named imports
		for _, named := range imp.Named {
			if named == componentName {
				importSource = imp.Source
				break
			}
		}
		if importSource != "" {
			break
		}
	}

	// If not imported, it's not memoized
	if importSource == "" {
		return false, nil
	}

	// Resolve the import path
	targetPath, err := resolver.Resolve(currentFile, importSource)
	if err != nil {
		// Can't resolve (external package or missing file) - assume not memoized
		return false, nil
	}

	// Load the target module
	targetModule, err := resolver.GetModule(targetPath)
	if err != nil {
		return false, err
	}

	// Analyze symbols if not already done
	analyzer.AnalyzeSymbols(targetModule)

	// Check if the component is memoized in the target module
	if symbol, exists := targetModule.Symbols[componentName]; exists {
		return symbol.IsMemoized, nil
	}

	return false, nil
}

// findUnstableProps finds all props that are inline objects/arrays/functions
func (r *MemoizedComponentUnstableProps) findUnstableProps(openingElement *parser.Node) []unstableProp {
	var unstableProps []unstableProp

	// Walk through the opening element to find jsx_attribute nodes
	openingElement.Walk(func(node *parser.Node) bool {
		if node.Type() != "jsx_attribute" {
			return true
		}

		// Get the value of this attribute
		for _, child := range node.Children() {
			if child.Type() == "jsx_expression" {
				// Check what's inside the expression
				for _, exprChild := range child.Children() {
					propType := ""
					switch exprChild.Type() {
					case "object":
						propType = "object"
					case "array":
						propType = "array"
					case "arrow_function":
						propType = "function"
					case "function":
						propType = "function"
					}

					if propType != "" {
						row, col := exprChild.StartPoint()
						unstableProps = append(unstableProps, unstableProp{
							propType: propType,
							line:     row + 1,
							column:   col,
						})
					}
				}
			}
		}

		return true
	})

	return unstableProps
}
