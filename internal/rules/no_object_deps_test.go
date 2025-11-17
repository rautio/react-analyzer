package rules

import (
	"os"
	"strings"
	"testing"

	"github.com/oskari/react-analyzer/internal/parser"
)

func TestNoObjectDeps_ValidDeps(t *testing.T) {
	// Load and parse the valid fixture
	content, err := os.ReadFile("../../test/fixtures/valid-object-deps.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("valid-object-deps.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Run the rule
	rule := &NoObjectDeps{}
	issues := rule.Check(ast, nil)

	// Should have NO issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, found %d:", len(issues))
		for _, issue := range issues {
			t.Errorf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}

func TestNoObjectDeps_ObjectInRender(t *testing.T) {
	// Load and parse the fixture with violations
	content, err := os.ReadFile("../../test/fixtures/object-in-render.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("object-in-render.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Run the rule
	rule := &NoObjectDeps{}
	issues := rule.Check(ast, nil)

	// Expected violations based on the fixture file:
	// Line 15: config in useEffect deps
	// Line 22: items in useEffect deps
	// Line 29: options in useMemo deps
	// Line 42: settings and preferences in useEffect deps (2 violations)
	// Line 51: theme in useEffect deps
	// Total: 6 violations

	expectedViolations := map[string]bool{
		"config":      false,
		"items":       false,
		"options":     false,
		"settings":    false,
		"preferences": false,
		"theme":       false,
	}

	if len(issues) == 0 {
		t.Fatal("Expected violations but found none")
	}

	// Verify we found all expected violations
	for _, issue := range issues {
		t.Logf("Found issue at line %d: %s", issue.Line, issue.Message)

		// Mark which violations we found
		for varName := range expectedViolations {
			// Check if this issue message contains the variable name
			if containsVar(issue.Message, varName) {
				expectedViolations[varName] = true
			}
		}
	}

	// Check that we found all expected violations
	for varName, found := range expectedViolations {
		if !found {
			t.Errorf("Expected to find violation for '%s' but didn't", varName)
		}
	}

	// Verify we have the right number of issues
	if len(issues) < 6 {
		t.Errorf("Expected at least 6 violations, found %d", len(issues))
	}
}

func TestNoObjectDeps_RuleName(t *testing.T) {
	rule := &NoObjectDeps{}
	if rule.Name() != "no-object-deps" {
		t.Errorf("Expected rule name 'no-object-deps', got '%s'", rule.Name())
	}
}

func TestNoObjectDeps_WithHooks(t *testing.T) {
	// Test with the existing with-hooks fixture
	content, err := os.ReadFile("../../test/fixtures/with-hooks.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("with-hooks.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Run the rule
	rule := &NoObjectDeps{}
	issues := rule.Check(ast, nil)

	// The with-hooks.tsx has:
	// const config = { theme: 'dark' };
	// useEffect(() => { ... }, [count, config]);
	// Should find 1 violation for 'config'

	if len(issues) != 1 {
		t.Errorf("Expected 1 violation, found %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}

	if len(issues) > 0 {
		issue := issues[0]
		if !containsVar(issue.Message, "config") {
			t.Errorf("Expected violation for 'config', got: %s", issue.Message)
		}
		if issue.Line != 10 {
			t.Errorf("Expected violation at line 10, got line %d", issue.Line)
		}
	}
}

// Helper function to check if a message mentions a variable
func containsVar(message, varName string) bool {
	// Check if message contains 'varName' (with quotes)
	return strings.Contains(message, "'"+varName+"'")
}
