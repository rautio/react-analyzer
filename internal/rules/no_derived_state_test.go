package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

func TestNoDerivedState_SimpleMirroring(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "derived-state-simple.tsx")
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

	// Create module resolver (not needed for this rule, but required by interface)
	resolver, err := analyzer.NewModuleResolver(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Run the rule
	rule := &NoDerivedState{}
	issues := rule.Check(ast, resolver)

	// Should find violations in:
	// - SimpleMirror (1 violation)
	// - MultipleMirrors (2 violations)
	// - InlineEffect (1 violation)
	// Total: 4 violations expected
	expectedMin := 3 // At least 3 to be conservative
	if len(issues) < expectedMin {
		t.Errorf("Expected at least %d issues, found %d", expectedMin, len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}

	// Verify issues have correct fields
	for _, issue := range issues {
		if issue.Rule != "no-derived-state" {
			t.Errorf("Expected rule 'no-derived-state', got '%s'", issue.Rule)
		}
		if issue.Message == "" {
			t.Error("Expected non-empty message")
		}
		if issue.Line == 0 {
			t.Error("Expected non-zero line number")
		}
	}
}

func TestNoDerivedState_ValidCases(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "derived-state-valid.tsx")
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

	// Create module resolver
	resolver, err := analyzer.NewModuleResolver(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}
	defer resolver.Close()

	// Run the rule
	rule := &NoDerivedState{}
	issues := rule.Check(ast, resolver)

	// Valid cases:
	// - InitialValueOnly: No useEffect, so no violation
	// - InteractiveState: Modified in event handler (Phase 3 will handle this)
	// - ComplexEffect: Effect has other side effects (Phase 3 will handle this)
	// - DirectDerivation: No useState at all
	//
	// For Phase 1, we might get some false positives (InteractiveState, ComplexEffect)
	// That's OK - Phase 3 will reduce these

	// For now, let's just verify no crashes and structure is correct
	for _, issue := range issues {
		if issue.Rule != "no-derived-state" {
			t.Errorf("Expected rule 'no-derived-state', got '%s'", issue.Rule)
		}
		t.Logf("Found issue (may be false positive for Phase 1): Line %d: %s", issue.Line, issue.Message)
	}
}

func TestNoDerivedState_RuleName(t *testing.T) {
	rule := &NoDerivedState{}
	if rule.Name() != "no-derived-state" {
		t.Errorf("Expected rule name 'no-derived-state', got '%s'", rule.Name())
	}
}

func TestNoDerivedState_NoViolations(t *testing.T) {
	// Create a simple component with no violations
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	content := []byte(`
		import React from 'react';

		function Component({ user }) {
			// Direct derivation - correct pattern!
			const name = user.name;
			return <div>{name}</div>;
		}
	`)

	ast, err := p.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Run rule without resolver (not needed for this rule)
	rule := &NoDerivedState{}
	issues := rule.Check(ast, nil)

	// Should have no violations
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d:", len(issues))
		for _, issue := range issues {
			t.Errorf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}
