# React Analyzer - Detailed Implementation Plan

## Overview

This plan builds on the successful POC (tree-sitter + Go) to create a production-ready static analysis tool. The implementation is structured in phases, with each phase delivering working features and comprehensive tests.

## Phase 0: Foundation & POC Validation ✅ COMPLETED

**Duration**: Completed (2.5 hours)
**Status**: ✅ Success - All performance targets exceeded

### Deliverables
- ✅ tree-sitter integration working
- ✅ Basic exhaustive-deps rule (100% accuracy)
- ✅ Performance validation (12,500 files/sec)
- ✅ Project structure established

---

## System Architecture & Technical Design

### Overall Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER INTERFACES                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ VS Code Ext  │  │  CLI Tool    │  │  LSP Server  │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────────┐
│                      ANALYSIS COORDINATOR                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  • Project Discovery                                      │   │
│  │  • File Queue Management                                  │   │
│  │  • Dependency Ordering                                    │   │
│  │  • Result Aggregation                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬─────────────────────────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
┌─────────▼────────┐  ┌──────▼─────────┐  ┌───▼──────────┐
│  PARSER LAYER    │  │  ANALYZER      │  │  RULE ENGINE │
│                  │  │  LAYER         │  │              │
│ ┌──────────────┐ │  │ ┌────────────┐ │  │ ┌──────────┐ │
│ │ tree-sitter  │ │  │ │ Scope      │ │  │ │ Registry │ │
│ │   Parser     │ │  │ │ Analysis   │ │  │ └──────────┘ │
│ └──────────────┘ │  │ └────────────┘ │  │              │
│                  │  │                │  │ ┌──────────┐ │
│ ┌──────────────┐ │  │ ┌────────────┐ │  │ │ Executor │ │
│ │  AST Cache   │ │  │ │ Data Flow  │ │  │ │ (Worker  │ │
│ │  (Incr.)     │ │  │ │ Analysis   │ │  │ │  Pool)   │ │
│ └──────────────┘ │  │ └────────────┘ │  │ └──────────┘ │
│                  │  │                │  │              │
│ ┌──────────────┐ │  │ ┌────────────┐ │  │ ┌──────────┐ │
│ │ Semantic     │ │  │ │ Import     │ │  │ │ Rules    │ │
│ │ Layer        │ │  │ │ Resolver   │ │  │ │ (3 MVP)  │ │
│ └──────────────┘ │  │ └────────────┘ │  │ └──────────┘ │
└──────────────────┘  │                │  └──────────────┘
                      │ ┌────────────┐ │
                      │ │ Component  │ │
                      │ │ Graph      │ │
                      │ └────────────┘ │
                      └────────────────┘
                             │
                   ┌─────────┴─────────┐
                   │                   │
          ┌────────▼────────┐  ┌───────▼──────────┐
          │  DIAGNOSTIC     │  │  AUTO-FIX        │
          │  SYSTEM         │  │  GENERATOR       │
          │                 │  │                  │
          │ • Location      │  │ • Code Actions   │
          │ • Severity      │  │ • Validation     │
          │ • Messages      │  │ • Preview        │
          └─────────────────┘  └──────────────────┘
```

### Core Data Structures

#### 1. AST Representation (`internal/ast/types.go`)

```go
// Node wraps tree-sitter node with additional metadata
type Node struct {
    tsNode    *sitter.Node      // Underlying tree-sitter node
    parent    *Node              // Parent node (for traversal)
    metadata  *NodeMetadata      // Cached analysis results
    source    []byte             // Source code reference
}

type NodeMetadata struct {
    // Semantic flags (computed lazily)
    isComponent       *bool
    isMemoized        *bool
    isHookCall        *bool

    // Scope information
    scopeID           ScopeID
    definedSymbols    []Symbol
    referencedSymbols []Symbol

    // React-specific
    hookDependencies  []string
    jsxProps          []PropNode
}

// AST represents a fully parsed file
type AST struct {
    root       *Node
    filePath   string
    language   Language
    tree       *sitter.Tree        // tree-sitter tree (for incremental updates)
    source     []byte

    // Caches
    nodeCache  map[uint32]*Node    // Cache nodes by tree-sitter ID
    scopeTree  *ScopeTree          // Scope hierarchy
    symbols    *SymbolTable        // All symbols in file
    imports    []ImportDecl        // Import declarations
    exports    []ExportDecl        // Export declarations

    // Metadata
    parseTime  time.Duration
    lastEdit   *Edit               // For incremental parsing
}

// Edit represents a source code change
type Edit struct {
    StartByte  uint32
    OldEndByte uint32
    NewEndByte uint32
    StartPoint Point
    OldEndPoint Point
    NewEndPoint Point
}

type Point struct {
    Row    uint32
    Column uint32
}
```

#### 2. Scope Analysis (`internal/analyzer/scope.go`)

```go
// ScopeTree represents the lexical scope hierarchy
type ScopeTree struct {
    root   *Scope
    scopes map[ScopeID]*Scope
}

type Scope struct {
    id       ScopeID
    kind     ScopeKind  // Global, Function, Block, etc.
    parent   *Scope
    children []*Scope

    // Symbol tracking
    symbols  map[string]*Symbol
    node     *ast.Node   // AST node that defines this scope

    // React-specific
    isComponent bool
    hookCalls   []*HookCall
}

type ScopeKind int
const (
    ScopeGlobal ScopeKind = iota
    ScopeModule
    ScopeFunction
    ScopeBlock
    ScopeClass
)

type Symbol struct {
    name       string
    kind       SymbolKind  // Var, Const, Let, Function, Class, Import
    scope      *Scope

    // Definition
    defNode    *ast.Node
    defRange   SourceRange

    // Usage tracking
    references []*Reference

    // Type information (optional, for future)
    typeInfo   *TypeInfo

    // Stability analysis
    isStable   *bool       // Cached: doesn't change between renders
    stability  Stability   // Stable, Unstable, Conditional, Unknown
}

type SymbolKind int
const (
    SymbolVar SymbolKind = iota
    SymbolConst
    SymbolLet
    SymbolFunction
    SymbolClass
    SymbolImport
    SymbolParam
)

type Reference struct {
    node    *ast.Node
    scope   *Scope
    kind    ReferenceKind  // Read, Write, ReadWrite
}

type Stability int
const (
    StabilityUnknown Stability = iota
    StabilityStable       // Never changes (const outside component, useMemo, etc.)
    StabilityUnstable     // Changes every render (inline object/array/function)
    StabilityConditional  // Depends on deps (useCallback/useMemo with deps)
    StabilityParam        // Function parameter (depends on caller)
)
```

#### 3. Data Flow Analysis (`internal/analyzer/dataflow.go`)

```go
// DataFlowGraph tracks how values flow through the program
type DataFlowGraph struct {
    nodes map[*ast.Node]*DFNode
    edges []*DFEdge
}

type DFNode struct {
    astNode   *ast.Node
    value     *Value
    kind      DFNodeKind
}

type DFNodeKind int
const (
    DFLiteral DFNodeKind = iota  // Object/array literal
    DFFunction                     // Function expression
    DFIdentifier                   // Variable reference
    DFMemberAccess                 // obj.prop
    DFCallExpression               // func()
    DFJSXElement
)

type Value struct {
    kind      ValueKind
    stability Stability

    // Tracking
    origin    *ast.Node       // Where value is created
    deps      []*Symbol       // What symbols this value depends on
}

type ValueKind int
const (
    ValueObject ValueKind = iota
    ValueArray
    ValueFunction
    ValuePrimitive
    ValueJSXElement
    ValueUnknown
)

type DFEdge struct {
    from  *DFNode
    to    *DFNode
    kind  DFEdgeKind
}

type DFEdgeKind int
const (
    DFAssignment DFEdgeKind = iota
    DFPropAccess
    DFReturn
    DFArgument
)
```

#### 4. Component Graph (`internal/analyzer/component_graph.go`)

```go
// ComponentGraph represents the component hierarchy and relationships
type ComponentGraph struct {
    components map[string]*Component
    edges      []*ComponentEdge

    // Reverse lookup
    fileToComponents map[string][]*Component
}

type Component struct {
    name      string
    filePath  string
    node      *ast.Node

    // Component metadata
    isMemoized   bool
    isForwardRef bool

    // Props
    props        []Prop
    propTypes    map[string]*PropType  // For future type analysis

    // Hooks used
    hooks        []*HookCall

    // Children (components this renders)
    renders      []*ComponentUsage

    // Parents (components that render this)
    renderedBy   []*ComponentUsage
}

type Prop struct {
    name      string
    node      *ast.Node
    required  bool

    // For destructured props
    isDestructured bool
    defaultValue   *ast.Node
}

type ComponentUsage struct {
    component   *Component  // The component being rendered
    callSite    *ast.Node   // JSX element node
    props       []PropUsage // Props passed at this call site
}

type PropUsage struct {
    name     string
    value    *Value           // Data flow value
    node     *ast.Node
    stability Stability
}

type HookCall struct {
    name         string
    node         *ast.Node
    dependencies []string  // For useEffect, useMemo, useCallback

    // Result tracking (for stability analysis)
    resultSymbol *Symbol
}

