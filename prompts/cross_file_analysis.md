# Cross-File Analysis: ESLint vs. Custom Tool

**Date:** 2025-11-13
**Context:** Evaluating cross-file React component tree analysis for performance linting

---

## The Core Requirement

**Your Goal:** Track state/props being passed through React component tree across many files

**Example Use Case:**
```
File: App.tsx
  → passes { user, setUser } to Dashboard.tsx
    → passes setUser to UserProfile.tsx
      → passes setUser to EditButton.tsx (3 levels deep!)

Issue: EditButton re-renders when user changes, even though it only uses setUser
Detection: Requires analyzing 4 files and building component tree
```

This requires:
1. Parsing all files in the project
2. Resolving import/export relationships
3. Building a component hierarchy graph
4. Tracking prop flow through the tree
5. Detecting memoization boundaries
6. Identifying unnecessary re-renders

---

## ESLint's Cross-File Capabilities

### What ESLint CAN Do

**TypeScript Project Service (v8):**
```javascript
// eslint.config.js
export default {
  languageOptions: {
    parserOptions: {
      projectService: true,  // ← Enables cross-file type analysis
      tsconfigRootDir: import.meta.dirname,
    },
  },
};
```

**Capabilities:**
- ✅ Access TypeScript's type checker across entire project
- ✅ Resolve imports and get type information from other files
- ✅ Use same APIs as VS Code (consistent with editor)
- ✅ Support monorepos with project references

**Example: Type-Aware Rule**
```javascript
module.exports = {
  create(context) {
    return {
      JSXAttribute(node) {
        const services = context.sourceCode.parserServices;
        const checker = services.program.getTypeChecker();

        // Get type of prop from another file!
        const propType = checker.getTypeAtLocation(node.value);

        // Check if prop value is from import
        const symbol = propType.getSymbol();
        if (symbol) {
          const declarations = symbol.getDeclarations();
          // declarations[0].getSourceFile() ← Different file!
        }
      }
    };
  }
};
```

### What ESLint CANNOT Do (Well)

**1. Cache Invalidation Across Files**

**Problem:**
```
File A changes → File B imports A → ESLint doesn't know B needs re-analysis
Result: Stale lint results in editor until you restart ESLint server
```

From Microsoft VSCode ESLint issue #1774:
> "ESLint doesn't provide a way for editors to know about cross-file dependencies, such as type information. This results in files receiving out-of-date type information when files they import from are changed."

**Workaround:** Run "Restart ESLint Server" command manually

**2. Full Project Graph in Real-Time**

ESLint processes files independently (even with TypeScript project service):
- Each file gets its own AST
- Rule sees one file at a time
- No built-in "component graph" data structure
- You'd need to build and maintain graph yourself

**3. Performance at Scale**

Type-aware linting requires TypeScript to build the entire project:
- First run: Full TypeScript compile (slow)
- Subsequent runs: Incremental, but still TypeScript overhead
- Large projects: Can take 30s-2min for full lint

---

## Someone Already Built It: Perf Fiscal

**NPM:** `eslint-plugin-perf-fiscal`
**GitHub:** https://github.com/ruidosujeira/perf-linter

### What It Does

> "Perf Fiscal is the first performance linting toolkit to correlate multi-file signals in real time, using the TypeScript checker to understand components, props, and async flows across your entire project."

**Features:**
- ✅ Cross-file TypeScript analysis engine
- ✅ Understands React component tree across files
- ✅ Tracks prop types (function vs. object vs. literal)
- ✅ Guards React memoization
- ✅ Flags unstable props and dependency arrays
- ✅ Whole-project analyzer that indexes exports and memo wrappers

**Example Rules:**
- Detect unstable props passed to memoized components
- Find missing React.memo on expensive components
- Identify cascading re-renders across component boundaries

### How It Works

**Architecture (inferred from description):**
```
1. Index Phase (uses TypeScript Program API):
   - Parse all files in project
   - Build component registry
   - Identify memo wrappers
   - Extract prop signatures

2. Analysis Phase (during ESLint run):
   - Each file is linted
   - Rule accesses pre-built index
   - Correlates current file with imported components
   - Detects cross-file performance issues

3. Output:
   - ESLint diagnostics with fix suggestions
```

**Key Insight:** They pre-build a project-wide index, then use it during linting. This sidesteps ESLint's single-file limitation.

---

## ESLint Cross-File Analysis: The Hard Truth

