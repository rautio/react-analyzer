package analyzer

import (
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestAnalyzeSymbols_ReactMemo(t *testing.T) {
	content := []byte(`
import { memo } from 'react';

export const MyComponent = memo(({ name }) => {
  return <div>{name}</div>;
});
`)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	module := &Module{
		FilePath: "test.tsx",
		AST:      ast,
		Symbols:  make(map[string]*Symbol),
	}

	AnalyzeSymbols(module)

	// Debug output
	t.Logf("Found %d symbols:", len(module.Symbols))
	for name, sym := range module.Symbols {
		t.Logf("  %s: IsMemoized=%v, IsExported=%v", name, sym.IsMemoized, sym.IsExported)
	}

	// Should find MyComponent
	symbol, exists := module.Symbols["MyComponent"]
	if !exists {
		t.Fatal("Expected to find MyComponent symbol")
	}

	if symbol.Type != SymbolComponent {
		t.Errorf("Expected SymbolComponent, got %v", symbol.Type)
	}

	if !symbol.IsMemoized {
		t.Error("Expected MyComponent to be memoized")
	}

	if !symbol.IsExported {
		t.Error("Expected MyComponent to be exported")
	}
}

func TestAnalyzeSymbols_FunctionComponent(t *testing.T) {
	content := []byte(`
function UserProfile({ user }) {
  return <div>{user.name}</div>;
}
`)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	module := &Module{
		FilePath: "test.tsx",
		AST:      ast,
		Symbols:  make(map[string]*Symbol),
	}

	AnalyzeSymbols(module)

	// Should find UserProfile
	symbol, exists := module.Symbols["UserProfile"]
	if !exists {
		t.Fatal("Expected to find UserProfile symbol")
	}

	if symbol.Type != SymbolComponent {
		t.Errorf("Expected SymbolComponent, got %v", symbol.Type)
	}

	if symbol.IsMemoized {
		t.Error("Expected UserProfile to NOT be memoized")
	}
}

func TestAnalyzeSymbols_RegularFunction(t *testing.T) {
	content := []byte(`
function formatUser(user) {
  return user.name.toUpperCase();
}
`)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	module := &Module{
		FilePath: "test.tsx",
		AST:      ast,
		Symbols:  make(map[string]*Symbol),
	}

	AnalyzeSymbols(module)

	// Should find formatUser as a function, not component
	symbol, exists := module.Symbols["formatUser"]
	if !exists {
		t.Fatal("Expected to find formatUser symbol")
	}

	if symbol.Type != SymbolFunction {
		t.Errorf("Expected SymbolFunction, got %v", symbol.Type)
	}
}

func TestLooksLikeComponent(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"MyComponent", true},
		{"UserProfile", true},
		{"formatUser", false},
		{"useState", false},
		{"Component", true},
		{"", false},
	}

	for _, tt := range tests {
		result := looksLikeComponent(tt.name)
		if result != tt.expected {
			t.Errorf("looksLikeComponent(%s) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}
