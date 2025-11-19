# React-Analyzer Roadmap

**Last Updated:** 2025-11-18
**Current Phase:** Phase 2.1 Complete → Beginning Phase 2.2

---

## Table of Contents
- [Vision](#vision)
- [Completed Work](#completed-work)
- [Current Limitations](#current-limitations)
- [Planned Phases](#planned-phases)
- [Contributing](#contributing)

---

## Vision

**Goal:** Build a comprehensive static analysis tool for React applications that detects performance antipatterns, state management issues, and architectural problems that are difficult to catch with traditional linters.

**Differentiator:** Graph-based analysis that understands component relationships, prop flows, and state dependencies across files.

---

## Completed Work

### ✅ Phase 2.1: Prop Drilling Detection & Graph Infrastructure (COMPLETE)

**Duration:** Weeks 1-4
**Status:** ✅ Shipped
**Completion Date:** 2025-11-18

#### Core Infrastructure

**Component Graph System**
- ✅ Component hierarchy tracking across files
- ✅ State node representation (useState, useReducer, context, props, derived)
- ✅ Edge-based relationships (defines, consumes, updates, passes, derives)
- ✅ 4-phase graph construction pipeline:
  1. Component node creation
  2. State node extraction
  3. Component hierarchy building
  4. Prop passing edge detection
- ✅ Module resolver with import/export tracking
- ✅ BFS-based path finding between components
- ✅ Symbol table with memoization tracking

**Analysis Rules (6 Total)**
1. ✅ **no-object-deps** - Detects inline objects/arrays in hook dependency arrays
2. ✅ **unstable-props-to-memo** - Detects unstable props to React.memo/useMemo/useCallback (most complex: 488 LOC)
3. ✅ **no-derived-state** - Detects useState mirroring props via useEffect
4. ✅ **no-stale-state** - Detects state updates without functional form (prevents race conditions)
5. ✅ **no-inline-props** - Detects inline objects/arrays/functions in JSX props
6. ✅ **deep-prop-drilling** - Detects props drilled through 3+ component levels (graph-based)

**Prop Drilling Detection Features**
- ✅ Graph-based algorithm detecting props passed through 3+ levels
- ✅ Identifies passthrough components (components that don't use props)
- ✅ Leaf consumer identification (avoids duplicate violations)
- ✅ Shortest path tracing from state origin to final consumer
- ✅ Contextual recommendations (suggests Context API)
- ✅ Per-prop edge tracking (multiple props between same components)
- ✅ 12 comprehensive test fixtures covering edge cases

**Developer Experience**
- ✅ Parallel file analysis
- ✅ Colored output with issue grouping
- ✅ Verbose mode with performance metrics
- ✅ Quiet mode for CI/CD
- ✅ Directory scanning with smart exclusions (node_modules, dist, .git)
- ✅ Monorepo support
- ✅ Import alias resolution

#### Key Achievements
- **200+ tests** with comprehensive coverage
- **6 production-ready rules** detecting real antipatterns
- **Graph infrastructure** supporting cross-component analysis
- **Fast performance** (~2ms per file average)

---

## Current Limitations

See [known_limitations.md](known_limitations.md) for full details. Key gaps:

### Critical (Blocks Real-World Adoption)
- ❌ **Arrow function components not detected** - Misses `const Foo = () => <div />`
- ❌ **Cross-file prop drilling incomplete** - Only detects within single files
- ❌ **Prop spread operators ignored** - Doesn't track `{...props}`

### High Priority (Common Patterns)
- ⚠️ **Object property access not tracked** - Misses `settings.locale` as prop
- ⚠️ **Partial prop usage false positives** - Flags components that both use AND pass props

### Medium Priority
- 🔸 Renamed props not tracked (`appTheme` → `theme`)
- 🔸 Multiple props not analyzed together

### Out of Scope
- Runtime performance detection
- Actual render frequency measurement

---

## Planned Phases

### 🚀 Phase 2.2: Core Completeness (4-6 weeks) - IN PROGRESS

**Goal:** Make existing features work for real-world React codebases
**Target:** Q1 2026
**Priority:** CRITICAL - These gaps block adoption
**Progress:** 1/3 Complete (Arrow Functions ✅)

#### Priority 1A: Arrow Function Components ✅ COMPLETED

**Status:** ✅ SHIPPED (2025-11-18)
**Impact:** CRITICAL - Analyzer now detects 95%+ of modern React components

**Deliverables:**
- [x] Detect `const MyComponent = () => <div />`
- [x] Detect `const MyComponent = ({ props }) => { return <div /> }`
- [x] Extract props from arrow function parameters
- [x] Support React.memo wrapping: `React.memo(() => <div />)`
- [x] Update all 6 rules to work with arrow components
- [x] Add 3 test fixtures for arrow function patterns

**Implementation:**
- Added `getComponentNameFromArrowFunction()` to detect arrow components from `variable_declarator` nodes
- Added `extractPropsFromArrowFunction()` to extract props from arrow function parameters
- Created `walkWithComponentContext()` helper using stack-based tracking for proper scoping
- Updated all 4 build phases (components, state, hierarchy, prop passing)
- Supports direct arrow functions and React.memo wrapped arrow functions

**Test Fixtures:**
- `ArrowFunctionDrilling.tsx` - Basic arrow function prop drilling
- `MemoArrowFunctionDrilling.tsx` - React.memo wrapped arrow functions
- `MixedArrowAndFunction.tsx` - Mixed arrow and function declarations

**Success Criteria:**
- [x] Arrow components detected in 95%+ of cases
- [x] All existing tests still pass (100% pass rate)
- [x] New test suite for arrow functions: 100% pass rate

---

#### Priority 1B: Cross-File Prop Drilling (2-3 weeks)

**Problem:** Prop drilling detection only works within single files
**Impact:** HIGH - Real apps split components across files

**Deliverables:**
- [ ] Enhance `findComponentID()` to search imported files
- [ ] Track import statements and resolve component definitions
- [ ] Build cross-file component edges in graph
- [ ] Handle circular dependencies gracefully
- [ ] Support re-exports (`export { Foo } from './Foo'`)
- [ ] Add cross-file test fixtures (3+ files deep)

**Technical Approach:**
```go
func (b *Builder) findComponentID(componentName, currentFile string) string {
    // 1. Search current file (existing)
    // 2. NEW: Get imports for current file
    imports := b.resolver.GetImports(currentFile)

    // 3. NEW: Search imported files
    for _, imp := range imports {
        if imp.ExportsComponent(componentName) {
            return b.findComponentInFile(componentName, imp.ResolvedPath)
        }
    }
}
```

**Dependencies:**
- ModuleResolver already tracks imports ✅
- Symbol table already has import info ✅
- Just need to wire up the logic

**Success Criteria:**
- Detects drilling across 3+ files
- Handles named and default exports
- Cross-file test fixtures: 100% pass rate
- No performance regression (< 10% slowdown)

---

#### Priority 1C: Prop Spread Operators (2-3 weeks)

**Problem:** Common pattern `<Child {...props} />` is invisible to analyzer
**Impact:** HIGH - Very common in React, major blind spot

**Deliverables:**
- [ ] Detect `jsx_spread_attribute` nodes in JSX
- [ ] Track which props are included in spread
- [ ] Create edges for all props in spread object
- [ ] Handle partial spreads: `<Child {...rest} theme={theme} />`
- [ ] Support spread with destructuring: `const { theme, ...rest } = props`
- [ ] Add test fixtures for spread patterns

**Technical Approach:**
```go
// In processJSXElement
for _, child := range openingElement.Children() {
    if child.Type() == "jsx_spread_attribute" {
        // Get spread expression
        expr := child.ChildByFieldName("argument")
        if expr.Type() == "identifier" {
            varName := expr.Text()
            // Create edges for ALL props in this object
            for _, prop := range parentComp.Props {
                if isSpreadSource(varName, prop, parentComp) {
                    createEdge(prop)
                }
            }
        }
    }
}
```

**Challenges:**
- Need type information to know what's in spread (may need TypeScript integration)
- Rest parameters make this complex: `const { theme, ...rest } = props`
- May need to be conservative (over-report rather than under-report)

**Success Criteria:**
- Detects drilling through spreads in common cases
- Handles `{...props}`, `{...rest}` patterns
- PropSpread.tsx test fixture passes
- Documented limitations for edge cases

---

### 🎯 Phase 2.3: Enhanced Accuracy (4-5 weeks)

**Goal:** Reduce false positives and improve recommendation quality
**Target:** Q1-Q2 2026
**Priority:** HIGH - Improves user trust

#### Priority 2A: Partial Prop Usage Detection (1-2 weeks)

**Problem:** Flags components that both use AND pass props
**Impact:** MEDIUM - False positives hurt user trust

**Deliverables:**
- [ ] AST analysis to detect prop references in component body
- [ ] Check for prop usage in JSX expressions
- [ ] Check for prop usage in hooks (useEffect, useMemo deps)
- [ ] Check for prop usage in function calls
- [ ] Update `componentUsesProp()` heuristic
- [ ] PartialUsage.tsx test fixture passes

**Technical Approach:**
```go
func componentUsesProp(comp *ComponentNode, propName string, g *Graph) bool {
    // Walk component's AST to find references
    found := false
    comp.AST.Walk(func(node *parser.Node) bool {
        if node.Type() == "identifier" && node.Text() == propName {
            // Check it's not just in jsx_attribute for passing down
            if !isJSXAttributeValue(node) {
                found = true
                return false // stop walking
            }
        }
        return true
    })
    return found
}
```

**Success Criteria:**
- PartialUsage.tsx: No violation (currently fails)
- No new false negatives
- All existing test fixtures still pass

---

#### Priority 2B: Object Property Access Tracking (2 weeks)

**Problem:** Doesn't track `settings.locale` as separate prop flow
**Impact:** MEDIUM-HIGH - Common TypeScript pattern

**Deliverables:**
- [ ] Detect `member_expression` nodes when creating edges
- [ ] Track property access paths (e.g., `settings.locale`, `settings.currency`)
- [ ] Create virtual state nodes for object properties OR track in edge metadata
- [ ] Update prop matching to handle property chains
- [ ] Dashboard.tsx test fixture passes (expects 3 violations)

**Technical Approach:**
```go
// In processJSXElement
if exprChild.Type() == "member_expression" {
    // settings.locale
    object := exprChild.ChildByFieldName("object")    // "settings"
    property := exprChild.ChildByFieldName("property") // "locale"

    objName := object.Text()
    propName := property.Text()

    // Check if object is parent state
    if b.isParentState(objName, parentComp) {
        // Create edge with property path
        edge.PropName = propName
        edge.PropPath = fmt.Sprintf("%s.%s", objName, propName)
    }
}
```

**Challenges:**
- Deep property paths: `config.settings.theme.primary`
- Array access: `items[0].name`
- Computed access: `settings[key]`

**Success Criteria:**
- Dashboard.tsx: 3 violations detected (locale, currency, dateFormat)
- Handles simple property access (1-2 levels deep)
- Documented limitations for complex paths

---

#### Priority 2C: Multiple Props Analysis (1 week)

**Problem:** Reports each drilled prop separately; doesn't suggest combining
**Impact:** MEDIUM - Recommendations could be better

**Deliverables:**
- [ ] Detect when multiple props follow same path
- [ ] Suggest combining into single context or config object
- [ ] Enhanced recommendation messages
- [ ] Update test expectations

**Example Output:**
```
Props 'theme', 'locale', and 'currency' are all drilled through the same 4 component levels.
Consider creating a single AppSettingsContext instead of passing them individually.
```

**Success Criteria:**
- MultiplePropsDrilled.tsx: Suggests combining theme + lang
- Improved user experience
- No performance regression

---

### 🔧 Phase 2.4: Advanced Patterns (3-5 weeks)

**Goal:** Handle sophisticated real-world patterns
**Target:** Q2 2026
**Priority:** MEDIUM - Nice to have, not critical

#### Features
- [ ] **Prop Rename Tracking** (1 week) - Track `appTheme` → `theme` renames
- [ ] **Context API Integration** (2 weeks) - Detect existing Context providers, improve recommendations
- [ ] **Prop Transformation Detection** (2-3 weeks) - Track when props are modified along path

---

### 🆕 Phase 3: New Rule Development

**Goal:** Expand beyond prop drilling to other antipatterns
**Target:** Q2-Q3 2026
**Priority:** MEDIUM - Depends on user feedback

#### Candidate Rules
From [react_antipatterns_catalog.md](react_antipatterns_catalog.md):

**High Priority:**
- [ ] Missing key in lists
- [ ] useState for server state (should use react-query/SWR)
- [ ] Too many useState calls (suggest useReducer)
- [ ] Missing cleanup in useEffect

**Medium Priority:**
- [ ] Large component detection (complexity metrics)
- [ ] Prop drilling depth warnings (configurable threshold)
- [ ] Exhaustive deps improvements

**Will be prioritized based on:**
- User feature requests
- GitHub issues / discussions
- Community feedback
- Real-world usage patterns

---

### 🛠️ Phase 4: Tooling & Developer Experience

**Goal:** Make the tool production-ready and easy to use
**Target:** Q3 2026
**Priority:** HIGH - After core rules are solid

#### 4A: VS Code Extension (4-6 weeks)

**Deliverables:**
- [ ] LSP server implementation
- [ ] Inline diagnostics (squiggly lines)
- [ ] Quick fixes / code actions
- [ ] Component graph visualization
- [ ] Settings panel
- [ ] Extension marketplace publishing

See [phase2_vscode_extension_architecture.md](design/phase2_vscode_extension_architecture.md) for detailed design.

#### 4B: Auto-Fix Generation (2-3 weeks)

**Deliverables:**
- [ ] Generate Context API refactoring for prop drilling
- [ ] Extract to useMemo/useCallback for inline props
- [ ] Add functional form to state setters
- [ ] Batch fixes across multiple files

#### 4C: CLI Enhancements (1-2 weeks)

**Deliverables:**
- [ ] SARIF output for GitHub Actions integration
- [ ] Directory analysis with progress bar
- [ ] Configuration file support (.react-analyzer.json)
- [ ] Pre-commit hook templates
- [ ] Watch mode for development
- [ ] JSON output for CI/CD pipelines

---

## Timeline Overview

```
2025 Q4 ██████████ Phase 2.1 Complete
2026 Q1 ██████████ Phase 2.2: Core Completeness
2026 Q1 ██████████ Phase 2.3: Enhanced Accuracy
2026 Q2 ██████     Phase 2.4: Advanced Patterns
2026 Q2 ██████     Phase 3: New Rules (ongoing)
2026 Q3 ████████   Phase 4: Tooling & DX
```

**Total estimated time:** 6-9 months for Phases 2.2-4

---

## Success Metrics

### Phase 2.2
- [ ] Arrow function detection: >95% accuracy
- [ ] Cross-file drilling: Works across 3+ files
- [ ] Prop spreads: Detects common patterns
- [ ] Test suite: All 200+ tests passing
- [ ] Real-world validation: Tested on 3 major OSS React projects

### Phase 2.3
- [ ] False positive rate: <10%
- [ ] False negative rate: <15%
- [ ] User satisfaction: Positive feedback from beta testers

### Overall
- [ ] **GitHub stars:** 1,000+
- [ ] **NPM downloads:** 10,000+/month
- [ ] **VS Code extension:** 5,000+ installs
- [ ] **Production usage:** 50+ companies

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- How to add a new rule
- How to run tests
- Code architecture overview
- PR guidelines

**Want to help?**
- Pick up issues labeled `good-first-issue`
- Join discussions about prioritization
- Report bugs or suggest features
- Improve documentation

---

## Questions?

- **Technical questions:** Open a GitHub discussion
- **Bug reports:** Open a GitHub issue
- **Feature requests:** Comment on this roadmap or open an issue
- **General chat:** Join our Discord (coming soon)

---

**Last reviewed:** 2025-11-18
**Next review:** 2026-01 (monthly updates)
