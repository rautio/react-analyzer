# React Analyzer - MVP Scope Reduction

## Goal
Build a minimal viable product that proves the core concept: **static analysis of React code using tree-sitter, with at least one working rule that detects real performance issues**.

## What to KEEP for MVP

### Core Engine (Absolutely Essential)

1. **Parser Layer**
   - tree-sitter integration
   - Basic AST wrapper with parent pointers
   - Semantic helpers: `IsComponent()`, `IsHookCall()`, `IsMemoized()`
   - Language detection (.tsx, .jsx)
   - **SKIP**: Incremental parsing (nice-to-have, not MVP-critical)

2. **Scope Analysis**
   - Build scope tree (global, function, block)
   - Symbol table (track declarations)
   - Basic symbol resolution (scope chain lookup)
   - **SKIP**: Full stability analysis (can simplify)

3. **Rule Engine**
   - Simple rule interface
   - Rule registry
   - Sequential execution (no worker pool initially)
   - Basic diagnostic output
   - **SKIP**: Phase-based execution (just run all rules)
   - **SKIP**: Panic recovery (nice-to-have)
   - **SKIP**: Timeout protection (add later)

4. **MVP Rules** (Pick 1-2)
   - **Rule #1: `no-object-deps`** (MUST HAVE)
     - Detect inline object/array in hook deps
     - Single-file analysis only
     - Simple pattern matching

   - **Rule #2: `no-unstable-props`** (OPTIONAL)
     - Detect inline object/array/function in JSX props
     - Single-file only
     - Simpler than cross-file version

   - **SKIP for MVP: `memo-unstable-props`**
     - Requires cross-file analysis
     - Requires import resolution
     - Requires component graph
     - TOO COMPLEX for MVP

5. **Basic CLI**
   - `react-analyzer analyze <file>`
   - Text output only (human-readable)
   - Exit code: 0 = no errors, 1 = errors found
   - **SKIP**: JSON/SARIF output formats
   - **SKIP**: Multiple files/directory analysis (start with single file)
   - **SKIP**: Progress reporting
   - **SKIP**: Parallel analysis

## What to DEFER (Post-MVP)

### Phase 1.5: Optimizations & Polish
1. **Incremental Parsing**
   - Not needed for CLI that analyzes once and exits
   - Critical for LSP server later

2. **Memory Management**
   - LRU caching
   - Memory budgets
   - Cleanup strategies
   - **Rationale**: CLI processes are short-lived

3. **Worker Pool / Parallel Execution**
   - Start with sequential rule execution
   - Add parallelism when we have >3 rules
   - **Rationale**: Premature optimization

4. **Advanced Error Handling**
   - Panic recovery per rule
   - Timeout protection
   - Error metrics
   - **Rationale**: Can crash for MVP, fix bugs instead

### Phase 2: Cross-File Analysis (Defer Entirely)