type ComponentEdge struct {
    from  *Component
    to    *Component
    props []PropUsage
}
```

#### 5. Import Resolution System (`internal/analyzer/imports.go`)

```go
// ImportResolver handles module resolution
type ImportResolver struct {
    projectRoot  string
    nodeModules  []string        // Paths to node_modules
    tsConfig     *TSConfig       // tsconfig.json paths
    packageJSON  *PackageJSON

    // Caches
    resolveCache map[string]string           // import specifier -> absolute path
    moduleCache  map[string]*ResolvedModule  // absolute path -> module info

    mu sync.RWMutex
}

type ResolvedModule struct {
    absPath    string
    ast        *ast.AST
    exports    []Export

    // Metadata
    isNodeModule bool
    packageName  string
}

type Export struct {
    kind    ExportKind
    name    string      // Named export name or "default"
    node    *ast.Node
    symbol  *Symbol
}

type ExportKind int
const (
    ExportNamed ExportKind = iota
    ExportDefault
    ExportNamespace
)

type ImportDecl struct {
    source      string        // "./Component" or "react"
    specifiers  []ImportSpec
    node        *ast.Node

    // Resolved
    resolved    *ResolvedModule
}

type ImportSpec struct {
    kind      ImportSpecKind
    imported  string  // Name in source module
    local     string  // Name in this file
    symbol    *Symbol // Local symbol
}

type ImportSpecKind int
const (
    ImportNamed ImportSpecKind = iota      // import { Foo }
    ImportDefault                           // import Foo
    ImportNamespace                         // import * as Foo
)

// TSConfig represents relevant tsconfig.json settings
type TSConfig struct {
    baseURL string
    paths   map[string][]string  // Path mappings: "@/*" -> ["src/*"]
}
```

### Parser Implementation Details

#### Incremental Parsing Strategy

```go
// ParseManager handles parsing lifecycle
type ParseManager struct {
    parser   *sitter.Parser
    language *sitter.Language

    // Document tracking
    documents map[string]*Document
    mu        sync.RWMutex
}

type Document struct {
    uri       string
    content   []byte
    version   int32

    ast       *ast.AST
    tree      *sitter.Tree  // Keep alive for incremental updates

    // Edit tracking
    pendingEdits []ast.Edit
    dirty        bool
}

// Incremental update algorithm
func (pm *ParseManager) ApplyEdit(uri string, edit Edit) error {
    doc := pm.documents[uri]

    // 1. Apply edit to tree-sitter tree
    tsEdit := &sitter.InputEdit{
        StartByte:   edit.StartByte,
        OldEndByte:  edit.OldEndByte,
        NewEndByte:  edit.NewEndByte,
        StartPoint:  sitter.Point{Row: edit.StartPoint.Row, Column: edit.StartPoint.Column},
        OldEndPoint: sitter.Point{Row: edit.OldEndPoint.Row, Column: edit.OldEndPoint.Column},
        NewEndPoint: sitter.Point{Row: edit.NewEndPoint.Row, Column: edit.NewEndPoint.Column},
    }
    doc.tree.Edit(tsEdit)

    // 2. Update content
    doc.content = applyEditToContent(doc.content, edit)

    // 3. Incremental re-parse (only changed subtrees)
    newTree := pm.parser.Parse(doc.tree, doc.content)

    // 4. Invalidate affected caches
    pm.invalidateAffectedCaches(doc, edit)

    // 5. Replace tree
    doc.tree.Close()
    doc.tree = newTree

    return nil
}

// Cache invalidation (surgical, not full rebuild)
func (pm *ParseManager) invalidateAffectedCaches(doc *Document, edit Edit) {
    // Find AST nodes in edited range
    affectedNodes := findNodesInRange(doc.ast.root, edit.Range())

    for _, node := range affectedNodes {
        // Clear cached metadata
        node.metadata = &ast.NodeMetadata{}

        // Invalidate parent scopes (up to function boundary)
        invalidateScopesUpToFunction(node)
    }
}
```

#### Semantic Analysis Algorithms

```go
// Semantic analyzer provides React-specific queries
type SemanticAnalyzer struct {
    ast *ast.AST
}

// IsComponent detects if a function is a React component
func (sa *SemanticAnalyzer) IsComponent(node *ast.Node) bool {
    // Check cache first
    if node.metadata.isComponent != nil {
        return *node.metadata.isComponent
    }

    // Algorithm:
    // 1. Must be a function (function decl or arrow function)
    if !sa.isFunction(node) {
        result := false
        node.metadata.isComponent = &result
        return false
    }

    // 2. Must return JSX
    hasJSXReturn := sa.functionReturnsJSX(node)

    // 3. Name must start with capital letter (convention)
    name := sa.getFunctionName(node)
    startsWithCapital := len(name) > 0 && unicode.IsUpper(rune(name[0]))

    result := hasJSXReturn && startsWithCapital
    node.metadata.isComponent = &result
    return result
}

// functionReturnsJSX traverses all return statements
func (sa *SemanticAnalyzer) functionReturnsJSX(funcNode *ast.Node) bool {
    // Find function body
    body := sa.findChild(funcNode, "statement_block")
    if body == nil {
        // Arrow function with expression body
        body = funcNode
    }

    // Traverse and find return statements
    returns := sa.findReturnStatements(body)

    for _, ret := range returns {
        // Check if return value is JSX
        returnValue := sa.findChild(ret, "jsx_element", "jsx_fragment")
        if returnValue != nil {
            return true
        }

        // Check for conditional JSX: return condition ? <A/> : <B/>
        ternary := sa.findChild(ret, "ternary_expression")
        if ternary != nil && sa.hasPot
enJSX(ternary) {
            return true
        }
    }

    return false
}

// IsMemoized detects React.memo, useMemo, useCallback wrapping
func (sa *SemanticAnalyzer) IsMemoized(node *ast.Node) bool {
    // Check cache
    if node.metadata.isMemoized != nil {
        return *node.metadata.isMemoized
    }

    // Pattern 1: React.memo(Component)
    parent := node.parent
    if sa.isCallExpression(parent) {
        callee := sa.findChild(parent, "member_expression")
        if callee != nil {
            obj := sa.getText(sa.findChild(callee, "object"))
            prop := sa.getText(sa.findChild(callee, "property"))
            if obj == "React" && prop == "memo" {
                result := true
                node.metadata.isMemoized = &result
                return true
            }
        }
    }

    // Pattern 2: const x = useMemo(() => ..., deps)
    // Pattern 3: const x = useCallback(() => ..., deps)
    if sa.isVariableDeclarator(parent) {
        init := sa.findChild(parent, "call_expression")
        if init != nil {
            callee := sa.getText(sa.findChild(init, "identifier"))
            if callee == "useMemo" || callee == "useCallback" {
                result := true
                node.metadata.isMemoized = &result
                return true
            }
        }
    }

    result := false
    node.metadata.isMemoized = &result
    return false
}
```

### Scope Analysis Implementation

```go
// ScopeBuilder constructs the scope tree from AST
type ScopeBuilder struct {
    ast       *ast.AST
    scopeTree *ScopeTree
    current   *Scope
    nextID    ScopeID
}

func (sb *ScopeBuilder) Build() *ScopeTree {
    sb.scopeTree = &ScopeTree{
        scopes: make(map[ScopeID]*Scope),
    }

    // Create global scope
    sb.current = &Scope{
        id:      sb.nextID,
        kind:    ScopeGlobal,
        symbols: make(map[string]*Symbol),
    }
    sb.scopeTree.root = sb.current
    sb.scopeTree.scopes[sb.current.id] = sb.current
    sb.nextID++

    // Traverse AST and build scopes
    sb.visit(sb.ast.root)

    return sb.scopeTree
}

func (sb *ScopeBuilder) visit(node *ast.Node) {
    switch node.Type() {
    case "function_declaration", "arrow_function", "function":
        sb.enterFunctionScope(node)
        sb.visitChildren(node)
        sb.exitScope()

    case "lexical_declaration":  // const, let
        sb.addDeclaration(node, false)

    case "variable_declaration":  // var
        sb.addDeclaration(node, true)  // var is function-scoped

    case "import_statement":
        sb.addImport(node)

    default:
        sb.visitChildren(node)
    }
}

func (sb *ScopeBuilder) enterFunctionScope(node *ast.Node) {
    newScope := &Scope{
        id:      sb.nextID,
        kind:    ScopeFunction,
        parent:  sb.current,
        node:    node,
        symbols: make(map[string]*Symbol),
    }
    sb.nextID++

    sb.current.children = append(sb.current.children, newScope)
    sb.scopeTree.scopes[newScope.id] = newScope
    sb.current = newScope

    // Add function parameters as symbols
    sb.addParameters(node)
}

func (sb *ScopeBuilder) addDeclaration(node *ast.Node, isFunctionScoped bool) {
    // Find the scope to add to
    targetScope := sb.current
    if isFunctionScoped {
        // var: hoist to function scope
        targetScope = sb.findFunctionScope()
    }

    // Extract declarators
    declarators := sb.findChildren(node, "variable_declarator")
    for _, decl := range declarators {
        nameNode := sb.findChild(decl, "identifier")
        if nameNode == nil {
            continue  // Destructuring - handle separately
        }

        name := sb.getText(nameNode)
        init := sb.findChild(decl, "call_expression", "arrow_function", "object", "array")

        symbol := &Symbol{
            name:       name,
            kind:       sb.symbolKindFromDecl(node),
            scope:      targetScope,
            defNode:    nameNode,
            defRange:   nameNode.Range(),
            references: []*Reference{},
        }

        // Analyze stability
        if init != nil {
            symbol.stability = sb.analyzeStability(init, targetScope)
        }

        targetScope.symbols[name] = symbol
    }
}

