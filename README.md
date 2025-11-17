# React Analyzer

[![codecov](https://codecov.io/gh/rautio/react-analyzer/branch/main/graph/badge.svg?token=DVC95OTN7M)](https://codecov.io/gh/rautio/react-analyzer)

Static analysis tool for detecting React performance issues and anti-patterns that traditional linters miss.

## Why React Analyzer?

React Analyzer catches performance issues before they reach production:

- **Infinite re-render loops** from unstable hook dependencies
- **Unnecessary re-renders** from inline objects in props and derived state anti-patterns
- **Broken memoization** when React.memo components receive unstable props
- **Stale closures and race conditions** from non-functional state updates
- **Derived state bugs** from useState mirroring props via useEffect

Built with Go and tree-sitter for blazing-fast analysis.

## Installation

Currently requires building from source:

```bash
git clone https://github.com/rautio/react-analyzer
cd react-analyzer
go build -o react-analyzer ./cmd/react-analyzer
```

Pre-built binaries coming soon.

## Quick Start

Analyze a React file:

```bash
react-analyzer src/App.tsx
```

**Output:**
```
✓ No issues found in src/App.tsx
Analyzed 1 file in 12ms
```

If issues are found:
```
src/Dashboard.tsx
  12:5 - [no-object-deps] Inline object in hook dependency array will cause infinite re-renders

✖ Found 1 issue in 1 file
Analyzed 1 file in 15ms
```

## Usage

### Command

```bash
react-analyzer [options] <file>
```

### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Show version number |
| `--verbose` | `-V` | Show detailed analysis output and performance metrics |
| `--quiet` | `-q` | Only show errors (suppress success messages and timing) |
| `--no-color` | | Disable colored output (useful for CI) |

### Examples

**Analyze a directory:**
```bash
react-analyzer src/components/
```
Output:
```
Analyzing 7 files...

src/components/Dashboard.tsx
  12:5 - [no-object-deps] Dependency 'config' is an object/array created in render

✖ Found 1 issue in 1 file (6 files clean)
Analyzed 7 files in 45ms
```

**Verbose mode (detailed metrics):**
```bash
react-analyzer --verbose src/components/
```
Output includes performance breakdown:
```
Analyzing 7 files...

Rules enabled: 3
  - no-object-deps
  - memoized-component-unstable-props
  - placeholder

... (issues) ...

✖ Found 1 issue in 1 file (6 files clean)
Analyzed 7 files in 45ms

Performance Summary:
  Time elapsed: 45ms (parse: 15ms, analyze: 28ms)
  Throughput: 156 files/sec

Rules executed:
  no-object-deps: 1 issue
  memoized-component-unstable-props: 0 issues
  placeholder: 0 issues
```

**Quiet mode (errors only):**
```bash
react-analyzer --quiet src/App.tsx
```
Only shows issues if found, no success message or timing.

**CI/CD integration:**
```bash
react-analyzer --no-color src/
if [ $? -ne 0 ]; then
  echo "React analysis failed"
  exit 1
fi
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No issues found |
| `1` | Issues found |
| `2` | Analysis error (file not found, parse error, etc.) |

## Supported Files

- `.tsx` - TypeScript with JSX
- `.jsx` - JavaScript with JSX
- `.ts` - TypeScript
- `.js` - JavaScript

## Rules

### unstable-props-to-memo

Detects when unstable props break memoization in React.memo components. **Requires cross-file analysis** to detect violations that ESLint cannot catch.

**✅ Currently Detects:**
- Unstable props passed to React.memo components (cross-file)

**⏳ Future Detection (TODO):**
- useMemo with unstable prop dependencies (partially implemented)
- useCallback with unstable prop dependencies (partially implemented)

**Bad: Unstable props to React.memo component**
```tsx
// MemoChild.tsx
export const MemoChild = memo(({ config }) => <div>...</div>);

// Parent.tsx
function Parent() {
  return <MemoChild config={{ theme: 'dark' }} />; // ❌ Detected! Breaks memoization
}
```

**Good: Stable props to React.memo component**
```tsx
// Parent.tsx
const CONFIG = { theme: 'dark' };  // Outside component

function Parent() {
  return <MemoChild config={CONFIG} />; // ✅ Memoization works!
}
```

### no-object-deps

Prevents inline objects/arrays in hook dependency arrays that cause infinite re-render loops.

**Bad:**
```tsx
function Component() {
  const config = { theme: 'dark' };
  useEffect(() => {
    applyConfig(config);
  }, [config]); // ❌ Runs every render!
}
```

**Good:**
```tsx
const CONFIG = { theme: 'dark' };
function Component() {
  useEffect(() => {
    applyConfig(CONFIG);
  }, []); // ✅ Runs once
}
```

### no-derived-state

Detects when useState is used to mirror props and sync via useEffect, causing unnecessary re-renders and complexity. This is a common anti-pattern that leads to stale state issues and performance problems.

**Bad: State mirroring props**
```tsx
function UserProfile({ user }) {
  const [name, setName] = useState(user);

  useEffect(() => {
    setName(user);  // ❌ Causes extra re-render every time user changes
  }, [user]);

  return <div>{name}</div>;
}
```

**Good: Direct derivation**
```tsx
function UserProfile({ user }) {
  const name = user;  // ✅ Simple, performant, no bugs!
  return <div>{name}</div>;
}
```

**Performance Impact:**
- Each prop change triggers **2 renders** instead of 1 (initial + effect)
- Adds unnecessary complexity and potential for bugs
- Can cause stale state issues with multiple prop dependencies

**When to use useState with props:**
- Initial value only (user controls state after): `useState(initialColor)`
- Form controls with reset functionality
- Interactive state that diverges from props

### no-stale-state

Prevents stale state bugs by requiring state updates that depend on previous state to use the functional form. This catches race conditions and stale closures in all contexts including async callbacks, timers, and effects.

**Bad: Direct state reference**
```tsx
function Counter() {
  const [count, setCount] = useState(0);

  const increment = () => setCount(count + 1);  // ❌ Stale closure!

  // Especially dangerous in async contexts
  setTimeout(() => setCount(count + 1), 1000);  // ❌ Will get stuck!

  return <button onClick={increment}>{count}</button>;
}
```

**Good: Functional form**
```tsx
function Counter() {
  const [count, setCount] = useState(0);

  const increment = () => setCount(prev => prev + 1);  // ✅ Always correct!

  // Safe in async contexts
  setTimeout(() => setCount(prev => prev + 1), 1000);  // ✅ Works correctly!

  return <button onClick={increment}>{count}</button>;
}
```

**Common Bug Scenarios:**
- **Timers get stuck:** `setInterval(() => setState(state + 1))` - timer stops at 1
- **Lost updates:** Multiple rapid clicks only increment by 1 instead of N
- **Race conditions:** Async operations overwrite each other's updates
- **Array accumulation fails:** `setItems([...items, x])` loses items with concurrent updates

**All contexts detected:**
- Regular functions ✓
- setTimeout/setInterval ✓
- Promise callbacks ✓
- async/await ✓
- useEffect ✓
- Event handlers ✓

### Planned Rules

- **`unstable-props-in-effects`** - Detect unstable props in useEffect/useLayoutEffect (lower severity)
- **`no-unstable-props`** - Detect inline objects/functions in JSX props
- **`exhaustive-deps`** - Comprehensive dependency checking
- **`require-memo-expensive-component`** - Suggest memoization

## Troubleshooting

### File not found
```
✖ Error: file not found: src/App.tsx
```
Check the file path is correct.

### Unsupported file type
```
✖ Error: unsupported file type: .json
Supported extensions: .tsx, .jsx, .ts, .js
```
React Analyzer only analyzes React/JavaScript/TypeScript files.

### Parse error
```
✖ Parse error in src/Broken.tsx:5:12
Cannot analyze file with syntax errors.
```
Fix syntax errors in your code first.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

TBD

---

**Questions?** Run `react-analyzer --help` or open an issue on GitHub.
