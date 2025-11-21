# Contributing to React-Analyzer

Thank you for your interest in contributing! This guide will help you get started.

---

## Table of Contents
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Adding a New Rule](#adding-a-new-rule)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)

---

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Basic understanding of React
- Familiarity with AST concepts (helpful but not required)

### Quick Start
```bash
# Clone the repository
git clone https://github.com/rautio/react-analyzer.git
cd react-analyzer

# Build the project
go build cmd/react-analyzer/main.go

# Run tests
go test ./...

# Run on a test file
./main test/fixtures/prop-drilling/SimpleDrilling.tsx
```

---

## Development Setup

### Project Structure
```
react-analyzer/
├── cmd/
│   └── react-analyzer/
│       └── main.go              # CLI entry point
├── internal/
│   ├── analyzer/                # Module resolution
│   ├── cli/                     # CLI logic
│   ├── graph/                   # Component graph
│   ├── parser/                  # Tree-sitter wrapper
│   └── rules/                   # Analysis rules
├── test/
│   └── fixtures/                # Test files
└── docs/                        # Documentation
```

### Building
```bash
# Development build
go build cmd/react-analyzer/main.go

# Production build
go build -ldflags="-s -w" cmd/react-analyzer/main.go

# Run without building
go run cmd/react-analyzer/main.go <path>
```

### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/rules/

# With verbose output
go test -v ./internal/rules/

# With coverage
go test -cover ./...
```

---

## Project Architecture

### Core Components

#### 1. Parser (`internal/parser/`)
Wraps tree-sitter for TypeScript/JSX parsing.

**Key files:**
- `parser.go` - Main parser interface
- `ast.go` - AST wrapper with helper methods
- `node.go` - Node wrapper with React-specific utilities

**Example:**
```go
p, _ := parser.NewParser()
ast, _ := p.ParseFile("App.tsx", content)
defer ast.Close()

// Use helper methods
if node.IsHookCall() {
    hookName := node.ChildByFieldName("function").Text()
}
```

---

#### 2. Graph (`internal/graph/`)
Builds component dependency graph.

**Key files:**
- `types.go` - Data structure definitions
- `graph.go` - Graph operations (AddNode, FindPath, etc.)
- `builder.go` - 4-phase graph construction
- `prop_drilling.go` - Prop drilling detection algorithm

**Construction phases:**
1. `buildComponentNodes()` - Extract all components
2. `buildStateNodes()` - Extract all state (useState, etc.)
3. `buildComponentHierarchy()` - Parent-child relationships
4. `buildPropPassingEdges()` - Prop flow tracking

**Example:**
```go
builder := graph.NewBuilder(resolver)
g, _ := builder.Build()

// Use graph
path := g.FindPathBetweenComponents(startID, endID)
consumers := findPropConsumers(propID, g)
```

---

#### 3. Rules (`internal/rules/`)
Analysis rules that detect antipatterns.

**Key files:**
- `rule.go` - Rule interfaces
- `registry.go` - Rule registration
- Individual rule files (e.g., `no_object_deps.go`)

**Rule types:**
- **AST-based rules:** Analyze single files
- **Graph-based rules:** Analyze cross-component patterns

---

#### 4. Module Resolver (`internal/analyzer/`)
Resolves imports and tracks cross-file dependencies.

**Key files:**
- `module_resolver.go` - Import resolution
- `symbol_table.go` - Track memoization, exports

---

## Adding a New Rule

### Step 1: Design the Rule

**Define:**
1. **Name:** Kebab-case (e.g., `no-missing-keys`)
2. **What it detects:** Specific antipattern
3. **Why it matters:** Performance/correctness impact
4. **Recommendation:** How to fix

**Example:**
```
Name: no-missing-keys
Detects: Lists without key prop
Impact: Poor reconciliation performance
Fix: Add unique key prop to list items
```

---

### Step 2: Create the Rule File

Create `internal/rules/no_missing_keys.go`:

```go
package rules

import (
    "fmt"
    "github.com/rautio/react-analyzer/internal/analyzer"
    "github.com/rautio/react-analyzer/internal/parser"
)

// NoMissingKeys detects list items without key prop
type NoMissingKeys struct{}

func (r *NoMissingKeys) Name() string {
    return "no-missing-keys"
}

func (r *NoMissingKeys) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
    var issues []Issue

    ast.Root.Walk(func(node *parser.Node) bool {
        // Detect pattern: array.map(() => <Component />)
        if r.isMapWithoutKeys(node) {
            line, col := node.StartPoint()
            issues = append(issues, Issue{
                Rule:     r.Name(),
                Message:  "List item missing 'key' prop. Add unique key for each item.",
                FilePath: ast.FilePath,
                Line:     line + 1,
                Column:   col,
            })
        }
        return true // Continue walking
    })

    return issues
}