1. **Import Resolution**
   - Relative imports (./Component)
   - tsconfig.json paths (@/*)
   - node_modules resolution
   - **Rationale**: Complex, not needed for MVP rules

2. **Component Graph**
   - Track component relationships
   - Prop flow analysis
   - **Rationale**: Only needed for `memo-unstable-props` rule

3. **Data Flow Analysis (Full)**
   - Comprehensive DFA is complex
   - Start with simple pattern matching
   - Add DFA when needed for advanced rules

### Phase 3: Configuration (Simplify Heavily)

**MVP Approach**: Hardcoded configuration
- Rules always enabled
- All errors reported
- No customization

**Defer to Post-MVP**:
- `.react-analyzer.json` config files
- Rule enable/disable
- Severity levels
- Include/exclude patterns
- Config hierarchy
- Presets

**Rationale**: Configuration is important but not critical to prove the concept works.

### Phase 4: IDE Integration (Defer Entirely)

1. **LSP Server**
   - Complex protocol implementation
   - Document synchronization
   - Debouncing
   - Incremental updates

2. **VS Code Extension**
   - Extension packaging
   - Binary bundling
   - Cross-platform builds

3. **Auto-Fix Generation**
   - Code action generation
   - Validation
   - Multi-edit coordination

**Rationale**: CLI proves the analysis works. IDE integration is just better UX.

### Phase 5: CI/CD (Defer Entirely)

1. **GitHub Actions integration**
2. **SARIF output**
3. **Pre-commit hooks**

**Rationale**: Can use basic CLI in CI. Fancy integrations are nice-to-have.

## Simplified MVP Architecture

```
┌─────────────────────────────────────────────┐
│              CLI (main.go)                  │
│  • Parse args                               │
│  • Read file                                │
│  • Print results                            │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           Analyzer                          │
│  • Coordinate analysis                      │
│  • Run rules sequentially                   │
└─────────────────┬───────────────────────────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
┌────▼─────┐  ┌──▼───┐  ┌────▼──────┐
│ Parser   │  │ Scope│  │ Rules     │
│          │  │ Build│  │ Registry  │
│ tree-    │  │      │  │           │
│ sitter   │  │      │  │ • no-     │
│          │  │      │  │   object- │
└──────────┘  └──────┘  │   deps    │
                        └───────────┘
                              │
                        ┌─────▼──────┐
                        │ Diagnostic │
                        └────────────┘
```

**That's it.** No component graph, no import resolver, no LSP server, no config system.

## MVP Timeline: 2-3 Weeks (Not 14!)

### Week 1: Core Infrastructure
- Parser abstraction over tree-sitter
- AST node wrapper
- Semantic helpers (`IsComponent`, `IsHookCall`, etc.)
- Basic scope builder
- Tests for parser and semantic layer

**Deliverable**: Can parse React files and answer semantic questions

### Week 2: Rule Engine + First Rule
- Rule interface
- Rule registry
- Simple executor (sequential)
- Implement `no-object-deps` rule
- Diagnostic system
- Tests for rule engine and rule

**Deliverable**: One working rule that finds real bugs

### Week 3: CLI + Polish
- Basic CLI argument parsing
- File reading
- Text output formatting
- End-to-end tests
- README with examples
- Fix bugs found during testing

**Deliverable**: Usable CLI tool

## Success Criteria for MVP

1. ✅ Can analyze a single React file
2. ✅ Detects inline objects in hook dependencies (no-object-deps rule)
3. ✅ Produces human-readable output
4. ✅ Exit code indicates pass/fail
5. ✅ Works on real-world React code (test on TodoMVC)
6. ✅ <10% false positive rate on test corpus
7. ✅ Performance: <10ms per file
8. ✅ Unit test coverage >80%

## What Gets Added Post-MVP (Prioritized)

### Priority 1: Usability (Weeks 4-5)
- Configuration file support
- Multiple file analysis
- Better error messages
- Second rule: `no-unstable-props`

### Priority 2: Performance (Week 6)
- Parallel rule execution
- Directory analysis with progress
- Incremental parsing (if needed)

### Priority 3: IDE Integration (Weeks 7-10)
- LSP server implementation
- VS Code extension
- Auto-fix suggestions

### Priority 4: Advanced Rules (Weeks 11-12)
- Import resolution
- Component graph
- Cross-file rule: `memo-unstable-props`

### Priority 5: CI/CD (Weeks 13-14)
- SARIF output
- GitHub Actions
- Pre-commit hooks

## Key Simplifications

| Component | Full Plan | MVP Simplification |
|-----------|-----------|-------------------|
| **Rules** | 3 rules, cross-file | 1 rule, single-file only |
| **Parser** | Incremental updates | Full parse only |
| **Execution** | Multi-phase, parallel | Sequential, single-phase |
| **Input** | Directory tree | Single file |
| **Output** | Text/JSON/SARIF | Text only |
| **Config** | File-based hierarchy | Hardcoded |
| **Error Handling** | Graceful, isolated | Basic, can crash |
| **Memory** | LRU cache, limits | No limits |
| **Analysis** | Cross-file DFA | Single-file pattern matching |
| **Scope Analysis** | Full stability tracking | Basic symbol resolution |

## Testing Strategy for MVP

### Unit Tests (~50 tests, not 200)
- Parser: 10 tests
- Semantic analysis: 10 tests
- Scope building: 10 tests
- Rule engine: 5 tests
- `no-object-deps` rule: 15 tests

### Integration Tests (~5 tests, not 50)
- End-to-end: Parse → Analyze → Output
- Real-world files from TodoMVC

### Performance Tests (~3 benchmarks)
- Small file (<100 lines)
- Medium file (233 lines - our POC file)
- Large file (1000 lines)

**Target**: >80% coverage (not 90%)

## Why This MVP is Better

1. **Faster to market**: 3 weeks vs 14 weeks
2. **Proves core value**: One rule that works is infinitely better than zero
3. **Validates approach**: Tests tree-sitter + React analysis concept
4. **Reduces risk**: Find problems early before building too much
5. **Easier to pivot**: Less code to throw away if we learn something
6. **Focus on quality**: One great rule > three mediocre rules
7. **Get feedback sooner**: Users can try it and give input

## Risk: What If We Cut Too Much?

**Mitigation**: The deferred features are not removed, just sequenced later.
- We're building a **foundation**, not the full system
- Each post-MVP phase adds one major feature
- Architecture supports future additions
- Can always accelerate if MVP proves successful

## Next Steps (If Approved)

1. Update implementation_plan.md to reflect MVP scope
2. Remove/defer complex sections:
   - Component Graph section
   - Full Data Flow Analysis
   - Import Resolution
   - Auto-Fix Generation
   - LSP Server Architecture
   - Configuration System (simplify to hardcoded)
3. Simplify Phase 1 to 3 weeks instead of 6
4. Defer Phase 2 and 3 entirely
5. Update success criteria to MVP-focused metrics

---

**Question for Review**: Does this MVP scope feel right? Too aggressive? Not aggressive enough?
