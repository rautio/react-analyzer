# Phase 2: Graph Data Structures Design

**Document Version:** 2.0
**Date:** 2025-11-18
**Status:** ✅ IMPLEMENTED - Phase 2.1 Complete
**Last Updated:** 2025-11-18
**Author:** React Analyzer Team

---

## Overview

This document defines the core data structures for Phase 2's state dependency graph. The graph enables tracking state flow, component relationships, and architectural patterns across an entire React application.

---

## Core Design Principles

1. **Separation of Concerns:** State tracking separate from component structure
2. **Extensibility:** Easy to add new node/edge types for future features
3. **Performance:** Designed for lazy construction and incremental updates
4. **Clarity:** Each node type has a single, clear responsibility

---

## Data Structure Hierarchy

```
Graph
├── StateNodes (map[string]*StateNode)
├── ComponentNodes (map[string]*ComponentNode)
└── Edges ([]Edge)
```

---

## 1. StateNode

Represents a single piece of state in the application.

### Structure

```go
package graph

// StateType categorizes the origin of state
type StateType string

const (
    StateTypeUseState    StateType = "useState"
    StateTypeUseReducer  StateType = "useReducer"
    StateTypeContext     StateType = "context"
    StateTypeProp        StateType = "prop"
    StateTypeDerived     StateType = "derived"     // useMemo, computed value
    StateTypeRedux       StateType = "redux"       // Phase 2.4
    StateTypeZustand     StateType = "zustand"     // Phase 2.4
    StateTypeJotai       StateType = "jotai"       // Phase 2.4
    StateTypeRecoil      StateType = "recoil"      // Phase 2.4
)

// DataType categorizes the type of data stored
type DataType string

const (
    DataTypePrimitive DataType = "primitive"  // string, number, boolean
    DataTypeObject    DataType = "object"
    DataTypeArray     DataType = "array"
    DataTypeFunction  DataType = "function"
    DataTypeUnknown   DataType = "unknown"    // Can't determine statically
)

// Location represents a position in source code
type Location struct {
    FilePath   string
    Component  string  // Component name where this exists
    Line       int
    Column     int
}

// StateNode represents a piece of state in the application
type StateNode struct {
    // Identity
    ID   string     // Unique identifier: "{filePath}:{component}:{name}"
    Name string     // Variable name: "count", "user", "theme"

    // Classification
    Type     StateType
    DataType DataType

    // Location
    Location Location

    // State Characteristics
    Mutable       bool     // Can this state change? (useState: true, prop: depends)
    InitialValue  string   // Initial value expression (if determinable)

    // Relationships
    Dependencies  []string // IDs of StateNodes this depends on
    UpdatedBy     []string // IDs of StateNodes that trigger updates to this

    // Update Tracking
    UpdateLocations []Location  // Where this state is updated (setter calls)

    // Metadata
    Annotations map[string]string  // Extensible metadata
}
```

### Examples

**useState Example:**
```typescript
// File: src/components/Counter.tsx
const [count, setCount] = useState(0);
```

```go
StateNode{
    ID:       "src/components/Counter.tsx:Counter:count",
    Name:     "count",
    Type:     StateTypeUseState,
    DataType: DataTypePrimitive,
    Location: Location{
        FilePath:  "src/components/Counter.tsx",
        Component: "Counter",
        Line:      5,
        Column:    9,
    },
    Mutable:      true,
    InitialValue: "0",
    Dependencies: []string{},
    UpdatedBy:    []string{},
    UpdateLocations: []Location{
        {FilePath: "src/components/Counter.tsx", Line: 12},
    },
}
```

**Prop Example:**
```typescript
// File: src/components/UserProfile.tsx
function UserProfile({ user }: { user: User }) {
    // ...
}
```