func (r *NoMissingKeys) isMapWithoutKeys(node *parser.Node) bool {
    // Implementation details...
    // Check if node is .map() call
    // Check if returned JSX has key prop
    return false
}
```

---

### Step 3: Add Tests

Create `internal/rules/no_missing_keys_test.go`:

```go
package rules

import (
    "testing"
    "github.com/rautio/react-analyzer/internal/parser"
)

func TestNoMissingKeys(t *testing.T) {
    tests := []struct {
        name          string
        code          string
        expectIssues  int
    }{
        {
            name: "missing key",
            code: `
                const items = [1, 2, 3];
                items.map(item => <div>{item}</div>);
            `,
            expectIssues: 1,
        },
        {
            name: "has key",
            code: `
                const items = [1, 2, 3];
                items.map(item => <div key={item}>{item}</div>);
            `,
            expectIssues: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p, _ := parser.NewParser()
            ast, _ := p.ParseFile("test.tsx", []byte(tt.code))
            defer ast.Close()

            rule := &NoMissingKeys{}
            issues := rule.Check(ast, nil)

            if len(issues) != tt.expectIssues {
                t.Errorf("Expected %d issues, got %d", tt.expectIssues, len(issues))
            }
        })
    }
}
```

---

### Step 4: Add Test Fixtures

Create `test/fixtures/missing-keys/HasKey.tsx`:
```tsx
const items = [1, 2, 3];
items.map(item => <div key={item}>{item}</div>);
```

Create `test/fixtures/missing-keys/MissingKey.tsx`:
```tsx
const items = [1, 2, 3];
items.map(item => <div>{item}</div>);  // ❌ Should detect
```

---

### Step 5: Register the Rule

Edit `internal/rules/registry.go`:

```go
func NewRegistry() *Registry {
    return &Registry{
        rules: []Rule{
            &NoObjectDeps{},
            &UnstablePropsToMemo{},
            &NoMissingKeys{},  // ← Add your rule here
            // ...
        },
    }
}
```

---

### Step 6: Add Documentation

Create `docs/rules/no-missing-keys.md`:

```markdown
# no-missing-keys

Detects list items rendered without a unique `key` prop.

## Why?

React uses keys to identify which items have changed, been added, or removed. Missing keys can cause:
- Poor reconciliation performance
- Unexpected behavior when list order changes
- State preservation issues

## Examples

❌ **Bad:**
\`\`\`tsx
items.map(item => <div>{item.name}</div>)
\`\`\`

✅ **Good:**
\`\`\`tsx
items.map(item => <div key={item.id}>{item.name}</div>)
\`\`\`

## Configuration

This rule has no configuration options.

## Related Rules

- no-array-index-key (detects using index as key)
```

---

### Step 7: Test Your Rule

```bash
# Run unit tests
go test ./internal/rules/ -run TestNoMissingKeys

# Build and test on fixture
go build cmd/react-analyzer/main.go
./main test/fixtures/missing-keys/MissingKey.tsx

# Should output:
# test/fixtures/missing-keys/MissingKey.tsx
#   2:1 - [no-missing-keys] List item missing 'key' prop...
```

---

### Step 8: Update VSCode Extension (Optional)

If your rule should display fix suggestions in the graph view, update `vscode-extension/src/views/GraphWebview.ts`:

```typescript
function getRuleInfo(ruleName) {
    const ruleInfoMap = {
        'no-missing-keys': {
            suggestion: 'Add a unique <code>key</code> prop to each item in the list. Use a stable identifier like <code>item.id</code>, not array index.',
            docsUrl: 'https://github.com/rautio/react-analyzer/blob/main/docs/rules/no-missing-keys.md'
        },
        // ... other rules
    };
    return ruleInfoMap[ruleName] || { suggestion: null, docsUrl: null };
}
```

---

## New Rule Checklist

Use this checklist when adding a new rule to ensure completeness:

### Required Files

- [ ] **Rule Implementation** (`internal/rules/your_rule_name.go`)
  - [ ] Implements `Name() string` method
  - [ ] Implements `Check(ast, resolver) []Issue` method
  - [ ] Implements `CheckGraph(graph) []Issue` if graph-based
  - [ ] Has descriptive comments explaining what it detects
  - [ ] Error messages are clear and actionable

