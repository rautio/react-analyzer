# Current State - What Works Today

**Last Updated:** 2025-11-18
**Version:** Phase 2.2 Complete! (Arrow Functions + Cross-File + Partial Spreads)

This document describes what the react-analyzer tool can do **right now**, helping you understand its current capabilities and limitations.

---

## TL;DR - Quick Overview

✅ **What Works:**
- 6 production-ready analysis rules
- **NEW:** Cross-file component analysis (prop drilling across files!)
- **NEW:** Arrow function components (`const Foo = () => <div />`)
- **NEW:** React.memo wrapped arrow functions
- **NEW:** Partial spread operator support (infrastructure in place)
- Function declaration components
- Named and default exports
- Explicit prop passing
- Fast performance (~2-3ms per file)

⚠️ **Partial Support:**
- Prop spread operators (works with destructured props, needs TypeScript analysis for full support)

❌ **What Doesn't Work Yet:**
- Full prop spread through non-destructured props
- Object property access tracking

**See:** [known_limitations.md](known_limitations.md) for full details.

---

## What's Working Today

### ✅ 1. Component Graph Infrastructure

The analyzer builds a complete graph of your React application:

**Components:**
- ✅ Detects function declaration components
- ✅ **NEW:** Detects arrow function components (`const Foo = () => ...`)
- ✅ **NEW:** Detects React.memo wrapped arrow functions
- ✅ Tracks component hierarchy (parent-child relationships)
- ✅ Extracts prop definitions from parameters (both function and arrow)
- ✅ Identifies React.memo usage

**State:**
- ✅ Detects useState hooks
- ✅ Detects useReducer hooks
- ✅ Detects Context usage
- ✅ Tracks which component owns which state

**Relationships:**
- ✅ "defines" edges (component → state it creates)
- ✅ "passes" edges (parent → child prop passing)
- ✅ Shortest path finding between components
- ✅ BFS traversal for dependency analysis

---

### ✅ 2. Six Production-Ready Rules

#### Rule 1: no-object-deps
Detects inline objects/arrays in hook dependency arrays.

**Example:**
```tsx
// ❌ Detected
useEffect(() => {
    fetchData();
}, [{ userId: 123 }]);  // New object every render!

// ✅ Suggested fix
const deps = useMemo(() => ({ userId: 123 }), []);
useEffect(() => {
    fetchData();
}, [deps]);
```

**Status:** ✅ Fully working

---

#### Rule 2: unstable-props-to-memo
Detects unstable props passed to React.memo, useMemo, or useCallback.

**Example:**
```tsx
// ❌ Detected
const MemoizedComponent = React.memo(ExpensiveComponent);

function Parent() {
    return <MemoizedComponent
        config={{ theme: 'dark' }}  // New object every render!
    />;
}

// ✅ Suggested fix
const config = useMemo(() => ({ theme: 'dark' }), []);
return <MemoizedComponent config={config} />;
```

**Status:** ✅ Fully working (most complex rule: 488 LOC)
**Cross-file:** ✅ Works across files (tracks memoization in symbol table)

---

#### Rule 3: no-derived-state
Detects useState mirroring props via useEffect (anti-pattern).

**Example:**
```tsx
// ❌ Detected
function UserProfile({ user }: Props) {
    const [localUser, setLocalUser] = useState(user);

    useEffect(() => {
        setLocalUser(user);  // Syncing prop to state
    }, [user]);
}

// ✅ Suggested fix
function UserProfile({ user }: Props) {
    // Just use the prop directly!
    return <div>{user.name}</div>;
}
```

**Status:** ✅ Fully working

---

#### Rule 4: no-stale-state
Detects state updates without functional form (causes race conditions).

**Example:**
```tsx
// ❌ Detected
const [count, setCount] = useState(0);

const increment = () => {
    setCount(count + 1);  // Uses stale closure value
};

// ✅ Suggested fix
const increment = () => {
    setCount(prev => prev + 1);  // Always has latest value
};
```

**Status:** ✅ Fully working

---

#### Rule 5: no-inline-props
Detects inline objects/arrays/functions in JSX props.

