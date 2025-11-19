package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// NoDerivedState detects when useState is used to mirror props and sync via useEffect
// This causes unnecessary re-renders and complexity
type NoDerivedState struct{}

// Name returns the rule identifier
func (r *NoDerivedState) Name() string {
	return "no-derived-state"
}

// Check analyzes an AST for derived state anti-patterns
func (r *NoDerivedState) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Acquire global tree-sitter lock before walking ASTs
	// Tree-sitter C library is not thread-safe - must serialize all AST operations
	if resolver != nil {
		resolver.LockTreeSitter()
		defer resolver.UnlockTreeSitter()
	}

	// Walk the AST to find function declarations (potential React components)
	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()
		if nodeType != "function_declaration" && nodeType != "arrow_function" && nodeType != "function" {
			return true
		}

		// Extract props from this component
		props := r.extractProps(node)
		if len(props) == 0 {
			return true // No props, can't have derived state from props
		}

		// Find all useState declarations in this component
		states := r.findUseStateDeclarations(node, node)
		if len(states) == 0 {
			return true // No state, nothing to check
		}

		// Find all useEffect calls in this component
		effects := r.findUseEffects(node)
		if len(effects) == 0 {
			return true // No effects, can't be syncing state
		}

		// Match useState + useEffect pairs
		violations := r.findViolations(props, states, effects, ast.FilePath)
		issues = append(issues, violations...)

		return true
	})

	return issues
}

// StateDeclaration represents a useState call
type StateDeclaration struct {
	StateName     string
	SetterName    string
	Initializer   *parser.Node
	ComponentNode *parser.Node // The component this state belongs to
	Line          uint32
	Column        uint32
}

// EffectDeclaration represents a useEffect call
type EffectDeclaration struct {
	Body *parser.Node
	Deps []string
	Line uint32
}

// extractProps extracts prop names from function parameters
func (r *NoDerivedState) extractProps(funcNode *parser.Node) []string {
	var props []string

	// Get parameters node
	paramsNode := funcNode.ChildByFieldName("parameters")
	if paramsNode == nil {
		// Try "parameter" for single-param arrow functions
		paramNode := funcNode.ChildByFieldName("parameter")
		if paramNode != nil {
			paramsNode = paramNode
		} else {
			return props
		}
	}

	// Walk parameters to find prop names
	paramsNode.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Destructured props: ({ user, isAdmin })
		if nodeType == "shorthand_property_identifier_pattern" {
			props = append(props, node.Text())
		}

		// Named props: (props) - we'd need to track props.user
		// For Phase 1, we'll just handle destructured props
		if nodeType == "identifier" {
			// This is the parameter name itself (e.g., "props")
			// We'll track this for Phase 2
			props = append(props, node.Text())
		}

		return true
	})

	return props
}