```go
StateNode{
    ID:       "src/components/UserProfile.tsx:UserProfile:user",
    Name:     "user",
    Type:     StateTypeProp,
    DataType: DataTypeObject,
    Location: Location{
        FilePath:  "src/components/UserProfile.tsx",
        Component: "UserProfile",
        Line:      3,
        Column:    23,
    },
    Mutable:      false,  // Props are immutable in child
    InitialValue: "",     // Unknown - comes from parent
    Dependencies: []string{},  // Will be populated in Pass 2
    UpdatedBy:    []string{},
}
```

**Derived State Example:**
```typescript
// File: src/components/ShoppingCart.tsx
const total = useMemo(() => items.reduce((sum, item) => sum + item.price, 0), [items]);
```

```go
StateNode{
    ID:       "src/components/ShoppingCart.tsx:ShoppingCart:total",
    Name:     "total",
    Type:     StateTypeDerived,
    DataType: DataTypePrimitive,
    Location: Location{
        FilePath:  "src/components/ShoppingCart.tsx",
        Component: "ShoppingCart",
        Line:      8,
        Column:    9,
    },
    Mutable:      false,
    InitialValue: "",
    Dependencies: []string{
        "src/components/ShoppingCart.tsx:ShoppingCart:items",
    },
    UpdatedBy:    []string{},  // Derived state doesn't get updated directly
}
```

---

## 2. ComponentNode

Represents a React component and its role in the component tree.

### Structure

```go
// ComponentType categorizes the component
type ComponentType string

const (
    ComponentTypeFunction    ComponentType = "function"
    ComponentTypeClass       ComponentType = "class"
    ComponentTypeForwardRef  ComponentType = "forwardRef"
    ComponentTypeMemo        ComponentType = "memo"
)

// ComponentNode represents a React component
type ComponentNode struct {
    // Identity
    ID   string     // Unique identifier: "{filePath}:{componentName}"
    Name string     // Component name: "UserProfile", "App"

    // Classification
    Type ComponentType

    // Location
    Location Location

    // Optimization Status
    IsMemoized     bool   // Wrapped in React.memo
    HasMemoHooks   bool   // Uses useMemo/useCallback

    // State Management
    StateNodes     []string  // IDs of state defined in this component
    ConsumedState  []string  // IDs of state used but not defined here

    // Component Tree Relationships
    Parent   string    // ID of parent component (empty if root)
    Children []string  // IDs of child components

    // Props
    Props          []PropDefinition
    PropsPassedTo  map[string][]string  // childID -> list of prop names passed

    // Performance Metrics (populated during analysis)
    EstimatedRenderFrequency string  // "high", "medium", "low"
    RerenderTriggers         []string  // StateNode IDs that cause re-render

    // Metadata
    IsRoot         bool   // Is this a root component? (no parent)
    Annotations    map[string]string
}

// PropDefinition describes a prop accepted by a component
type PropDefinition struct {
    Name         string
    Type         string     // TypeScript type if available
    Required     bool
    DefaultValue string
}
```

### Examples

**Simple Component:**
```typescript
// File: src/components/Button.tsx
function Button({ label, onClick }: ButtonProps) {
    return <button onClick={onClick}>{label}</button>;
}
```

```go
ComponentNode{
    ID:   "src/components/Button.tsx:Button",
    Name: "Button",
    Type: ComponentTypeFunction,
    Location: Location{
        FilePath:  "src/components/Button.tsx",
        Component: "Button",
        Line:      3,
        Column:    10,
    },
    IsMemoized:     false,
    HasMemoHooks:   false,
    StateNodes:     []string{},
    ConsumedState:  []string{
        "src/components/Button.tsx:Button:label",    // Prop
        "src/components/Button.tsx:Button:onClick",  // Prop
    },
    Parent:   "",  // Will be populated when parent renders this
    Children: []string{},
    Props: []PropDefinition{
        {Name: "label", Type: "string", Required: true},
        {Name: "onClick", Type: "() => void", Required: true},
    },
    PropsPassedTo:            map[string][]string{},
    EstimatedRenderFrequency: "medium",
    RerenderTriggers:         []string{},
}
```

