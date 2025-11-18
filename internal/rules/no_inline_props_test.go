package rules

import (
	"strings"
	"testing"
)

func TestNoInlineProps_SimplePatterns(t *testing.T) {
	ast := parseTestFixture(t, "no-inline-props-simple.tsx")

	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	// Count violations by type instead of checking exact line numbers
	objectCount := 0
	arrayCount := 0
	funcCount := 0

	for _, issue := range issues {
		if strings.Contains(issue.Message, "inline object") {
			objectCount++
		} else if strings.Contains(issue.Message, "inline array") {
			arrayCount++
		} else if strings.Contains(issue.Message, "inline function") {
			funcCount++
		}

		// Verify all issues have correct rule name
		if issue.Rule != "no-inline-props" {
			t.Errorf("Expected rule name 'no-inline-props', got '%s'", issue.Rule)
		}
	}

	// Expected violations by type:
	// - Objects: 4 (3 from ObjectProps + 1 from NestedInline)
	// - Arrays: 3 (from ArrayProps)
	// - Functions: 5 (3 arrow functions + 2 function expressions)
	if objectCount != 4 {
		t.Errorf("Expected 4 object violations, got %d", objectCount)
	}
	if arrayCount != 3 {
		t.Errorf("Expected 3 array violations, got %d", arrayCount)
	}
	if funcCount != 5 {
		t.Errorf("Expected 5 function violations, got %d", funcCount)
	}

	// Total should be 12
	totalExpected := 12
	if len(issues) != totalExpected {
		t.Errorf("Expected %d total issues, got %d", totalExpected, len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}

func TestNoInlineProps_ValidPatterns(t *testing.T) {
	ast := parseTestFixture(t, "no-inline-props-valid.tsx")

	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	// Should have zero violations
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues in valid file, got %d:", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}
}

func TestNoInlineProps_MessageFormat(t *testing.T) {
	code := `
import React from 'react';

function Test() {
  return <Button onClick={() => click()} />;
}
`

	ast := parseTestCode(t, code)

	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	if len(issues) == 0 {
		t.Fatal("Expected at least one issue")
	}

	issue := issues[0]

	// Check message format
	expectedSubstrings := []string{
		"onClick",
		"inline function",
		"new reference every render",
		"useCallback",
	}

	for _, substr := range expectedSubstrings {
		if !stringContains(issue.Message, substr) {
			t.Errorf("Expected message to contain '%s', got: %s", substr, issue.Message)
		}
	}
}

func TestNoInlineProps_ObjectProps(t *testing.T) {
	code := `
import React from 'react';

function Test() {
  return <Component config={{ theme: 'dark' }} />;
}
`

	ast := parseTestCode(t, code)

	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]

	if !stringContains(issue.Message, "config") {
		t.Errorf("Expected message to mention 'config' prop")
	}

	if !stringContains(issue.Message, "object") {
		t.Errorf("Expected message to mention 'object'")
	}

	if !stringContains(issue.Message, "useMemo") {
		t.Errorf("Expected message to suggest 'useMemo'")
	}
}

func TestNoInlineProps_ArrayProps(t *testing.T) {
	code := `
import React from 'react';

function Test() {
  return <List items={[1, 2, 3]} />;
}
`

	ast := parseTestCode(t, code)

	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]

	if !stringContains(issue.Message, "items") {
		t.Errorf("Expected message to mention 'items' prop")
	}

	if !stringContains(issue.Message, "array") {
		t.Errorf("Expected message to mention 'array'")
	}

	if !stringContains(issue.Message, "useMemo") {
		t.Errorf("Expected message to suggest 'useMemo'")
	}
}

func TestNoInlineProps_RuleName(t *testing.T) {
	rule := &NoInlineProps{}
	if rule.Name() != "no-inline-props" {
		t.Errorf("Expected rule name 'no-inline-props', got '%s'", rule.Name())
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
