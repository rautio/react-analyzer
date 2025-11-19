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
// TODO: This function is partially implemented and needs additional work:
//   - Currently only detects function declarations (function Foo() {}), not arrow functions (const Foo = () => {})
//   - Local detection logic exists but needs debugging
//   - Cross-file detection stub exists but is not implemented
//   - Full implementation blocked on AST traversal enhancements for arrow functions in variable declarations
//   - See test fixtures: LocalUseMemoViolation.tsx, ChildWithUseMemo.tsx, ParentWithUseMemoChild.tsx
func (r *UnstablePropsToMemo) checkMemoHookDeps(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	// TODO: Remove this early return once implementation is complete
	// For now, we skip this detection to avoid false negatives
	return nil

	// TODO: Uncomment and debug the implementation below
	/*
		var issues []Issue

		// Get the module for this file
		module, err := resolver.GetModule(ast.FilePath)
		if err != nil {
			return nil
		}

		// Ensure symbols are analyzed
		analyzer.AnalyzeSymbols(module)

		// Map to store which props are used in memo hooks: componentName -> propNames
		componentMemoDeps := make(map[string]map[string][]memoDependency)

		// Walk the AST to find function declarations (potential React components)
		ast.Root.Walk(func(node *parser.Node) bool {
			nodeType := node.Type()
			if nodeType != "function_declaration" && nodeType != "arrow_function" && nodeType != "function" {
				return true
			}

			// Get component name
			componentName := r.getComponentNameFromFunction(node)
			if componentName == "" {
				return true
			}

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

					// Store this memo dependency for cross-file analysis
					if componentMemoDeps[componentName] == nil {
						componentMemoDeps[componentName] = make(map[string][]memoDependency)
					}

					line, col := dep.StartPoint()
					componentMemoDeps[componentName][depName] = append(
						componentMemoDeps[componentName][depName],
						memoDependency{
							hookName: hookName,
							line:     line + 1,
							column:   col,
						},
					)
				}

				return true
			})

			return true
		})

		// Now do cross-file analysis: find where these components are used
		// and check if parents pass unstable values to the props
		for componentName, propDeps := range componentMemoDeps {
			// Check if this component is exported
			symbol, exists := module.Symbols[componentName]
			if !exists || !symbol.IsExported {
				// Not exported, so it's only used locally in this file
				// Check for local usage
				issues = append(issues, r.checkLocalUsage(ast, componentName, propDeps)...)
				continue
			}

			// Component is exported - find files that import it
			issues = append(issues, r.checkCrossFileUsage(ast.FilePath, componentName, propDeps, resolver)...)
		}

		return issues
	*/
}

// memoDependency represents a prop used in useMemo/useCallback
type memoDependency struct {
	hookName string
	line     uint32
	column   uint32
}

// checkLocalUsage checks if a component is used locally with unstable props
func (r *UnstablePropsToMemo) checkLocalUsage(ast *parser.AST, componentName string, propDeps map[string][]memoDependency) []Issue {
	var issues []Issue

	// Walk AST to find JSX usage of this component
	ast.Root.Walk(func(node *parser.Node) bool {
		if node.Type() != "jsx_element" && node.Type() != "jsx_self_closing_element" {
			return true
		}

		openingElement := r.getOpeningElement(node)
		if openingElement == nil {
			return true
		}

		usedComponentName := r.getComponentName(openingElement)
		if usedComponentName != componentName {
			return true
		}

		// Found usage - check if any of the props are unstable
		for propName, deps := range propDeps {
			propValue := r.getPropValueByName(openingElement, propName)
			if propValue == nil {
				continue
			}

			// Check if the prop value is unstable
			if IsUnstableValue(propValue) {
				for _, dep := range deps {
					propLine, propCol := propValue.StartPoint()
					issues = append(issues, Issue{
						Rule:     r.Name(),
						Message:  fmt.Sprintf("Passing unstable %s to prop '%s' breaks %s in '%s'", r.getValueType(propValue), propName, dep.hookName, componentName),
						FilePath: ast.FilePath,
						Line:     propLine + 1,
						Column:   propCol,
					})
				}
			}
		}

		return true
	})

	return issues
}

// checkCrossFileUsage checks if a component is used in other files with unstable props
func (r *UnstablePropsToMemo) checkCrossFileUsage(currentFile, componentName string, propDeps map[string][]memoDependency, resolver *analyzer.ModuleResolver) []Issue {
	// TODO: Implement full cross-file analysis with reverse import lookup
	// This requires:
	//   1. Building a reverse import index (which files import this component?)
	//   2. For each importing file, finding JSX usage of the component
	//   3. Checking if unstable values are passed to the props that are used in memo hooks
	//   4. Similar pattern to React.memo detection but in reverse direction
	//
	// Complexity: Medium-High
	// Benefit: Detects cross-file useMemo/useCallback violations (similar to React.memo)
	//
	// For now, local usage detection (same file) is implemented above

	return nil
}

// getComponentNameFromFunction extracts component name from a function node
func (r *UnstablePropsToMemo) getComponentNameFromFunction(funcNode *parser.Node) string {
	if funcNode.Type() == "function_declaration" {
		nameNode := funcNode.ChildByFieldName("name")
		if nameNode != nil {
			return nameNode.Text()
		}
	}

	// For arrow functions, we'd need parent context
	// This is a limitation we can address later
	return ""
}

// getPropValueByName finds a prop's value by name in a JSX opening element
func (r *UnstablePropsToMemo) getPropValueByName(openingElement *parser.Node, propName string) *parser.Node {
	var result *parser.Node

	openingElement.Walk(func(node *parser.Node) bool {
		if node.Type() != "jsx_attribute" {
			return true
		}

		// Check if this is the prop we're looking for
		nameNode := node.ChildByFieldName("name")
		if nameNode == nil || nameNode.Text() != propName {
			return true
		}

		// Get the value
		for _, child := range node.Children() {
			if child.Type() == "jsx_expression" {
				// Get the expression inside the braces
				for _, exprChild := range child.Children() {
					if exprChild.Type() != "{" && exprChild.Type() != "}" {
						result = exprChild
						return false // Stop walking
					}
				}
			}
		}

		return false // Stop walking
	})

	return result
}

// getValueType returns a human-readable type name for an unstable value
func (r *UnstablePropsToMemo) getValueType(node *parser.Node) string {
	switch node.Type() {
	case "object":
		return "object"
	case "array":
		return "array"
	case "arrow_function", "function":
		return "function"
	default:
		return "value"
	}
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
	var importedName string // The actual name exported from the module
	for _, imp := range module.Imports {
		// Check default import
		if imp.Default == componentName {
			importSource = imp.Source
			importedName = componentName
			break
		}
		// Check named imports (use LocalName for lookup, ImportedName for target module)
		for _, named := range imp.Named {
			if named.LocalName == componentName {
				importSource = imp.Source
				importedName = named.ImportedName
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

	// Note: GetModule() already calls AnalyzeSymbols(), so symbols are populated

	// Check if the component is memoized in the target module
	// Use the imported name (not the local alias) to look up in the target module
	if symbol, exists := targetModule.Symbols[importedName]; exists {
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