**Example:**
```tsx
// ❌ Detected
<ExpensiveComponent
    config={{ theme: 'dark' }}           // New object
    items={[1, 2, 3]}                    // New array
    onClick={() => console.log('hi')}    // New function
/>

// ✅ Suggested fix
const config = useMemo(() => ({ theme: 'dark' }), []);
const items = useMemo(() => [1, 2, 3], []);
const onClick = useCallback(() => console.log('hi'), []);

<ExpensiveComponent config={config} items={items} onClick={onClick} />
```

**Status:** ✅ Fully working

---

#### Rule 6: deep-prop-drilling (NEW!)
Detects props drilled through 3+ component levels.

**Example:**
```tsx
// ❌ Detected
function App() {
    const [theme, setTheme] = useState('dark');
    return <Parent theme={theme} />;
}

function Parent({ theme }) {
    return <Child theme={theme} />;  // Passthrough
}

function Child({ theme }) {
    return <Display theme={theme} />;  // Passthrough
}

function Display({ theme }) {
    return <div className={theme} />;  // Finally uses it
}

// ✅ Suggested fix
const ThemeContext = createContext('light');

function App() {
    const [theme, setTheme] = useState('dark');
    return (
        <ThemeContext.Provider value={theme}>
            <Parent />
        </ThemeContext.Provider>
    );
}

function Display() {
    const theme = useContext(ThemeContext);
    return <div className={theme} />;
}
```

**Status:** ✅ Working within single files
**Limitation:** ❌ Doesn't work across files yet

---

### ✅ 3. Developer Experience Features

**CLI:**
- ✅ Colored output with issue grouping by file
- ✅ Verbose mode with performance metrics (`-V` flag)
- ✅ Quiet mode for CI/CD (`-q` flag)
- ✅ Directory scanning with smart exclusions
- ✅ Supports `.tsx`, `.jsx`, `.ts`, `.js` files
- ✅ Exit code 0 (no issues) or 1 (issues found)

**Performance:**
- ✅ Parallel file analysis
- ✅ Fast: ~2ms per file average
- ✅ Scales to large codebases

**Configuration:**
- ✅ Monorepo support
- ✅ Import alias resolution
- ✅ TypeScript and JavaScript support

---

## What's NOT Working Yet

See [known_limitations.md](known_limitations.md) for full details. Here are the critical gaps:

### ✅ ~~1. Arrow Function Components~~ FIXED!

**Status:** ✅ Completed in Phase 2.2A (2025-11-18)

All arrow function patterns now work:
```tsx
// ✅ ALL NOW DETECTED
const MyComponent = ({ theme }) => {
    return <div className={theme}>Content</div>;
};

const MemoComponent = memo(() => <div />);

function MyComponent({ theme }) {
    return <div className={theme}>Content</div>;
}
```

---

### ✅ ~~1. Cross-File Prop Drilling~~ FIXED!

**Status:** ✅ Completed in Phase 2.2B (2025-11-18)

Cross-file prop drilling now works perfectly:
```tsx
// App.tsx
function App() {
    const [theme] = useState('dark');
    return <Dashboard theme={theme} />;  // ✅ Detected
}

// Dashboard.tsx
export function Dashboard({ theme }) {
    return <Sidebar theme={theme} />;  // ✅ NOW DETECTED (cross-file!)
}

// Sidebar.tsx
export function Sidebar({ theme }) {
    return <div className={theme} />;
}
```

**Result:** ✅ Detects complete chain: App → Dashboard → Sidebar

---

### ❌ 1. Prop Spread Operators (HIGH)

**Impact:** Very common pattern, major blind spot

```tsx
function Container(props) {
    return <Panel {...props} />;  // ❌ NOT tracked
}
```

**Workaround:** Pass props explicitly instead of spreading.

**Fix:** Planned for Phase 2.2 (Q1 2026)

---

## How to Use Effectively Today

### 1. Best for Single-File Components

**Works Great:**
```tsx
// MyComponent.tsx - All components in one file
function App() {
    const [theme, setTheme] = useState('dark');
    return <Dashboard theme={theme} />;
}

function Dashboard({ theme }) {
    return <Sidebar theme={theme} />;
}

function Sidebar({ theme }) {
    return <ThemeToggle theme={theme} />;
}

function ThemeToggle({ theme }) {
    return <button className={theme}>Toggle</button>;
}
```

