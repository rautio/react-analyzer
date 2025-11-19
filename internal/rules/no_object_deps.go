package rules

import (
	"fmt"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
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

	// Acquire global tree-sitter lock before walking ASTs
	// Tree-sitter C library is not thread-safe - must serialize all AST operations
	if resolver != nil {
		resolver.LockTreeSitter()
		defer resolver.UnlockTreeSitter()
	}

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
			depType := dep.Type()
			depText := dep.Text()

			// Case 1: Inline object/array literal (e.g., [{ foo: 'bar' }], [[1, 2, 3]])
			if depType == "object" || depType == "array" {
				line, col := dep.StartPoint()
				issues = append(issues, Issue{
					Rule:     r.Name(),
					Message:  fmt.Sprintf("Inline %s literal in dependency array will cause infinite re-renders", depType),
					FilePath: filePath,
					Line:     line + 1,
					Column:   col,
				})
				continue
			}

			// Case 2: Simple identifier (check this before member expressions)
			if depType == "identifier" && problematicVars[depText] {
				line, col := dep.StartPoint()
				issues = append(issues, Issue{
					Rule:     r.Name(),
					Message:  fmt.Sprintf("Dependency '%s' is an object/array created in render and will cause infinite re-renders", depText),
					FilePath: filePath,
					Line:     line + 1,
					Column:   col,
				})
				continue
			}

			// Case 3: Property access (e.g., config.theme, user?.name)
			// Extract the base identifier from member expressions
			if depType == "member_expression" || depType == "subscript_expression" {
				baseIdentifier := r.getBaseIdentifier(dep)
				if baseIdentifier != "" && problematicVars[baseIdentifier] {
					line, col := dep.StartPoint()
					issues = append(issues, Issue{
						Rule:     r.Name(),
						Message:  fmt.Sprintf("Dependency '%s' accesses object/array '%s' created in render and will cause infinite re-renders", depText, baseIdentifier),
						FilePath: filePath,
						Line:     line + 1,
						Column:   col,
					})
				}
			}
		}

		return true
	})

	return issues
}

// getBaseIdentifier extracts the base identifier from a member expression
// e.g., "config.theme" → "config", "user?.name" → "user", "items[0]" → "items"
func (r *NoObjectDeps) getBaseIdentifier(node *parser.Node) string {
	nodeType := node.Type()

	// If it's already an identifier, return its text
	if nodeType == "identifier" {
		return node.Text()
	}

	// For member expressions (e.g., config.theme, user?.name)
	if nodeType == "member_expression" {
		// Get the object part (left side)
		objectNode := node.ChildByFieldName("object")
		if objectNode != nil {
			// Recursively extract the base identifier
			return r.getBaseIdentifier(objectNode)
		}
	}

	// For subscript expressions (e.g., items[0])
	if nodeType == "subscript_expression" {
		// Get the object part
		objectNode := node.ChildByFieldName("object")
		if objectNode != nil {
			return r.getBaseIdentifier(objectNode)
		}
	}

	// Not a member/subscript expression or couldn't extract
	return ""
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
