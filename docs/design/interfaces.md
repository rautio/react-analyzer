# React Analyzer - CLI Interface Design

## Overview

The `react-analyzer` CLI is a command-line tool for static analysis of React code. The MVP focuses on simplicity: analyze a single file and report performance issues.

## Design Principles

1. **Zero Configuration** - Works out of the box
2. **Familiar Patterns** - Follows conventions from ESLint, Prettier, etc.
3. **Clear Output** - Easy to understand what's wrong and where
4. **Scriptable** - Exit codes and output format suitable for CI/CD
5. **Fast** - Instant feedback (<100ms for typical files)

---

## Command Structure

### Basic Usage

```bash
react-analyzer <file>
```

### Examples

```bash
# Analyze a single file
react-analyzer src/App.tsx

# Show version
react-analyzer --version

# Show help
react-analyzer --help

# Verbose output
react-analyzer --verbose src/components/Dashboard.tsx

# Quiet mode (errors only, no warnings)
react-analyzer --quiet src/App.tsx
```

---

## Command Reference

### Main Command

```
react-analyzer [options] <file>

Analyze a React/TypeScript file for performance issues.

ARGUMENTS:
  <file>                Path to the file to analyze (.tsx, .jsx, .ts, .js)

OPTIONS:
  -h, --help            Show this help message
  -v, --version         Show version number
  -V, --verbose         Show detailed analysis information
  -q, --quiet           Only show errors (suppress warnings and info)
      --no-color        Disable colored output
      --format <type>   Output format (default: text)
                        Available: text (MVP only; json/sarif post-MVP)

EXIT CODES:
  0                     No issues found
  1                     Issues found
  2                     Analysis error (invalid file, syntax error, etc.)

EXAMPLES:
  react-analyzer src/App.tsx
  react-analyzer --verbose src/components/UserList.tsx
  react-analyzer --quiet src/Dashboard.tsx
```

---

## Output Format

### Default Output (Text)

#### Success (No Issues)

```bash
$ react-analyzer src/App.tsx
✓ No issues found in src/App.tsx
```

**Exit Code**: 0

#### Issues Found

```bash
$ react-analyzer src/Dashboard.tsx

src/Dashboard.tsx
  12:5  error  Inline object in hook dependency array will cause infinite re-renders  no-object-deps
  24:8  error  Inline object in hook dependency array will cause infinite re-renders  no-object-deps

✖ 2 errors

```

**Exit Code**: 1

**Format Explanation**:
- `src/Dashboard.tsx` - File being analyzed
- `12:5` - Line 12, Column 5
- `error` - Severity level
- Message - Description of the issue
- `no-object-deps` - Rule ID

#### Verbose Output

```bash
$ react-analyzer --verbose src/Dashboard.tsx

React Analyzer v0.1.0
Analyzing: src/Dashboard.tsx
File size: 1,234 bytes
Language: TypeScript + JSX

Rules enabled:
  ✓ no-object-deps      Prevent inline objects in hook dependencies

Parsing... done (1.2ms)
Analysis... done (0.8ms)

src/Dashboard.tsx
  12:5  error  Inline object in hook dependency array will cause infinite re-renders  no-object-deps

    10 | function Dashboard() {
    11 |   const [data, setData] = useState([]);
  > 12 |   useEffect(() => {
       |     ^
    13 |     fetchData();
    14 |   }, [{ page: 1 }]);  // ← Inline object causes infinite re-renders
    15 |

    Suggestion: Extract to a constant or use useMemo
      const options = useMemo(() => ({ page: 1 }), []);
      useEffect(() => { ... }, [options]);

✖ 1 error

Analysis time: 2.0ms
```

#### Quiet Mode (Errors Only)

```bash
$ react-analyzer --quiet src/App.tsx

src/App.tsx
  12:5  error  Inline object in hook dependency array will cause infinite re-renders  no-object-deps

✖ 1 error
```

**Suppresses**: Warnings, info messages, success messages

#### Syntax Error