func (sb *ScopeBuilder) analyzeStability(node *ast.Node, scope *Scope) Stability {
    switch node.Type() {
    case "object", "array":
        // Inline object/array literals are unstable
        return StabilityUnstable

    case "arrow_function", "function":
        // Inline functions are unstable
        return StabilityUnstable

    case "call_expression":
        callee := sb.getText(sb.findChild(node, "identifier"))

        // useMemo/useCallback with empty deps = stable
        if callee == "useMemo" || callee == "useCallback" {
            deps := sb.findDependencyArray(node)
            if deps != nil && len(deps) == 0 {
                return StabilityStable
            }
            return StabilityConditional
        }

        return StabilityUnknown

    case "identifier":
        // Look up symbol and check its stability
        sym := sb.resolveSymbol(sb.getText(node), scope)
        if sym != nil {
            return sym.stability
        }
        return StabilityUnknown

    default:
        return StabilityUnknown
    }
}

// Symbol resolution with scope chain traversal
func (sb *ScopeBuilder) resolveSymbol(name string, scope *Scope) *Symbol {
    // Search up the scope chain
    current := scope
    for current != nil {
        if sym, found := current.symbols[name]; found {
            return sym
        }
        current = current.parent
    }
    return nil
}
```

### Rule Execution Pipeline

```go
// RuleExecutor manages parallel rule execution
type RuleExecutor struct {
    registry  *RuleRegistry
    workerPool *WorkerPool

    // Execution phases
    phases []ExecutionPhase
}

type ExecutionPhase int
const (
    PhaseLocal ExecutionPhase = iota  // Single-node analysis
    PhaseFile                           // Whole-file analysis
    PhaseGraph                          // Cross-file analysis
)

type WorkerPool struct {
    workers   int
    taskQueue chan *AnalysisTask
    results   chan *RuleResult
    wg        sync.WaitGroup
}

type AnalysisTask struct {
    rule    Rule
    context *AnalysisContext
    node    *ast.Node  // Nil for file/graph-level rules
    phase   ExecutionPhase
}

type RuleResult struct {
    ruleID      string
    diagnostics []Diagnostic
    duration    time.Duration
    err         error
}

// Execute runs all enabled rules on the given context
func (re *RuleExecutor) Execute(ctx *AnalysisContext) ([]Diagnostic, error) {
    allDiagnostics := []Diagnostic{}

    // Execute in phases (ensures dependencies are met)
    for _, phase := range []ExecutionPhase{PhaseLocal, PhaseFile, PhaseGraph} {
        diagnostics, err := re.executePhase(ctx, phase)
        if err != nil {
            return nil, err
        }
        allDiagnostics = append(allDiagnostics, diagnostics...)
    }

    return allDiagnostics, nil
}

func (re *RuleExecutor) executePhase(ctx *AnalysisContext, phase ExecutionPhase) ([]Diagnostic, error) {
    // Get rules for this phase
    rules := re.registry.GetRulesForPhase(phase)

    // Start workers
    wp := NewWorkerPool(runtime.NumCPU())
    wp.Start()

    // Enqueue tasks
    for _, rule := range rules {
        if phase == PhaseLocal {
            // Local phase: analyze each relevant node
            nodes := re.findRelevantNodes(ctx.AST, rule.NodeTypes())
            for _, node := range nodes {
                wp.Submit(&AnalysisTask{
                    rule:    rule,
                    context: ctx,
                    node:    node,
                    phase:   phase,
                })
            }
        } else {
            // File/Graph phase: one task per rule
            wp.Submit(&AnalysisTask{
                rule:    rule,
                context: ctx,
                node:    nil,
                phase:   phase,
            })
        }
    }

    // Collect results
    wp.Close()
    results := wp.Results()

    // Aggregate diagnostics
    allDiagnostics := []Diagnostic{}
    for _, result := range results {
        if result.err != nil {
            // Log error but continue (isolation)
            log.Errorf("Rule %s failed: %v", result.ruleID, result.err)
            continue
        }
        allDiagnostics = append(allDiagnostics, result.diagnostics...)
    }

    return allDiagnostics, nil
}

// Worker pool implementation
func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()

    for task := range wp.taskQueue {
        // Recover from panics (rule isolation)
        func() {
            defer func() {
                if r := recover(); r != nil {
                    wp.results <- &RuleResult{
                        ruleID: task.rule.ID(),
                        err:    fmt.Errorf("panic: %v", r),
                    }
                }
            }()

            start := time.Now()

            // Execute rule
            var diagnostics []Diagnostic
            if task.node != nil {
                diagnostics = task.rule.Analyze(task.context, task.node)
            } else {
                diagnostics = task.rule.AnalyzeFile(task.context)
            }

            wp.results <- &RuleResult{
                ruleID:      task.rule.ID(),
                diagnostics: diagnostics,
                duration:    time.Since(start),
            }
        }()
    }
}
```

### Memory Management Strategy

```go
// Memory management for long-lived LSP server

type MemoryManager struct {
    // LRU cache for parsed ASTs
    astCache *lru.Cache

    // Weak references to inactive documents
    inactiveDocs map[string]*WeakRef

    // Memory budget
    maxMemoryMB int64
    currentMB   int64
    mu          sync.RWMutex
}

// Cache policy: Keep recently used, evict old
func (mm *MemoryManager) GetOrParse(uri string, parser *ParseManager) (*ast.AST, error) {
    // Check cache
    if cached, ok := mm.astCache.Get(uri); ok {
        return cached.(*ast.AST), nil
    }

    // Parse
    ast, err := parser.ParseFile(uri)
    if err != nil {
        return nil, err
    }

    // Estimate memory usage
    estimatedMB := mm.estimateASTSize(ast)

    // Evict if needed
    mm.mu.Lock()
    for mm.currentMB+estimatedMB > mm.maxMemoryMB {
        mm.evictOldest()
    }
    mm.currentMB += estimatedMB
    mm.mu.Unlock()

    // Cache
    mm.astCache.Add(uri, ast)

    return ast, nil
}

func (mm *MemoryManager) estimateASTSize(ast *ast.AST) int64 {
    // Rough estimate: 1KB per node + source size
    nodeCount := mm.countNodes(ast.root)
    sourceSize := int64(len(ast.source))
    return (nodeCount * 1024) + sourceSize
}

// Tree-sitter tree lifecycle management
type TreeManager struct {
    trees map[string]*sitter.Tree
    mu    sync.Mutex
}

func (tm *TreeManager) Close(uri string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    if tree, ok := tm.trees[uri]; ok {
        tree.Close()  // Free C memory
        delete(tm.trees, uri)
    }
}

// Cleanup strategy for LSP server
func (lsp *LSPServer) CleanupInactiveDocuments() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        lsp.mm.mu.Lock()

        now := time.Now()
        for uri, doc := range lsp.documents {
            // Close documents inactive for >10 minutes
            if now.Sub(doc.lastAccess) > 10*time.Minute {
                lsp.closeDocument(uri)
            }
        }

        lsp.mm.mu.Unlock()
    }
}
```

### Auto-Fix Generation System

```go
// AutoFixGenerator creates code actions for diagnostics
type AutoFixGenerator struct {
    ast *ast.AST
}

type CodeAction struct {
    title   string
    kind    CodeActionKind
    edits   []TextEdit

    // Validation
    validated bool
    safe      bool  // Won't break code
}

type CodeActionKind string
const (
    QuickFix CodeActionKind = "quickfix"
    Refactor                = "refactor"
)

type TextEdit struct {
    range   SourceRange
    newText string
}

// Example: Extract inline object to useMemo
func (afg *AutoFixGenerator) ExtractToUseMemo(node *ast.Node) (*CodeAction, error) {
    // 1. Find the inline object/array
    // Node is something like: <Child config={{ theme: 'dark' }} />

    // 2. Determine insertion point (before the component function)
    component := afg.findContainingComponent(node)
    if component == nil {
        return nil, errors.New("not inside a component")
    }

    // 3. Generate variable name
    varName := afg.generateUniqueName("config", component)

    // 4. Extract dependencies
    deps := afg.extractDependencies(node)
    depsArray := afg.formatDepsArray(deps)

    // 5. Generate useMemo code
    objectCode := afg.getText(node)
    useMemoCode := fmt.Sprintf("const %s = useMemo(() => %s, [%s]);",
        varName, objectCode, depsArray)

    // 6. Create edits
    edits := []TextEdit{
        {
            // Insert useMemo before component
            range:   afg.findInsertionPoint(component),
            newText: useMemoCode + "\n  ",
        },
        {
            // Replace inline object with variable reference
            range:   node.Range(),
            newText: varName,
        },
    }

    action := &CodeAction{
        title: fmt.Sprintf("Extract to useMemo(%s)", varName),
        kind:  QuickFix,
        edits: edits,
    }

    // 7. Validate (parse result, ensure no syntax errors)
    if err := afg.validateAction(action); err != nil {
        action.safe = false
        return action, err
    }

    action.safe = true
    action.validated = true

    return action, nil
}

