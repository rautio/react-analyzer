# React Analyzer

[![codecov](https://codecov.io/gh/rautio/react-analyzer/branch/main/graph/badge.svg?token=DVC95OTN7M)](https://codecov.io/gh/rautio/react-analyzer)

Static analysis tool for detecting React performance issues and anti-patterns that traditional linters miss.

## Why React Analyzer?

React Analyzer catches performance issues before they reach production:

- **Infinite re-render loops** from unstable hook dependencies
- **Unnecessary re-renders** from inline objects in props
- **Broken memoization** when React.memo components receive unstable props
- **Missing dependencies** in useEffect, useMemo, and useCallback

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

### no-object-deps (In Development)

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

### Planned Rules

- **`no-unstable-props`** - Detect inline objects/functions in JSX props
- **`memo-unstable-props`** - Validate React.memo effectiveness
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