```bash
$ react-analyzer src/Broken.tsx

✖ Parse error in src/Broken.tsx:5:12
  Unexpected token '}'

Cannot analyze file with syntax errors.
```

**Exit Code**: 2

#### File Not Found

```bash
$ react-analyzer src/Missing.tsx

✖ Error: File not found: src/Missing.tsx
```

**Exit Code**: 2

#### Unsupported File Type

```bash
$ react-analyzer src/data.json

✖ Error: Unsupported file type: .json
  Supported: .tsx, .jsx, .ts, .js
```

**Exit Code**: 2

---

## Color Output

### Default (Colored)

Uses ANSI color codes for better readability:

- ✓ **Green**: Success messages
- ✖ **Red**: Errors
- ⚠ **Yellow**: Warnings (post-MVP)
- **Cyan**: File paths
- **Gray**: Line numbers, rule IDs
- **White/Default**: Messages

### No Color Mode

```bash
react-analyzer --no-color src/App.tsx
```

Strips all ANSI color codes. Useful for:
- CI/CD logs that don't support colors
- Piping output to files
- Accessibility

---

## Integration Examples

### CI/CD (GitHub Actions)

```yaml
name: React Analysis

on: [push, pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install react-analyzer
        run: |
          curl -L https://github.com/rautio/react-analyzer/releases/latest/download/react-analyzer-linux-amd64 -o react-analyzer
          chmod +x react-analyzer

      - name: Analyze components
        run: |
          ./react-analyzer src/App.tsx
          ./react-analyzer src/components/Dashboard.tsx
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Get staged .tsx and .jsx files
FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(tsx|jsx)$')

if [ -z "$FILES" ]; then
  exit 0
fi

# Analyze each file
for file in $FILES; do
  react-analyzer "$file"
  if [ $? -ne 0 ]; then
    echo "❌ react-analyzer found issues. Commit aborted."
    exit 1
  fi
done

echo "✅ All files passed react-analyzer"
exit 0
```

### NPM Script

```json
{
  "scripts": {
    "analyze": "react-analyzer src/App.tsx",
    "analyze:all": "find src -name '*.tsx' -exec react-analyzer {} +"
  }
}
```

### Shell Script (Batch Analysis - Manual for MVP)

```bash
#!/bin/bash
# analyze-all.sh

EXIT_CODE=0

for file in src/**/*.tsx; do
  echo "Analyzing $file..."
  react-analyzer "$file"

  if [ $? -ne 0 ]; then
    EXIT_CODE=1
  fi
done

exit $EXIT_CODE
```

---

## Help Text

### `react-analyzer --help`

```
react-analyzer v0.1.0

Static analysis tool for React performance issues.

USAGE:
    react-analyzer [OPTIONS] <file>

ARGUMENTS:
    <file>    Path to React/TypeScript file to analyze

OPTIONS:
    -h, --help         Show this help message
    -v, --version      Show version number
    -V, --verbose      Show detailed output
    -q, --quiet        Only show errors
        --no-color     Disable colored output

EXAMPLES:
    react-analyzer src/App.tsx
    react-analyzer --verbose src/components/Dashboard.tsx

RULES:
    no-object-deps     Prevent inline objects in hook dependencies
                       Inline objects/arrays in useEffect, useMemo, or
                       useCallback dependencies cause infinite re-renders.

EXIT CODES:
    0    No issues found
    1    Issues found
    2    Analysis error

For more information, visit: https://github.com/rautio/react-analyzer
```

### `react-analyzer --version`

```
react-analyzer 0.1.0
```

---

## Implementation API (Go)

### Main Function Structure

