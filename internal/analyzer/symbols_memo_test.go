package analyzer

import (
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestAnalyzeSymbols_ReactMemo_SameFile(t *testing.T) {
	code := `
import React from 'react';

const MemoChild = React.memo(function Child({ config }: any) {
  return <div>{config}</div>;
});

const AnotherMemo = React.memo(({ data }: any) => {
  return <span>{data}</span>;
});

function RegularComponent() {
  return <div />;
}
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Create module
	module := &Module{
		FilePath: "test.tsx",
		AST:      ast,
		Symbols:  make(map[string]*Symbol),
	}

	// Debug: print AST structure
	t.Log("AST structure:")
	ast.Root.Walk(func(node *parser.Node) bool {
		if node.Type() == "variable_declaration" || node.Type() == "lexical_declaration" {
			line, _ := node.StartPoint()
			t.Logf("  Found %s at line %d", node.Type(), line)
		}
		return true
	})

	// Analyze symbols
	AnalyzeSymbols(module)

	// Debug: print all symbols found
	t.Logf("Found %d symbols:", len(module.Symbols))
	for name, symbol := range module.Symbols {
		t.Logf("  %s: Type=%v, IsMemoized=%v", name, symbol.Type, symbol.IsMemoized)
	}

	// Check MemoChild is detected as memoized
	if symbol, exists := module.Symbols["MemoChild"]; !exists {
		t.Error("MemoChild symbol not found")
	} else {
		if !symbol.IsMemoized {
			t.Error("MemoChild should be marked as memoized")
		}
		if symbol.Type != SymbolComponent {
			t.Errorf("MemoChild should be SymbolComponent, got %v", symbol.Type)
		}
	}

	// Check AnotherMemo is detected as memoized
	if symbol, exists := module.Symbols["AnotherMemo"]; !exists {
		t.Error("AnotherMemo symbol not found")
	} else {
		if !symbol.IsMemoized {
			t.Error("AnotherMemo should be marked as memoized")
		}
		if symbol.Type != SymbolComponent {
			t.Errorf("AnotherMemo should be SymbolComponent, got %v", symbol.Type)
		}
	}

	// Check RegularComponent is NOT marked as memoized
	if symbol, exists := module.Symbols["RegularComponent"]; !exists {
		t.Error("RegularComponent symbol not found")
	} else {
		if symbol.IsMemoized {
			t.Error("RegularComponent should NOT be marked as memoized")
		}
	}
}