// findUseStateDeclarations finds all useState calls in a component
func (r *NoDerivedState) findUseStateDeclarations(componentNode *parser.Node, ownerNode *parser.Node) []StateDeclaration {
	var states []StateDeclaration

	componentNode.Walk(func(node *parser.Node) bool {
		// Look for: const [state, setState] = useState(initialValue)
		if node.Type() != "variable_declarator" {
			return true
		}

		// Check if this is useState
		valueNode := node.ChildByFieldName("value")
		if valueNode == nil || valueNode.Type() != "call_expression" {
			return true
		}

		// Check if the function being called is useState
		funcNode := valueNode.ChildByFieldName("function")
		if funcNode == nil || funcNode.Text() != "useState" {
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

		// Get initializer (argument to useState)
		args := valueNode.ChildByFieldName("arguments")
		if args == nil {
			return true
		}

		// Get first argument
		argChildren := args.NamedChildren()
		var initializer *parser.Node
		if len(argChildren) > 0 {
			initializer = argChildren[0]
		}

		line, col := node.StartPoint()
		states = append(states, StateDeclaration{
			StateName:     stateName,
			SetterName:    setterName,
			Initializer:   initializer,
			ComponentNode: ownerNode,
			Line:          line + 1,
			Column:        col,
		})

		return true
	})

	return states
}

// findUseEffects finds all useEffect calls in a component
func (r *NoDerivedState) findUseEffects(componentNode *parser.Node) []EffectDeclaration {
	var effects []EffectDeclaration

	componentNode.Walk(func(node *parser.Node) bool {
		if !node.IsHookCall() {
			return true
		}

		// Check if it's useEffect (handles both bare and namespaced hooks)
		if node.GetHookName() != "useEffect" {
			return true
		}

		// Get arguments
		argsNode := node.ChildByFieldName("arguments")
		if argsNode == nil {
			return true
		}

		argChildren := argsNode.NamedChildren()
		if len(argChildren) < 1 {
			return true
		}

		// First argument is the effect callback
		callback := argChildren[0]
		var body *parser.Node

		// Get the body of the callback
		if callback.Type() == "arrow_function" {
			body = callback.ChildByFieldName("body")
		} else if callback.Type() == "function" {
			body = callback.ChildByFieldName("body")
		}

		// Get dependency array (second argument)
		var deps []string
		if len(argChildren) >= 2 {
			depsArray := argChildren[1]
			if depsArray.Type() == "array" {
				for _, depNode := range depsArray.NamedChildren() {
					deps = append(deps, depNode.Text())
				}
			}
		}

		line, _ := node.StartPoint()
		effects = append(effects, EffectDeclaration{
			Body: body,
			Deps: deps,
			Line: line + 1,
		})

		return true
	})

	return effects
}

// findViolations matches useState + useEffect pairs that form the anti-pattern
func (r *NoDerivedState) findViolations(
	props []string,
	states []StateDeclaration,
	effects []EffectDeclaration,
	filePath string,
) []Issue {
	var issues []Issue

	for _, state := range states {
		// Check if initializer references a prop
		propUsed := r.initializerUsesProps(state.Initializer, props)
		if propUsed == "" {
			continue // Initializer doesn't use props
		}

		// Find matching useEffect
		for _, effect := range effects {
			// Check if effect body calls the setter
			if !r.effectCallsSetter(effect.Body, state.SetterName) {
				continue
			}

			// Check if effect depends on the prop or any nested property
			if !r.effectDependsOnProp(effect.Deps, propUsed) {
				continue
			}

			// Phase 3: Check if this is the controlled/uncontrolled hybrid pattern
			// If the setter is used outside the effect (e.g., in event handlers),
			// this is likely intentional (form control pattern)
			if r.isSetterUsedOutsideEffect(state.ComponentNode, state.SetterName, effect.Body) {
				continue // Skip - this is the valid hybrid pattern
			}

			// Found a violation!
			// Use the actual initializer expression for the suggestion
			suggestion := state.Initializer.Text()
			if suggestion == "" {
				suggestion = propUsed // Fallback to prop name
			}

			issues = append(issues, Issue{
				Rule: r.Name(),
				Message: fmt.Sprintf(
					"State '%s' mirrors prop and causes unnecessary re-renders. Replace with: const %s = %s;",
					state.StateName,
					state.StateName,
					suggestion,
				),
				FilePath: filePath,
				Line:     state.Line,
				Column:   state.Column,
			})
		}
	}

	return issues
}

// initializerUsesProps checks if useState initializer references any props
// Returns the prop name if found, empty string otherwise
func (r *NoDerivedState) initializerUsesProps(initializer *parser.Node, props []string) string {
	if initializer == nil {
		return ""
	}

	// Phase 1: Simple case - direct prop reference
	// useState(user) where user is a prop
	if initializer.Type() == "identifier" {
		propName := initializer.Text()
		if contains(props, propName) {
			return propName
		}
	}

	// Phase 2: Nested properties and computed values
	// Check if any child identifier references a prop
	propUsed := r.findPropInExpression(initializer, props)
	if propUsed != "" {
		return propUsed
	}

	return ""
}

// findPropInExpression recursively searches for prop references in complex expressions
func (r *NoDerivedState) findPropInExpression(node *parser.Node, props []string) string {
	if node == nil {
		return ""
	}

	// Check current node
	if node.Type() == "identifier" {
		text := node.Text()
		if contains(props, text) {
			return text
		}
	}

	// Recursively check children
	for _, child := range node.NamedChildren() {
		if propUsed := r.findPropInExpression(child, props); propUsed != "" {
			return propUsed
		}
	}

	return ""
}

// effectCallsSetter checks if the effect body calls the state setter
func (r *NoDerivedState) effectCallsSetter(body *parser.Node, setterName string) bool {
	if body == nil {
		return false
	}

	found := false
	body.Walk(func(node *parser.Node) bool {
		// Look for call_expression
		if node.Type() == "call_expression" {
			funcNode := node.ChildByFieldName("function")
			if funcNode != nil && funcNode.Text() == setterName {
				found = true
				return false // Stop walking
			}
		}
		return true
	})

	return found
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// effectDependsOnProp checks if the effect depends on the prop or any nested property
// e.g., prop "user" matches dependencies ["user"], ["user.name"], ["user.age.value"]
func (r *NoDerivedState) effectDependsOnProp(deps []string, propName string) bool {
	for _, dep := range deps {
		// Exact match: user === user
		if dep == propName {
			return true
		}
		// Nested property: user.name starts with "user."
		if len(dep) > len(propName) && dep[:len(propName)] == propName && dep[len(propName)] == '.' {
			return true
		}
	}
	return false
}

// isSetterUsedOutsideEffect checks if the state setter is called anywhere in the component
// outside of the specific useEffect. This indicates the controlled/uncontrolled hybrid pattern
// (e.g., form controls where user can modify state via event handlers)
func (r *NoDerivedState) isSetterUsedOutsideEffect(componentNode *parser.Node, setterName string, effectBody *parser.Node) bool {
	if componentNode == nil || effectBody == nil {
		return false
	}

	found := false
	componentNode.Walk(func(node *parser.Node) bool {
		// Skip if we're inside the effect body
		if r.isNodeInsideEffect(node, effectBody) {
			return true // Continue walking but skip this subtree
		}

		// Look for call_expression
		if node.Type() == "call_expression" {
			funcNode := node.ChildByFieldName("function")
			if funcNode != nil && funcNode.Text() == setterName {
				found = true
				return false // Stop walking
			}
		}
		return true
	})

	return found
}

// isNodeInsideEffect checks if a node is inside the effect body
func (r *NoDerivedState) isNodeInsideEffect(node *parser.Node, effectBody *parser.Node) bool {
	if node == nil || effectBody == nil {
		return false
	}

	// Get the position ranges
	nodeStartRow, nodeStartCol := node.StartPoint()
	nodeEndRow, nodeEndCol := node.EndPoint()
	effectStartRow, effectStartCol := effectBody.StartPoint()
	effectEndRow, effectEndCol := effectBody.EndPoint()

	// Check if node is within effect body bounds
	// Node starts after effect starts
	nodeStartsAfter := nodeStartRow > effectStartRow || (nodeStartRow == effectStartRow && nodeStartCol >= effectStartCol)
	// Node ends before effect ends
	nodeEndsBefore := nodeEndRow < effectEndRow || (nodeEndRow == effectEndRow && nodeEndCol <= effectEndCol)

	return nodeStartsAfter && nodeEndsBefore
}
