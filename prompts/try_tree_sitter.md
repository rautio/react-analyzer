# Tree-sitter Parser Evaluation Plan

**Date:** 2025-11-13
**Status:** Ready for POC Testing
**Estimated Time:** 2-4 hours

---

## TL;DR Recommendation

**Use tree-sitter as the primary JavaScript/TypeScript/JSX parser.**

- ✅ Production-proven at GitHub scale (6M+ repos, 40k requests/min)
- ✅ Incremental parsing built-in (critical for IDE responsiveness)
- ✅ Native TypeScript + JSX support
- ✅ Designed specifically for code analysis tools
- ✅ 36x faster than traditional parsers in production benchmarks
- ⚠️ Uses CGO (minor cross-compilation complexity, but solved problem)

---

## Parser Options Comparison

| Parser | Pros | Cons | Verdict |
|--------|------|------|---------|
| **tree-sitter** | Production-proven, incremental parsing, TS/JSX support, huge community | Uses CGO | ✅ **RECOMMENDED** |
| **go-fAST** | Pure Go, simple deployment | Newer/less mature, unclear TS/JSX support | 🔶 Backup option |
| **esbuild** | Super fast, great TS support | AST not exposed in public API | ❌ Not viable |
| **Goja/Otto** | Pure Go | ES5 only, insufficient React support | ❌ Not viable |
| **Babel via Node** | Most complete JS parser | Slow startup, complex deployment | ❌ Not viable |

---

## Architecture: Parser Abstraction Layer

**Key Principle:** Abstract the parser behind an interface so you can swap implementations without rewriting analysis rules.

### Parser Interface

```go
// internal/ast/parser.go

package ast

import (
    "context"
)

// Parser interface - swap implementations without changing analysis code
type Parser interface {
    ParseFile(path string, content []byte) (*AST, error)
    ParseTypeScript(content []byte) (*AST, error)
    ParseJSX(content []byte) (*AST, error)
    Close() error
}

// Language represents supported languages
type Language int

const (
    JavaScript Language = iota
    TypeScript
    TSX
    JSX
)

// Normalized AST (your own format, independent of parser)
type AST struct {
    Root       Node
    SourceFile string
    Language   Language
}

// Node interface - common abstraction across all parsers
type Node interface {
    Type() NodeType
    Range() SourceRange
    Children() []Node
    Text() string

    // React-specific semantic methods
    IsComponent() bool
    IsFunctionDeclaration() bool
    IsHookCall() bool
    GetJSXProps() []PropNode
    GetFunctionParams() []ParamNode
}

type NodeType string

const (
    NodeProgram           NodeType = "program"
    NodeFunctionDecl      NodeType = "function_declaration"
    NodeArrowFunction     NodeType = "arrow_function"
    NodeCallExpression    NodeType = "call_expression"
    NodeJSXElement        NodeType = "jsx_element"
    NodeJSXAttribute      NodeType = "jsx_attribute"
    NodeIdentifier        NodeType = "identifier"
    NodeArrayExpression   NodeType = "array_expression"
    NodeObjectExpression  NodeType = "object_expression"
)

type SourceRange struct {
    Start Position
    End   Position
}

type Position struct {
    Line   int
    Column int
    Offset int
}

// React-specific node types
type PropNode interface {
    Node
    Name() string
    Value() Node
}

type ParamNode interface {
    Node
    Name() string
    Type() string // TypeScript type annotation if present
}
```

### Tree-sitter Implementation

