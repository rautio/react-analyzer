package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oskari/react-analyzer/internal/parser"
)

// ModuleResolver resolves import paths and manages parsed modules
type ModuleResolver struct {
	modules map[string]*Module // Cache of parsed modules (key: absolute path)
	baseDir string             // Project root directory
	parser  *parser.TreeSitterParser
}

// NewModuleResolver creates a new module resolver
func NewModuleResolver(baseDir string) (*ModuleResolver, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}

	p, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	return &ModuleResolver{
		modules: make(map[string]*Module),
		baseDir: absBase,
		parser:  p,
	}, nil
}

// Close cleans up the resolver resources
func (r *ModuleResolver) Close() error {
	return r.parser.Close()
}

// Resolve converts an import path to an absolute file path
func (r *ModuleResolver) Resolve(fromFile string, importPath string) (string, error) {
	// Skip external packages (no relative path)
	if !strings.HasPrefix(importPath, ".") {
		return "", fmt.Errorf("external package: %s", importPath)
	}

	// Get the directory of the importing file
	fromDir := filepath.Dir(fromFile)

	// Resolve relative path
	targetPath := filepath.Join(fromDir, importPath)
	targetPath = filepath.Clean(targetPath)

	// Try different extensions
	extensions := []string{".tsx", ".ts", ".jsx", ".js"}

	// First, try with the path as-is (might already have extension)
	if _, err := os.Stat(targetPath); err == nil {
		return filepath.Abs(targetPath)
	}

	// Try adding extensions
	for _, ext := range extensions {
		testPath := targetPath + ext
		if _, err := os.Stat(testPath); err == nil {
			return filepath.Abs(testPath)
		}
	}

	// Try as directory with index file
	for _, ext := range extensions {
		testPath := filepath.Join(targetPath, "index"+ext)
		if _, err := os.Stat(testPath); err == nil {
			return filepath.Abs(testPath)
		}
	}

	return "", fmt.Errorf("cannot resolve: %s from %s", importPath, fromFile)
}

// GetModule returns a module, parsing it if necessary
func (r *ModuleResolver) GetModule(filePath string) (*Module, error) {
	// Normalize path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Check cache
	if mod, exists := r.modules[absPath]; exists {
		return mod, nil
	}

	// Parse the module
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", absPath, err)
	}

	ast, err := r.parser.ParseFile(absPath, content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %v", absPath, err)
	}

	// Extract imports
	imports := ExtractImports(ast)

	// Create module
	module := &Module{
		FilePath: absPath,
		AST:      ast,
		Imports:  imports,
		Symbols:  make(map[string]*Symbol),
	}

	// Cache it
	r.modules[absPath] = module

	return module, nil
}

// GetModules returns all cached modules
func (r *ModuleResolver) GetModules() map[string]*Module {
	return r.modules
}