// Dependency extraction (simplified control flow analysis)
func (afg *AutoFixGenerator) extractDependencies(node *ast.Node) []*Symbol {
    deps := []*Symbol{}
    seen := make(map[string]bool)

    afg.visitIdentifiers(node, func(id *ast.Node) {
        name := afg.getText(id)

        // Skip if already added
        if seen[name] {
            return
        }

        // Resolve symbol
        scope := afg.getScopeForNode(id)
        symbol := afg.resolveSymbol(name, scope)

        if symbol != nil {
            // Only include if defined in component scope
            if afg.isInComponentScope(symbol) {
                deps = append(deps, symbol)
                seen[name] = true
            }
        }
    })

    return deps
}

// Validation: Parse modified code
func (afg *AutoFixGenerator) validateAction(action *CodeAction) error {
    // Apply edits to source
    modifiedSource := afg.applyEdits(afg.ast.source, action.edits)

    // Try to parse
    parser := NewParser()
    defer parser.Close()

    _, err := parser.Parse(modifiedSource)
    if err != nil {
        return fmt.Errorf("auto-fix produces invalid syntax: %w", err)
    }

    return nil
}
```

### LSP Server Architecture

```go
// LSP server implementation details
type LSPServer struct {
    // State
    initialized   bool
    capabilities  protocol.ServerCapabilities
    documents     map[string]*Document

    // Analysis engine
    analyzer      *Analyzer
    parseManager  *ParseManager
    memManager    *MemoryManager

    // Communication
    conn          jsonrpc2.Conn

    // Request handling
    reqQueue      chan *Request
    debouncer     *Debouncer  // Debounce didChange events

    // Concurrency
    mu            sync.RWMutex
    wg            sync.WaitGroup
}

// Request lifecycle
type Request struct {
    id        interface{}
    method    string
    params    interface{}
    createdAt time.Time
}

// Debouncer prevents excessive re-analysis on rapid edits
type Debouncer struct {
    timers map[string]*time.Timer
    delay  time.Duration
    mu     sync.Mutex
}

func (d *Debouncer) Debounce(key string, fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Cancel existing timer
    if timer, ok := d.timers[key]; ok {
        timer.Stop()
    }

    // Create new timer
    d.timers[key] = time.AfterFunc(d.delay, func() {
        fn()
        d.mu.Lock()
        delete(d.timers, key)
        d.mu.Unlock()
    })
}

// Document synchronization
func (lsp *LSPServer) handleDidChange(params *protocol.DidChangeTextDocumentParams) error {
    uri := params.TextDocument.URI

    // Apply content changes
    doc := lsp.documents[uri]
    for _, change := range params.ContentChanges {
        // Incremental or full sync
        if change.Range != nil {
            // Incremental update
            edit := convertToEdit(change)
            lsp.parseManager.ApplyEdit(uri, edit)
        } else {
            // Full document sync
            doc.content = []byte(change.Text)
            lsp.parseManager.ResetDocument(uri, doc.content)
        }
    }

    doc.version = params.TextDocument.Version

    // Debounced analysis (wait for user to stop typing)
    lsp.debouncer.Debounce(uri, func() {
        lsp.analyzeAndPublishDiagnostics(uri)
    })

    return nil
}

// Diagnostic publishing
func (lsp *LSPServer) analyzeAndPublishDiagnostics(uri string) {
    // 1. Get or parse AST
    ast, err := lsp.memManager.GetOrParse(uri, lsp.parseManager)
    if err != nil {
        lsp.logError("Parse failed for %s: %v", uri, err)
        return
    }

    // 2. Build analysis context
    ctx := &AnalysisContext{
        AST:        ast,
        FilePath:   uri,
        ScopeTree:  buildScopeTree(ast),
        // ... other context
    }

    // 3. Run rules
    diagnostics, err := lsp.analyzer.Analyze(ctx)
    if err != nil {
        lsp.logError("Analysis failed for %s: %v", uri, err)
        return
    }

    // 4. Convert to LSP diagnostics
    lspDiags := convertToLSPDiagnostics(diagnostics)

    // 5. Publish
    lsp.conn.Notify(context.Background(), "textDocument/publishDiagnostics", &protocol.PublishDiagnosticsParams{
        URI:         uri,
        Version:     lsp.documents[uri].version,
        Diagnostics: lspDiags,
    })
}

// Code actions (quick fixes)
func (lsp *LSPServer) handleCodeAction(params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
    uri := params.TextDocument.URI

    // Get AST
    ast, err := lsp.memManager.GetOrParse(uri, lsp.parseManager)
    if err != nil {
        return nil, err
    }

    // Get diagnostics in range
    diagnostics := filterDiagnosticsInRange(params.Context.Diagnostics, params.Range)

    // Generate code actions for each diagnostic
    actions := []protocol.CodeAction{}
    generator := NewAutoFixGenerator(ast)

    for _, diag := range diagnostics {
        // Find node at diagnostic location
        node := findNodeAtPosition(ast, diag.Range.Start)

        // Generate fix based on rule ID
        var codeAction *CodeAction
        switch diag.Source {
        case "no-object-deps":
            codeAction, _ = generator.ExtractToUseMemo(node)
        case "no-unstable-props":
            codeAction, _ = generator.ExtractToVariable(node)
        // ... other rules
        }

        if codeAction != nil && codeAction.safe {
            actions = append(actions, convertToLSPCodeAction(codeAction, diag))
        }
    }

    return actions, nil
}

// Concurrency model: Request handler
func (lsp *LSPServer) requestHandler() {
    for req := range lsp.reqQueue {
        // Process request based on method
        switch req.method {
        case "textDocument/didOpen":
            lsp.handleDidOpen(req.params.(*protocol.DidOpenTextDocumentParams))
        case "textDocument/didChange":
            lsp.handleDidChange(req.params.(*protocol.DidChangeTextDocumentParams))
        case "textDocument/codeAction":
            // Handle synchronously (user waiting)
            actions, err := lsp.handleCodeAction(req.params.(*protocol.CodeActionParams))
            lsp.sendResponse(req.id, actions, err)
        }
    }
}
```

### Configuration System

```go
// Configuration hierarchy: User settings → Workspace → File → Defaults

type Config struct {
    // Rule configuration
    Rules map[string]*RuleConfig `json:"rules"`

    // Global settings
    Severity      SeverityLevel     `json:"severity"`      // Minimum severity to report
    Include       []string          `json:"include"`       // File patterns to analyze
    Exclude       []string          `json:"exclude"`       // File patterns to skip

    // Performance
    MaxWorkers    int               `json:"maxWorkers"`    // Parallel analysis workers
    MemoryLimitMB int64             `json:"memoryLimitMB"` // Memory budget

    // Experimental
    CrossFile     bool              `json:"crossFile"`     // Enable cross-file analysis
    TypeAnalysis  bool              `json:"typeAnalysis"`  // Use TypeScript types

    // Source
    source        ConfigSource
}

type RuleConfig struct {
    Enabled  bool                   `json:"enabled"`
    Severity SeverityLevel          `json:"severity"`
    Options  map[string]interface{} `json:"options"`  // Rule-specific options
}

type SeverityLevel string
const (
    SeverityOff     SeverityLevel = "off"
    SeverityHint                  = "hint"
    SeverityInfo                  = "info"
    SeverityWarning               = "warning"
    SeverityError                 = "error"
)

type ConfigSource int
const (
    ConfigDefault ConfigSource = iota
    ConfigFile                  // .react-analyzer.json
    ConfigWorkspace             // VS Code workspace settings
    ConfigUser                  // VS Code user settings
)

// ConfigManager handles config loading and merging
type ConfigManager struct {
    defaults      *Config
    fileConfigs   map[string]*Config  // Per-directory configs
    workspaceConf *Config
    userConf      *Config

    mu sync.RWMutex
}

// GetConfigForFile returns merged config for a specific file
func (cm *ConfigManager) GetConfigForFile(filePath string) *Config {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    // Build config hierarchy
    merged := cm.defaults.Clone()

    // Apply workspace config
    if cm.workspaceConf != nil {
        merged.Merge(cm.workspaceConf)
    }

    // Apply directory-specific config (walk up directory tree)
    dir := filepath.Dir(filePath)
    for {
        if cfg, ok := cm.fileConfigs[dir]; ok {
            merged.Merge(cfg)
            break
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            break  // Reached root
        }
        dir = parent
    }

    // Apply user config (lowest priority)
    if cm.userConf != nil {
        merged.Merge(cm.userConf)
    }

    return merged
}

// Config file format (.react-analyzer.json)
/*
{
  "$schema": "https://react-analyzer.dev/schema.json",
  "extends": ["react-analyzer:recommended"],

  "rules": {
    "no-object-deps": "error",
    "no-unstable-props": {
      "severity": "warning",
      "options": {
        "checkInlineCallbacks": true,
        "ignoredProps": ["style", "className"]
      }
    },
    "memo-unstable-props": "off"
  },

  "include": ["src/**/*.tsx", "src/**/*.jsx"],
  "exclude": ["**/*.test.tsx", "**/*.stories.tsx"],

  "severity": "warning",
  "maxWorkers": 4,
  "crossFile": true
}
*/

// Preset configurations
var PresetRecommended = &Config{
    Rules: map[string]*RuleConfig{
        "no-object-deps":      {Enabled: true, Severity: SeverityError},
        "no-unstable-props":   {Enabled: true, Severity: SeverityWarning},
        "memo-unstable-props": {Enabled: true, Severity: SeverityError},
    },
    Severity:      SeverityWarning,
    CrossFile:     true,
    TypeAnalysis:  false,
}

