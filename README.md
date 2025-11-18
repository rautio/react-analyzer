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

## Usage

### Command

```bash
react-analyzer [options] <file>
```

### Options

| Option       | Short | Description                                             |
| ------------ | ----- | ------------------------------------------------------- |
| `--help`     | `-h`  | Show help message                                       |
| `--version`  | `-v`  | Show version number                                     |
| `--verbose`  | `-V`  | Show detailed analysis output and performance metrics   |
| `--quiet`    | `-q`  | Only show errors (suppress success messages and timing) |
| `--no-color` |       | Disable colored output (useful for CI)                  |

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

**CI/CD integration:**

```bash
react-analyzer --no-color src/
if [ $? -ne 0 ]; then
  echo "React analysis failed"
  exit 1
fi
```

## Exit Codes

| Code | Meaning                                            |
| ---- | -------------------------------------------------- |
| `0`  | No issues found                                    |
| `1`  | Issues found                                       |
| `2`  | Analysis error (file not found, parse error, etc.) |

## Supported Files

- `.tsx` - TypeScript with JSX
- `.jsx` - JavaScript with JSX
- `.ts` - TypeScript
- `.js` - JavaScript

## Rules

React Analyzer includes several rules to catch common React performance issues and anti-patterns. Each rule has detailed documentation with examples and explanations.

| Rule                     | Description                                                                                              | Documentation                          |
| ------------------------ | -------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| `unstable-props-to-memo` | Detects unstable props breaking memoization (React.memo, useMemo, useCallback). **Cross-file analysis**. | [docs](docs/rules/unstable-props-to-memo.md) |
| `no-object-deps`         | Prevents unstable object/array dependencies causing infinite re-render loops                             | [docs](docs/rules/no-object-deps.md)   |
| `no-derived-state`       | Detects useState mirroring props via useEffect (unnecessary re-renders)                                  | [docs](docs/rules/no-derived-state.md) |
| `no-stale-state`         | Prevents stale closures by requiring functional state updates                                            | [docs](docs/rules/no-stale-state.md)   |
| `no-inline-props`        | Detects inline objects/arrays/functions in JSX props breaking memoization                                | [docs](docs/rules/no-inline-props.md)  |

### Planned Rules

- **`unstable-props-in-effects`** - Detect unstable props in useEffect/useLayoutEffect (lower severity)
- **`exhaustive-deps`** - Comprehensive dependency checking
- **`require-memo-expensive-component`** - Suggest memoization for expensive components

## Path Aliases

React Analyzer automatically detects and supports path aliases for cross-file analysis:

### Auto-detection from `tsconfig.json`

If your project has a `tsconfig.json` with path mappings, React Analyzer will automatically use them:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  }
}
```

### Custom Configuration with `.reactanalyzer.json`

For non-TypeScript projects or to override `tsconfig.json`, create a `.reactanalyzer.json` file:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/": ["src/"],
      "@components/": ["src/components/"],
      "~/": ["./"]
    }
  }
}
```

**Priority:** `.reactanalyzer.json` > `tsconfig.json`

**Supported Import Formats:**
```tsx
import Button from '@/components/Button';       // ✅ Aliased import
import { utils } from '@utils/helpers';         // ✅ Aliased with nested path
import Component from './Component';            // ✅ Relative import (always supported)
import React from 'react';                      // ✅ External package (skipped)
```

See `.reactanalyzer.example.json` for a complete configuration example.

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

---

**Questions?** Run `react-analyzer --help` or open an issue on GitHub.