```go
package main

import (
    "fmt"
    "os"

    "github.com/rautio/react-analyzer/internal/analyzer"
    "github.com/rautio/react-analyzer/internal/cli"
)

func main() {
    // Parse CLI arguments
    args := cli.ParseArgs(os.Args[1:])

    // Handle version/help
    if args.ShowVersion {
        fmt.Println("react-analyzer", Version)
        os.Exit(0)
    }

    if args.ShowHelp {
        cli.PrintHelp()
        os.Exit(0)
    }

    // Validate input file
    if err := cli.ValidateFile(args.FilePath); err != nil {
        cli.PrintError(err, args.NoColor)
        os.Exit(2)
    }

    // Run analysis
    result, err := analyzer.AnalyzeFile(args.FilePath)
    if err != nil {
        cli.PrintError(err, args.NoColor)
        os.Exit(2)
    }

    // Print results
    cli.PrintResults(result, args)

    // Exit with appropriate code
    if result.HasErrors() {
        os.Exit(1)
    }

    os.Exit(0)
}
```

### CLI Arguments Structure

```go
// internal/cli/args.go

type Args struct {
    FilePath    string

    // Flags
    ShowHelp    bool
    ShowVersion bool
    Verbose     bool
    Quiet       bool
    NoColor     bool
    Format      string  // "text" (MVP only)
}

func ParseArgs(args []string) *Args {
    // Parse using flag package or cobra
    // ...
}
```

### Result Structure

```go
// internal/analyzer/result.go

type AnalysisResult struct {
    FilePath    string
    Diagnostics []Diagnostic
    ParseTime   time.Duration
    AnalysisTime time.Duration
}

type Diagnostic struct {
    Location    Location
    Severity    Severity
    Message     string
    RuleID      string
    Suggestion  string  // Optional fix suggestion
}

type Location struct {
    FilePath    string
    StartLine   int
    StartColumn int
    EndLine     int
    EndColumn   int
}

type Severity int

const (
    SeverityError Severity = iota
    SeverityWarning  // Post-MVP
    SeverityInfo     // Post-MVP
)

func (r *AnalysisResult) HasErrors() bool {
    for _, d := range r.Diagnostics {
        if d.Severity == SeverityError {
            return true
        }
    }
    return false
}
```

### Formatter Interface

```go
// internal/cli/formatter.go

type Formatter interface {
    Format(result *AnalysisResult, opts *FormatOptions) string
}

type FormatOptions struct {
    Verbose bool
    Quiet   bool
    NoColor bool
}

type TextFormatter struct{}

func (f *TextFormatter) Format(result *AnalysisResult, opts *FormatOptions) string {
    // Build output string
    // ...
}
```

---

## Error Messages

### User-Friendly Error Messages

```go
// internal/cli/errors.go

var ErrorMessages = map[string]string{
    "file_not_found": "File not found: %s\nCheck that the file path is correct.",

    "unsupported_extension": "Unsupported file type: %s\nSupported extensions: .tsx, .jsx, .ts, .js",

    "parse_error": "Parse error in %s:%d:%d\n%s\n\nCannot analyze file with syntax errors.\nFix the syntax error and try again.",

    "read_error": "Cannot read file: %s\n%s",

    "internal_error": "Internal error: %s\nThis is a bug. Please report it at:\nhttps://github.com/rautio/react-analyzer/issues",
}

func FormatError(errorType string, args ...interface{}) string {
    template := ErrorMessages[errorType]
    return fmt.Sprintf(template, args...)
}
```

---

## Future Enhancements (Post-MVP)

### Multiple Files

```bash
# Analyze multiple files
react-analyzer src/App.tsx src/Dashboard.tsx

# Analyze directory (recursive)
react-analyzer src/

# With glob pattern
react-analyzer 'src/**/*.tsx'
```

### Configuration File

```bash
# Use config file
react-analyzer --config .react-analyzer.json src/

# Generate default config
react-analyzer --init
```

### Output Formats

```bash
# JSON output
react-analyzer --format json src/App.tsx > results.json

# SARIF (for GitHub Code Scanning)
react-analyzer --format sarif src/ > results.sarif
```

### Rule Selection

```bash
# Run specific rules only
react-analyzer --rule no-object-deps src/App.tsx

# Disable specific rules
react-analyzer --disable no-unstable-props src/App.tsx
```

### Fix Mode (Auto-fix)