var PresetStrict = &Config{
    Rules: map[string]*RuleConfig{
        "no-object-deps":      {Enabled: true, Severity: SeverityError},
        "no-unstable-props":   {Enabled: true, Severity: SeverityError},
        "memo-unstable-props": {Enabled: true, Severity: SeverityError},
    },
    Severity:      SeverityHint,
    CrossFile:     true,
    TypeAnalysis:  true,
}

// Config validation
func (c *Config) Validate() error {
    // Check rule IDs are valid
    validRules := GetRegisteredRuleIDs()
    for ruleID := range c.Rules {
        if !contains(validRules, ruleID) {
            return fmt.Errorf("unknown rule: %s", ruleID)
        }
    }

    // Check severity levels
    validSeverities := []SeverityLevel{SeverityOff, SeverityHint, SeverityInfo, SeverityWarning, SeverityError}
    if !contains(validSeverities, c.Severity) {
        return fmt.Errorf("invalid severity: %s", c.Severity)
    }

    // Validate file patterns (glob syntax)
    for _, pattern := range append(c.Include, c.Exclude...) {
        if _, err := filepath.Match(pattern, "test.tsx"); err != nil {
            return fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
        }
    }

    return nil
}
```

### Error Handling & Recovery Strategy

```go
// Error types hierarchy
type ErrorKind int
const (
    ErrKindParse ErrorKind = iota  // Syntax error in source
    ErrKindAnalysis                 // Analysis failed
    ErrKindRule                     // Rule execution error
    ErrKindImport                   // Import resolution failed
    ErrKindInternal                 // Internal error (bug)
    ErrKindTimeout                  // Operation timed out
)

type AnalysisError struct {
    kind     ErrorKind
    message  string
    filePath string
    location *SourceRange
    cause    error

    // Recovery
    recoverable bool
    suggestion  string
}

func (e *AnalysisError) Error() string {
    if e.filePath != "" {
        return fmt.Sprintf("%s:%d:%d: %s", e.filePath, e.location.Start.Line, e.location.Start.Column, e.message)
    }
    return e.message
}

// Error handling strategies

// 1. Parse Errors: Graceful degradation
func (p *Parser) ParseWithRecovery(content []byte) (*AST, error) {
    tree := p.parser.Parse(nil, content)

    // Check for syntax errors
    if tree.RootNode().HasError() {
        // Still return partial AST
        ast := &AST{
            root:   wrapNode(tree.RootNode()),
            tree:   tree,
            source: content,
        }

        // Collect error nodes
        errors := findErrorNodes(tree.RootNode())

        return ast, &AnalysisError{
            kind:        ErrKindParse,
            message:     fmt.Sprintf("Syntax error: found %d error(s)", len(errors)),
            recoverable: true,
            suggestion:  "Fix syntax errors to enable full analysis",
        }
    }

    // Success
    return &AST{root: wrapNode(tree.RootNode()), tree: tree, source: content}, nil
}

// 2. Rule Errors: Isolation (don't crash other rules)
func (wp *WorkerPool) worker() {
    defer wp.wg.Done()

    for task := range wp.taskQueue {
        func() {
            // Panic recovery
            defer func() {
                if r := recover(); r != nil {
                    stack := debug.Stack()
                    wp.results <- &RuleResult{
                        ruleID: task.rule.ID(),
                        err: &AnalysisError{
                            kind:        ErrKindRule,
                            message:     fmt.Sprintf("Rule panic: %v", r),
                            recoverable: false,
                        },
                    }

                    // Log for debugging
                    log.Errorf("Rule %s panicked: %v\n%s", task.rule.ID(), r, stack)
                }
            }()

            // Timeout protection
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

            resultChan := make(chan *RuleResult, 1)
            go func() {
                diagnostics := task.rule.Analyze(task.context, task.node)
                resultChan <- &RuleResult{
                    ruleID:      task.rule.ID(),
                    diagnostics: diagnostics,
                }
            }()

            select {
            case result := <-resultChan:
                wp.results <- result
            case <-ctx.Done():
                // Timeout
                wp.results <- &RuleResult{
                    ruleID: task.rule.ID(),
                    err: &AnalysisError{
                        kind:        ErrKindTimeout,
                        message:     "Rule execution timed out (5s)",
                        recoverable: true,
                        suggestion:  "Rule may be stuck in infinite loop",
                    },
                }
            }
        }()
    }
}

// 3. Import Resolution: Fallback strategies
func (ir *ImportResolver) Resolve(importPath string, fromFile string) (*ResolvedModule, error) {
    // Try cache first
    cacheKey := fromFile + ":" + importPath
    if cached, ok := ir.resolveCache[cacheKey]; ok {
        return ir.moduleCache[cached], nil
    }

    // Try resolution strategies in order
    strategies := []ResolutionStrategy{
        ir.resolveRelative,       // ./Component
        ir.resolveTSConfig,       // @/components/Component (tsconfig paths)
        ir.resolveNodeModules,    // react-router-dom
    }

    var lastErr error
    for _, strategy := range strategies {
        absPath, err := strategy(importPath, fromFile)
        if err != nil {
            lastErr = err
            continue
        }

        // Success - parse module
        module, err := ir.loadModule(absPath)
        if err != nil {
            lastErr = err
            continue
        }

        // Cache and return
        ir.resolveCache[cacheKey] = absPath
        ir.moduleCache[absPath] = module
        return module, nil
    }

    // All strategies failed
    return nil, &AnalysisError{
        kind:        ErrKindImport,
        message:     fmt.Sprintf("Cannot resolve import: %s", importPath),
        filePath:    fromFile,
        cause:       lastErr,
        recoverable: true,
        suggestion:  "Check import path and file exists",
    }
}

// 4. LSP Error Handling: Never crash server
func (lsp *LSPServer) handleRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
    // Catch all panics
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("Handler panic for %s: %v\n%s", req.Method, r, debug.Stack())

            // Send error response
            conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
                Code:    jsonrpc2.CodeInternalError,
                Message: "Internal server error",
            })
        }
    }()

    // Handle method
    switch req.Method {
    case "textDocument/didOpen":
        var params protocol.DidOpenTextDocumentParams
        if err := json.Unmarshal(*req.Params, &params); err != nil {
            return nil, &jsonrpc2.Error{
                Code:    jsonrpc2.CodeInvalidParams,
                Message: "Invalid parameters",
            }
        }
        return nil, lsp.handleDidOpen(&params)

    // ... other methods

    default:
        return nil, &jsonrpc2.Error{
            Code:    jsonrpc2.CodeMethodNotFound,
            Message: fmt.Sprintf("Method not found: %s", req.Method),
        }
    }
}

// 5. Logging & Observability
type Logger struct {
    level LogLevel
    out   io.Writer
}

type LogLevel int
const (
    LogDebug LogLevel = iota
    LogInfo
    LogWarn
    LogError
)

func (l *Logger) Error(format string, args ...interface{}) {
    l.log(LogError, format, args...)
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
    if level < l.level {
        return
    }

    msg := fmt.Sprintf(format, args...)
    timestamp := time.Now().Format(time.RFC3339)
    levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR"}[level]

    fmt.Fprintf(l.out, "[%s] %s: %s\n", timestamp, levelStr, msg)
}

// Error metrics (for monitoring)
type ErrorMetrics struct {
    parseErrors     atomic.Int64
    analysisErrors  atomic.Int64
    ruleErrors      atomic.Int64
    timeouts        atomic.Int64

    mu sync.Mutex
    recentErrors []*AnalysisError  // Last 100 errors
}