**Component with State:**
```typescript
// File: src/components/Counter.tsx
const Counter = memo(function Counter() {
    const [count, setCount] = useState(0);
    return <button onClick={() => setCount(count + 1)}>{count}</button>;
});
```

```go
ComponentNode{
    ID:   "src/components/Counter.tsx:Counter",
    Name: "Counter",
    Type: ComponentTypeMemo,
    Location: Location{
        FilePath:  "src/components/Counter.tsx",
        Component: "Counter",
        Line:      3,
        Column:    28,
    },
    IsMemoized:   true,
    HasMemoHooks: false,
    StateNodes: []string{
        "src/components/Counter.tsx:Counter:count",
    },
    ConsumedState: []string{
        "src/components/Counter.tsx:Counter:count",
    },
    Parent:   "",
    Children: []string{},
    Props:    []PropDefinition{},
    PropsPassedTo:            map[string][]string{},
    EstimatedRenderFrequency: "high",  // Re-renders on every count change
    RerenderTriggers: []string{
        "src/components/Counter.tsx:Counter:count",
    },
}
```

**Parent Component with Prop Drilling:**
```typescript
// File: src/components/App.tsx
function App() {
    const [theme, setTheme] = useState('dark');
    return <Layout theme={theme}><Content theme={theme} /></Layout>;
}
```

```go
ComponentNode{
    ID:   "src/components/App.tsx:App",
    Name: "App",
    Type: ComponentTypeFunction,
    Location: Location{
        FilePath:  "src/components/App.tsx",
        Component: "App",
        Line:      3,
        Column:    10,
    },
    IsMemoized:   false,
    HasMemoHooks: false,
    StateNodes: []string{
        "src/components/App.tsx:App:theme",
    },
    ConsumedState: []string{
        "src/components/App.tsx:App:theme",
    },
    Parent:   "",
    Children: []string{
        "src/components/Layout.tsx:Layout",
        "src/components/Content.tsx:Content",
    },
    Props: []PropDefinition{},
    PropsPassedTo: map[string][]string{
        "src/components/Layout.tsx:Layout":   {"theme"},
        "src/components/Content.tsx:Content": {"theme"},
    },
    EstimatedRenderFrequency: "low",
    RerenderTriggers: []string{
        "src/components/App.tsx:App:theme",
    },
}
```

---

## 3. Edge

Represents relationships between nodes in the graph.

### Structure

```go
// EdgeType categorizes the relationship
type EdgeType string

const (
    EdgeTypeDefines  EdgeType = "defines"   // Component defines state
    EdgeTypeConsumes EdgeType = "consumes"  // Component consumes state
    EdgeTypeUpdates  EdgeType = "updates"   // Component updates state
    EdgeTypePasses   EdgeType = "passes"    // Component passes state as prop
    EdgeTypeDerives  EdgeType = "derives"   // State derived from other state
    EdgeTypeRenders  EdgeType = "renders"   // Component renders child component
)

// Edge represents a directed relationship in the graph
type Edge struct {
    // Identity
    ID         string    // Unique identifier: "{sourceID}-{type}-{targetID}"

    // Endpoints
    SourceID   string    // Node ID (StateNode or ComponentNode)
    TargetID   string    // Node ID (StateNode or ComponentNode)

    // Classification
    Type       EdgeType

    // Weight/Importance
    Weight     float64   // 0.0-1.0, higher = more significant

    // Context
    PropName   string    // If type is "passes", what prop name is used?
    Location   Location  // Where this relationship occurs in code

    // Metadata
    Annotations map[string]string
}
```

### Examples

**Component Defines State:**
```go
Edge{
    ID:       "src/components/Counter.tsx:Counter-defines-src/components/Counter.tsx:Counter:count",
    SourceID: "src/components/Counter.tsx:Counter",
    TargetID: "src/components/Counter.tsx:Counter:count",
    Type:     EdgeTypeDefines,
    Weight:   1.0,  // Primary definition
    Location: Location{
        FilePath: "src/components/Counter.tsx",
        Line:     5,
    },
}
```

