package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestNoStaleState_SimplePatterns(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "no-stale-state-simple.tsx")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the file
	ast, err := p.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Run the rule
	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	// Expected violations:
	// Counter: 4 violations (lines 6, 7, 8, 9)
	// Toggle: 1 violation (line 16)
	// ItemList: 3 violations (lines 23, 24, 25)
	// UserProfile: 2 violations (lines 32, 33)
	// MultipleUpdates: 2 violations (lines 41, 42)
	// ComplexHandler: 2 violations (lines 56, 59)
	// Total: 14 violations expected

	expectedMin := 14
	if len(issues) < expectedMin {
		t.Errorf("Expected at least %d issues, found %d", expectedMin, len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	} else {
		t.Logf("Successfully detected %d violations:", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}

	// Verify issues have correct fields
	for _, issue := range issues {
		if issue.Rule != "no-stale-state" {
			t.Errorf("Expected rule 'no-stale-state', got '%s'", issue.Rule)
		}
		if issue.Message == "" {
			t.Error("Expected non-empty message")
		}
		if issue.Line == 0 {
			t.Error("Expected non-zero line number")
		}
	}
}

func TestNoStaleState_ValidPatterns(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "no-stale-state-valid.tsx")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the file
	ast, err := p.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Run the rule
	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	// Should have no violations - all patterns are valid
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues in valid file, got %d:", len(issues))
		for _, issue := range issues {
			t.Errorf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}

func TestNoStaleState_RuleName(t *testing.T) {
	rule := &NoStaleState{}
	if rule.Name() != "no-stale-state" {
		t.Errorf("Expected rule name 'no-stale-state', got '%s'", rule.Name())
	}
}

func TestNoStaleState_NoViolations(t *testing.T) {
	// Create a simple component with no violations
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	content := []byte(`
		import { useState } from 'react';

		function Component() {
			const [count, setCount] = useState(0);

			// Functional form - correct!
			const increment = () => setCount(prev => prev + 1);

			return <div>{count}</div>;
		}
	`)

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Run rule
	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	// Should have no violations
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d:", len(issues))
		for _, issue := range issues {
			t.Errorf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}