func (em *ErrorMetrics) RecordError(err *AnalysisError) {
    // Increment counter
    switch err.kind {
    case ErrKindParse:
        em.parseErrors.Add(1)
    case ErrKindAnalysis:
        em.analysisErrors.Add(1)
    case ErrKindRule:
        em.ruleErrors.Add(1)
    case ErrKindTimeout:
        em.timeouts.Add(1)
    }

    // Store recent error
    em.mu.Lock()
    em.recentErrors = append(em.recentErrors, err)
    if len(em.recentErrors) > 100 {
        em.recentErrors = em.recentErrors[1:]
    }
    em.mu.Unlock()
}
```

---

## Phase 1: Core Analysis Engine (Weeks 1-6)

### Week 1-2: Parser Abstraction Layer

**Goal**: Create a clean abstraction over tree-sitter that provides React-specific semantics

#### Implementation Tasks

1. **Parser Interface** (`internal/ast/parser.go`)
   ```go
   type Parser interface {
       ParseFile(path string, content []byte) (*AST, error)
       ParseTypeScript(content []byte) (*AST, error)
       ParseJSX(content []byte) (*AST, error)
       Close() error
   }
   ```

2. **tree-sitter Implementation** (`internal/ast/treesitter/parser.go`)
   - Wrap tree-sitter parser
   - Handle language detection (.tsx, .jsx, .ts, .js)
   - Implement incremental parsing
   - Memory management (defer tree.Close())

3. **AST Node Abstraction** (`internal/ast/node.go`)
   ```go
   type Node interface {
       Type() NodeType
       Range() SourceRange
       Children() []Node
       Text() string

       // React semantics
       IsComponent() bool
       IsFunctionDeclaration() bool
       IsHookCall() bool
       IsMemoized() bool
       GetJSXProps() []PropNode
       GetFunctionParams() []ParamNode
       GetHookDependencies() []string
   }
   ```

4. **Semantic Analysis Layer** (`internal/ast/treesitter/semantic.go`)
   - Implement IsComponent() - detect functions returning JSX
   - Implement IsHookCall() - functions starting with "use"
   - Implement IsMemoized() - React.memo/useMemo/useCallback detection
   - Implement GetHookDependencies() - extract dependency arrays

#### Testing Strategy

**Unit Tests** (`test/ast/`)
- `parser_test.go`:
  - ✅ Parse valid TypeScript/JSX/TSX files
  - ✅ Handle syntax errors gracefully
  - ✅ Auto-detect language from file extension
  - ✅ Memory cleanup (no leaks)

- `semantic_test.go`:
  - ✅ IsComponent() correctly identifies function components
  - ✅ IsComponent() correctly identifies arrow function components
  - ✅ IsComponent() returns false for non-components
  - ✅ IsHookCall() identifies standard hooks (useState, useEffect, etc.)
  - ✅ IsHookCall() identifies custom hooks
  - ✅ IsMemoized() detects React.memo wrapper
  - ✅ IsMemoized() detects useMemo/useCallback
  - ✅ GetHookDependencies() extracts dependency arrays
  - ✅ GetHookDependencies() handles empty arrays
  - ✅ GetJSXProps() extracts prop names and values

**Test Fixtures** (`test/fixtures/ast/`)
```
test/fixtures/ast/
├── valid_component.tsx          # Simple function component
├── arrow_component.tsx          # Arrow function component
├── memoized_component.tsx       # React.memo wrapped
├── complex_hooks.tsx            # Multiple hooks with dependencies
├── jsx_props.tsx                # Various prop types
├── typescript_generics.tsx      # Complex TypeScript
└── syntax_error.tsx             # Invalid syntax for error handling
```

**Performance Tests** (`test/benchmarks/parser_bench_test.go`)
```go
func BenchmarkParseSmallFile(b *testing.B) {
    // 30-line component
    // Target: <100μs
}

func BenchmarkParseMediumFile(b *testing.B) {
    // 233-line component
    // Target: <2ms
}

func BenchmarkParseLargeFile(b *testing.B) {
    // 1000-line component
    // Target: <10ms
}

func BenchmarkIncrementalParse(b *testing.B) {
    // Edit and re-parse
    // Target: <100μs (10x faster than full parse)
}
```

**Success Criteria**
- ✅ All unit tests pass (>95% code coverage)
- ✅ Parse 30-line file in <100μs
- ✅ Parse 233-line file in <2ms
- ✅ Semantic analysis accuracy >95%
- ✅ Zero memory leaks (use go test -race)

---

### Week 3-4: Rule Engine Framework

**Goal**: Build pluggable rule system that can execute rules in parallel

#### Implementation Tasks

1. **Rule Interface** (`internal/rules/rule.go`)
   ```go
   type Rule interface {
       ID() string
       Name() string
       Description() string
       Category() RuleCategory
       Severity() Severity

       Initialize(config RuleConfig) error
       Analyze(ctx *AnalysisContext, node ASTNode) []Diagnostic

       NodeTypes() []string
       RequiresFullGraph() bool
   }
   ```

2. **Rule Registry** (`internal/rules/registry.go`)
   - Register/unregister rules
   - Rule discovery
   - Configuration management

3. **Rule Executor** (`internal/rules/executor.go`)
   - Worker pool for parallel execution
   - Phase-based execution (local → file → graph)
   - Result aggregation

4. **Analysis Context** (`internal/analyzer/context.go`)
   ```go
   type AnalysisContext struct {
       ComponentGraph  *ComponentGraph
       DependencyMap   map[string][]string
       TypeRegistry    *TypeRegistry
       ImportMap       map[string]ImportInfo
   }
   ```

5. **Diagnostic System** (`internal/rules/diagnostic.go`)
   - Diagnostic creation
   - Code action (auto-fix) support
   - Severity levels

#### Testing Strategy

**Unit Tests** (`test/rules/`)
- `registry_test.go`:
  - ✅ Register rule successfully
  - ✅ Prevent duplicate rule IDs
  - ✅ Retrieve rule by ID
  - ✅ List all registered rules
  - ✅ Load configuration from YAML

- `executor_test.go`:
  - ✅ Execute single rule
  - ✅ Execute multiple rules in parallel
  - ✅ Aggregate diagnostics correctly
  - ✅ Handle rule panics gracefully
  - ✅ Phase-based execution (local → file → graph)
  - ✅ Worker pool scales with CPU count

- `diagnostic_test.go`:
  - ✅ Create diagnostic with location
  - ✅ Attach code actions
  - ✅ Format diagnostic messages
  - ✅ Sort diagnostics by location

**Integration Tests** (`test/integration/rule_engine_test.go`)
```go
func TestRuleEngineEndToEnd(t *testing.T) {
    // Create mock rule
    // Register rule
    // Execute on test file
    // Verify diagnostics
}

func TestRuleEngineParallelExecution(t *testing.T) {
    // Register 10 rules
    // Execute on large file
    // Verify all rules run
    // Verify parallelism (time < sequential time)
}
```

**Performance Tests** (`test/benchmarks/executor_bench_test.go`)
```go
func BenchmarkSingleRuleExecution(b *testing.B) {
    // Target: <1ms per file
}

func BenchmarkParallelRuleExecution(b *testing.B) {
    // 10 rules on 100 files
    // Target: <10s total
}
```

**Success Criteria**
- ✅ Rule registration works
- ✅ Parallel execution 5x faster than sequential
- ✅ Zero rule interference (isolation)
- ✅ Error handling prevents crashes

---

### Week 5-6: MVP Rules Implementation

**Goal**: Implement 3 core rules with auto-fix support

#### Implementation Tasks

1. **Rule 1: no-object-deps** (`internal/rules/no_object_deps.go`)
   - Detect object/array literals in hook dependencies
   - Trace dependency to definition
   - Suggest useMemo/constant extraction
   - Auto-fix: Extract to const (when possible)

2. **Rule 2: no-unstable-props** (`internal/rules/no_unstable_props.go`)
   - Detect inline objects/arrays/functions in JSX props
   - Check if child is memoized (optional enhancement)
   - Suggest memoization or extraction
   - Auto-fix: Extract to const/useMemo/useCallback

3. **Rule 3: memo-unstable-props** (`internal/rules/memo_unstable_props.go`)
   - Resolve component imports
   - Check if component is React.memo wrapped
   - Analyze all props passed to component
   - Detect unstable props
   - Suggest stabilization
   - **Requires**: Import resolution, cross-file analysis

4. **Import Resolver** (`internal/analyzer/import_resolver.go`)
   - Resolve relative imports (./components/Foo)
   - Handle named exports/default exports
   - Cache resolved modules

#### Testing Strategy

**Unit Tests per Rule** (`test/rules/`)

**no_object_deps_test.go**:
```go
func TestDetectInlineObjectDeps(t *testing.T) {
    // const config = { theme: 'dark' };
    // useEffect(() => {}, [config]);
    // Expected: ERROR
}

func TestDetectInlineArrayDeps(t *testing.T) {
    // const items = [1, 2, 3];
    // useEffect(() => {}, [items]);
    // Expected: ERROR
}

func TestAllowMemoizedDeps(t *testing.T) {
    // const config = useMemo(() => ({ theme: 'dark' }), []);
    // useEffect(() => {}, [config]);
    // Expected: NO ERROR
}

func TestAllowConstantDeps(t *testing.T) {
    // const CONFIG = { theme: 'dark' }; // outside component
    // useEffect(() => {}, []);
    // Expected: NO ERROR
}

func TestAutoFixExtractToConst(t *testing.T) {
    // Test code action generation
    // Verify fix suggestion
}
```

**no_unstable_props_test.go**:
```go
func TestDetectInlineObjectProp(t *testing.T) {
    // <Child config={{ theme: 'dark' }} />
    // Expected: WARNING
}

func TestDetectInlineArrayProp(t *testing.T) {
    // <Child items={[1, 2, 3]} />
    // Expected: WARNING
}

func TestDetectInlineFunction(t *testing.T) {
    // <Child onClick={() => {}} />
    // Expected: WARNING
}

func TestAllowStableProps(t *testing.T) {
    // const config = useMemo(...);
    // <Child config={config} />
    // Expected: NO ERROR
}
```

**memo_unstable_props_test.go**:
```go
func TestDetectMemoizedChild(t *testing.T) {
    // Resolve import, check React.memo
}

func TestDetectUnstablePropsToMemoChild(t *testing.T) {
    // Parent passes inline object to memo child
    // Expected: ERROR
}

func TestCrossFileAnalysis(t *testing.T) {
    // Parent.tsx imports Child.tsx
    // Child is React.memo
    // Parent passes unstable prop
    // Expected: ERROR with cross-file context
}
```

**End-to-End Tests** (`test/e2e/`)

Create real React component files and test complete analysis:

```
test/e2e/fixtures/
├── scenario1_unstable_deps/
│   └── App.tsx                  # Component with unstable hook deps
├── scenario2_unstable_props/
│   └── Parent.tsx               # Component with inline props
├── scenario3_memo_breaking/
│   ├── Parent.tsx               # Passes unstable props
│   └── Child.tsx                # React.memo component
└── scenario4_complex_project/
    ├── components/
    │   ├── Header.tsx
    │   ├── Dashboard.tsx
    │   └── UserList.tsx
    └── utils/
        └── helpers.ts