### It's Possible But Painful

**Yes, you CAN build cross-file analysis in ESLint:**

```javascript
// Simplified architecture
const projectIndex = new Map(); // Global state (shared across all files)

module.exports = {
  create(context) {
    const filename = context.getFilename();

    return {
      Program(node) {
        // On first pass, build index
        if (!projectIndex.has('built')) {
          buildProjectIndex(context.parserServices.program);
        }

        // Use index to analyze current file
        const componentInfo = projectIndex.get(filename);
        const imports = resolveImports(node, projectIndex);

        // Detect cross-file issues
        checkPropFlow(componentInfo, imports);
      }
    };
  }
};

function buildProjectIndex(program) {
  // Use TypeScript Program API to get all source files
  const sourceFiles = program.getSourceFiles();

  for (const file of sourceFiles) {
    // Parse each file
    // Extract components
    // Build graph
  }

  projectIndex.set('built', true);
}
```

### The Problems

**1. ESLint's Architecture Fights You**

ESLint is designed for **isolated file analysis**:
- Each file processed independently
- No official "project-level state" API
- Rules can't easily share data across files
- Cache invalidation doesn't understand cross-file deps

**2. Performance Penalty**

Building project index on every ESLint run:
- First file: Scan entire project (expensive!)
- Subsequent files: Use cached index (but when to invalidate?)
- File change: Need to rebuild parts of index (complex logic)

**3. Editor Integration Issues**

The VSCode ESLint extension doesn't know about cross-file dependencies:
- Change file A → Only re-lints file A
- File B (which imports A) → Shows stale results
- User must manually restart ESLint server

**4. You're Fighting the Tool**

ESLint's design philosophy:
- Fast, incremental, per-file analysis
- Simple caching (file hash → lint result)
- Parallelizable (each file independent)

Your requirement:
- Slow, whole-project, cross-file analysis
- Complex caching (file + all dependencies)
- Sequential (need full graph first)

---

## Custom Tool Advantages for Cross-File Analysis

### Why Go Tool Makes Sense Now

Given your **core requirement** is cross-file analysis, a custom tool has major advantages:

**1. Architecture Designed for It**

```go
type ProjectAnalyzer struct {
    // First-class concepts
    componentGraph  *ComponentGraph
    dependencyGraph *DependencyGraph
    propFlowGraph   *PropFlowGraph
}

func (pa *ProjectAnalyzer) Analyze() {
    // Natural flow:
    // 1. Parse all files
    // 2. Build graphs
    // 3. Analyze graph for issues
    // 4. Return diagnostics
}
```

No fighting against single-file architecture.

**2. Efficient Incremental Updates**

```go
func (pa *ProjectAnalyzer) OnFileChange(path string) {
    // 1. Find affected nodes in graph
    affectedFiles := pa.dependencyGraph.GetDependents(path)

    // 2. Re-parse only affected files
    for _, file := range affectedFiles {
        pa.reparseFile(file)
    }

    // 3. Update graph edges
    pa.componentGraph.UpdateEdges(affectedFiles)

    // 4. Re-analyze affected subtree
    return pa.analyzeSubgraph(affectedFiles)
}
```

You control the dependency invalidation logic.

**3. Performance Control**

```go
// Parallel parsing
func (pa *ProjectAnalyzer) ParseProject(files []string) {
    results := make(chan *ParseResult, len(files))

    // Parse files in parallel
    for _, file := range files {
        go func(f string) {
            results <- pa.parseFile(f)
        }(file)
    }

    // Collect results
    for range files {
        result := <-results
        pa.addToGraph(result)
    }
}
```

No TypeScript compiler overhead, full control over concurrency.

**4. Rich Data Structures**

```go
type ComponentNode struct {
    Name       string
    FilePath   string
    Props      []PropInfo
    Children   []*ComponentNode
    Parent     *ComponentNode

    // Performance metadata
    IsMemoized         bool
    PropsAreStable     bool
    RenderComplexity   int

    // Cross-file tracking
    ImportedFrom       string
    ExportedAs         string
    UsedBy             []*ComponentNode
}

type PropFlow struct {
    Source      *ComponentNode
    Destination *ComponentNode
    PropName    string
    IsStable    bool
    Depth       int  // How many levels deep
}
```

Model exactly what you need without ESLint's AST limitations.

**5. Precise Diagnostics**

