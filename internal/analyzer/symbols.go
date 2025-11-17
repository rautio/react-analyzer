package analyzer

import (
	"strings"

	"github.com/rautio/react-analyzer/internal/parser"
)

// AnalyzeSymbols extracts symbols from a module's AST
func AnalyzeSymbols(module *Module) {
	if module.Symbols == nil {
		module.Symbols = make(map[string]*Symbol)
	}

	// Find all exports and declarations
	module.AST.Root.Walk(func(node *parser.Node) bool {
		analyzeNode(node, module)
		return true
	})
}

// analyzeNode examines a single AST node for symbols
func analyzeNode(node *parser.Node, module *Module) {
	nodeType := node.Type()

	switch nodeType {
	case "export_statement":
		handleExportStatement(node, module)
	case "variable_declaration":
		handleVariableDeclaration(node, module)
	case "function_declaration":
		handleFunctionDeclaration(node, module)
	case "class_declaration":
		handleClassDeclaration(node, module)
	}
}

// handleExportStatement processes export statements
func handleExportStatement(node *parser.Node, module *Module) {
	// Check for: export const Foo = ...
	// Check for: export function Foo() { ... }
	// Check for: export default Foo

	for _, child := range node.NamedChildren() {
		switch child.Type() {
		case "lexical_declaration", "variable_declaration":
			handleVariableDeclaration(child, module)
			markAsExported(child, module)
		case "function_declaration":
			handleFunctionDeclaration(child, module)
			markAsExported(child, module)
		case "class_declaration":
			handleClassDeclaration(child, module)
			markAsExported(child, module)
		}
	}
}

// handleVariableDeclaration processes variable declarations
func handleVariableDeclaration(node *parser.Node, module *Module) {
	// Look for: const MyComponent = ...
	// Check if it's a React component or React.memo(...)

	for _, child := range node.NamedChildren() {
		if child.Type() == "variable_declarator" {
			nameNode := child.ChildByFieldName("name")
			valueNode := child.ChildByFieldName("value")

			if nameNode == nil {
				continue
			}

			name := nameNode.Text()

			// Determine symbol type and properties
			symbol := &Symbol{
				Name: name,
				Type: SymbolVariable,
				Node: child,
			}

			// Check if value is React.memo(...)
			if valueNode != nil && isReactMemo(valueNode) {
				symbol.Type = SymbolComponent
				symbol.IsMemoized = true
			} else if valueNode != nil && isArrowFunction(valueNode) && looksLikeComponent(name) {
				// Arrow function component: const Foo = () => <div>...</div>
				symbol.Type = SymbolComponent
			}

			module.Symbols[name] = symbol
		}
	}
}

// handleFunctionDeclaration processes function declarations
func handleFunctionDeclaration(node *parser.Node, module *Module) {
	// function MyComponent() { return <div>...</div> }

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	name := nameNode.Text()

	symbolType := SymbolFunction
	if looksLikeComponent(name) {
		symbolType = SymbolComponent
	}

	symbol := &Symbol{
		Name: name,
		Type: symbolType,
		Node: node,
	}

	module.Symbols[name] = symbol
}

// handleClassDeclaration processes class declarations
func handleClassDeclaration(node *parser.Node, module *Module) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	name := nameNode.Text()

	symbol := &Symbol{
		Name: name,
		Type: SymbolClass,
		Node: node,
	}

	// Class components are components
	if looksLikeComponent(name) {
		symbol.Type = SymbolComponent
	}

	module.Symbols[name] = symbol
}

// markAsExported marks symbols from a declaration as exported
func markAsExported(node *parser.Node, module *Module) {
	// Find all symbols declared in this node and mark them as exported
	node.Walk(func(n *parser.Node) bool {
		if n.Type() == "variable_declarator" {
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				name := nameNode.Text()
				if symbol, exists := module.Symbols[name]; exists {
					symbol.IsExported = true
				}
			}
		}
		return true
	})
}

// isReactMemo checks if a node is a call to React.memo()
func isReactMemo(node *parser.Node) bool {
	if node.Type() != "call_expression" {
		return false
	}

	funcNode := node.ChildByFieldName("function")
	if funcNode == nil {
		return false
	}

	funcText := funcNode.Text()
	// Check for: memo(...), React.memo(...), or react.memo(...)
	return funcText == "memo" ||
		strings.HasSuffix(funcText, ".memo") ||
		funcText == "React.memo"
}

// isArrowFunction checks if a node is an arrow function
func isArrowFunction(node *parser.Node) bool {
	return node.Type() == "arrow_function"
}

// looksLikeComponent checks if a name looks like a React component
// (starts with uppercase letter)
func looksLikeComponent(name string) bool {
	if len(name) == 0 {
		return false
	}
	firstChar := name[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}
