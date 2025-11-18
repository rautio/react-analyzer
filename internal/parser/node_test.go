package parser

import (
	"testing"
)

// TestIsHookCall_Namespaced tests detection of React.useState, React.useEffect etc.
func TestIsHookCall_Namespaced(t *testing.T) {
	code := `
import React from 'react';

function Component() {
  const [count, setCount] = React.useState(0);

  React.useEffect(() => {
    console.log(count);
  }, [count]);

  const value = React.useMemo(() => count * 2, [count]);

  return <div>{value}</div>;
}
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Find all hook calls
	var hookCalls []*Node
	ast.Root.Walk(func(node *Node) bool {
		if node.IsHookCall() {
			hookCalls = append(hookCalls, node)
		}
		return true
	})

	// Should find 3 hooks: React.useState, React.useEffect, React.useMemo
	if len(hookCalls) != 3 {
		t.Errorf("Expected 3 hook calls, got %d", len(hookCalls))
	}

	// Check hook names
	expectedNames := []string{"useState", "useEffect", "useMemo"}
	for i, hook := range hookCalls {
		name := hook.GetHookName()
		if name != expectedNames[i] {
			t.Errorf("Hook %d: expected name %s, got %s", i, expectedNames[i], name)
		}
	}
}

// TestIsHookCall_Mixed tests both bare and namespaced hooks
func TestIsHookCall_Mixed(t *testing.T) {
	code := `
import React, { useState, useEffect } from 'react';

function Component() {
  const [a, setA] = useState(0);
  const [b, setB] = React.useState(0);

  useEffect(() => console.log(a), [a]);
  React.useEffect(() => console.log(b), [b]);
}
`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	ast, err := p.ParseFile("test.tsx", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	// Find all hook calls
	hookCount := 0
	ast.Root.Walk(func(node *Node) bool {
		if node.IsHookCall() {
			hookCount++
		}
		return true
	})

	// Should find 4 hooks: 2 useState (bare + namespaced) + 2 useEffect (bare + namespaced)
	if hookCount != 4 {
		t.Errorf("Expected 4 hook calls, got %d", hookCount)
	}
}

// TestGetHookName tests extracting hook names from both styles
func TestGetHookName(t *testing.T) {
	tests := []struct {
		code         string
		expectedName string
	}{
		{
			code:         "useState(0)",
			expectedName: "useState",
		},
		{
			code:         "React.useState(0)",
			expectedName: "useState",
		},
		{
			code:         "useEffect(() => {}, [])",
			expectedName: "useEffect",
		},
		{
			code:         "React.useEffect(() => {}, [])",
			expectedName: "useEffect",
		},
	}

	for _, tt := range tests {
		p, err := NewParser()
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		ast, err := p.ParseFile("test.tsx", []byte(tt.code))
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", tt.code, err)
		}

		// Find the call_expression
		var callNode *Node
		ast.Root.Walk(func(node *Node) bool {
			if node.Type() == "call_expression" {
				callNode = node
				return false
			}
			return true
		})

		if callNode == nil {
			t.Fatalf("No call_expression found in: %s", tt.code)
		}

		name := callNode.GetHookName()
		if name != tt.expectedName {
			t.Errorf("For %s: expected name %s, got %s", tt.code, tt.expectedName, name)
		}

		p.Close()
		ast.Close()
	}
}
