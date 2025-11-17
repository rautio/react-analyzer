package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rautio/react-analyzer/internal/parser"
)

func TestNoInlineProps_SimplePatterns(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "no-inline-props-simple.tsx")
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
	rule := &NoInlineProps{}
	issues := rule.Check(ast, nil)

	// Expected violations: 12
	// ObjectProps: 3 (lines 7, 8, 9)
	// ArrayProps: 3 (lines 18, 19, 20)
	// FunctionProps: 3 (lines 33, 34, 35)
	// FunctionExpressions: 2 (lines 46, 47)
	// NestedInline: 1 (line 58 for object - nested values inside are not JSX props)
	expectedCount := 12
	if len(issues) != expectedCount {
		t.Errorf("Expected %d issues, got %d", expectedCount, len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
	}

	// Verify specific violations
	expectedLines := []uint32{7, 8, 9, 18, 19, 20, 33, 34, 35, 46, 47, 58}
	actualLines := make([]uint32, len(issues))
	for i, issue := range issues {
		actualLines[i] = issue.Line
	}

	if len(actualLines) == len(expectedLines) {
		for i, expected := range expectedLines {
			if actualLines[i] != expected {
				t.Errorf("Expected issue at line %d, got line %d", expected, actualLines[i])
			}
		}
	}

	// Verify rule name
	if len(issues) > 0 && issues[0].Rule != "no-inline-props" {
		t.Errorf("Expected rule name 'no-inline-props', got '%s'", issues[0].Rule)
	}
}

func TestNoInlineProps_ValidPatterns(t *testing.T) {
	// Get absolute path to fixtures
	fixturesDir, err := filepath.Abs("../../test/fixtures")
	if err != nil {
		t.Fatalf("Failed to get fixtures directory: %v", err)
	}

	testFile := filepath.Join(fixturesDir, "no-inline-props-valid.tsx")
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

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the code
	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	defer ast.Close()

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

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the code
	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	defer ast.Close()

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

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Parse the code
	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	defer ast.Close()

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