```bash
# Show suggested fixes
react-analyzer --fix-dry-run src/App.tsx

# Apply fixes automatically
react-analyzer --fix src/App.tsx
```

### Watch Mode

```bash
# Watch file for changes
react-analyzer --watch src/App.tsx
```

---

## Testing the CLI

### Manual Testing Checklist

```bash
# Valid file, no issues
react-analyzer test/fixtures/clean.tsx
# Expected: ✓ No issues found, exit 0

# Valid file, with issues
react-analyzer test/fixtures/with-issues.tsx
# Expected: Show diagnostics, exit 1

# File not found
react-analyzer missing.tsx
# Expected: Error message, exit 2

# Invalid file type
react-analyzer package.json
# Expected: Unsupported file type error, exit 2

# Syntax error
react-analyzer test/fixtures/syntax-error.tsx
# Expected: Parse error, exit 2

# Help
react-analyzer --help
# Expected: Help text, exit 0

# Version
react-analyzer --version
# Expected: Version number, exit 0

# Verbose
react-analyzer --verbose test/fixtures/with-issues.tsx
# Expected: Detailed output with timing

# Quiet
react-analyzer --quiet test/fixtures/with-issues.tsx
# Expected: Only errors, no extra output

# No color
react-analyzer --no-color test/fixtures/with-issues.tsx
# Expected: No ANSI codes in output
```

### Automated Tests

```go
// internal/cli/cli_test.go

func TestCLI_NoIssues(t *testing.T) {
    output, exitCode := runCLI("test/fixtures/clean.tsx")

    assert.Equal(t, 0, exitCode)
    assert.Contains(t, output, "No issues found")
}

func TestCLI_WithIssues(t *testing.T) {
    output, exitCode := runCLI("test/fixtures/with-issues.tsx")

    assert.Equal(t, 1, exitCode)
    assert.Contains(t, output, "no-object-deps")
}

func TestCLI_FileNotFound(t *testing.T) {
    output, exitCode := runCLI("missing.tsx")

    assert.Equal(t, 2, exitCode)
    assert.Contains(t, output, "File not found")
}

func TestCLI_VerboseMode(t *testing.T) {
    output, exitCode := runCLI("--verbose", "test/fixtures/clean.tsx")

    assert.Contains(t, output, "Analysis time:")
    assert.Contains(t, output, "Rules enabled:")
}
```

---

## Success Criteria

The CLI is considered complete when:

1. ✅ Can analyze a single .tsx/.jsx file
2. ✅ Shows clear, actionable error messages
3. ✅ Exit codes are correct (0, 1, 2)
4. ✅ Help text is comprehensive
5. ✅ Works on Windows, macOS, Linux
6. ✅ Handles edge cases gracefully (missing file, syntax error, etc.)
7. ✅ Performance: <100ms for typical files
8. ✅ All manual test cases pass
9. ✅ Automated tests cover main flows

---

## Distribution

### Binary Releases

```
react-analyzer-v0.1.0-darwin-amd64       # macOS Intel
react-analyzer-v0.1.0-darwin-arm64       # macOS Apple Silicon
react-analyzer-v0.1.0-linux-amd64        # Linux
react-analyzer-v0.1.0-windows-amd64.exe  # Windows
```

### Installation

```bash
# macOS (Homebrew - future)
brew install react-analyzer

# Linux (curl)
curl -L https://github.com/rautio/react-analyzer/releases/latest/download/react-analyzer-linux-amd64 -o /usr/local/bin/react-analyzer
chmod +x /usr/local/bin/react-analyzer

# Windows (Chocolatey - future)
choco install react-analyzer

# Go get (for development)
go install github.com/rautio/react-analyzer@latest
```

---

## Summary

The MVP CLI is designed to be:
- **Simple**: One command, one file
- **Clear**: Obvious what's wrong and where
- **Fast**: Instant feedback
- **Familiar**: Follows ESLint/Prettier patterns
- **Extensible**: Easy to add features post-MVP

This interface provides a solid foundation that can grow with the tool while keeping the initial implementation focused and achievable.
