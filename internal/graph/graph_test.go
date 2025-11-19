package graph

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g == nil {
		t.Fatal("NewGraph returned nil")
	}

	if g.StateNodes == nil {
		t.Error("StateNodes map not initialized")
	}

	if g.ComponentNodes == nil {
		t.Error("ComponentNodes map not initialized")
	}

	if g.Edges == nil {
		t.Error("Edges slice not initialized")
	}

	if g.EdgesBySource == nil {
		t.Error("EdgesBySource map not initialized")
	}

	if g.EdgesByTarget == nil {
		t.Error("EdgesByTarget map not initialized")
	}
}

func TestAddStateNode(t *testing.T) {
	g := NewGraph()

	state := &StateNode{
		ID:   "state:App:count:abc123",
		Name: "count",
		Type: StateTypeUseState,
	}

	g.AddStateNode(state)

	if len(g.StateNodes) != 1 {
		t.Errorf("Expected 1 state node, got %d", len(g.StateNodes))
	}

	retrieved, exists := g.StateNodes[state.ID]
	if !exists {
		t.Error("State node not found in graph")
	}

	if retrieved.Name != "count" {
		t.Errorf("Expected name 'count', got '%s'", retrieved.Name)
	}
}

func TestAddComponentNode(t *testing.T) {
	g := NewGraph()

	comp := &ComponentNode{
		ID:   "component:App:xyz789",
		Name: "App",
		Type: ComponentTypeFunction,
	}

	g.AddComponentNode(comp)

	if len(g.ComponentNodes) != 1 {
		t.Errorf("Expected 1 component node, got %d", len(g.ComponentNodes))
	}

	retrieved, exists := g.ComponentNodes[comp.ID]
	if !exists {
		t.Error("Component node not found in graph")
	}

	if retrieved.Name != "App" {
		t.Errorf("Expected name 'App', got '%s'", retrieved.Name)
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph()

	edge := Edge{
		ID:       "edge:defines:comp1:state1",
		SourceID: "comp1",
		TargetID: "state1",
		Type:     EdgeTypeDefines,
	}

	g.AddEdge(edge)

	if len(g.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(g.Edges))
	}

	// Check source index
	sourceEdges := g.GetOutgoingEdges("comp1")
	if len(sourceEdges) != 1 {
		t.Errorf("Expected 1 outgoing edge from comp1, got %d", len(sourceEdges))
	}

	// Check target index
	targetEdges := g.GetIncomingEdges("state1")
	if len(targetEdges) != 1 {
		t.Errorf("Expected 1 incoming edge to state1, got %d", len(targetEdges))
	}
}

func TestGetEdgesByType(t *testing.T) {
	g := NewGraph()

	g.AddEdge(Edge{ID: "e1", SourceID: "c1", TargetID: "s1", Type: EdgeTypeDefines})
	g.AddEdge(Edge{ID: "e2", SourceID: "c2", TargetID: "s1", Type: EdgeTypeConsumes})
	g.AddEdge(Edge{ID: "e3", SourceID: "c1", TargetID: "s2", Type: EdgeTypeDefines})

	definesEdges := g.GetEdgesByType(EdgeTypeDefines)
	if len(definesEdges) != 2 {
		t.Errorf("Expected 2 'defines' edges, got %d", len(definesEdges))
	}

	consumesEdges := g.GetEdgesByType(EdgeTypeConsumes)
	if len(consumesEdges) != 1 {
		t.Errorf("Expected 1 'consumes' edge, got %d", len(consumesEdges))
	}
}

func TestGetComponentDepth(t *testing.T) {
	g := NewGraph()

	// Build a simple tree: App -> Parent -> Child
	app := &ComponentNode{ID: "app", Name: "App", Parent: ""}
	parent := &ComponentNode{ID: "parent", Name: "Parent", Parent: "app"}
	child := &ComponentNode{ID: "child", Name: "Child", Parent: "parent"}

	g.AddComponentNode(app)
	g.AddComponentNode(parent)
	g.AddComponentNode(child)

	if depth := g.GetComponentDepth("app"); depth != 0 {
		t.Errorf("Expected depth 0 for App, got %d", depth)
	}

	if depth := g.GetComponentDepth("parent"); depth != 1 {
		t.Errorf("Expected depth 1 for Parent, got %d", depth)
	}

	if depth := g.GetComponentDepth("child"); depth != 2 {
		t.Errorf("Expected depth 2 for Child, got %d", depth)
	}
}

func TestFindPathBetweenComponents(t *testing.T) {
	g := NewGraph()

	// Build a tree: App -> Parent -> Child
	app := &ComponentNode{ID: "app", Name: "App", Parent: "", Children: []string{"parent"}}
	parent := &ComponentNode{ID: "parent", Name: "Parent", Parent: "app", Children: []string{"child"}}
	child := &ComponentNode{ID: "child", Name: "Child", Parent: "parent", Children: []string{}}

	g.AddComponentNode(app)
	g.AddComponentNode(parent)
	g.AddComponentNode(child)

	path := g.FindPathBetweenComponents("app", "child")
	if path == nil {
		t.Fatal("Expected a path, got nil")
	}

	if len(path) != 3 {
		t.Errorf("Expected path length 3, got %d", len(path))
	}

	if path[0] != "app" || path[1] != "parent" || path[2] != "child" {
		t.Errorf("Unexpected path: %v", path)
	}
}

func TestGenerateIDs(t *testing.T) {
	stateID1 := GenerateStateID("App", "count", "/path/to/file.tsx", 10)
	stateID2 := GenerateStateID("App", "count", "/path/to/file.tsx", 10)
	stateID3 := GenerateStateID("App", "count", "/path/to/file.tsx", 11) // Different line

	// Same inputs should generate same ID
	if stateID1 != stateID2 {
		t.Error("Same inputs generated different state IDs")
	}

	// Different inputs should generate different IDs
	if stateID1 == stateID3 {
		t.Error("Different inputs generated same state ID")
	}

	// Test component ID
	compID1 := GenerateComponentID("App", "/path/to/file.tsx", 5)
	compID2 := GenerateComponentID("App", "/path/to/file.tsx", 5)
	compID3 := GenerateComponentID("Dashboard", "/path/to/file.tsx", 5)

	if compID1 != compID2 {
		t.Error("Same inputs generated different component IDs")
	}

	if compID1 == compID3 {
		t.Error("Different inputs generated same component ID")
	}
}

func TestIsReactComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"PascalCase component", "App", true},
		{"PascalCase long name", "DashboardLayout", true},
		{"camelCase function", "handleClick", false},
		{"lowercase", "button", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReactComponent(tt.input)
			if result != tt.expected {
				t.Errorf("isReactComponent(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
