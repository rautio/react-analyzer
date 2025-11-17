package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oskari/react-analyzer/internal/analyzer"
	"github.com/oskari/react-analyzer/internal/parser"
)

func TestMemoizedComponentUnstableProps_WithViolations(t *testing.T) {
	// Get absolute paths to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	parentFile := filepath.Join(fixturesDir, "ParentWithUnstableProps.tsx")
	content, err := os.ReadFile(parentFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the parent file
	ast, err := p.ParseFile(parentFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Create module resolver
	resolver, err := analyzer.NewModuleResolver(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Run the rule
	rule := &MemoizedComponentUnstableProps{}
	issues := rule.Check(ast, resolver)

	// Should find violations (inline object, array, function passed to MemoChild, and object to AnotherMemo)
	if len(issues) < 3 {
		t.Errorf("Expected at least 3 issues (object, array, function), found %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}

	// Verify issues have correct fields
	for _, issue := range issues {
		if issue.Rule != "memoized-component-unstable-props" {
			t.Errorf("Expected rule 'memoized-component-unstable-props', got '%s'", issue.Rule)
		}
		if issue.Message == "" {
			t.Error("Expected non-empty message")
		}
		if issue.Line == 0 {
			t.Error("Expected non-zero line number")
		}
	}
}

func TestMemoizedComponentUnstableProps_WithStableProps(t *testing.T) {
	// Get absolute paths to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	parentFile := filepath.Join(fixturesDir, "ParentWithStableProps.tsx")
	content, err := os.ReadFile(parentFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the parent file
	ast, err := p.ParseFile(parentFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Create module resolver
	resolver, err := analyzer.NewModuleResolver(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Run the rule
	rule := &MemoizedComponentUnstableProps{}
	issues := rule.Check(ast, resolver)

	// Should have NO issues - all props are properly memoized
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, found %d:", len(issues))
		for _, issue := range issues {
			t.Errorf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}

func TestMemoizedComponentUnstableProps_RuleName(t *testing.T) {
	rule := &MemoizedComponentUnstableProps{}
	if rule.Name() != "memoized-component-unstable-props" {
		t.Errorf("Expected rule name 'memoized-component-unstable-props', got '%s'", rule.Name())
	}
}

func TestMemoizedComponentUnstableProps_NoResolver(t *testing.T) {
	// Create a simple AST
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	content := []byte(`import React from 'react'; function App() { return <div />; }`)
	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Run rule without resolver
	rule := &MemoizedComponentUnstableProps{}
	issues := rule.Check(ast, nil)

	// Should return empty (no resolver available)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when no resolver provided, got %d", len(issues))
	}
}