```go
// internal/ast/treesitter/parser.go

package treesitter

import (
    "context"
    "fmt"

    "your-project/internal/ast"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/javascript"
    "github.com/smacker/go-tree-sitter/typescript/tsx"
    "github.com/smacker/go-tree-sitter/typescript/typescript"
)

type TreeSitterParser struct {
    parser      *sitter.Parser
    jsLanguage  *sitter.Language
    tsLanguage  *sitter.Language
    tsxLanguage *sitter.Language
}

func NewParser() (*TreeSitterParser, error) {
    parser := sitter.NewParser()

    return &TreeSitterParser{
        parser:      parser,
        jsLanguage:  javascript.GetLanguage(),
        tsLanguage:  typescript.GetLanguage(),
        tsxLanguage: tsx.GetLanguage(),
    }, nil
}

func (p *TreeSitterParser) ParseFile(path string, content []byte) (*ast.AST, error) {
    // Detect language from file extension
    lang := detectLanguage(path)

    switch lang {
    case ast.TypeScript:
        return p.ParseTypeScript(content)
    case ast.TSX:
        return p.ParseJSX(content) // TSX uses same grammar
    case ast.JavaScript:
        fallthrough
    default:
        return p.parseWithLanguage(content, p.jsLanguage, lang, path)
    }
}

func (p *TreeSitterParser) ParseTypeScript(content []byte) (*ast.AST, error) {
    return p.parseWithLanguage(content, p.tsLanguage, ast.TypeScript, "")
}

func (p *TreeSitterParser) ParseJSX(content []byte) (*ast.AST, error) {
    return p.parseWithLanguage(content, p.tsxLanguage, ast.TSX, "")
}

func (p *TreeSitterParser) parseWithLanguage(
    content []byte,
    lang *sitter.Language,
    astLang ast.Language,
    sourcePath string,
) (*ast.AST, error) {
    p.parser.SetLanguage(lang)

    tree, err := p.parser.ParseCtx(context.Background(), nil, content)
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }

    // Don't close tree yet - we need it for the AST
    // Caller must call tree.Close() or AST.Close()

    root := p.convertNode(tree.RootNode(), content)

    return &ast.AST{
        Root:       root,
        SourceFile: sourcePath,
        Language:   astLang,
        tree:       tree, // Keep reference for cleanup
    }, nil
}

func (p *TreeSitterParser) convertNode(node *sitter.Node, content []byte) ast.Node {
    return &TreeSitterNode{
        node:    node,
        content: content,
    }
}

func (p *TreeSitterParser) Close() error {
    // Parser doesn't need cleanup, but implement for interface
    return nil
}

func detectLanguage(path string) ast.Language {
    switch {
    case strings.HasSuffix(path, ".tsx"):
        return ast.TSX
    case strings.HasSuffix(path, ".ts"):
        return ast.TypeScript
    case strings.HasSuffix(path, ".jsx"):
        return ast.JSX
    default:
        return ast.JavaScript
    }
}
```

### Tree-sitter Node Wrapper

