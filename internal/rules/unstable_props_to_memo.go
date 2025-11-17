package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// UnstablePropsToMemo detects when unstable props are passed to memoized contexts:
// 1. Parent passes unstable props to React.memo components
// 2. useMemo/useCallback depend on unstable props from parent
// This breaks memoization and causes unnecessary re-renders
type UnstablePropsToMemo struct{}

// Name returns the rule identifier
func (r *UnstablePropsToMemo) Name() string {
	return "unstable-props-to-memo"
}

// Check analyzes an AST for unstable props breaking memoization
func (r *UnstablePropsToMemo) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Skip if resolver is not available (shouldn't happen, but safety check)
	if resolver == nil {
		return nil
	}

	// Detection 1: React.memo components with unstable props
	issues = append(issues, r.checkMemoComponentProps(ast, resolver)...)

	// Detection 2: useMemo/useCallback with unstable prop dependencies
	issues = append(issues, r.checkMemoHookDeps(ast, resolver)...)

	return issues
}

// checkMemoComponentProps finds JSX elements rendering memoized components with unstable props
func (r *UnstablePropsToMemo) checkMemoComponentProps(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

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

// checkMemoHookDeps finds useMemo/useCallback hooks with unstable prop dependencies
func (r *UnstablePropsToMemo) checkMemoHookDeps(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Walk the AST to find function declarations (potential React components)
	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()
		if nodeType != "function_declaration" && nodeType != "arrow_function" && nodeType != "function" {
			return true
		}

		// Only analyze React components (functions that return JSX)
		// For now, we'll check all functions - can optimize later

		// Find useMemo/useCallback calls in this function
		node.Walk(func(hookNode *parser.Node) bool {
			if !hookNode.IsHookCall() {
				return true
			}

			// Check if it's useMemo or useCallback
			callee := r.getCalleeNode(hookNode)
			if callee == nil {
				return true
			}

			hookName := callee.Text()
			if hookName != "useMemo" && hookName != "useCallback" {
				return true
			}

			// Get dependency array
			deps := hookNode.GetDependencyArray()
			if deps == nil {
				return true
			}

			// Check each dependency to see if it's a prop
			for _, dep := range deps.GetArrayElements() {
				depName := dep.Text()

				// Check if this dependency is a prop (function parameter)
				if !IsPropIdentifier(depName, node) {
					continue // Not a prop, skip
				}

				// Now we need to check if the parent component passes an unstable value to this prop
				// This requires finding where this component is used and checking the prop value
				// For now, we'll report a warning that a prop is being used in memo deps
				// In the future, we can do full cross-file analysis like React.memo

				line, col := dep.StartPoint()
				issues = append(issues, Issue{
					Rule:     r.Name(),
					Message:  fmt.Sprintf("%s depends on prop '%s' which may be unstable from parent", hookName, depName),
					FilePath: ast.FilePath,
					Line:     line + 1,
					Column:   col,
				})
			}

			return true
		})

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
func (r *UnstablePropsToMemo) getOpeningElement(node *parser.Node) *parser.Node {
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
func (r *UnstablePropsToMemo) getComponentName(openingElement *parser.Node) string {
	for _, child := range openingElement.Children() {
		nodeType := child.Type()
		if nodeType == "identifier" || nodeType == "jsx_identifier" {
			return child.Text()
		}
	}
	return ""
}

// isComponentMemoized checks if a component is wrapped in React.memo
func (r *UnstablePropsToMemo) isComponentMemoized(componentName string, currentFile string, resolver *analyzer.ModuleResolver) (bool, error) {
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
func (r *UnstablePropsToMemo) findUnstableProps(openingElement *parser.Node) []unstableProp {
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

// getCalleeNode gets the function being called from a call_expression
func (r *UnstablePropsToMemo) getCalleeNode(callNode *parser.Node) *parser.Node {
	if callNode.Type() != "call_expression" {
		return nil
	}

	return callNode.ChildByFieldName("function")
}