```go
func (pa *ProjectAnalyzer) DetectPropsDrilling() []Diagnostic {
    var issues []Diagnostic

    for _, flow := range pa.propFlowGraph.GetAllFlows() {
        if flow.Depth >= 3 {  // 3+ levels
            issues = append(issues, Diagnostic{
                Message: fmt.Sprintf(
                    "Prop '%s' drilled through %d components: %s → %s → ... → %s",
                    flow.PropName,
                    flow.Depth,
                    flow.Source.Name,
                    // intermediate components
                    flow.Destination.Name,
                ),
                // Include full path in diagnostic
                RelatedLocations: flow.GetAllLocations(),
            })
        }
    }

    return issues
}
```

Show user the ENTIRE prop flow across files.

---

## The Verdict: Custom Tool for Your Use Case

### Given Your Requirements

**Core need:** Build React component tree spanning many files to track prop/state flow

**Recommendation:** **Build the custom Go tool**

### Why ESLint Doesn't Fit

| Requirement | ESLint | Custom Go Tool |
|-------------|--------|----------------|
| Build component graph across files | ⚠️ Possible but painful | ✅ Natural fit |
| Track props through 3+ levels | ⚠️ Can do with hacks | ✅ First-class support |
| Incremental updates on file change | ❌ Poor cache invalidation | ✅ You control it |
| Performance (100k LOC) | ❌ TypeScript overhead | ✅ Optimized for it |
| Editor integration | ❌ Stale cross-file info | ✅ Build custom LSP |
| Data model | ❌ Fight against AST | ✅ Your domain model |

### What You'd Lose (vs. ESLint)

- ❌ Free IDE integration (need to build LSP)
- ❌ Familiar ecosystem (ESLint plugins)
- ❌ Auto-fix infrastructure (need to build)
- ❌ Configuration familiarity (.eslintrc)

**But:** These are solvable with time investment.

### What You'd Gain

- ✅ Architecture designed for cross-file analysis
- ✅ Full control over performance
- ✅ Rich domain model (component graph, prop flow)
- ✅ Precise diagnostics across file boundaries
- ✅ Incremental analysis that actually works
- ✅ Fast (Go + tree-sitter vs. Node.js + TypeScript)

---

## Hybrid Approach (Best of Both Worlds)

### Phase 1: ESLint Plugin for Single-File Rules

**Use ESLint for rules that don't need cross-file analysis:**

```javascript
// eslint-plugin-react-analyzer
module.exports = {
  rules: {
    'no-inline-props': singleFileRule,      // ✅ ESLint is great for this
    'exhaustive-deps': singleFileRule,      // ✅ Already exists!
    'no-unstable-deps': singleFileRule,     // ✅ Single-file detection
  }
};
```

**Time to value:** 2-3 weeks
**Coverage:** ~40% of performance issues

### Phase 2: Custom Go Tool for Cross-File Analysis

**Use Go tool for rules requiring component graph:**

```go
// react-analyzer CLI
rules := []Rule{
    &PropsDrillingRule{threshold: 3},           // ✅ Needs full graph
    &CascadingRerendersRule{},                  // ✅ Needs parent-child tracking
    &UnmemoizedChildOfInstableParentRule{},     // ✅ Cross-file memoization
}
```

**Time to value:** 6-12 weeks (but worth it for your use case)
**Coverage:** The other 60% (the hard stuff)

### Integration

**Output ESLint-compatible diagnostics from Go tool:**

```go
// Go tool outputs JSON in ESLint format
type ESLintDiagnostic struct {
    FilePath string   `json:"filePath"`
    Messages []struct {
        RuleId   string `json:"ruleId"`
        Severity int    `json:"severity"`
        Message  string `json:"message"`
        Line     int    `json:"line"`
        Column   int    `json:"column"`
    } `json:"messages"`
}
```

**Use from VS Code:**

```typescript
// VS Code extension can run both tools
const eslintResults = await runESLint(file);
const goToolResults = await runGoAnalyzer(projectRoot);

// Merge results
const allDiagnostics = [...eslintResults, ...goToolResults];
```

---

## Real-World Example: Props Drilling Detection

### ESLint Approach (Painful)

