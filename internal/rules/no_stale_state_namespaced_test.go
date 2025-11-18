package rules

import (
	"strings"
	"testing"
)

func TestNoStaleState_NamespacedHooks(t *testing.T) {
	code := `
import React from 'react';

function Counter() {
  const [count, setCount] = React.useState(0);

  const increment = () => {
    setCount(count + 1);
  };

  return <button onClick={increment}>{count}</button>;
}
`

	ast := parseTestCode(t, code)

	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
		return
	}

	issue := issues[0]

	// Should detect line 8: setCount(count + 1)
	if issue.Line != 8 {
		t.Errorf("Expected issue on line 8, got line %d", issue.Line)
	}

	if !strings.Contains(issue.Message, "count") {
		t.Errorf("Expected message to mention 'count'")
	}

	if !strings.Contains(issue.Message, "functional form") {
		t.Errorf("Expected message to mention 'functional form'")
	}
}

func TestNoStaleState_MixedHooks(t *testing.T) {
	code := `
import React, { useState } from 'react';

function Component() {
  const [a, setA] = useState(0);
  const [b, setB] = React.useState(0);

  const update = () => {
    setA(a + 1);  // bare import
    setB(b + 1);  // namespaced import
  };
}
`

	ast := parseTestCode(t, code)

	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	// Should detect both violations
	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
		return
	}

	// Check that both state variables are mentioned
	foundA := false
	foundB := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "'a'") {
			foundA = true
		}
		if strings.Contains(issue.Message, "'b'") {
			foundB = true
		}
	}

	if !foundA {
		t.Errorf("Expected to find issue for state 'a'")
	}
	if !foundB {
		t.Errorf("Expected to find issue for state 'b'")
	}
}
