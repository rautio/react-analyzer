package parser

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
)

// TreeSitterParser implements Parser using tree-sitter
type TreeSitterParser struct {
	parser   *sitter.Parser
	language *sitter.Language
}

// NewParser creates a new tree-sitter parser for TypeScript/JSX
func NewParser() (*TreeSitterParser, error) {
	parser := sitter.NewParser()
	language := tsx.GetLanguage()

	parser.SetLanguage(language)

	return &TreeSitterParser{
		parser:   parser,
		language: language,
	}, nil
}

// ParseFile parses a source file and returns an AST
func (p *TreeSitterParser) ParseFile(filePath string, content []byte) (*AST, error) {
	tree, err := p.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	if tree == nil {
		return nil, fmt.Errorf("failed to parse file")
	}

	root := tree.RootNode()
	if root == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	// Check for syntax errors
	if root.HasError() {
		return nil, fmt.Errorf("syntax error in file")
	}

	return &AST{
		Root:     wrapNode(root, content),
		FilePath: filePath,
		Language: "tsx",
		tree:     tree,
	}, nil
}

// Close cleans up the parser resources
func (p *TreeSitterParser) Close() error {
	// Parser cleanup is handled by GC in go-tree-sitter
	return nil
}

// CloseAST cleans up tree-sitter tree resources
func (ast *AST) Close() {
	if ast.tree != nil {
		ast.tree.Close()
		ast.tree = nil
	}
}