```

**e2e_test.go**:
```go
func TestScenario1_UnstableDeps(t *testing.T) {
    result := analyzeProject("test/e2e/fixtures/scenario1_unstable_deps")

    assert.Len(t, result.Diagnostics, 1)
    assert.Equal(t, "no-object-deps", result.Diagnostics[0].RuleID)
    assert.Contains(t, result.Diagnostics[0].Message, "will change every render")
}

func TestScenario3_MemoBreaking(t *testing.T) {
    result := analyzeProject("test/e2e/fixtures/scenario3_memo_breaking")

    // Should detect that Parent breaks Child's memo
    assert.Len(t, result.Diagnostics, 1)
    assert.Equal(t, "memo-unstable-props", result.Diagnostics[0].RuleID)

    // Verify cross-file context
    assert.Equal(t, "Parent.tsx", result.Diagnostics[0].Location.File)
    assert.Contains(t, result.Diagnostics[0].Message, "React.memo on Child")
}

func TestScenario4_ComplexProject(t *testing.T) {
    result := analyzeProject("test/e2e/fixtures/scenario4_complex_project")

    // Should find multiple issues across files
    assert.GreaterOrEqual(t, len(result.Diagnostics), 3)

    // Verify all rules executed
    ruleIDs := extractRuleIDs(result.Diagnostics)
    assert.Contains(t, ruleIDs, "no-object-deps")
    assert.Contains(t, ruleIDs, "no-unstable-props")
}
```

**Real-World Testing** (`test/real_world/`)

Test against actual open-source React projects:

```go
func TestRealWorld_TodoMVC(t *testing.T) {
    // Clone TodoMVC React implementation
    // Run analyzer
    // Verify reasonable number of diagnostics
    // Manual review of false positives
}

func TestRealWorld_ReactRouter(t *testing.T) {
    // Test on react-router examples
}

func TestRealWorld_MaterialUI(t *testing.T) {
    // Test on Material-UI components
    // Expect many issues (large codebase)
}
```

**Performance Tests** (`test/benchmarks/rules_bench_test.go`)
```go
func BenchmarkNoObjectDeps_SmallFile(b *testing.B) {
    // Target: <500μs per file
}

func BenchmarkNoUnstableProps_MediumFile(b *testing.B) {
    // Target: <1ms per file
}

func BenchmarkMemoUnstableProps_WithImports(b *testing.B) {
    // Includes import resolution
    // Target: <5ms per file
}

func BenchmarkFullAnalysis_100Files(b *testing.B) {
    // All rules on 100 files
    // Target: <10s
}
```

**Success Criteria**
- ✅ All 3 rules implemented and tested
- ✅ Unit test coverage >90%
- ✅ E2E tests pass on all scenarios
- ✅ False positive rate <10% (manual review)
- ✅ Performance targets met
- ✅ Auto-fix suggestions work correctly

---

## Phase 2: VS Code Extension (Weeks 7-10)

### Week 7-8: LSP Server Implementation

**Goal**: Build Language Server Protocol server that communicates with VS Code

#### Implementation Tasks

1. **LSP Server** (`cmd/lsp/main.go`)
   - JSON-RPC communication over stdio
   - Document synchronization
   - Incremental updates

2. **Protocol Handlers** (`internal/lsp/handlers.go`)
   - `textDocument/didOpen`
   - `textDocument/didChange`
   - `textDocument/publishDiagnostics`
   - `textDocument/codeAction`

3. **Go Engine Bridge** (`internal/lsp/engine.go`)
   - Long-lived process
   - Request queue
   - Response correlation

#### Testing Strategy

**Unit Tests** (`test/lsp/`)
- `server_test.go`:
  - ✅ Initialize LSP server
  - ✅ Handle didOpen notification
  - ✅ Handle didChange notification
  - ✅ Publish diagnostics
  - ✅ Provide code actions
  - ✅ Handle concurrent requests

**Integration Tests** (`test/lsp_integration/`)
```go
func TestLSP_FullLifecycle(t *testing.T) {
    // Start LSP server
    // Send initialize request
    // Open document
    // Receive diagnostics
    // Request code actions
    // Apply fix
    // Verify diagnostics cleared
}

func TestLSP_IncrementalUpdates(t *testing.T) {
    // Open document
    // Make edits
    // Verify only changed portions re-analyzed
}
```

**Performance Tests**
```go
func BenchmarkLSP_DidChange(b *testing.B) {
    // Target: <100ms response time
}

func BenchmarkLSP_CodeAction(b *testing.B) {
    // Target: <50ms response time
}
```

**Success Criteria**
- ✅ LSP server responds to all requests
- ✅ Diagnostics update on file change
- ✅ Response time <100ms (p95)
- ✅ No crashes on malformed requests

---

### Week 9-10: VS Code Extension

**Goal**: Package LSP server as VS Code extension

#### Implementation Tasks

1. **Extension Client** (`extensions/vscode/src/extension.ts`)
   - Activate extension
   - Start LSP server (Go binary)
   - Handle configuration

2. **Engine Process Manager** (`extensions/vscode/src/engine.ts`)
   - Spawn Go binary as subprocess
   - Handle stdio communication
   - Restart on crash

3. **Configuration** (`extensions/vscode/package.json`)
   - Extension metadata
   - Configuration schema
   - Commands

4. **Binary Bundling**
   - Cross-compile Go binary (Linux, macOS, Windows)
   - Bundle with extension
   - Auto-select correct binary

#### Testing Strategy

**Unit Tests** (TypeScript: `extensions/vscode/src/test/`)
```typescript
describe('Extension Activation', () => {
    it('should activate successfully', async () => {
        // Test extension loads
    });

    it('should start LSP server', async () => {
        // Test Go binary spawns
    });
});

describe('Diagnostics', () => {
    it('should show diagnostics on file open', async () => {
        // Open fixture file
        // Wait for diagnostics
        // Verify diagnostic count and content
    });
});
```

**E2E Tests** (VS Code extension test runner)
```typescript
describe('E2E: User Workflow', () => {
    it('should detect issue and apply fix', async () => {
        // Open test file with unstable prop
        // Wait for diagnostic to appear
        // Trigger code action
        // Apply fix
        // Verify diagnostic disappears
        // Verify code changed correctly
    });
});
```

**Manual Testing Checklist**
- [ ] Install extension in VS Code
- [ ] Open React project
- [ ] See diagnostics appear
- [ ] Hover over diagnostic - see description
- [ ] Trigger quick fix (Cmd+.)
- [ ] Apply fix - code updates correctly
- [ ] Save file - diagnostic disappears
- [ ] Open settings - configuration works
- [ ] Disable rule - diagnostic disappears
- [ ] Test on Windows/Mac/Linux

**Performance Testing**
- Open large React project (1000+ files)
- Measure time to first diagnostic
- Target: <10s for full project scan
- Target: <500ms for single file change

**Success Criteria**
- ✅ Extension installs successfully
- ✅ Diagnostics appear within 10s
- ✅ Code actions work correctly
- ✅ No crashes or hangs
- ✅ Works on all platforms

---

## Phase 3: CLI & CI Integration (Weeks 11-14)

### Week 11-12: CLI Tool

**Goal**: Build command-line interface for batch analysis

#### Implementation Tasks

1. **CLI Framework** (`cmd/react-analyzer/main.go`)
   - Use Cobra for commands
   - Subcommands: analyze, init, rules, fix

2. **Batch Analysis** (`internal/cli/analyze.go`)
   - Discover React files in directory
   - Parallel analysis
   - Progress reporting

3. **Output Formatters** (`internal/cli/formatters/`)
   - Text output (human-readable)
   - JSON output (machine-readable)
   - SARIF output (GitHub Code Scanning)

4. **Auto-fix Mode** (`internal/cli/fix.go`)
   - Apply auto-fixes to files
   - Dry-run mode
   - Backup original files

#### Testing Strategy

**Unit Tests** (`test/cli/`)
```go
func TestCLI_Analyze_SingleFile(t *testing.T) {
    output := runCLI("analyze", "test/fixtures/component.tsx")
    assert.Contains(t, output, "1 error")
}

func TestCLI_Analyze_Directory(t *testing.T) {
    output := runCLI("analyze", "test/fixtures/project")
    assert.Contains(t, output, "src/App.tsx")
}

func TestCLI_OutputFormat_JSON(t *testing.T) {
    output := runCLI("analyze", "--format=json", "test/fixtures/component.tsx")
    var result AnalysisResult
    json.Unmarshal([]byte(output), &result)
    assert.Len(t, result.Diagnostics, 1)
}

func TestCLI_Fix_DryRun(t *testing.T) {
    output := runCLI("analyze", "--fix", "--dry-run", "test/fixtures/component.tsx")
    // Verify file not modified
}

func TestCLI_Fix_Apply(t *testing.T) {
    tempFile := copyFixture("component.tsx")
    runCLI("analyze", "--fix", tempFile)
    // Verify file modified correctly
}
```

**Integration Tests** (`test/cli_integration/`)
```go
func TestCLI_FullProject(t *testing.T) {
    // Create temporary project
    // Run analysis
    // Verify exit code
    // Verify output format
}

func TestCLI_GitIntegration(t *testing.T) {
    // Analyze only changed files (git diff)
}
```

**Performance Tests**
```go
func BenchmarkCLI_SmallProject(b *testing.B) {
    // 10 files, 1k LOC
    // Target: <1s
}

