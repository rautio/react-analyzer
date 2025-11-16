package parser

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Parser interface for parsing source files
type Parser interface {
	ParseFile(filePath string, content []byte) (*AST, error)
	Close() error
}

// AST represents a parsed file
type AST struct {
	Root     *Node
	FilePath string
	Language string
	tree     *sitter.Tree // Keep for cleanup
}

// Node represents an AST node
type Node struct {
	tsNode  *sitter.Node
	content []byte
}