```go
// internal/ast/treesitter/node.go

package treesitter

import (
    "strings"

    "your-project/internal/ast"
    sitter "github.com/smacker/go-tree-sitter"
)

type TreeSitterNode struct {
    node    *sitter.Node
    content []byte
}

func (n *TreeSitterNode) Type() ast.NodeType {
    return ast.NodeType(n.node.Type())
}

func (n *TreeSitterNode) Range() ast.SourceRange {
    start := n.node.StartPoint()
    end := n.node.EndPoint()

    return ast.SourceRange{
        Start: ast.Position{
            Line:   int(start.Row) + 1, // tree-sitter uses 0-based
            Column: int(start.Column) + 1,
            Offset: int(n.node.StartByte()),
        },
        End: ast.Position{
            Line:   int(end.Row) + 1,
            Column: int(end.Column) + 1,
            Offset: int(n.node.EndByte()),
        },
    }
}

func (n *TreeSitterNode) Children() []ast.Node {
    count := int(n.node.ChildCount())
    children := make([]ast.Node, count)

    for i := 0; i < count; i++ {
        child := n.node.Child(i)
        children[i] = &TreeSitterNode{
            node:    child,
            content: n.content,
        }
    }

    return children
}

func (n *TreeSitterNode) Text() string {
    return n.node.Content(n.content)
}

// React-specific methods
func (n *TreeSitterNode) IsComponent() bool {
    nodeType := n.node.Type()

    // Function declaration or arrow function that returns JSX
    if nodeType == "function_declaration" || nodeType == "arrow_function" {
        // Check if it returns JSX (has jsx_element in body)
        return n.containsJSX()
    }

    return false
}

func (n *TreeSitterNode) IsFunctionDeclaration() bool {
    nodeType := n.node.Type()
    return nodeType == "function_declaration" ||
           nodeType == "arrow_function" ||
           nodeType == "function_expression"
}

func (n *TreeSitterNode) IsHookCall() bool {
    if n.node.Type() != "call_expression" {
        return false
    }

    // Get function name
    funcNode := n.node.ChildByFieldName("function")
    if funcNode == nil {
        return false
    }

    funcName := funcNode.Content(n.content)

    // Hook names start with "use"
    return strings.HasPrefix(funcName, "use")
}

func (n *TreeSitterNode) GetJSXProps() []ast.PropNode {
    // Only valid for jsx_element or jsx_self_closing_element
    nodeType := n.node.Type()
    if nodeType != "jsx_element" && nodeType != "jsx_self_closing_element" {
        return nil
    }

    var props []ast.PropNode

    // Find jsx_opening_element
    for i := 0; i < int(n.node.ChildCount()); i++ {
        child := n.node.Child(i)
        if child.Type() == "jsx_opening_element" {
            // Get all jsx_attribute children
            for j := 0; j < int(child.ChildCount()); j++ {
                attr := child.Child(j)
                if attr.Type() == "jsx_attribute" {
                    props = append(props, &TreeSitterPropNode{
                        TreeSitterNode: TreeSitterNode{
                            node:    attr,
                            content: n.content,
                        },
                    })
                }
            }
        }
    }

    return props
}

func (n *TreeSitterNode) GetFunctionParams() []ast.ParamNode {
    // Find formal_parameters node
    var params []ast.ParamNode

    paramsNode := n.node.ChildByFieldName("parameters")
    if paramsNode == nil {
        return params
    }

    // Extract each parameter
    for i := 0; i < int(paramsNode.ChildCount()); i++ {
        child := paramsNode.Child(i)
        if child.Type() == "required_parameter" || child.Type() == "optional_parameter" {
            params = append(params, &TreeSitterParamNode{
                TreeSitterNode: TreeSitterNode{
                    node:    child,
                    content: n.content,
                },
            })
        }
    }

    return params
}

func (n *TreeSitterNode) containsJSX() bool {
    // Recursive check for JSX elements
    if strings.HasPrefix(n.node.Type(), "jsx_") {
        return true
    }

    for i := 0; i < int(n.node.ChildCount()); i++ {
        child := &TreeSitterNode{
            node:    n.node.Child(i),
            content: n.content,
        }
        if child.containsJSX() {
            return true
        }
    }

    return false
}

// TreeSitterPropNode implements PropNode
type TreeSitterPropNode struct {
    TreeSitterNode
}

func (p *TreeSitterPropNode) Name() string {
    nameNode := p.node.ChildByFieldName("name")
    if nameNode != nil {
        return nameNode.Content(p.content)
    }
    return ""
}

func (p *TreeSitterPropNode) Value() ast.Node {
    valueNode := p.node.ChildByFieldName("value")
    if valueNode != nil {
        return &TreeSitterNode{
            node:    valueNode,
            content: p.content,
        }
    }
    return nil
}

// TreeSitterParamNode implements ParamNode
type TreeSitterParamNode struct {
    TreeSitterNode
}

func (p *TreeSitterParamNode) Name() string {
    // Pattern: identifier or destructuring
    patternNode := p.node.ChildByFieldName("pattern")
    if patternNode != nil {
        return patternNode.Content(p.content)
    }
    return ""
}

func (p *TreeSitterParamNode) Type() string {
    // TypeScript type annotation
    typeNode := p.node.ChildByFieldName("type")
    if typeNode != nil {
        return typeNode.Content(p.content)
    }
    return ""
}
```