**Analysis:**
```bash
./react-analyzer MyComponent.tsx
```

**Output:**
```
✖ Found 1 issue in 1 file
  MyComponent.tsx
    8:31 - [deep-prop-drilling] Prop 'theme' is drilled through 3 component levels...
```

---

### 2. Use Function Declarations

**Works Great:**
```tsx
// ✅ These all work
function MyComponent(props) { }
function MyComponent({ prop1, prop2 }) { }
export default function MyComponent() { }
```

**Doesn't Work:**
```tsx
// ❌ These don't work yet
const MyComponent = (props) => { };
const MyComponent = React.memo(() => { });
export const MyComponent: React.FC = () => { };
```

---

### 3. Pass Props Explicitly

**Works Great:**
```tsx
<Child theme={theme} config={config} />
```

**Doesn't Work:**
```tsx
<Child {...props} />
<Child {...rest} theme={theme} />
```

---

### 4. Use Direct Variable References

**Works Great:**
```tsx
const [theme, setTheme] = useState('dark');
return <Child theme={theme} />;
```

**Doesn't Work:**
```tsx
const [settings, setSettings] = useState({ theme: 'dark' });
return <Child theme={settings.theme} />;  // Property access not tracked
```

---

## Testing Your Codebase

### Step 1: Run on a Single File
```bash
./react-analyzer src/components/App.tsx
```

### Step 2: Run on a Directory
```bash
./react-analyzer src/
```

**Output includes:**
- Number of files analyzed
- Issues grouped by file
- Performance metrics

### Step 3: Interpret Results

**Exit Codes:**
- `0` - No issues found
- `1` - Issues found
- `2` - Error (invalid path, no files found, etc.)

**Verbose Mode:**
```bash
./react-analyzer src/ -V
```

Shows:
- Rules enabled
- Files analyzed
- Parse/analyze timing
- Per-rule statistics

---

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: React Analyzer

on: [pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build React Analyzer
        run: go build cmd/react-analyzer/main.go
      - name: Run Analysis
        run: ./main src/
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Running React Analyzer..."
./react-analyzer src/

if [ $? -ne 0 ]; then
    echo "❌ React Analyzer found issues. Fix them before committing."
    exit 1
fi

echo "✅ React Analyzer passed!"
```

---

## Performance Characteristics

**Benchmarks** (on typical React codebase):
- Single file: ~2ms
- 100 files: ~200ms
- 1000 files: ~2s

**Scales well** because:
- Parallel file processing
- Efficient AST parsing
- Minimal memory overhead

**Not optimized for:**
- Monorepos with 10,000+ React files
- Repeated analyses (no caching yet)

---

## Success Stories

The tool works great for:

1. **Small-Medium Single-File Components**
   - ✅ Detects prop drilling
   - ✅ Finds memo issues
   - ✅ Catches stale state bugs

2. **Code Review**
   - ✅ Run on changed files in PR
   - ✅ Catch issues before merge

3. **Learning Tool**
   - ✅ Understand React antipatterns
   - ✅ See recommendations

4. **Incremental Adoption**
   - ✅ Fix one rule at a time
   - ✅ Gradual improvement

---

## What's Coming Next

See [ROADMAP.md](ROADMAP.md) for full details.

**Phase 2.2 (Q1 2026):**
- ✅ Arrow function components
- ✅ Cross-file prop drilling
- ✅ Prop spread operators

**Phase 2.3 (Q1-Q2 2026):**
- ✅ Partial prop usage detection
- ✅ Object property access
- ✅ Better recommendations

---

## Getting Help

**Found a bug?**
- Check [known_limitations.md](known_limitations.md) first
- Search GitHub issues
- Open a new issue with reproduction

**Want a feature?**
- Check [ROADMAP.md](ROADMAP.md) to see if it's planned
- Open a discussion to suggest new features

**Want to contribute?**
- See [CONTRIBUTING.md](CONTRIBUTING.md)
- Pick up a `good-first-issue`

---

**Last Updated:** 2025-11-18
**Next Review:** 2026-01