- [ ] **Unit Tests** (`internal/rules/your_rule_name_test.go`)
  - [ ] Tests for true positives (should detect violation)
  - [ ] Tests for true negatives (should not detect)
  - [ ] Tests for edge cases
  - [ ] Minimum 3-5 test cases
  - [ ] All tests pass with `go test ./internal/rules/`

- [ ] **Test Fixtures** (`test/fixtures/your-rule-name/`)
  - [ ] At least one positive example (triggers rule)
  - [ ] At least one negative example (doesn't trigger rule)
  - [ ] Files are properly formatted and runnable
  - [ ] Include comments explaining what should be detected

- [ ] **Documentation** (`docs/rules/your-rule-name.md`)
  - [ ] Rule name and brief description
  - [ ] "Why This Matters" section explaining the problem
  - [ ] Minimum 2 "❌ Incorrect Code" examples
  - [ ] Minimum 2 "✅ Correct Code" examples with explanations
  - [ ] Configuration options (if any)
  - [ ] "When to Disable" section
  - [ ] Related rules section
  - [ ] References to React docs or articles

### Integration

- [ ] **Registry** (`internal/rules/registry.go`)
  - [ ] Rule added to `NewRegistry()` function
  - [ ] Rule appears in correct order (group with similar rules)

- [ ] **VSCode Extension** (`vscode-extension/src/views/GraphWebview.ts`)
  - [ ] Added to `getRuleInfo()` function with suggestion
  - [ ] Documentation URL points to correct GitHub path
  - [ ] Suggestion is concise (1-2 sentences max)

### Testing

- [ ] **Unit Tests**
  - [ ] `go test ./internal/rules/ -run TestYourRule` passes
  - [ ] Tests cover both positive and negative cases
  - [ ] No false positives or false negatives

- [ ] **Fixture Tests**
  - [ ] `./main test/fixtures/your-rule-name/` runs successfully
  - [ ] Detects expected violations
  - [ ] Doesn't produce unexpected violations

- [ ] **Real-World Testing**
  - [ ] Tested on at least one real React codebase
  - [ ] Verified no crashes or panics
  - [ ] Performance is acceptable (< 1s for small files)

### Documentation

- [ ] **Rule Documentation**
  - [ ] Follows template from existing rules
  - [ ] Examples are clear and realistic
  - [ ] Code examples are properly formatted
  - [ ] Grammar and spelling checked

- [ ] **Contributing Guide**
  - [ ] No updates needed (checklist covers everything)

- [ ] **CURRENT_STATE.md** (optional)
  - [ ] Update if rule adds new capability
  - [ ] Add to "Implemented Rules" section

### Code Quality

- [ ] **Go Standards**
  - [ ] Code formatted with `gofmt`
  - [ ] No linter warnings
  - [ ] Exported types have doc comments
  - [ ] Error messages follow project style (lowercase, no period)

- [ ] **Performance**
  - [ ] No obvious performance issues
  - [ ] Avoids redundant tree walking
  - [ ] Doesn't hold unnecessary references

### Validation Example

Here's a complete checklist for a hypothetical rule:

```
Rule: no-missing-keys

✅ internal/rules/no_missing_keys.go (185 lines)
✅ internal/rules/no_missing_keys_test.go (5 test cases)
✅ test/fixtures/missing-keys/HasKey.tsx
✅ test/fixtures/missing-keys/MissingKey.tsx
✅ test/fixtures/missing-keys/IndexAsKey.tsx
✅ docs/rules/no-missing-keys.md
✅ Registry updated (internal/rules/registry.go:318)
✅ VSCode extension updated (GraphWebview.ts:1022-1025)
✅ All tests pass (go test ./...)
✅ Tested on test/fixtures/ directory
✅ Tested on real React app (create-react-app)
✅ Code formatted (gofmt -w .)
✅ Documentation reviewed
```

---

## Graph-Based Rules

For rules that need cross-component analysis:

```go
type MyGraphRule struct{}

// Implement both Rule and GraphRule interfaces
func (r *MyGraphRule) Name() string {
    return "my-graph-rule"
}

// Check is called per-file
func (r *MyGraphRule) Check(ast *parser.AST, resolver *analyzer.ModuleResolver) []Issue {
    return nil // No per-file analysis
}

// CheckGraph is called after graph is built
func (r *MyGraphRule) CheckGraph(g *graph.Graph) []Issue {
    var issues []Issue

    // Analyze the full graph
    for _, comp := range g.ComponentNodes {
        // Cross-component analysis...
    }

    return issues
}
```

**Example:** See `internal/rules/deep_prop_drilling.go` for a complete graph-based rule.

---

## Testing

### Unit Tests
Test individual functions and components:

```go
func TestMyFunction(t *testing.T) {
    result := MyFunction(input)
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Tests
Test complete analysis flow:

```go
func TestAnalyzeFile(t *testing.T) {
    content, _ := os.ReadFile("test/fixtures/example.tsx")
    p, _ := parser.NewParser()
    ast, _ := p.ParseFile("example.tsx", content)

    rule := &MyRule{}
    issues := rule.Check(ast, nil)

    // Assert issues
}
```

### Fixture Tests
Test on real React code:

```bash
./main test/fixtures/prop-drilling/
```

---

## Code Style

### Go Conventions
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` to format code
- Write descriptive variable names
- Add comments for exported types and functions

### Project Conventions
- **File names:** Snake case (e.g., `no_object_deps.go`)
- **Rule names:** Kebab case (e.g., `no-object-deps`)
- **Test files:** `*_test.go`
- **Error messages:** Start with lowercase, no period

**Example error messages:**
```go
// ✅ Good
"prop 'theme' is drilled through 3 component levels"

// ❌ Bad
"Prop 'theme' is drilled through 3 component levels."
```

---

## Submitting Changes

### Before Submitting

1. **Run tests:**
   ```bash
   go test ./...
   ```

2. **Format code:**
   ```bash
   gofmt -w .
   ```

3. **Test on real code:**
   ```bash
   ./main test/fixtures/
   ```

4. **Update documentation:**
   - Add rule documentation if new rule
   - Update ROADMAP.md if relevant
   - Update CURRENT_STATE.md if adding capability

### Pull Request Process

1. **Fork and create branch:**
   ```bash
   git checkout -b feature/my-new-rule
   ```

2. **Make changes and commit:**
   ```bash
   git add .
   git commit -m "Add no-missing-keys rule"
   ```

3. **Push and open PR:**
   ```bash
   git push origin feature/my-new-rule
   ```
   Then open PR on GitHub

4. **PR template:**
   ```markdown
   ## Description
   Adds detection for list items without key prop

   ## Type of Change
   - [ ] Bug fix
   - [x] New rule
   - [ ] Enhancement
   - [ ] Documentation

   ## Testing
   - [x] Added unit tests
   - [x] Added test fixtures
   - [x] Tested on real codebase

   ## Checklist
   - [x] Tests pass
   - [x] Code formatted with gofmt
   - [x] Documentation updated
   ```

---

## Getting Help

- **Questions:** Open a GitHub discussion
- **Bugs:** Open a GitHub issue
- **Feature requests:** Check ROADMAP.md first, then discuss
- **Chat:** Join our Discord (coming soon)

---

## Good First Issues

Looking for something to work on? Check issues labeled `good-first-issue`:

**Example good first issues:**
- Add new test fixtures
- Improve error messages
- Add documentation
- Fix typos
- Enhance existing rules

---

## Architecture Deep Dive

### How Analysis Works

1. **CLI parses arguments** (`internal/cli/runner.go`)
2. **Find files to analyze** (directory scanning)
3. **For each file:**
   - Parse with tree-sitter → AST
   - Run AST-based rules
   - Add to module resolver
4. **Build component graph** (`internal/graph/builder.go`)
5. **Run graph-based rules** on complete graph
6. **Output results** grouped by file

### Key Design Decisions

**Why tree-sitter?**
- Fast, incremental parsing
- Battle-tested on millions of codebases
- Supports all JS/TS syntax

**Why build a graph?**
- Enables cross-component analysis
- Required for prop drilling detection
- Foundation for future features (re-render analysis, etc.)

**Why Go?**
- Fast compilation and execution
- Great tooling (testing, profiling)
- Easy distribution (single binary)
- Tree-sitter has Go bindings

---

## Resources

- **[ROADMAP.md](ROADMAP.md)** - Future plans
- **[CURRENT_STATE.md](CURRENT_STATE.md)** - What works today
- **[known_limitations.md](known_limitations.md)** - Current limitations
- **Design docs:** `docs/design/`
- **Rule docs:** `docs/rules/`

---

**Thank you for contributing!** 🎉

Every contribution, no matter how small, helps make React development better for everyone.