---

## POC Implementation Plan

### Step 1: Setup (15 minutes)

```bash
# Create new Go module
mkdir react-analyzer-poc
cd react-analyzer-poc
go mod init github.com/yourorg/react-analyzer

# Install tree-sitter dependencies
go get github.com/smacker/go-tree-sitter
go get github.com/smacker/go-tree-sitter/javascript
go get github.com/smacker/go-tree-sitter/typescript/tsx
go get github.com/smacker/go-tree-sitter/typescript/typescript
```

### Step 2: Create Test File (10 minutes)

```bash
# Create test React component
cat > test_component.tsx <<'EOF'
import React, { useEffect, useState, useMemo } from 'react';

interface User {
  id: string;
  name: string;
}

interface Props {
  userId: string;
  onUserClick: (id: string) => void;
}

function UserProfile({ userId, onUserClick }: Props) {
  const [user, setUser] = useState<User | null>(null);

  // BAD: Missing dependency 'userId'
  useEffect(() => {
    fetchUser(userId).then(setUser);
  }, []);

  // BAD: Inline object in prop
  return (
    <div style={{ padding: 10 }}>
      <h1>{user?.name}</h1>
      <button onClick={() => onUserClick(userId)}>
        Click
      </button>
    </div>
  );
}

export default UserProfile;
EOF
```

### Step 3: Simple Parser Test (30 minutes)

```go
// main.go
package main

import (
    "context"
    "fmt"
    "os"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/typescript/tsx"
)

func main() {
    // Read test file
    content, err := os.ReadFile("test_component.tsx")
    if err != nil {
        panic(err)
    }

    // Create parser
    parser := sitter.NewParser()
    parser.SetLanguage(tsx.GetLanguage())

    // Parse
    tree, err := parser.ParseCtx(context.Background(), nil, content)
    if err != nil {
        panic(err)
    }
    defer tree.Close()

    // Print AST
    root := tree.RootNode()
    fmt.Printf("Root node type: %s\n", root.Type())
    fmt.Printf("Root has %d children\n", root.ChildCount())

    // Find all function declarations
    fmt.Println("\nFunction Declarations:")
    findFunctions(root, content)

    // Find all hook calls
    fmt.Println("\nHook Calls:")
    findHookCalls(root, content)

    // Find all JSX elements
    fmt.Println("\nJSX Elements:")
    findJSXElements(root, content)
}

func findFunctions(node *sitter.Node, content []byte) {
    if node.Type() == "function_declaration" {
        nameNode := node.ChildByFieldName("name")
        if nameNode != nil {
            fmt.Printf("  - %s (line %d)\n",
                nameNode.Content(content),
                node.StartPoint().Row + 1)
        }
    }

    for i := 0; i < int(node.ChildCount()); i++ {
        findFunctions(node.Child(i), content)
    }
}

func findHookCalls(node *sitter.Node, content []byte) {
    if node.Type() == "call_expression" {
        funcNode := node.ChildByFieldName("function")
        if funcNode != nil {
            funcName := funcNode.Content(content)
            if len(funcName) > 3 && funcName[:3] == "use" {
                fmt.Printf("  - %s (line %d)\n",
                    funcName,
                    node.StartPoint().Row + 1)

                // Print arguments
                argsNode := node.ChildByFieldName("arguments")
                if argsNode != nil {
                    fmt.Printf("    Args: %s\n", argsNode.Content(content))
                }
            }
        }
    }

    for i := 0; i < int(node.ChildCount()); i++ {
        findHookCalls(node.Child(i), content)
    }
}

func findJSXElements(node *sitter.Node, content []byte) {
    nodeType := node.Type()
    if nodeType == "jsx_element" || nodeType == "jsx_self_closing_element" {
        // Get element name
        var name string
        for i := 0; i < int(node.ChildCount()); i++ {
            child := node.Child(i)
            if child.Type() == "jsx_opening_element" {
                identNode := child.ChildByFieldName("name")
                if identNode != nil {
                    name = identNode.Content(content)
                }
            }
        }

        fmt.Printf("  - <%s> (line %d)\n",
            name,
            node.StartPoint().Row + 1)
    }

    for i := 0; i < int(node.ChildCount()); i++ {
        findJSXElements(node.Child(i), content)
    }
}
```

