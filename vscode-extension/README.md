# React Analyzer for VS Code

Static analysis extension for detecting React performance issues and antipatterns.

## Features

- **Real-time diagnostics**: See performance issues directly in your editor with squiggly underlines
- **Analyze on save**: Automatically analyze files when you save them
- **Workspace analysis**: Analyze your entire project with a single command
- **6 production-ready rules**:
  - `no-object-deps`: Detect inline objects/arrays in hook dependencies
  - `unstable-props-to-memo`: Detect unstable props passed to memoized components
  - `no-derived-state`: Detect useState mirroring props (anti-pattern)
  - `no-stale-state`: Detect state updates without functional form
  - `no-inline-props`: Detect inline objects/arrays/functions in JSX props
  - `deep-prop-drilling`: Detect props drilled through 3+ component levels

## Commands

- `React Analyzer: Analyze Current File` - Analyze the currently open file
- `React Analyzer: Analyze Workspace` - Analyze all React files in the workspace
- `React Analyzer: Clear All Diagnostics` - Clear all diagnostic markers

## Configuration

- `reactAnalyzer.enabled` - Enable/disable the extension (default: `true`)
- `reactAnalyzer.analyzeOnSave` - Run analysis when files are saved (default: `true`)
- `reactAnalyzer.cliPath` - Path to react-analyzer CLI binary (leave empty to use bundled binary)

## Usage

### Analyze Current File

1. Open a React file (`.tsx`, `.jsx`, `.ts`, `.js`)
2. Run command: `React Analyzer: Analyze Current File`
3. Issues will appear as squiggly underlines in the editor

### Analyze on Save

By default, files are automatically analyzed when saved. You can disable this in settings:

```json
{
  "reactAnalyzer.analyzeOnSave": false
}
```

### View Issues

Issues appear in:
- **Editor**: Squiggly underlines with hover messages
- **Problems panel**: All issues across your workspace

### Issue Severity

- **Error** (red): Bugs that can cause runtime issues
  - `no-stale-state`: Can cause race conditions
  - `no-derived-state`: Can cause sync issues

- **Warning** (yellow): Performance issues
  - `no-object-deps`: Causes infinite re-renders
  - `unstable-props-to-memo`: Breaks memoization
  - `no-inline-props`: Creates unnecessary re-renders

- **Information** (blue): Code quality suggestions
  - `deep-prop-drilling`: Consider using Context API

## Development

### Prerequisites

- Go 1.20+ (for building the CLI)
- Node.js 16+ (for the extension)

### Building

```bash
# Build the CLI
cd ..
go build -o react-analyzer cmd/react-analyzer/main.go

# Build the extension
cd vscode-extension
npm install
npm run compile
```

### Testing

Press F5 in VS Code to launch the Extension Development Host and test the extension.

## License

MIT
