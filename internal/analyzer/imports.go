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
func extractNamedImports(node *parser.Node) []NamedImport {
	var imports []NamedImport

	node.Walk(func(n *parser.Node) bool {
		// Look for import_specifier nodes
		// Structure: import { Foo as Bar } creates:
		//   import_specifier
		//     - identifier: "Foo" (imported name)
		//     - identifier: "Bar" (local alias)
		if n.Type() == "import_specifier" {
			var identifiers []string
			for _, child := range n.Children() {
				if child.Type() == "identifier" {
					identifiers = append(identifiers, child.Text())
				}
			}

			if len(identifiers) > 0 {
				namedImport := NamedImport{
					ImportedName: identifiers[0],
					LocalName:    identifiers[0], // Default to same name
				}
				// If there's a second identifier, it's the alias
				if len(identifiers) > 1 {
					namedImport.LocalName = identifiers[1]
				}
				imports = append(imports, namedImport)
			}
		}
		return true
	})

	return imports
}