**Run it:**
```bash
go run main.go
```

**Expected Output:**
```
Root node type: program
Root has X children

Function Declarations:
  - UserProfile (line 14)

Hook Calls:
  - useEffect (line 18)
    Args: (() => { ... }, [])
  - useState (line 16)
    Args: (<User | null>(null))

JSX Elements:
  - <div> (line 23)
  - <h1> (line 24)
  - <button> (line 25)
```

### Step 4: Implement Missing Dependency Detection (45 minutes)

```go
// exhaustive_deps.go
package main

import (
    "fmt"
    "strings"

    sitter "github.com/smacker/go-tree-sitter"
)

type HookDependencyChecker struct {
    content []byte
}

type DependencyIssue struct {
    HookName     string
    Line         int
    Missing      []string
    Unnecessary  []string
}

func (c *HookDependencyChecker) CheckHook(node *sitter.Node) *DependencyIssue {
    if node.Type() != "call_expression" {
        return nil
    }

    funcNode := node.ChildByFieldName("function")
    if funcNode == nil {
        return nil
    }

    hookName := funcNode.Content(c.content)

    // Only check these hooks
    if hookName != "useEffect" && hookName != "useMemo" && hookName != "useCallback" {
        return nil
    }

    argsNode := node.ChildByFieldName("arguments")
    if argsNode == nil {
        return nil
    }

    // Get callback (first argument) and dependency array (second argument)
    var callbackNode, depsArrayNode *sitter.Node
    argCount := 0

    for i := 0; i < int(argsNode.ChildCount()); i++ {
        child := argsNode.Child(i)
        if child.Type() != "," {
            if argCount == 0 {
                callbackNode = child
            } else if argCount == 1 {
                depsArrayNode = child
            }
            argCount++
        }
    }

    if callbackNode == nil {
        return nil
    }

    // Find all identifiers referenced in callback
    referencedVars := c.findReferencedVariables(callbackNode)

    // Parse declared dependencies
    declaredDeps := c.parseDependencyArray(depsArrayNode)

    // Compare
    missing := difference(referencedVars, declaredDeps)
    unnecessary := difference(declaredDeps, referencedVars)

    if len(missing) > 0 || len(unnecessary) > 0 {
        return &DependencyIssue{
            HookName:    hookName,
            Line:        int(node.StartPoint().Row) + 1,
            Missing:     missing,
            Unnecessary: unnecessary,
        }
    }

    return nil
}

func (c *HookDependencyChecker) findReferencedVariables(node *sitter.Node) []string {
    vars := make(map[string]bool)
    c.findIdentifiers(node, vars)

    // Convert to slice
    result := make([]string, 0, len(vars))
    for v := range vars {
        // Filter out known globals and React APIs
        if !c.isKnownGlobal(v) {
            result = append(result, v)
        }
    }

    return result
}

func (c *HookDependencyChecker) findIdentifiers(node *sitter.Node, vars map[string]bool) {
    if node.Type() == "identifier" {
        // Check if it's a variable reference (not a property name or function declaration)
        parent := node.Parent()
        if parent != nil && parent.Type() == "member_expression" {
            // Skip property names (e.g., in obj.prop, skip 'prop')
            propertyNode := parent.ChildByFieldName("property")
            if propertyNode != nil && propertyNode.Equal(node) {
                return
            }
        }

        varName := node.Content(c.content)
        vars[varName] = true
    }

    for i := 0; i < int(node.ChildCount()); i++ {
        c.findIdentifiers(node.Child(i), vars)
    }
}

func (c *HookDependencyChecker) parseDependencyArray(node *sitter.Node) []string {
    if node == nil || node.Type() != "array" {
        return []string{}
    }

    deps := []string{}

    for i := 0; i < int(node.ChildCount()); i++ {
        child := node.Child(i)
        if child.Type() == "identifier" {
            deps = append(deps, child.Content(c.content))
        }
    }

    return deps
}

func (c *HookDependencyChecker) isKnownGlobal(name string) bool {
    globals := map[string]bool{
        "console":  true,
        "window":   true,
        "document": true,
        "fetch":    true,
        "setTimeout": true,
        "setInterval": true,
        // Add React-specific
        "React": true,
        // Add common functions (you'd expand this)
        "fetchUser": true,
        "setUser": true, // setState functions (heuristic: starts with 'set')
    }

    // Heuristic: setState functions
    if strings.HasPrefix(name, "set") && len(name) > 3 {
        next := name[3]
        if next >= 'A' && next <= 'Z' {
            return true // Likely a setState function
        }
    }

    return globals[name]
}

func difference(a, b []string) []string {
    bMap := make(map[string]bool)
    for _, item := range b {
        bMap[item] = true
    }

    result := []string{}
    for _, item := range a {
        if !bMap[item] {
            result = append(result, item)
        }
    }

    return result
}

// Add to main.go:
func analyzeHookDependencies(root *sitter.Node, content []byte) {
    checker := &HookDependencyChecker{content: content}

    var walk func(*sitter.Node)
    walk = func(node *sitter.Node) {
        if issue := checker.CheckHook(node); issue != nil {
            fmt.Printf("\n[ERROR] %s at line %d\n", issue.HookName, issue.Line)
            if len(issue.Missing) > 0 {
                fmt.Printf("  Missing dependencies: %v\n", issue.Missing)
            }
            if len(issue.Unnecessary) > 0 {
                fmt.Printf("  Unnecessary dependencies: %v\n", issue.Unnecessary)
            }
        }

        for i := 0; i < int(node.ChildCount()); i++ {
            walk(node.Child(i))
        }
    }

    walk(root)
}
```