func BenchmarkCLI_MediumProject(b *testing.B) {
    // 100 files, 10k LOC
    // Target: <10s
}

func BenchmarkCLI_LargeProject(b *testing.B) {
    // 1000 files, 100k LOC
    // Target: <60s
}
```

**Success Criteria**
- ✅ CLI commands work correctly
- ✅ All output formats valid
- ✅ Auto-fix applies correctly
- ✅ Performance targets met
- ✅ Exit codes correct (0 = no errors, 1 = errors found)

---

### Week 13-14: CI/CD Integration

**Goal**: Enable react-analyzer in CI pipelines

#### Implementation Tasks

1. **GitHub Actions** (`.github/workflows/examples/`)
   - Example workflow file
   - SARIF upload
   - PR comments

2. **Pre-commit Hook** (`scripts/pre-commit.sh`)
   - Analyze staged files
   - Block commit on errors

3. **Documentation** (`docs/ci-integration.md`)
   - GitHub Actions guide
   - GitLab CI guide
   - Pre-commit hook guide

#### Testing Strategy

**Integration Tests**
```go
func TestCI_GitHubActions(t *testing.T) {
    // Run in GitHub Actions environment
    // Verify SARIF output
}

func TestCI_ExitCodes(t *testing.T) {
    // No errors: exit 0
    // Warnings only: exit 0
    // Errors: exit 1
}
```

**E2E CI Testing**
- Create test repository
- Set up GitHub Action
- Push code with issues
- Verify:
  - [ ] Action runs
  - [ ] Diagnostics uploaded
  - [ ] PR comment appears
  - [ ] Check fails on errors

**Success Criteria**
- ✅ GitHub Action example works
- ✅ SARIF format valid
- ✅ Pre-commit hook blocks bad commits
- ✅ Documentation clear

---

## Comprehensive Testing Strategy Summary

### Test Pyramid

```
                 /\
                /  \
               /E2E \          10 tests - Real projects
              /------\
             /        \
            / Integration\    50 tests - Multi-component
           /------------\
          /              \
         /  Unit Tests    \   200 tests - Individual functions
        /------------------\
```

### Test Categories

#### 1. Unit Tests (200+ tests)
**Coverage Target**: >90%

**Locations**:
- `test/ast/` - Parser and semantic analysis
- `test/rules/` - Individual rule logic
- `test/analyzer/` - Context building
- `test/lsp/` - LSP protocol handlers
- `test/cli/` - CLI commands

**Run Frequency**: On every commit (CI)

**Duration Target**: <5s for full suite

#### 2. Integration Tests (50+ tests)
**Coverage**: Component interaction

**Locations**:
- `test/integration/` - Rule engine + parser
- `test/lsp_integration/` - LSP server + engine
- `test/cli_integration/` - CLI + analysis engine

**Run Frequency**: On PR, before merge

**Duration Target**: <30s

#### 3. End-to-End Tests (10+ tests)
**Coverage**: Complete workflows

**Locations**:
- `test/e2e/` - Full analysis scenarios
- `extensions/vscode/src/test/` - VS Code extension

**Run Frequency**: Nightly, before release

**Duration Target**: <5min

#### 4. Real-World Tests (5+ projects)
**Coverage**: Production React codebases

**Test Projects**:
- TodoMVC (React implementation)
- react-router examples
- Material-UI components
- Internal company projects
- Synthetic test project (all edge cases)

**Metrics**:
- False positive rate
- Detection rate
- Performance on large codebases

**Run Frequency**: Weekly, before major releases

**Duration Target**: <10min per project

#### 5. Performance Tests (Benchmarks)
**Coverage**: Speed and memory

**Locations**:
- `test/benchmarks/` - All performance tests

**Metrics**:
- Parse time (μs per file)
- Analysis time (ms per file)
- Memory usage (MB per file)
- Throughput (files per second)
- Incremental update time

**Run Frequency**: On performance-related PRs, weekly

**Baseline Tracking**: Store results, alert on regressions >10%

#### 6. Fuzz Testing
**Coverage**: Edge cases and crashes

```go
func FuzzParser(f *testing.F) {
    // Fuzz with random TypeScript/JSX
    // Should not crash
}

func FuzzRuleEngine(f *testing.F) {
    // Fuzz rule inputs
    // Should not panic
}
```

**Run Frequency**: Nightly

**Duration**: 1 hour (continuous fuzzing)

### Test Infrastructure

#### CI/CD Pipeline
```yaml
name: Tests
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./... -short -cover
      - target: <5s, coverage >90%

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./test/integration/... -v
      - target: <30s

  e2e-tests:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - run: go test ./test/e2e/... -v
      - target: <5min per platform

  performance-tests:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./test/benchmarks/... -bench=. -benchmem
      - run: scripts/check-performance-regression.sh

  real-world-tests:
    runs-on: ubuntu-latest
    steps:
      - run: scripts/test-real-world-projects.sh
      - run: scripts/analyze-false-positives.sh
```

#### Quality Gates
Before merging PR:
- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Code coverage >90%
- ✅ No performance regressions >10%
- ✅ go test -race passes (no data races)
- ✅ Linters pass (golangci-lint)

Before release:
- ✅ All E2E tests pass on all platforms
- ✅ Real-world tests show <10% false positive rate
- ✅ Performance benchmarks meet targets
- ✅ Manual testing checklist complete

### Test Data Management

#### Fixture Organization
```
test/fixtures/
├── parser/               # Valid/invalid syntax
├── rules/                # Specific rule test cases
│   ├── no-object-deps/
│   ├── no-unstable-props/
│   └── memo-unstable-props/
├── projects/             # Mini React projects
│   ├── small/           # 10 files
│   ├── medium/          # 100 files
│   └── large/           # 1000 files
└── real-world/          # Cloned open-source projects
    ├── todomvc/
    ├── react-router-examples/
    └── material-ui-components/
```

#### Golden Files
Store expected output for regression testing:
```
test/golden/
├── diagnostics/
│   ├── scenario1.json    # Expected diagnostics
│   └── scenario2.json
└── fixes/
    ├── unstable-deps-fix.tsx    # Expected code after auto-fix
    └── memo-breaking-fix.tsx
```

### Testing Tools

**Go Testing Tools**:
- `testing` - Standard library
- `testify/assert` - Assertions
- `testify/require` - Fatal assertions
- `go-cmp` - Deep comparison
- `goleak` - Goroutine leak detection

**Performance Tools**:
- `testing.B` - Benchmarking
- `pprof` - CPU/memory profiling
- `trace` - Execution tracing

**VS Code Extension Testing**:
- `@vscode/test-electron` - Extension test runner
- `mocha` - Test framework
- `chai` - Assertions

### Success Metrics

**Phase 1 (Core Engine)**:
- Unit test coverage: >90%
- All 3 MVP rules: 100% tests passing
- Performance targets: All met
- False positive rate: <10%

**Phase 2 (VS Code)**:
- Extension installs: Success rate >95%
- Diagnostic accuracy: >90%
- Response time: <100ms (p95)
- Crash rate: <0.1%

**Phase 3 (CLI/CI)**:
- CLI exit codes: 100% correct
- SARIF format: 100% valid
- GitHub Actions: Works on public repos
- Performance: Meets all targets

---

## Risk Mitigation & Monitoring

### Continuous Monitoring

1. **Performance Regression Detection**
   - Store benchmark results on main branch
   - Alert if >10% slower
   - Block merge if >20% slower

2. **False Positive Tracking**
   - User feedback mechanism
   - Weekly review of dismissed diagnostics
   - Adjust rules based on data

3. **Crash Reporting**
   - Opt-in telemetry
   - Stack trace collection
   - Weekly review

### Quality Assurance Process

**Before Each Release**:
1. Run full test suite (all categories)
2. Manual testing on 3 real projects
3. Performance profiling
4. False positive review
5. Documentation update
6. Changelog generation

---

## Phase Timeline Summary

| Phase | Duration | Key Deliverables | Tests |
|-------|----------|------------------|-------|
| **Phase 0** | ✅ Done | POC, tree-sitter validation | Benchmark tests |
| **Phase 1** | Weeks 1-6 | Parser, rules, engine | 200+ unit, 50+ integration, 10+ e2e |
| **Phase 2** | Weeks 7-10 | LSP server, VS Code extension | LSP tests, extension e2e |
| **Phase 3** | Weeks 11-14 | CLI, CI/CD integration | CLI tests, CI validation |
| **Total** | 14 weeks | Production-ready tool | Full test coverage |

---

## Next Steps

### Immediate Actions (Week 1)
1. Set up project structure (`cmd/`, `internal/`, `test/`)
2. Set up CI pipeline (GitHub Actions)
3. Create test fixtures directory structure
4. Implement parser interface
5. Write first unit tests
6. Set up benchmarking infrastructure

### Success Definition
After 14 weeks, we will have:
- ✅ Production-ready static analysis tool
- ✅ 3 working rules with <10% false positive rate
- ✅ VS Code extension installable from marketplace
- ✅ CLI tool for CI/CD integration
- ✅ Comprehensive test coverage (>90%)
- ✅ All performance targets met
- ✅ Documentation complete

This implementation plan provides a structured, test-driven approach to building react-analyzer with confidence in quality, performance, and reliability at every step.
