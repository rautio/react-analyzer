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
	modules      map[string]*Module // Cache of parsed modules (key: absolute path)
	mu           sync.RWMutex       // Protects modules map for concurrent access
	treeSitterMu sync.Mutex         // GLOBAL lock - protects ALL tree-sitter operations (parsing + AST walking). Tree-sitter C library is not thread-safe.
	baseDir      string             // Project root directory
	parser       *parser.TreeSitterParser
	aliasCache   map[string]map[string]string // Per-directory alias cache: dir -> aliases
	aliasCacheMu sync.RWMutex                 // Protects aliasCache for concurrent access
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
		modules:    make(map[string]*Module),
		baseDir:    absBase,
		parser:     p,
		aliasCache: make(map[string]map[string]string),
	}, nil
}

// Close cleans up the resolver resources
func (r *ModuleResolver) Close() error {
	return r.parser.Close()
}

// getOrLoadAliasesForFile returns the path aliases for a given file
// Uses nearest config file (walks up directory tree) with caching
func (r *ModuleResolver) getOrLoadAliasesForFile(filePath string) map[string]string {
	dir := filepath.Dir(filePath)

	// Fast path: check cache with read lock
	r.aliasCacheMu.RLock()
	if aliases, ok := r.aliasCache[dir]; ok {
		r.aliasCacheMu.RUnlock()
		return aliases
	}
	r.aliasCacheMu.RUnlock()

	// Slow path: find and load nearest config
	configDir := r.findNearestConfigDir(dir)

	var aliases map[string]string
	if configDir != "" {
		// Check if we already loaded this config directory
		r.aliasCacheMu.RLock()
		if cachedAliases, ok := r.aliasCache[configDir]; ok {
			r.aliasCacheMu.RUnlock()
			// Cache for the current directory too
			r.aliasCacheMu.Lock()
			r.aliasCache[dir] = cachedAliases
			r.aliasCacheMu.Unlock()
			return cachedAliases
		}
		r.aliasCacheMu.RUnlock()

		// Load aliases for this config directory
		aliases, _ = LoadPathAliases(configDir)
	}

	if aliases == nil {
		aliases = make(map[string]string)
	}

	// Cache the result (even if empty)
	r.aliasCacheMu.Lock()
	r.aliasCache[dir] = aliases
	if configDir != "" && configDir != dir {
		// Also cache for the config directory itself
		r.aliasCache[configDir] = aliases
	}
	r.aliasCacheMu.Unlock()

	return aliases
}

// findNearestConfigDir walks up the directory tree to find the nearest config file
// Returns the directory containing the config, or empty string if not found
func (r *ModuleResolver) findNearestConfigDir(startDir string) string {
	dir := startDir

	for {
		// Check for .reactanalyzer.json (higher priority)
		reactConfigPath := filepath.Join(dir, ".reactanalyzer.json")
		if _, err := os.Stat(reactConfigPath); err == nil {
			return dir
		}

		// Check for tsconfig.json
		tsconfigPath := filepath.Join(dir, "tsconfig.json")
		if _, err := os.Stat(tsconfigPath); err == nil {
			return dir
		}

		// Stop at baseDir (project root)
		if dir == r.baseDir {
			break
		}

		// Move up one directory
		parent := filepath.Dir(dir)

		// Stop at filesystem root
		if parent == dir {
			break
		}

		dir = parent
	}

	return ""
}

// Resolve converts an import path to an absolute file path
func (r *ModuleResolver) Resolve(fromFile string, importPath string) (string, error) {
	var targetPath string

	// Try alias resolution first for non-relative imports
	if !strings.HasPrefix(importPath, ".") {
		// Get aliases for the source file (uses nearest config with caching)
		aliases := r.getOrLoadAliasesForFile(fromFile)

		// Check if this matches any alias
		if aliasPrefix, aliasTarget, ok := FindLongestMatchingAlias(importPath, aliases); ok {
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

	// Acquire global tree-sitter lock for ALL operations (parsing + AST walking)
	// The tree-sitter C library is not thread-safe, so we must serialize all operations
	r.treeSitterMu.Lock()
	defer r.treeSitterMu.Unlock()

	ast, err := r.parser.ParseFile(absPath, content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %v", absPath, err)
	}

	// Extract imports (walks AST)
	imports := ExtractImports(ast)

	// Create module
	module := &Module{
		FilePath: absPath,
		AST:      ast,
		Imports:  imports,
		Symbols:  make(map[string]*Symbol),
	}

	// Analyze symbols in this module (walks AST)
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

// GetPathAliases returns the path aliases for the base directory
// This loads aliases directly without using the cache (for display purposes)
func (r *ModuleResolver) GetPathAliases() map[string]string {
	aliases, _ := LoadPathAliases(r.baseDir)
	if aliases == nil {
		return make(map[string]string)
	}
	return aliases
}

// LockTreeSitter acquires the global tree-sitter lock
// Must be called before any AST operations outside of GetModule
func (r *ModuleResolver) LockTreeSitter() {
	r.treeSitterMu.Lock()
}

// UnlockTreeSitter releases the global tree-sitter lock
func (r *ModuleResolver) UnlockTreeSitter() {
	r.treeSitterMu.Unlock()
}
