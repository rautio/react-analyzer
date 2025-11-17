package analyzer

import (
	"github.com/rautio/react-analyzer/internal/parser"
)

// ExtractImports finds all import statements in an AST
func ExtractImports(ast *parser.AST) []Import {
	var imports []Import

	ast.Root.Walk(func(node *parser.Node) bool {
		if node.Type() != "import_statement" {
			return true
		}

		imp := parseImport(node)
		if imp != nil {
			imports = append(imports, *imp)
		}

		return true
	})

	return imports
}

// parseImport extracts import information from an import_statement node
func parseImport(node *parser.Node) *Import {
	imp := &Import{}

	// Get the source (string node containing string_fragment)
	for _, child := range node.Children() {
		if child.Type() == "string" {
			// Find string_fragment child
			for _, strChild := range child.Children() {
				if strChild.Type() == "string_fragment" {
					imp.Source = strChild.Text()
					break
				}
			}
		}
	}

	// Get the import clause (what's being imported)
	for _, child := range node.Children() {
		if child.Type() == "import_clause" {
			parseImportClause(child, imp)
			break
		}
	}

	return imp
}

// parseImportClause extracts import details from import_clause node
func parseImportClause(clause *parser.Node, imp *Import) {
	for _, child := range clause.Children() {
		switch child.Type() {
		case "identifier":
			// Default import
			imp.Default = child.Text()

		case "named_imports":
			// Named imports: { foo, bar }
			imp.Named = extractNamedImports(child)

		case "namespace_import":
			// Namespace import: * as Utils
			for _, nsChild := range child.Children() {
				if nsChild.Type() == "identifier" {
					imp.Namespace = nsChild.Text()
				}
			}
		}
	}
}

// extractNamedImports gets the list of named imports from a named_imports node
func extractNamedImports(node *parser.Node) []string {
	var names []string

	node.Walk(func(n *parser.Node) bool {
		// Look for import_specifier nodes
		if n.Type() == "import_specifier" {
			// Get the name (could be aliased)
			for _, child := range n.Children() {
				if child.Type() == "identifier" {
					names = append(names, child.Text())
					break
				}
			}
		}
		return true
	})

	return names
}