```javascript
// Attempt to detect props drilling in ESLint
const componentMap = new Map(); // Global state (fragile!)

module.exports = {
  meta: { /* ... */ },
  create(context) {
    return {
      Program(node) {
        // Need to parse ALL files first
        const program = context.parserServices?.program;
        if (!program) return;

        // Traverse all source files (expensive!)
        const sourceFiles = program.getSourceFiles();

        // Build component map
        for (const file of sourceFiles) {
          // ... complex traversal logic
        }

        // Now analyze current file's props
        const currentComponent = extractComponent(node);
        const propFlow = tracePropFlow(currentComponent, componentMap);

        if (propFlow.depth >= 3) {
          context.report({
            node,
            message: `Prop drilling detected: ${propFlow.path.join(' → ')}`
          });
        }
      }
    };
  }
};
```

**Problems:**
- Global state shared across files (race conditions?)
- Re-builds component map on EVERY file
- Expensive TypeScript API calls
- Cache invalidation nightmare

### Custom Tool Approach (Natural)

```go
type PropsDrillingAnalyzer struct {
    graph *ComponentGraph
}

func (a *PropsDrillingAnalyzer) Analyze() []Diagnostic {
    var issues []Diagnostic

    // Graph already built, just traverse it
    for _, component := range a.graph.GetAllComponents() {
        for _, prop := range component.Props {
            // Trace prop backwards to source
            path := a.tracePropToSource(prop, component)

            if len(path) >= 3 {
                issues = append(issues, Diagnostic{
                    RuleID:  "props-drilling",
                    Message: a.formatPropPath(path),
                    Locations: a.getLocationsForPath(path),
                    Suggestions: []CodeAction{
                        {
                            Title: "Use Context API",
                            Edits: a.generateContextRefactor(path),
                        },
                    },
                })
            }
        }
    }

    return issues
}

func (a *PropsDrillingAnalyzer) tracePropToSource(
    prop PropInfo,
    component *ComponentNode,
) []*ComponentNode {
    path := []*ComponentNode{component}
    current := component

    // Walk up the component tree
    for current.Parent != nil {
        current = current.Parent

        // Check if parent passes this prop
        if current.PassesProp(prop.Name) {
            path = append([]*ComponentNode{current}, path...)
        } else {
            break // Found the source
        }
    }

    return path
}
```

**Advantages:**
- Clean, readable code
- No global state hacks
- Efficient (graph already built)
- Easy to test

---

## Decision Matrix

### Use ESLint If:

- ✅ Most rules are single-file analysis
- ✅ Performance requirements are modest (<10k LOC)
- ✅ You want to ship in 2-3 weeks
- ✅ You're okay with limitations

### Use Custom Go Tool If:

- ✅ **Core requirement is cross-file analysis** ← **YOUR CASE**
- ✅ Need to build component/dependency graphs
- ✅ Track data flow across many files
- ✅ Performance at scale is critical (100k+ LOC)
- ✅ Can invest 3-6 months in development
- ✅ Want full control over analysis and caching

---

## Recommended Path Forward

### For Your Use Case (Cross-File Component Tree Analysis)

**Phase 0: Validate with Perf Fiscal (1 week)**

```bash
npm install --save-dev eslint-plugin-perf-fiscal

# Test it on your codebase
# See if it catches the issues you care about
# If yes → use it! Don't build custom tool.
# If no → proceed to Phase 1
```

**Phase 1: POC with Go + Tree-sitter (2-3 weeks)**

Build proof-of-concept:
- Parse React files
- Build component graph
- Implement ONE cross-file rule (props drilling)
- Validate performance

**Phase 2: MVP Custom Tool (8-12 weeks)**

If POC succeeds:
- Build full parser abstraction
- Implement component graph builder
- Add 3-4 cross-file rules
- Build VS Code LSP integration

**Phase 3: Hybrid Ecosystem (4-6 weeks)**

- Keep using ESLint for single-file rules
- Use custom tool for cross-file analysis
- Integrate both in VS Code extension

---

## Conclusion

**For your specific requirement (cross-file React component tree analysis), ESLint is the wrong tool.**

ESLint is excellent for:
- Single-file rules
- Fast, incremental linting
- Standard JS/TS best practices

But it's **not designed** for:
- Building project-wide graphs
- Tracking data flow across many files
- Complex dependency analysis

**Build the custom Go tool.** It's more work upfront, but it's the right architecture for your problem.

---

## Next Steps

1. ✅ Test Perf Fiscal (might solve your problem!)
2. ✅ If not, proceed with Go tool POC
3. ✅ Validate cross-file detection works
4. ✅ Decide: full custom tool or hybrid approach

**Want me to help with any of these steps?**
