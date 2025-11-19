package graph

import (
	"os"
	"testing"

	"github.com/rautio/react-analyzer/internal/analyzer"
)

func TestBuilder_Build_SimplePropDrilling(t *testing.T) {
	// Test the full graph building process with a simple prop drilling example
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"App.tsx": `
import React, { useState } from 'react';

function App() {
  const [count, setCount] = useState(0);
  return <Parent count={count} />;
}

function Parent({ count }) {
  return <Child count={count} />;
}

function Child({ count }) {
  return <div>{count}</div>;
}
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify components were detected
	if len(graph.ComponentNodes) == 0 {
		t.Fatal("Expected components to be detected")
	}

	// Verify state nodes were detected
	if len(graph.StateNodes) == 0 {
		t.Fatal("Expected state nodes to be detected")
	}

	// Verify edges were created
	if len(graph.Edges) == 0 {
		t.Fatal("Expected edges to be created")
	}

	// Check for specific components
	var foundApp, foundParent, foundChild bool
	for _, comp := range graph.ComponentNodes {
		switch comp.Name {
		case "App":
			foundApp = true
		case "Parent":
			foundParent = true
		case "Child":
			foundChild = true
		}
	}

	if !foundApp {
		t.Error("Expected to find App component")
	}
	if !foundParent {
		t.Error("Expected to find Parent component")
	}
	if !foundChild {
		t.Error("Expected to find Child component")
	}
}

func TestBuilder_Build_ComponentWithState(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"Component.tsx": `
import React, { useState } from 'react';

function Counter() {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('test');

  return <div>{count} {name}</div>;
}
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Find Counter component
	var counter *ComponentNode
	for _, comp := range graph.ComponentNodes {
		if comp.Name == "Counter" {
			counter = comp
			break
		}
	}

	if counter == nil {
		t.Fatal("Counter component not found")
	}

	// Verify it has 2 state nodes
	if len(counter.StateNodes) != 2 {
		t.Errorf("Expected 2 state nodes, got %d", len(counter.StateNodes))
	}

	// Verify state nodes exist in graph
	var foundCount, foundName bool
	for _, stateNode := range graph.StateNodes {
		if stateNode.Name == "count" {
			foundCount = true
			if stateNode.Type != "useState" {
				t.Errorf("Expected count to be useState type, got %s", stateNode.Type)
			}
		}
		if stateNode.Name == "name" {
			foundName = true
		}
	}

	if !foundCount {
		t.Error("Expected to find 'count' state node")
	}
	if !foundName {
		t.Error("Expected to find 'name' state node")
	}
}

func TestBuilder_Build_MemoizedComponent(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"Memo.tsx": `
import React from 'react';

const MemoComponent = React.memo(({ value }) => {
  return <div>{value}</div>;
});
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Find MemoComponent
	var memoComp *ComponentNode
	for _, comp := range graph.ComponentNodes {
		if comp.Name == "MemoComponent" {
			memoComp = comp
			break
		}
	}

	if memoComp == nil {
		t.Fatal("MemoComponent not found")
	}

	if !memoComp.IsMemoized {
		t.Error("Expected MemoComponent to be marked as memoized")
	}
}

func TestBuilder_Build_ComponentHierarchy(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"Hierarchy.tsx": `
import React from 'react';

function Parent() {
  return (
    <div>
      <ChildA />
      <ChildB />
    </div>
  );
}

function ChildA() {
  return <div>A</div>;
}

function ChildB() {
  return <div>B</div>;
}
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Find Parent component
	var parent *ComponentNode
	for _, comp := range graph.ComponentNodes {
		if comp.Name == "Parent" {
			parent = comp
			break
		}
	}

	if parent == nil {
		t.Fatal("Parent component not found")
	}

	// Verify Parent has 2 children
	if len(parent.Children) != 2 {
		t.Errorf("Expected Parent to have 2 children, got %d", len(parent.Children))
	}

	// Verify children have parent set
	for _, comp := range graph.ComponentNodes {
		if comp.Name == "ChildA" || comp.Name == "ChildB" {
			if comp.Parent == "" {
				t.Errorf("Expected %s to have parent set", comp.Name)
			}
		}
	}
}

func TestBuilder_Build_PropsPassing(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"Props.tsx": `
import React, { useState } from 'react';

function Parent() {
  const [value, setValue] = useState(0);
  return <Child value={value} />;
}

function Child({ value }) {
  return <div>{value}</div>;
}
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Find Parent component
	var parent *ComponentNode
	for _, comp := range graph.ComponentNodes {
		if comp.Name == "Parent" {
			parent = comp
			break
		}
	}

	if parent == nil {
		t.Fatal("Parent component not found")
	}

	// Verify Parent has props passed to children
	if len(parent.PropsPassedTo) == 0 {
		t.Error("Expected Parent to pass props to children")
	}

	// Check for 'passes' edges
	var foundPassesEdge bool
	for _, edge := range graph.Edges {
		if edge.Type == "passes" && edge.PropName == "value" {
			foundPassesEdge = true
			break
		}
	}

	if !foundPassesEdge {
		t.Error("Expected to find 'passes' edge for value prop")
	}
}

func TestBuilder_Build_EmptyFile(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"Empty.tsx": ``,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()

	// Should not error on empty file
	if err != nil {
		t.Fatalf("Build() should not error on empty file: %v", err)
	}

	// Graph should be empty but valid
	if graph == nil {
		t.Fatal("Expected graph to be non-nil")
	}
}

func TestBuilder_Build_NoComponents(t *testing.T) {
	resolver, cleanup := setupTestResolver(t, map[string]string{
		"NoComponents.tsx": `
const x = 1;
const y = 2;
`,
	})
	defer cleanup()

	builder := NewBuilder(resolver)
	graph, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Should have empty components but no error
	if len(graph.ComponentNodes) != 0 {
		t.Error("Expected no components for non-component file")
	}
}

// Helper function to set up a test resolver with in-memory files
func setupTestResolver(t *testing.T, files map[string]string) (*analyzer.ModuleResolver, func()) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()

	// Write all files to temp directory
	for filename, content := range files {
		path := tempDir + "/" + filename
		if err := writeTestFile(t, path, content); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create resolver
	resolver, err := analyzer.NewModuleResolver(tempDir)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	// Parse all files by calling GetModule
	for filename := range files {
		path := tempDir + "/" + filename
		if _, err := resolver.GetModule(path); err != nil {
			t.Fatalf("Failed to parse file %s: %v", filename, err)
		}
	}

	cleanup := func() {
		resolver.Close()
	}

	return resolver, cleanup
}

func writeTestFile(t *testing.T, path string, content string) error {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
