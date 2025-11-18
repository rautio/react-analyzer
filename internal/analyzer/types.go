package analyzer

import "github.com/rautio/react-analyzer/internal/parser"

// NamedImport represents a single named import with optional alias
type NamedImport struct {
	ImportedName string // The name being imported from the module
	LocalName    string // The local name (alias), same as ImportedName if no alias
}

// Import represents an import statement
type Import struct {
	Source    string        // Import path: "./MyComponent", "react", etc.
	Default   string        // Default import: "React" in "import React from 'react'"
	Named     []NamedImport // Named imports: e.g., {ImportedName: "MemoChild", LocalName: "FastChild"}
	Namespace string        // Namespace: "Utils" in "import * as Utils from './utils'"
}

// Symbol represents a named entity in a module
type Symbol struct {
	Name       string
	Type       SymbolType
	Node       *parser.Node
	IsMemoized bool // For components: wrapped in React.memo
	IsExported bool
}

// SymbolType categorizes symbols
type SymbolType int

const (
	SymbolUnknown   SymbolType = iota
	SymbolComponent            // React component
	SymbolFunction             // Regular function
	SymbolVariable             // Variable
	SymbolClass                // Class
)

// Module represents a parsed file with metadata
type Module struct {
	FilePath string
	AST      *parser.AST
	Imports  []Import
	Symbols  map[string]*Symbol
}
