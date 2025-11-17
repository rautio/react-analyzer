package rules

import (
	"fmt"

	"github.com/oskari/react-analyzer/internal/analyzer"
	"github.com/oskari/react-analyzer/internal/parser"
)

// NoObjectDeps checks for object/array literals in hook dependency arrays
type NoObjectDeps struct{}

// Name returns the rule identifier
func (r *NoObjectDeps) Name() string {
	return "no-object-deps"
}

// Check analyzes an AST for object/array dependencies in React hooks
func (r *NoObjectDeps) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
	var issues []Issue

	// Walk the AST to find function declarations (potential React components)
	ast.Root.Walk(func(node *parser.Node) bool {
		// Look for function declarations and arrow functions
		nodeType := node.Type()
		if nodeType == "function_declaration" || nodeType == "arrow_function" || nodeType == "function" {
			// Analyze this function for object dependencies
			componentIssues := r.checkFunction(node, ast.FilePath)
			issues = append(issues, componentIssues...)
		}
		return true
	})

	return issues
}

// checkFunction analyzes a single function for object/array deps violations
func (r *NoObjectDeps) checkFunction(funcNode *parser.Node, filePath string) []Issue {
	var issues []Issue

	// Step 1: Find all variable declarations in this function that are object/array literals
	problematicVars := r.findProblematicVars(funcNode)

	// Step 2: Find all hook calls and check their dependencies
	funcNode.Walk(func(node *parser.Node) bool {
		if !node.IsHookCall() {
			return true
		}

		// Get the dependency array
		deps := node.GetDependencyArray()
		if deps == nil {
			return true
		}

		// Check each dependency
		for _, dep := range deps.GetArrayElements() {
			// Check if this dependency is a problematic variable
			depText := dep.Text()
			if problematicVars[depText] {
				line, col := dep.StartPoint()
				issues = append(issues, Issue{
					Rule:     r.Name(),
					Message:  fmt.Sprintf("Dependency '%s' is an object/array created in render and will cause infinite re-renders", depText),
					FilePath: filePath,
					Line:     line + 1, // Convert to 1-indexed
					Column:   col,
				})
			}
		}

		return true
	})

	return issues
}

// findProblematicVars finds variables initialized with object/array literals
func (r *NoObjectDeps) findProblematicVars(funcNode *parser.Node) map[string]bool {
	problematicVars := make(map[string]bool)

	funcNode.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Look for variable declarations
		if nodeType == "lexical_declaration" || nodeType == "variable_declaration" {
			// Get all variable declarators
			for _, child := range node.NamedChildren() {
				if child.Type() == "variable_declarator" {
					// Get the variable name
					nameNode := child.ChildByFieldName("name")
					if nameNode == nil {
						continue
					}

					// Get the initializer (value)
					valueNode := child.ChildByFieldName("value")
					if valueNode == nil {
						continue
					}

					// Check if value is an object or array literal
					valueType := valueNode.Type()
					if valueType == "object" || valueType == "array" {
						varName := nameNode.Text()
						problematicVars[varName] = true
					}
				}
			}
		}

		return true
	})

	return problematicVars
}
