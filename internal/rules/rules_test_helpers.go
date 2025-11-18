package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

// parseTestFixture is a helper to reduce boilerplate in tests.
// It parses a fixture file from test/fixtures directory.
func parseTestFixture(t *testing.T, filename string) *parser.AST {
	t.Helper()

	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, filename)
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	t.Cleanup(func() { p.Close() })

	ast, err := p.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	t.Cleanup(func() { ast.Close() })

	return ast
}

// parseTestCode parses inline code strings for tests.
// Useful for quick unit tests without creating fixture files.
func parseTestCode(t *testing.T, code string) *parser.AST {
	t.Helper()

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	t.Cleanup(func() { p.Close() })

	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	t.Cleanup(func() { ast.Close() })

	return ast
}
