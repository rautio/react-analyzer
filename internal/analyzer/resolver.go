package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rautio/react-analyzer/internal/parser"
)

// ModuleResolver resolves import paths and manages parsed modules
type ModuleResolver struct {
	modules  map[string]*Module // Cache of parsed modules (key: absolute path)
	mu       sync.RWMutex       // Protects modules map for concurrent access
	parserMu sync.Mutex         // Protects parser (tree-sitter is not thread-safe)
	baseDir  string             // Project root directory
	parser   *parser.TreeSitterParser
	aliases  map[string]string // Path aliases: "@/" -> "/absolute/path/to/src/"
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

	// Load path aliases from config files (tsconfig.json, .reactanalyzer.json)
	aliases, err := LoadPathAliases(absBase)
	if err != nil {
		// Config loading is optional - continue without aliases if it fails
		aliases = make(map[string]string)
	}

	return &ModuleResolver{
		modules: make(map[string]*Module),
		baseDir: absBase,
		parser:  p,
		aliases: aliases,
	}, nil
}

// Close cleans up the resolver resources
func (r *ModuleResolver) Close() error {
	return r.parser.Close()
}

// Resolve converts an import path to an absolute file path
func (r *ModuleResolver) Resolve(fromFile string, importPath string) (string, error) {
	var targetPath string

	// Try alias resolution first for non-relative imports
	if !strings.HasPrefix(importPath, ".") {
		// Check if this matches any alias
		if aliasPrefix, aliasTarget, ok := FindLongestMatchingAlias(importPath, r.aliases); ok {
			// Replace alias prefix with target path
			relativePath := strings.TrimPrefix(importPath, aliasPrefix)
			targetPath = filepath.Join(aliasTarget, relativePath)
			targetPath = filepath.Clean(targetPath)
		} else {
			// No alias match - treat as external package
			return "", fmt.Errorf("external package: %s", importPath)
		}
	} else {
		// Relative path - resolve from importing file's directory
		fromDir := filepath.Dir(fromFile)
		targetPath = filepath.Join(fromDir, importPath)
		targetPath = filepath.Clean(targetPath)
	}

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
// Thread-safe: uses read lock for cache lookup, write lock for cache update
func (r *ModuleResolver) GetModule(filePath string) (*Module, error) {
	// Normalize path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Check cache with read lock (allows concurrent reads)
	r.mu.RLock()
	if mod, exists := r.modules[absPath]; exists {
		r.mu.RUnlock()
		return mod, nil
	}
	r.mu.RUnlock()

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", absPath, err)
	}

	// Parse with mutex protection (tree-sitter parser is not thread-safe)
	r.parserMu.Lock()
	ast, err := r.parser.ParseFile(absPath, content)
	r.parserMu.Unlock()

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

	// Analyze symbols in this module (populate Symbols map)
	AnalyzeSymbols(module)

	// Cache it with write lock
	r.mu.Lock()
	// Double-check: another goroutine might have cached it while we were parsing
	if existing, exists := r.modules[absPath]; exists {
		r.mu.Unlock()
		return existing, nil
	}
	r.modules[absPath] = module
	r.mu.Unlock()

	return module, nil
}

// GetModules returns all cached modules
// Thread-safe: returns a copy to prevent external modification
func (r *ModuleResolver) GetModules() map[string]*Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*Module, len(r.modules))
	for k, v := range r.modules {
		result[k] = v
	}
	return result
}