**Component Passes Prop:**
```go
Edge{
    ID:       "src/components/App.tsx:App-passes-src/components/Layout.tsx:Layout",
    SourceID: "src/components/App.tsx:App",
    TargetID: "src/components/Layout.tsx:Layout",
    Type:     EdgeTypePasses,
    Weight:   0.8,
    PropName: "theme",
    Location: Location{
        FilePath: "src/components/App.tsx",
        Line:     8,
    },
}
```

**State Derives from Other State:**
```go
Edge{
    ID:       "src/components/ShoppingCart.tsx:ShoppingCart:total-derives-src/components/ShoppingCart.tsx:ShoppingCart:items",
    SourceID: "src/components/ShoppingCart.tsx:ShoppingCart:total",
    TargetID: "src/components/ShoppingCart.tsx:ShoppingCart:items",
    Type:     EdgeTypeDerives,
    Weight:   1.0,
    Location: Location{
        FilePath: "src/components/ShoppingCart.tsx",
        Line:     8,
    },
}
```

---

## 4. Graph

The container for all nodes and edges.

### Structure

```go
// Graph represents the complete state dependency graph
type Graph struct {
    // Nodes
    StateNodes     map[string]*StateNode      // ID -> StateNode
    ComponentNodes map[string]*ComponentNode  // ID -> ComponentNode

    // Edges
    Edges []Edge

    // Indices for fast lookups
    EdgesBySource map[string][]Edge  // sourceID -> edges
    EdgesByTarget map[string][]Edge  // targetID -> edges
    EdgesByType   map[EdgeType][]Edge

    // Metadata
    ProjectRoot string
    AnalyzedAt  time.Time
    Version     string  // Graph schema version
}

// NewGraph creates an empty graph
func NewGraph(projectRoot string) *Graph {
    return &Graph{
        StateNodes:     make(map[string]*StateNode),
        ComponentNodes: make(map[string]*ComponentNode),
        Edges:          []Edge{},
        EdgesBySource:  make(map[string][]Edge),
        EdgesByTarget:  make(map[string][]Edge),
        EdgesByType:    make(map[EdgeType][]Edge),
        ProjectRoot:    projectRoot,
        AnalyzedAt:     time.Now(),
        Version:        "1.0.0",
    }
}
```

### Core Methods

