# Phase 2: Prop Drilling Detection Algorithm

**Document Version:** 2.0
**Date:** 2025-11-18
**Status:** Phase 2.1 Complete - Planning Phase 2.2+
**Last Updated:** 2025-11-18

---

## Overview

This document defines the algorithm and implementation approach for detecting prop drilling in React applications. Prop drilling occurs when props are passed through 3+ component levels, with intermediate components not using the props themselves.

---

## What is Prop Drilling?

### Definition

**Prop drilling** is the anti-pattern of passing props through multiple component levels where intermediate components don't use the props - they only pass them down to descendants.

### Example

```typescript
// Level 0: App defines theme
function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;
}

// Level 1: Dashboard doesn't use theme, just passes it
function Dashboard({ theme }: { theme: string }) {
    return <div><Sidebar theme={theme} /></div>;
}

// Level 2: Sidebar doesn't use theme, just passes it
function Sidebar({ theme }: { theme: string }) {
    return <div><ThemeToggle theme={theme} /></div>;
}

// Level 3: ThemeToggle finally uses theme
function ThemeToggle({ theme }: { theme: string }) {
    return <button className={theme}>Toggle</button>;  // Uses theme!
}
```

**Drilling depth:** 3 levels (App → Dashboard → Sidebar → ThemeToggle)
**Passthrough components:** Dashboard, Sidebar (don't use `theme`)

### Why It's Problematic

1. **Tight coupling:** Intermediate components depend on props they don't need
2. **Maintenance burden:** Changing the prop requires updating all intermediate components
3. **Component reusability:** Intermediate components can't be used without the drilled prop
4. **Performance:** All intermediate components re-render when prop changes

### Solution

Use Context API:

```typescript
// Create context
const ThemeContext = createContext<string>('light');

// App provides theme
function App() {
    const [theme, setTheme] = useState('dark');
    return (
        <ThemeContext.Provider value={theme}>
            <Dashboard />
        </ThemeContext.Provider>
    );
}

// Intermediate components don't need theme prop
function Dashboard() {
    return <div><Sidebar /></div>;
}

function Sidebar() {
    return <div><ThemeToggle /></div>;
}

// ThemeToggle consumes theme directly
function ThemeToggle() {
    const theme = useContext(ThemeContext);
    return <button className={theme}>Toggle</button>;
}
```

---

## Detection Algorithm

### High-Level Algorithm

```
1. Build component hierarchy tree
2. For each prop in the tree:
   a. Find where prop is defined (origin)
   b. Find where prop is used (consumers)
   c. Trace path from origin to each consumer
   d. Count passthrough components
   e. If passthrough count >= 2, report violation
```

### Detailed Steps

#### Step 1: Build Component Tree

Using the graph structures from `phase2_graph_data_structures.md`:

```go
type ComponentTree struct {
    Root     *ComponentNode
    Nodes    map[string]*ComponentNode
    Edges    []ComponentEdge
}

type ComponentEdge struct {
    Parent   string  // ComponentNode ID
    Child    string  // ComponentNode ID
    Props    []PropPass
}

type PropPass struct {
    Name        string
    IsUsed      bool   // Does child actually use this prop?
    LineNumber  int
}
```

#### Step 2: Identify Prop Origins

For each component, find props that are defined (not received):

```go
func findPropOrigins(graph *Graph) map[string]*StateNode {
    origins := make(map[string]*StateNode)

    for _, stateNode := range graph.StateNodes {
        // Props that are defined as state are origins
        if stateNode.Type == StateTypeUseState ||
           stateNode.Type == StateTypeUseReducer ||
           stateNode.Type == StateTypeContext {
            origins[stateNode.ID] = stateNode
        }
    }

    return origins
}
```

#### Step 3: Identify Prop Consumers

For each prop, find components that actually use it (not just pass it):

```go
func findPropConsumers(propID string, graph *Graph) []*ComponentNode {
    consumers := []*ComponentNode{}

    // Find all components that consume this state
    for _, edge := range graph.GetIncomingEdges(propID) {
        if edge.Type == EdgeTypeConsumes {
            component := graph.ComponentNodes[edge.SourceID]

            // Check if component actually uses the prop (not just passes it)
            if componentUsesProp(component, propID, graph) {
                consumers = append(consumers, component)
            }
        }
    }

    return consumers
}

func componentUsesProp(comp *ComponentNode, propID string, graph *Graph) bool {
    // A component "uses" a prop if:
    // 1. It references the prop in its render logic (not just JSX attributes)
    // 2. It uses the prop in a hook dependency
    // 3. It passes the prop to a non-component function

    // Check if prop is in ConsumedState but not in PropsPassedTo
    for _, consumedState := range comp.ConsumedState {
        if consumedState == propID {
            // Check if it's only passed to children
            for _, propsToChild := range comp.PropsPassedTo {
                for _, propName := range propsToChild {
                    if propName == extractPropName(propID) {
                        // Prop is passed to child, check if also used locally
                        // ... (detailed implementation)
                    }
                }
            }
        }
    }

    return true  // Simplified - full implementation in code
}
```

#### Step 4: Trace Prop Path

Find the path from origin to consumer:

```go
type PropPath struct {
    Origin          *StateNode
    Consumer        *ComponentNode
    PassthroughPath []*ComponentNode  // Components in between
    Depth           int
}

func tracePropPath(origin *StateNode, consumer *ComponentNode, graph *Graph) *PropPath {
    path := &PropPath{
        Origin:          origin,
        Consumer:        consumer,
        PassthroughPath: []*ComponentNode{},
    }

    // BFS to find path
    queue := []struct {
        comp *ComponentNode
        path []*ComponentNode
    }{{comp: consumer, path: []*ComponentNode{}}}

    visited := make(map[string]bool)

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if visited[current.comp.ID] {
            continue
        }
        visited[current.comp.ID] = true

        // Check if this component's parent is the origin
        if current.comp.Parent != "" {
            parent := graph.ComponentNodes[current.comp.Parent]

            // Check if parent defines the state
            for _, stateID := range parent.StateNodes {
                if stateID == origin.ID {
                    // Found the path!
                    path.PassthroughPath = current.path
                    path.Depth = len(current.path)
                    return path
                }
            }

            // Continue BFS upward
            newPath := append([]*ComponentNode{parent}, current.path...)
            queue = append(queue, struct {
                comp *ComponentNode
                path []*ComponentNode
            }{comp: parent, path: newPath})
        }
    }

    return path
}
```

#### Step 5: Detect Violations

```go
type PropDrillingViolation struct {
    PropName           string
    Origin             Location
    Consumer           Location
    PassthroughCount   int
    PassthroughComponents []ComponentReference
    Depth              int
    Recommendation     string
}

type ComponentReference struct {
    Name     string
    FilePath string
    Line     int
}

func detectPropDrilling(graph *Graph) []PropDrillingViolation {
    violations := []PropDrillingViolation{}

    origins := findPropOrigins(graph)

    for propID, origin := range origins {
        consumers := findPropConsumers(propID, graph)

        for _, consumer := range consumers {
            path := tracePropPath(origin, consumer, graph)

            // Violation if drilling depth >= 3 (2+ passthrough components)
            if path.Depth >= 3 {
                violation := PropDrillingViolation{
                    PropName:           origin.Name,
                    Origin:             origin.Location,
                    Consumer:           consumer.Location,
                    PassthroughCount:   len(path.PassthroughPath) - 1,
                    PassthroughComponents: buildComponentReferences(path.PassthroughPath),
                    Depth:              path.Depth,
                    Recommendation:     generateRecommendation(origin, path),
                }

                violations = append(violations, violation)
            }
        }
    }

    return violations
}

func generateRecommendation(origin *StateNode, path *PropPath) string {
    // Simple recommendation for Phase 2.1
    return fmt.Sprintf(
        "Consider using Context API to avoid passing '%s' through %d component levels. " +
        "Create a context at %s and consume it directly in %s.",
        origin.Name,
        path.Depth,
        origin.Location.Component,
        path.Consumer.Name,
    )
}
```

---

## Implementation Phases

### ✅ Phase 2.1: Basic Detection (COMPLETED)

**Status:** ✅ SHIPPED (2025-11-18)
**Duration:** Weeks 1-4

**Delivered Features:**
- ✅ Detect drilling depth >= 3 levels
- ✅ Identify passthrough components (components that don't use props)
- ✅ Basic Context API recommendations
- ✅ Leaf consumer filtering (avoid duplicate violations)
- ✅ Per-prop edge tracking (multiple props between same components)
- ✅ 12 comprehensive test fixtures

**Actual Output Example:**

```
/Users/oskari/repos/react-analyzer/test/fixtures/prop-drilling/SimpleDrilling.tsx
  8:31 - [deep-prop-drilling] Prop 'count' is drilled through 4 component levels (Parent → Child). Consider using Context API to avoid passing 'count' through 4 component levels. Create a context at App and consume it directly in Display.

✖ Found 1 issue in 1 file
Analyzed 1 file in 1ms
```

**Known Limitations (addressed in Phase 2.2+):**
- Only works within single files (not cross-file)
- Only detects function declaration components (not arrow functions)
- Doesn't handle prop spreads (`{...props}`)
- Doesn't track object property access (`settings.locale`)
- False positives for partial usage (component uses AND passes prop)

**See:** [CURRENT_STATE.md](../CURRENT_STATE.md) for full details on what's working.

---

### 🚀 Phase 2.2: Core Completeness (PLANNED)

**Status:** 📋 Planned for Q1 2026
**Duration:** 4-6 weeks
**Priority:** CRITICAL - These gaps block real-world adoption

**Target Capabilities:**

#### 2.2.A: Arrow Function Components (1-2 weeks)
- Detect `const MyComponent = () => <div />`
- Extract props from arrow function parameters
- Support React.memo wrapping: `React.memo(() => <div />)`
- Update all rules to work with arrow components

**Impact:** Unblocks 60-80% of modern React codebases

#### 2.2.B: Cross-File Prop Drilling (2-3 weeks)
- Enhance `findComponentID()` to search imported files
- Track import statements and resolve component definitions
- Build cross-file component edges in graph
- Handle circular dependencies and re-exports

**Impact:** Makes detection work for real-world multi-file apps

#### 2.2.C: Prop Spread Operators (2-3 weeks)
- Detect `jsx_spread_attribute` nodes
- Track props included in spreads
- Create edges for all props in spread object
- Handle partial spreads: `<Child {...rest} theme={theme} />`

**Impact:** Handles very common React pattern

**See:** [ROADMAP.md - Phase 2.2](../ROADMAP.md#phase-22-core-completeness-4-6-weeks) for detailed plans.

---

### 🎯 Phase 2.3: Enhanced Accuracy (PLANNED)

**Status:** 📋 Planned for Q1-Q2 2026
**Duration:** 4-5 weeks

**Additional capabilities:**
- Detect when component both uses AND passes prop (reduce false positives)
- Track object property access (`settings.locale`)
- Detect multiple props drilled together (suggest combining into context)
- Enhanced recommendations based on usage patterns

**See:** [ROADMAP.md - Phase 2.3](../ROADMAP.md#phase-23-enhanced-accuracy-4-5-weeks) for details.

---

### 🔧 Phase 2.4: Advanced Patterns (PLANNED)

**Status:** 📋 Planned for Q2 2026
**Duration:** 3-5 weeks

**Additional capabilities:**
- Track prop renames along the path
- Detect existing Context providers (improve recommendations)
- Track prop transformations along the path
- Suggest specific Context patterns based on use case

**See:** [ROADMAP.md - Phase 2.4](../ROADMAP.md#phase-24-advanced-patterns-3-5-weeks) for details.

---

## AST Patterns to Detect

### 1. Component Prop Definitions

**Function Components:**
```typescript
// Destructured props
function Dashboard({ theme }: Props) { }

// Props object
function Dashboard(props: Props) { }

// Inline type
function Dashboard({ theme }: { theme: string }) { }
```

**AST Pattern:**
```
function_declaration
  name: identifier "Dashboard"
  parameters
    parameter
      pattern: object_pattern
        shorthand_property_identifier_pattern "theme"
```

### 2. Prop Usage in Render

**Direct usage:**
```typescript
<div className={theme}>
```

**AST Pattern:**
```
jsx_expression
  identifier "theme"
```

**Passthrough:**
```typescript
<Sidebar theme={theme} />
```

**AST Pattern:**
```
jsx_element
  jsx_opening_element
    jsx_attribute
      property_identifier "theme"
      jsx_expression
        identifier "theme"
```

### 3. Prop Usage in Hooks

```typescript
useEffect(() => {
    console.log(theme);
}, [theme]);
```

**AST Pattern:**
```
call_expression
  function: identifier "useEffect"
  arguments
    arrow_function
      body: ... references to "theme"
    array
      identifier "theme"
```

---

## Rule Implementation

### File: `internal/rules/deep_prop_drilling.go`

```go
package rules

import (
    "fmt"
    "github.com/rautio/react-analyzer/internal/analyzer"
    "github.com/rautio/react-analyzer/internal/graph"
    "github.com/rautio/react-analyzer/internal/parser"
)

type DeepPropDrilling struct {
    graph *graph.Graph
}

func (r *DeepPropDrilling) Name() string {
    return "deep-prop-drilling"
}

func (r *DeepPropDrilling) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
    // This rule operates on the graph, not individual ASTs
    // It's called once after the graph is built
    return nil
}

// CheckGraph is called after graph construction
func (r *DeepPropDrilling) CheckGraph(g *graph.Graph) []Issue {
    r.graph = g
    violations := r.detectPropDrilling()

    issues := []Issue{}
    for _, v := range violations {
        issues = append(issues, Issue{
            Rule:     r.Name(),
            FilePath: v.Origin.FilePath,
            Line:     v.Origin.Line,
            Column:   v.Origin.Column,
            Message:  r.formatMessage(v),
        })
    }

    return issues
}

func (r *DeepPropDrilling) detectPropDrilling() []graph.PropDrillingViolation {
    // Implementation from algorithm above
    return graph.DetectPropDrilling(r.graph)
}

func (r *DeepPropDrilling) formatMessage(v graph.PropDrillingViolation) string {
    path := ""
    for i, comp := range v.PassthroughComponents {
        if i > 0 {
            path += " → "
        }
        path += comp.Name
    }

    return fmt.Sprintf(
        "Prop '%s' is drilled through %d component levels (%s). %s",
        v.PropName,
        v.Depth,
        path,
        v.Recommendation,
    )
}
```

---

## Test Cases

### Test Fixture 1: Simple Drilling

**File: `test/fixtures/prop-drilling/SimpleDrilling.tsx`**

```typescript
// Level 0: Origin
function App() {
    const [count, setCount] = useState(0);
    return <Parent count={count} />;
}

// Level 1: Passthrough
function Parent({ count }: { count: number }) {
    return <Child count={count} />;
}

// Level 2: Passthrough
function Child({ count }: { count: number }) {
    return <Display count={count} />;
}

// Level 3: Consumer
function Display({ count }: { count: number }) {
    return <div>{count}</div>;  // Uses count
}
```

**Expected:** 1 violation (depth: 3, passthrough: Parent, Child)

### Test Fixture 2: No Drilling (Direct Usage)

```typescript
function App() {
    const [count, setCount] = useState(0);
    return <Parent count={count} />;
}

function Parent({ count }: { count: number }) {
    return <div>{count}</div>;  // Uses count directly
}
```

**Expected:** 0 violations

### Test Fixture 3: Multiple Props Drilled

```typescript
function App() {
    const [theme, setTheme] = useState('dark');
    const [lang, setLang] = useState('en');
    return <Parent theme={theme} lang={lang} />;
}

function Parent({ theme, lang }: Props) {
    return <Child theme={theme} lang={lang} />;
}

function Child({ theme, lang }: Props) {
    return <Display theme={theme} lang={lang} />;
}

function Display({ theme, lang }: Props) {
    return <div className={theme}>{lang}</div>;
}
```

**Expected:** 2 violations (one for theme, one for lang)

### Test Fixture 4: Partial Usage (Mixed)

```typescript
function App() {
    const [theme, setTheme] = useState('dark');
    return <Parent theme={theme} />;
}

function Parent({ theme }: { theme: string }) {
    // Uses theme locally AND passes it
    const styles = { background: theme };
    return <div style={styles}><Child theme={theme} /></div>;
}

function Child({ theme }: { theme: string }) {
    return <Display theme={theme} />;
}

function Display({ theme }: { theme: string }) {
    return <div className={theme}></div>;
}
```

**Expected:** 0 violations (Parent uses theme, so it's not pure passthrough)

---

## Performance Considerations

### Graph Construction

Prop drilling detection requires the full component graph:
- Build graph lazily (only when needed)
- Cache graph between analyses
- Invalidate cache on file changes

### Optimization Strategies

1. **Incremental Analysis:** Only re-analyze affected subtrees when files change
2. **Parallel Processing:** Analyze different prop paths concurrently
3. **Early Termination:** Stop tracing if depth > 10 (extremely deep drilling is rare)

---

## VS Code Integration

### Tree View Display

Show drilling violations in component tree:

```
📦 App
  ↳ ⚠️ Parent (passes 'theme' without using)
     ↳ ⚠️ Child (passes 'theme' without using)
        ↳ Display (uses 'theme')
```

### Code Actions (Quick Fixes)

When user clicks on violation:

```typescript
// Quick Fix: Extract to Context
vscode.languages.registerCodeActionsProvider('typescriptreact', {
    provideCodeActions(document, range, context) {
        if (isDrillingViolation(context.diagnostics)) {
            return [{
                title: "Extract to Context API",
                kind: vscode.CodeActionKind.QuickFix,
                command: {
                    command: 'react-analyzer.extractToContext',
                    arguments: [violation]
                }
            }];
        }
    }
});
```

---

## Completed Work (Phase 2.1)

1. ✅ **Graph building implemented** - Full 4-phase pipeline
2. ✅ **Prop tracing implemented** - BFS-based path finding
3. ✅ **Test fixtures created** - 12 comprehensive scenarios
4. ✅ **Rule added to registry** - `deep-prop-drilling` rule
5. ✅ **Leaf consumer filtering** - Avoids duplicate violations
6. ✅ **Per-prop edge tracking** - Multiple props between components

## Next Steps (Phase 2.2+)

See [ROADMAP.md](../ROADMAP.md) for detailed plans.

**Immediate Next (Phase 2.2 - Q1 2026):**
1. Arrow function component detection
2. Cross-file prop drilling
3. Prop spread operator support

**Future (Phase 2.3+ - Q2 2026):**
4. VS Code extension with visualization
5. Enhanced accuracy (reduce false positives)
6. Advanced pattern detection

---

**End of Document**
**Last Updated:** 2025-11-18
**Status:** Phase 2.1 Complete ✅
