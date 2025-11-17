package rules

import (
	"os"
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	// Should have at least one rule (no-object-deps)
	if registry.Count() == 0 {
		t.Error("Expected at least one rule in registry")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	count := registry.Count()
	if count < 1 {
		t.Errorf("Expected at least 1 rule, got %d", count)
	}
}

func TestRegistry_GetRules(t *testing.T) {
	registry := NewRegistry()

	rules := registry.GetRules()
	if len(rules) == 0 {
		t.Error("Expected rules, got empty list")
	}

	// Verify all rules have names
	for _, rule := range rules {
		if rule.Name() == "" {
			t.Error("Expected rule to have a name")
		}
	}
}

func TestRegistry_GetRule(t *testing.T) {
	registry := NewRegistry()

	// Get existing rule
	rule, found := registry.GetRule("no-object-deps")
	if !found {
		t.Error("Expected to find 'no-object-deps' rule")
	}
	if rule == nil {
		t.Error("Expected rule, got nil")
	}
	if rule.Name() != "no-object-deps" {
		t.Errorf("Expected 'no-object-deps', got '%s'", rule.Name())
	}

	// Get non-existent rule
	_, found = registry.GetRule("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent rule")
	}
}

func TestRegistry_RunAll(t *testing.T) {
	registry := NewRegistry()

	// Load and parse a test file
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

	// Run all rules
	issues := registry.RunAll(ast, nil)

	// Should find at least one issue (from no-object-deps rule)
	if len(issues) == 0 {
		t.Error("Expected to find issues, got none")
	}

	// Verify all issues have required fields
	for _, issue := range issues {
		if issue.Rule == "" {
			t.Error("Expected issue to have rule name")
		}
		if issue.Message == "" {
			t.Error("Expected issue to have message")
		}
		if issue.Line == 0 {
			t.Error("Expected issue to have line number")
		}
	}
}

func TestRegistry_RunAllOnCleanFile(t *testing.T) {
	registry := NewRegistry()

	// Load and parse a clean test file
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

	// Run all rules
	issues := registry.RunAll(ast, nil)

	// Should find no issues
	if len(issues) != 0 {
		t.Errorf("Expected no issues on clean file, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  - %s: %s", issue.Rule, issue.Message)
		}
	}
}