```go
// AddStateNode adds a state node to the graph
func (g *Graph) AddStateNode(node *StateNode) {
    g.StateNodes[node.ID] = node
}

// AddComponentNode adds a component node to the graph
func (g *Graph) AddComponentNode(node *ComponentNode) {
    g.ComponentNodes[node.ID] = node
}

// AddEdge adds an edge and updates indices
func (g *Graph) AddEdge(edge Edge) {
    g.Edges = append(g.Edges, edge)

    // Update indices
    g.EdgesBySource[edge.SourceID] = append(g.EdgesBySource[edge.SourceID], edge)
    g.EdgesByTarget[edge.TargetID] = append(g.EdgesByTarget[edge.TargetID], edge)
    g.EdgesByType[edge.Type] = append(g.EdgesByType[edge.Type], edge)
}

// GetOutgoingEdges returns all edges from a node
func (g *Graph) GetOutgoingEdges(nodeID string) []Edge {
    return g.EdgesBySource[nodeID]
}

// GetIncomingEdges returns all edges to a node
func (g *Graph) GetIncomingEdges(nodeID string) []Edge {
    return g.EdgesByTarget[nodeID]
}

// GetEdgesByType returns all edges of a specific type
func (g *Graph) GetEdgesByType(edgeType EdgeType) []Edge {
    return g.EdgesByType[edgeType]
}

// GetComponentChildren returns all child components
func (g *Graph) GetComponentChildren(componentID string) []*ComponentNode {
    var children []*ComponentNode
    for _, edge := range g.GetOutgoingEdges(componentID) {
        if edge.Type == EdgeTypeRenders {
            if child, ok := g.ComponentNodes[edge.TargetID]; ok {
                children = append(children, child)
            }
        }
    }
    return children
}

// GetComponentState returns all state defined in a component
func (g *Graph) GetComponentState(componentID string) []*StateNode {
    var stateNodes []*StateNode
    for _, edge := range g.GetOutgoingEdges(componentID) {
        if edge.Type == EdgeTypeDefines {
            if state, ok := g.StateNodes[edge.TargetID]; ok {
                stateNodes = append(stateNodes, state)
            }
        }
    }
    return stateNodes
}

// TraceStatePropagation follows state through the component tree
func (g *Graph) TraceStatePropagation(stateID string) []*ComponentNode {
    affected := make(map[string]*ComponentNode)
    g.traceStatePropagationHelper(stateID, affected)

    result := make([]*ComponentNode, 0, len(affected))
    for _, comp := range affected {
        result = append(result, comp)
    }
    return result
}

func (g *Graph) traceStatePropagationHelper(nodeID string, affected map[string]*ComponentNode) {
    // Find all components that consume this state
    for _, edge := range g.GetIncomingEdges(nodeID) {
        if edge.Type == EdgeTypeConsumes {
            if comp, ok := g.ComponentNodes[edge.SourceID]; ok {
                if _, seen := affected[comp.ID]; !seen {
                    affected[comp.ID] = comp

                    // Recursively check components this one passes state to
                    for _, childEdge := range g.GetOutgoingEdges(comp.ID) {
                        if childEdge.Type == EdgeTypePasses {
                            g.traceStatePropagationHelper(childEdge.TargetID, affected)
                        }
                    }
                }
            }
        }
    }
}
```

---

## 5. ID Generation Strategy

Consistent ID generation is critical for graph integrity.

### ID Patterns

**StateNode ID:**
```
{filePath}:{componentName}:{stateName}
Example: "src/components/Counter.tsx:Counter:count"
```

**ComponentNode ID:**
```
{filePath}:{componentName}
Example: "src/components/Counter.tsx:Counter"
```

**Edge ID:**
```
{sourceID}-{edgeType}-{targetID}
Example: "src/components/App.tsx:App-passes-src/components/Layout.tsx:Layout"
```

### ID Generation Functions

```go
func GenerateStateNodeID(filePath, componentName, stateName string) string {
    return fmt.Sprintf("%s:%s:%s", filePath, componentName, stateName)
}

func GenerateComponentNodeID(filePath, componentName string) string {
    return fmt.Sprintf("%s:%s", filePath, componentName)
}

func GenerateEdgeID(sourceID string, edgeType EdgeType, targetID string) string {
    return fmt.Sprintf("%s-%s-%s", sourceID, edgeType, targetID)
}
```

---

## 6. Serialization

The graph must be serializable for caching and sharing.

### JSON Format

```go
// MarshalJSON serializes the graph to JSON
func (g *Graph) MarshalJSON() ([]byte, error) {
    type Alias Graph
    return json.MarshalIndent((*Alias)(g), "", "  ")
}

// UnmarshalJSON deserializes the graph from JSON
func (g *Graph) UnmarshalJSON(data []byte) error {
    type Alias Graph
    aux := (*Alias)(g)
    if err := json.Unmarshal(data, aux); err != nil {
        return err
    }

    // Rebuild indices
    g.rebuildIndices()
    return nil
}

func (g *Graph) rebuildIndices() {
    g.EdgesBySource = make(map[string][]Edge)
    g.EdgesByTarget = make(map[string][]Edge)
    g.EdgesByType = make(map[EdgeType][]Edge)

    for _, edge := range g.Edges {
        g.EdgesBySource[edge.SourceID] = append(g.EdgesBySource[edge.SourceID], edge)
        g.EdgesByTarget[edge.TargetID] = append(g.EdgesByTarget[edge.TargetID], edge)
        g.EdgesByType[edge.Type] = append(g.EdgesByType[edge.Type], edge)
    }
}

// SaveToFile saves the graph to a JSON file
func (g *Graph) SaveToFile(filepath string) error {
    data, err := g.MarshalJSON()
    if err != nil {
        return err
    }
    return os.WriteFile(filepath, data, 0644)
}

// LoadFromFile loads the graph from a JSON file
func LoadFromFile(filepath string) (*Graph, error) {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, err
    }

    var g Graph
    if err := g.UnmarshalJSON(data); err != nil {
        return nil, err
    }

    return &g, nil
}
```

