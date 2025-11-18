package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// NoStaleState detects when state setters update state based on previous value
// without using the functional form, which causes race conditions and stale closures
type NoStaleState struct{}

// Name returns the rule identifier
func (r *NoStaleState) Name() string {
	return "no-stale-state"
}

// Check analyzes an AST for stale state anti-patterns
func (r *NoStaleState) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Step 1: Find all useState declarations and build state mapping
	stateMap := r.findUseStateDeclarations(ast.Root)

	// Step 2: Find all setState calls and check for violations
	ast.Root.Walk(func(node *parser.Node) bool {
		if node.Type() != "call_expression" {
			return true
		}

		// Get the function being called
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			return true
		}

		functionName := funcNode.Text()

		// Is this a setState call?
		if stateInfo, found := stateMap[functionName]; found {
			// Get the argument to setState
			argsNode := node.ChildByFieldName("arguments")
			if argsNode == nil {
				return true
			}

			argChildren := argsNode.NamedChildren()
			if len(argChildren) == 0 {
				return true
			}

			argument := argChildren[0]

			// Step 3: Check if it's already in functional form
			if r.isFunctionalForm(argument) {
				return true // Already correct
			}

			// Step 4: Check if argument references the state variable
			if r.argumentReferencesState(argument, stateInfo.StateName) {
				// Found a violation!
				line, col := node.StartPoint()

				issues = append(issues, Issue{
					Rule: r.Name(),
					Message: fmt.Sprintf(
						"State '%s' is updated using its current value without the functional form, which can cause race conditions. Replace with: %s(prev => %s)",
						stateInfo.StateName,
						stateInfo.SetterName,
						r.generateSuggestion(argument, stateInfo.StateName),
					),
					FilePath: ast.FilePath,
					Line:     line + 1,
					Column:   col,
				})
			}
		}

		return true
	})

	return issues
}

// StateInfo holds information about a useState declaration
type StateInfo struct {
	StateName  string
	SetterName string
	Line       uint32
}

// findUseStateDeclarations finds all useState calls and extracts state/setter names
func (r *NoStaleState) findUseStateDeclarations(root *parser.Node) map[string]StateInfo {
	stateMap := make(map[string]StateInfo)

	root.Walk(func(node *parser.Node) bool {
		// Look for: const [state, setState] = useState(initialValue)
		if node.Type() != "variable_declarator" {
			return true
		}

		// Check if this is useState
		valueNode := node.ChildByFieldName("value")
		if valueNode == nil || valueNode.Type() != "call_expression" {
			return true
		}

		// Check if the function being called is useState (handles both bare and namespaced hooks)
		if valueNode.GetHookName() != "useState" {
			return true
		}

		// Get the destructured names [state, setState]
		nameNode := node.ChildByFieldName("name")
		if nameNode == nil || nameNode.Type() != "array_pattern" {
			return true
		}

		// Extract state name and setter name
		children := nameNode.NamedChildren()
		if len(children) < 2 {
			return true
		}

		stateName := children[0].Text()
		setterName := children[1].Text()

		line, _ := node.StartPoint()

		// Store mapping: setterName -> StateInfo
		stateMap[setterName] = StateInfo{
			StateName:  stateName,
			SetterName: setterName,
			Line:       line + 1,
		}

		return true
	})

	return stateMap
}

// isFunctionalForm checks if the argument is already an arrow function or function
func (r *NoStaleState) isFunctionalForm(node *parser.Node) bool {
	nodeType := node.Type()
	return nodeType == "arrow_function" || nodeType == "function"
}

// argumentReferencesState checks if the expression references the state variable
func (r *NoStaleState) argumentReferencesState(node *parser.Node, stateName string) bool {
	if node == nil {
		return false
	}

	// Check current node
	if node.Type() == "identifier" && node.Text() == stateName {
		return true
	}

	// Recursively check all children
	for _, child := range node.NamedChildren() {
		if r.argumentReferencesState(child, stateName) {
			return true
		}
	}

	return false
}

// generateSuggestion creates the functional form suggestion
// Replaces all references to the state variable with 'prev'
func (r *NoStaleState) generateSuggestion(expr *parser.Node, stateName string) string {
	return r.replaceStateWithPrev(expr, stateName)
}

// replaceStateWithPrev recursively replaces state variable references with 'prev'
func (r *NoStaleState) replaceStateWithPrev(node *parser.Node, stateName string) string {
	if node == nil {
		return ""
	}

	// If this is an identifier matching the state name, replace it
	if node.Type() == "identifier" && node.Text() == stateName {
		return "prev"
	}

	// For nodes with no children, return the text as-is
	children := node.Children()
	if len(children) == 0 {
		return node.Text()
	}

	// Recursively process children
	var result string
	for _, child := range children {
		result += r.replaceStateWithPrev(child, stateName)
	}

	return result
}
