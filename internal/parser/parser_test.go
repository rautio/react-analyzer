package parser

import (
	"os"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	if parser == nil {
		t.Fatal("Parser is nil")
	}
}

func TestParseSimpleComponent(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	content, err := os.ReadFile("../../test/fixtures/simple.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	ast, err := parser.ParseFile("simple.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	if ast.Root == nil {
		t.Fatal("AST root is nil")
	}

	if ast.Root.Type() != "program" {
		t.Errorf("Expected root type 'program', got '%s'", ast.Root.Type())
	}
}

func TestFindHookCalls(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	content, err := os.ReadFile("../../test/fixtures/with-hooks.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	ast, err := parser.ParseFile("with-hooks.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Find all hook calls
	var hookCalls []*Node
	ast.Root.Walk(func(node *Node) bool {
		if node.IsHookCall() {
			hookCalls = append(hookCalls, node)
		}
		return true // Continue traversal
	})

	// Should find useState and useEffect
	if len(hookCalls) != 2 {
		t.Errorf("Expected 2 hook calls, found %d", len(hookCalls))
	}

	// Check hook names
	expectedHooks := map[string]bool{
		"useState":  false,
		"useEffect": false,
	}

	for _, hook := range hookCalls {
		funcNode := hook.ChildByFieldName("function")
		if funcNode != nil {
			hookName := funcNode.Text()
			if _, exists := expectedHooks[hookName]; exists {
				expectedHooks[hookName] = true
			}
		}
	}

	for hookName, found := range expectedHooks {
		if !found {
			t.Errorf("Expected to find %s hook call", hookName)
		}
	}
}

func TestGetDependencyArray(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	content, err := os.ReadFile("../../test/fixtures/with-hooks.tsx")
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	ast, err := parser.ParseFile("with-hooks.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	defer ast.Close()

	// Find useEffect call
	var useEffectCall *Node
	ast.Root.Walk(func(node *Node) bool {
		if node.IsHookCall() {
			funcNode := node.ChildByFieldName("function")
			if funcNode != nil && funcNode.Text() == "useEffect" {
				useEffectCall = node
				return false // Stop traversal
			}
		}
		return true
	})

	if useEffectCall == nil {
		t.Fatal("Could not find useEffect call")
	}

	// Get dependency array
	deps := useEffectCall.GetDependencyArray()
	if deps == nil {
		t.Fatal("Could not find dependency array")
	}

	// Get array elements
	elements := deps.GetArrayElements()
	if len(elements) != 2 {
		t.Errorf("Expected 2 dependencies, found %d", len(elements))
	}

	// Check dependency names
	expectedDeps := []string{"count", "config"}
	for i, elem := range elements {
		if i < len(expectedDeps) {
			if elem.Text() != expectedDeps[i] {
				t.Errorf("Expected dependency '%s', got '%s'", expectedDeps[i], elem.Text())
			}
		}
	}
}

func TestNodeMethods(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	content := []byte(`function test() { return <div>Hello</div>; }`)
	ast, err := parser.ParseFile("test.tsx", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	defer ast.Close()

	root := ast.Root
	if root == nil {
		t.Fatal("Root is nil")
	}

	// Test Type()
	if root.Type() != "program" {
		t.Errorf("Expected type 'program', got '%s'", root.Type())
	}

	// Test Children()
	children := root.Children()
	if len(children) == 0 {
		t.Error("Expected children, got none")
	}

	// Test NamedChildren()
	namedChildren := root.NamedChildren()
	if len(namedChildren) == 0 {
		t.Error("Expected named children, got none")
	}

	// Test StartPoint()
	row, col := root.StartPoint()
	if row != 0 || col != 0 {
		t.Errorf("Expected start point (0, 0), got (%d, %d)", row, col)
	}
}