---

## 7. Performance Considerations

### Lazy Construction

Don't build the entire graph upfront:

```go
type GraphBuilder struct {
    graph     *Graph
    analyzed  map[string]bool  // Track what's been analyzed
    pending   []string          // Files pending analysis
}

func (gb *GraphBuilder) AnalyzeComponent(filePath string) {
    if gb.analyzed[filePath] {
        return  // Already analyzed
    }

    // Parse and analyze
    // ...

    gb.analyzed[filePath] = true
}
```

### Incremental Updates

When files change, only update affected parts:

```go
func (g *Graph) InvalidateFile(filePath string) {
    // Remove all nodes from this file
    for id, node := range g.ComponentNodes {
        if node.Location.FilePath == filePath {
            delete(g.ComponentNodes, id)
        }
    }

    for id, node := range g.StateNodes {
        if node.Location.FilePath == filePath {
            delete(g.StateNodes, id)
        }
    }

    // Remove edges involving this file
    g.Edges = filterEdges(g.Edges, func(e Edge) bool {
        return !strings.HasPrefix(e.SourceID, filePath) &&
               !strings.HasPrefix(e.TargetID, filePath)
    })

    // Rebuild indices
    g.rebuildIndices()
}
```

---

## 8. Testing Strategy

### Unit Tests

Test individual node creation:

```go
func TestStateNode_Creation(t *testing.T) {
    node := &StateNode{
        ID:       "test.tsx:Component:count",
        Name:     "count",
        Type:     StateTypeUseState,
        DataType: DataTypePrimitive,
    }

    assert.Equal(t, "count", node.Name)
    assert.Equal(t, StateTypeUseState, node.Type)
}
```

### Integration Tests

Test graph operations:

```go
func TestGraph_PropagationTracking(t *testing.T) {
    g := NewGraph(".")

    // Create nodes
    parent := &ComponentNode{ID: "Parent.tsx:Parent"}
    child := &ComponentNode{ID: "Child.tsx:Child"}
    state := &StateNode{ID: "Parent.tsx:Parent:theme"}

    g.AddComponentNode(parent)
    g.AddComponentNode(child)
    g.AddStateNode(state)

    // Add edges
    g.AddEdge(Edge{
        SourceID: parent.ID,
        TargetID: state.ID,
        Type:     EdgeTypeDefines,
    })
    g.AddEdge(Edge{
        SourceID: parent.ID,
        TargetID: child.ID,
        Type:     EdgeTypePasses,
        PropName: "theme",
    })

    // Verify propagation
    affected := g.TraceStatePropagation(state.ID)
    assert.Len(t, affected, 1)
    assert.Equal(t, child.ID, affected[0].ID)
}
```

---

## 9. Next Steps

1. **Implement in `internal/graph/`** package
2. **Create test fixtures** representing realistic component hierarchies
3. **Integrate with existing analyzer** to populate graph
4. **Add visualization helpers** for debugging

---

## Appendix A: Example Graph JSON