**Add to main():**
```go
// After other analyses
fmt.Println("\nHook Dependency Issues:")
analyzeHookDependencies(root, content)
```

**Expected Output:**
```
Hook Dependency Issues:

[ERROR] useEffect at line 18
  Missing dependencies: [userId]
```

### Step 5: Benchmark Performance (30 minutes)

```go
// benchmark_test.go
package main

import (
    "context"
    "os"
    "testing"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/typescript/tsx"
)

func BenchmarkParsing(b *testing.B) {
    content, _ := os.ReadFile("test_component.tsx")

    parser := sitter.NewParser()
    parser.SetLanguage(tsx.GetLanguage())

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        tree, _ := parser.ParseCtx(context.Background(), nil, content)
        tree.Close()
    }
}

func BenchmarkFullAnalysis(b *testing.B) {
    content, _ := os.ReadFile("test_component.tsx")

    parser := sitter.NewParser()
    parser.SetLanguage(tsx.GetLanguage())

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        tree, _ := parser.ParseCtx(context.Background(), nil, content)

        // Run all analysis
        checker := &HookDependencyChecker{content: content}
        var walk func(*sitter.Node)
        walk = func(node *sitter.Node) {
            checker.CheckHook(node)
            for j := 0; j < int(node.ChildCount()); j++ {
                walk(node.Child(j))
            }
        }
        walk(tree.RootNode())

        tree.Close()
    }
}
```

**Run benchmarks:**
```bash
go test -bench=. -benchmem
```

**Target Performance:**
- Parsing: <1ms per component
- Full analysis: <5ms per component
- 100 components: <500ms

---

## Success Criteria for POC

After completing the POC, tree-sitter is viable if:

- ✅ Can parse real TypeScript/JSX without errors
- ✅ Can extract hooks, components, and JSX correctly
- ✅ Can detect at least one anti-pattern (exhaustive-deps)
- ✅ Performance: parse 100 components in <1 second
- ✅ Memory: reasonable memory usage (<100MB for 100 files)

If any criterion fails, fall back to evaluating **go-fAST** or reconsider architecture.

---

## Next Steps After POC

### If POC Succeeds

1. **Week 1-2:** Build parser abstraction layer
2. **Week 2-3:** Implement remaining MVP rules
3. **Week 3-4:** Build rule engine framework
4. **Week 4-6:** VS Code extension integration

### If POC Has Issues

**Performance issues:**
- Try incremental parsing (tree-sitter's killer feature)
- Profile and optimize traversal
- Consider parallel parsing

**API issues:**
- Extend wrapper to hide tree-sitter specifics
- Build helper functions for common patterns

**Compatibility issues:**
- Check tree-sitter grammar versions
- Test with more diverse React code samples
- Consider contributing grammar fixes upstream

---

## Cross-Compilation Setup

tree-sitter uses CGO, so cross-compilation requires setup:

```bash
# macOS → Linux
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  CC=x86_64-linux-musl-gcc \
  go build -ldflags="-linkmode external -extldflags -static"

# Use GitHub Actions for multi-platform builds
# .github/workflows/build.yml
name: Build
on: [push]
jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go build
```

**Docker alternative:**
```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go build -o react-analyzer

FROM alpine:latest
COPY --from=builder /app/react-analyzer /usr/local/bin/
CMD ["react-analyzer"]
```

---

## Resources

**Tree-sitter Documentation:**
- Main docs: https://tree-sitter.github.io/tree-sitter/
- Go bindings: https://github.com/smacker/go-tree-sitter
- TypeScript grammar: https://github.com/tree-sitter/tree-sitter-typescript
- Playground: https://tree-sitter.github.io/tree-sitter/playground

**Queries & Patterns:**
Tree-sitter has a query system for pattern matching:
```scheme
; Find all useEffect calls
(call_expression
  function: (identifier) @hook-name
  (#eq? @hook-name "useEffect")
  arguments: (arguments
    .
    (_)
    .
    (array) @deps))
```

This can simplify rule implementation - explore after POC.

---

## Alternative: go-fAST (Backup Plan)

If tree-sitter doesn't work out:

```bash
go get github.com/T14Raptor/go-fAST

# Test basic parsing
import "github.com/T14Raptor/go-fAST"

func testGoFAST() {
    code := `function App() { return <div>Hi</div>; }`
    ast, err := gofast.Parse(code)
    // ... evaluate
}
```

**Evaluate:**
- TypeScript support?
- JSX support?
- AST completeness?
- Performance?
- Documentation quality?

---

## Questions to Answer During POC

- [ ] How complete is the TypeScript AST? (type annotations accessible?)
- [ ] How are JSX props represented in the tree?
- [ ] Can we distinguish useState setters from regular functions?
- [ ] How do we handle destructured props?
- [ ] Performance on 1000-line file?
- [ ] Memory usage on 100 files?
- [ ] Incremental parsing: how much faster?
- [ ] Error recovery: does it parse files with syntax errors?

---

## Timeline

| Phase | Duration | Goal |
|-------|----------|------|
| Setup | 15 min | Install dependencies |
| Basic parsing | 30 min | Parse test file successfully |
| AST exploration | 30 min | Understand tree structure |
| Simple rule | 45 min | Implement exhaustive-deps detection |
| Benchmarking | 30 min | Validate performance |
| **Total** | **~2.5 hours** | **Go/No-go decision** |

---

## Decision Point

After POC completion:

**GO:** tree-sitter meets all criteria → Proceed with full implementation
**NO-GO:** Issues found → Evaluate alternatives or adjust architecture

Document findings and share with team for final decision.

---

**Ready to start? Run through steps 1-5 and report back!**
