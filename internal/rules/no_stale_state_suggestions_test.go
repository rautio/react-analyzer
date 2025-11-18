package rules

import (
	"strings"
	"testing"
)

// TestNoStaleState_SuggestionCorrectness verifies that suggestions replace state variable with 'prev'
func TestNoStaleState_SuggestionCorrectness(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		expectedInMsg    string // What we expect to see in the suggestion
		shouldNotContain string // What should NOT be in the suggestion
	}{
		{
			name: "simple addition",
			code: `
import React from 'react';
function Counter() {
  const [count, setCount] = React.useState(0);
  setCount(count + 1);
}`,
			expectedInMsg:    "prev+1",  // Whitespace may vary
			shouldNotContain: "count+1", // Should NOT contain the old variable
		},
		{
			name: "simple subtraction",
			code: `
import React from 'react';
function Counter() {
  const [count, setCount] = React.useState(0);
  setCount(count - 1);
}`,
			expectedInMsg:    "prev-1",  // Whitespace may vary
			shouldNotContain: "count-1", // Should NOT contain the old variable
		},
		{
			name: "boolean toggle",
			code: `
import React from 'react';
function Toggle() {
  const [isOpen, setIsOpen] = React.useState(false);
  setIsOpen(!isOpen);
}`,
			expectedInMsg:    "!prev",
			shouldNotContain: "!isOpen",
		},
		{
			name: "array spread",
			code: `
import React from 'react';
function List() {
  const [items, setItems] = React.useState([]);
  setItems([...items, 'new']);
}`,
			expectedInMsg:    "[...prev,'new']", // Whitespace may vary
			shouldNotContain: "[...items,",
		},
		{
			name: "array slice",
			code: `
import React from 'react';
function List() {
  const [items, setItems] = React.useState([]);
  setItems(items.slice(1));
}`,
			expectedInMsg:    "prev.slice(1)",
			shouldNotContain: "items.slice",
		},
		{
			name: "object spread with property update",
			code: `
import React from 'react';
function UserProfile() {
  const [user, setUser] = React.useState({});
  setUser({ ...user, age: user.age + 1 });
}`,
			expectedInMsg:    "{...prev,age:prev.age+1}", // Whitespace may vary
			shouldNotContain: "user.age",
		},
		{
			name: "multiplication",
			code: `
import React from 'react';
function Counter() {
  const [count, setCount] = React.useState(0);
  setCount(count * 2);
}`,
			expectedInMsg:    "prev*2", // Whitespace may vary
			shouldNotContain: "count*2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := parseTestCode(t, tt.code)

			rule := &NoStaleState{}
			issues := rule.Check(ast, nil)

			if len(issues) == 0 {
				t.Fatal("Expected at least one issue")
			}

			issue := issues[0]

			// Check that the suggestion contains the expected text
			if !strings.Contains(issue.Message, tt.expectedInMsg) {
				t.Errorf("Expected message to contain '%s', got: %s", tt.expectedInMsg, issue.Message)
			}

			// Check that the suggestion does NOT contain the old stale reference
			if strings.Contains(issue.Message, tt.shouldNotContain) {
				t.Errorf("Message should NOT contain '%s', but got: %s", tt.shouldNotContain, issue.Message)
			}
		})
	}
}

// TestNoStaleState_ComplexSuggestions tests more complex scenarios
func TestNoStaleState_ComplexSuggestions(t *testing.T) {
	code := `
import React from 'react';

function Component() {
  const [count, setCount] = React.useState(0);
  const [items, setItems] = React.useState([]);
  const [user, setUser] = React.useState({ name: '', age: 0 });

  // Multiple references to the same state variable
  setCount(count + count);

  // Array filter
  setItems(items.filter(x => x.active));

  // Object with nested property
  setUser({ ...user, name: user.name.toUpperCase() });
}
`

	ast := parseTestCode(t, code)

	rule := &NoStaleState{}
	issues := rule.Check(ast, nil)

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("  Line %d: %s", issue.Line, issue.Message)
		}
		return
	}

	// Check first issue (count + count) - whitespace may vary
	if !strings.Contains(issues[0].Message, "prev+prev") && !strings.Contains(issues[0].Message, "prev + prev") {
		t.Errorf("Expected 'prev+prev' in first issue, got: %s", issues[0].Message)
	}

	// Check second issue (items.filter)
	if !strings.Contains(issues[1].Message, "prev.filter") {
		t.Errorf("Expected 'prev.filter' in second issue, got: %s", issues[1].Message)
	}

	// Check third issue (user spread with nested property)
	if !strings.Contains(issues[2].Message, "prev.name") {
		t.Errorf("Expected 'prev.name' in third issue, got: %s", issues[2].Message)
	}
}