```json
{
  "projectRoot": "/Users/oskari/repos/my-app",
  "analyzedAt": "2025-11-18T12:00:00Z",
  "version": "1.0.0",
  "stateNodes": {
    "src/App.tsx:App:theme": {
      "id": "src/App.tsx:App:theme",
      "name": "theme",
      "type": "useState",
      "dataType": "primitive",
      "location": {
        "filePath": "src/App.tsx",
        "component": "App",
        "line": 5,
        "column": 9
      },
      "mutable": true,
      "initialValue": "'dark'",
      "dependencies": [],
      "updatedBy": [],
      "updateLocations": [
        {
          "filePath": "src/App.tsx",
          "line": 12
        }
      ]
    }
  },
  "componentNodes": {
    "src/App.tsx:App": {
      "id": "src/App.tsx:App",
      "name": "App",
      "type": "function",
      "location": {
        "filePath": "src/App.tsx",
        "component": "App",
        "line": 3,
        "column": 10
      },
      "isMemoized": false,
      "hasMemoHooks": false,
      "stateNodes": ["src/App.tsx:App:theme"],
      "consumedState": ["src/App.tsx:App:theme"],
      "parent": "",
      "children": ["src/Layout.tsx:Layout"],
      "props": [],
      "propsPassedTo": {
        "src/Layout.tsx:Layout": ["theme"]
      }
    }
  },
  "edges": [
    {
      "id": "src/App.tsx:App-defines-src/App.tsx:App:theme",
      "sourceID": "src/App.tsx:App",
      "targetID": "src/App.tsx:App:theme",
      "type": "defines",
      "weight": 1.0,
      "location": {
        "filePath": "src/App.tsx",
        "line": 5
      }
    },
    {
      "id": "src/App.tsx:App-passes-src/Layout.tsx:Layout",
      "sourceID": "src/App.tsx:App",
      "targetID": "src/Layout.tsx:Layout",
      "type": "passes",
      "weight": 0.8,
      "propName": "theme",
      "location": {
        "filePath": "src/App.tsx",
        "line": 8
      }
    }
  ]
}
```

---

## Appendix B: Implementation Status

### ✅ Phase 2.1 Complete (2025-11-18)

All core data structures have been implemented and are working in production:

**StateNode:** ✅ Fully implemented
- All state types supported (useState, useReducer, Context, Props, Derived)
- Location tracking working
- Integration with graph complete

**ComponentNode:** ✅ Fully implemented
- Component hierarchy tracking working
- Prop extraction working (function declarations only)
- Memoization detection working
- PropsPassedTo tracking working

**Edge Types:** ✅ All implemented
- EdgeTypeDefines: Component defines state ✅
- EdgeTypePasses: Component passes prop to child ✅
- EdgeTypeConsumes: Component consumes state ✅
- EdgeTypeUpdates: Component updates state ✅
- EdgeTypeDerives: State derived from other state ✅

**Graph:** ✅ Fully implemented
- 4-phase construction pipeline working
- BFS path finding working
- Edge indexing for performance working
- Integration with module resolver working

**Implementation Details:**
- Location: `internal/graph/`
- Files: `types.go`, `graph.go`, `builder.go`, `prop_drilling.go`
- Test Coverage: Comprehensive
- Used By: 6 production rules including `deep-prop-drilling`

### 🚧 Known Limitations (Planned for Phase 2.2+)

**ComponentNode:**
- ❌ Arrow function components not detected yet
- ❌ Cross-file component resolution incomplete

**Edge Creation:**
- ❌ Spread operators not tracked
- ❌ Object property access not tracked
- ❌ Renamed props not tracked

**See:** [known_limitations.md](../known_limitations.md) for full details.

---

## Appendix C: Future Enhancements

### Phase 2.2: Core Completeness
- Arrow function component detection
- Cross-file edge creation
- Spread operator support
- Enhanced prop tracking

### Phase 2.3: Re-render Cascade Analysis
- Add `RerenderCascade` field to StateNode
- Track estimated re-render count and depth
- Detect unnecessary re-render chains

### Phase 2.4: Context Analysis
- Add `ContextNode` type
- Track Provider/Consumer relationships
- Detect Context drilling anti-patterns

### Phase 3: Global State Support
- Extend StateType for Redux/Zustand/Jotai/etc.
- Add `StoreNode` for global stores
- Track selector usage and dependencies

---

**End of Document**
**Last Updated:** 2025-11-18
**Status:** ✅ Phase 2.1 Complete
