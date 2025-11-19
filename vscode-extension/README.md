# React Analyzer for VS Code

Static analysis extension for detecting React performance issues and antipatterns.

## Features

- **Real-time diagnostics**: See performance issues directly in your editor with squiggly underlines
- **Component Tree View**: Visual sidebar showing component hierarchy and state flow
  - Hierarchical component tree with parent-child relationships
  - State nodes showing where state is defined and consumed
  - Click any item to navigate to its definition
  - Refresh button to re-analyze the project
- **Enhanced diagnostics with Related Information**: Click through the full chain of related code locations
- **Analyze on save**: Automatically analyze files when you save them
- **Workspace analysis**: Analyze your entire project with a single command
- **6 production-ready rules**:
  - `no-object-deps`: Detect inline objects/arrays in hook dependencies
  - `unstable-props-to-memo`: Detect unstable props passed to memoized components
  - `no-derived-state`: Detect useState mirroring props (anti-pattern)
  - `no-stale-state`: Detect state updates without functional form
  - `no-inline-props`: Detect inline objects/arrays/functions in JSX props
  - `deep-prop-drilling`: Detect props drilled through 3+ component levels (configurable)

## Commands

- `React Analyzer: Analyze Current File` - Analyze the currently open file
- `React Analyzer: Analyze Workspace` - Analyze all React files in the workspace
- `React Analyzer: Clear All Diagnostics` - Clear all diagnostic markers
- `Refresh Component Tree` - Re-analyze and update the component tree view

## Configuration

### VS Code Settings

- `reactAnalyzer.enabled` - Enable/disable the extension (default: `true`)
- `reactAnalyzer.analyzeOnSave` - Run analysis when files are saved (default: `true`)
- `reactAnalyzer.cliPath` - Path to react-analyzer CLI binary (leave empty to use bundled binary)

### Project Configuration

Create a `.rarc` (or `.reactanalyzerrc.json`) file in your project root to customize rules and path aliases:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  },
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 3
      }
    }
  }
}
```

See the [Configuration Guide](../docs/CONFIG.md) for more details.

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
- **Problems panel**: All issues across your workspace with clickable Related Information

### Component Tree View

The Component Tree View appears in the Explorer sidebar:

1. **View the tree**: Open the Explorer sidebar and find the "React Components" section
2. **Navigate**:
   - Expand "Components" to see your component hierarchy
   - Expand "State" to see all state nodes in your project
   - Click any item to jump to its definition in code
3. **Refresh**: Click the refresh icon in the tree view toolbar to re-analyze

The tree shows:
- Parent-child component relationships
- Which components are memoized `[memo]`
- State nodes defined in each component `(2 state)`
- State type indicators `[useState]`, `[context]`, etc.

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
