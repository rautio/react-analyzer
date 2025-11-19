# Phase 2: VS Code Extension Architecture

**Document Version:** 1.0
**Date:** 2025-11-18
**Status:** Draft - Design Phase
**Assumes:** No prior VS Code extension development experience

---

## Overview

This document provides a comprehensive guide to building the React Analyzer VS Code extension. It covers the complete architecture, VS Code APIs, project structure, and implementation patterns.

---

## Table of Contents

1. [VS Code Extension Fundamentals](#1-vs-code-extension-fundamentals)
2. [Project Structure](#2-project-structure)
3. [Architecture Overview](#3-architecture-overview)
4. [Core Components](#4-core-components)
5. [Communication with Go CLI](#5-communication-with-go-cli)
6. [UI Components](#6-ui-components)
7. [VS Code API Usage](#7-vs-code-api-usage)
8. [Development Workflow](#8-development-workflow)
9. [Testing Strategy](#9-testing-strategy)
10. [Publishing](#10-publishing)

---

## 1. VS Code Extension Fundamentals

### What is a VS Code Extension?

A VS Code extension is a Node.js/TypeScript application that runs in VS Code's extension host process and can:
- Add commands to the command palette
- Create custom UI panels (tree views, webviews)
- Decorate the editor (underlines, colors, icons)
- Respond to file changes
- Integrate with language servers

### Extension Activation

Extensions are **lazy-loaded** - they only activate when specific events occur:

```json
{
  "activationEvents": [
    "workspaceContains:**/*.tsx",  // When workspace has .tsx files
    "workspaceContains:**/*.jsx",  // When workspace has .jsx files
    "onCommand:react-analyzer.analyzeProject",  // When command is executed
    "onLanguage:typescriptreact",  // When editing .tsx file
    "onLanguage:javascriptreact"   // When editing .jsx file
  ]
}
```

### Extension Lifecycle

```
1. User opens VS Code workspace
2. VS Code scans for extensions
3. Extension waits for activation event
4. activate() function called
5. Extension runs until VS Code closes
6. deactivate() function called (cleanup)
```

### Key Concepts

**Extension Context:** Container for extension's state and resources
```typescript
export function activate(context: vscode.ExtensionContext) {
    // context.subscriptions: Register disposables (cleaned up on deactivate)
    // context.globalState: Persistent key-value storage
    // context.workspaceState: Workspace-specific storage
    // context.extensionPath: Path to extension directory
}
```

**Commands:** Actions users can invoke
```typescript
const disposable = vscode.commands.registerCommand(
    'react-analyzer.analyzeProject',
    async () => {
        // Command implementation
    }
);
context.subscriptions.push(disposable);
```

**Tree Views:** Hierarchical data display (like file explorer)
```typescript
const treeDataProvider = new ComponentTreeProvider();
vscode.window.createTreeView('react-analyzer.componentTree', {
    treeDataProvider: treeDataProvider
});
```

**Webviews:** Custom HTML/JS UI panels
```typescript
const panel = vscode.window.createWebviewPanel(
    'react-analyzer.graph',
    'State Dependency Graph',
    vscode.ViewColumn.Two,
    {
        enableScripts: true
    }
);
```

---

## 2. Project Structure

### Directory Layout

```
react-analyzer/
├── cmd/react-analyzer/       # Go CLI (existing)
├── internal/                 # Go internals (existing)
├── vscode-extension/         # NEW: VS Code extension
│   ├── src/                  # TypeScript source
│   │   ├── extension.ts      # Entry point (activate/deactivate)
│   │   ├── commands/         # Command implementations
│   │   │   ├── analyzeProject.ts
│   │   │   ├── analyzeFile.ts
│   │   │   └── simulateStateChange.ts
│   │   ├── providers/        # Tree view and other providers
│   │   │   ├── ComponentTreeProvider.ts
│   │   │   ├── DiagnosticsProvider.ts
│   │   │   └── CodeActionProvider.ts
│   │   ├── views/            # Webview components
│   │   │   ├── GraphView.ts
│   │   │   └── DashboardView.ts
│   │   ├── cli/              # Go CLI integration
│   │   │   ├── CliRunner.ts
│   │   │   └── GraphParser.ts
│   │   ├── models/           # Data models (matches Go graph)
│   │   │   ├── Graph.ts
│   │   │   ├── StateNode.ts
│   │   │   ├── ComponentNode.ts
│   │   │   └── Edge.ts
│   │   ├── ui/               # Reusable UI components
│   │   │   └── webview/      # React components for webviews
│   │   │       ├── GraphVisualization.tsx
│   │   │       └── ComponentExplorer.tsx
│   │   └── utils/            # Utilities
│   │       ├── logger.ts
│   │       └── fileWatcher.ts
│   ├── media/                # Icons, CSS, images
│   ├── out/                  # Compiled JavaScript (gitignored)
│   ├── package.json          # Extension manifest
│   ├── tsconfig.json         # TypeScript configuration
│   ├── .vscodeignore         # Files to exclude from .vsix
│   └── README.md             # Extension documentation
```

### File Purposes

**src/extension.ts** - Entry point
- Exports `activate()` and `deactivate()`
- Registers all commands, views, providers
- Sets up file watchers

**package.json** - Extension manifest
- Defines commands, views, configuration
- Specifies activation events
- Lists dependencies

**tsconfig.json** - TypeScript config
- Target: ES2020
- Module: CommonJS
- Strict mode enabled

---

## 3. Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│  VS Code UI                                         │
│  ├── Command Palette                                │
│  ├── Editor Decorations                             │
│  ├── Tree View (Component Explorer)                 │
│  └── Webview Panel (Graph Visualization)            │
└────────────────┬────────────────────────────────────┘
                 │ VS Code Extension API
┌────────────────▼────────────────────────────────────┐
│  Extension Host (Node.js/TypeScript)                │
│  ├── Command Handlers                               │
│  ├── Tree View Providers                            │
│  ├── Diagnostic Provider                            │
│  ├── File Watcher                                   │
│  └── CLI Runner                                     │
└────────────────┬────────────────────────────────────┘
                 │ Child Process / File System
┌────────────────▼────────────────────────────────────┐
│  Go CLI (react-analyzer)                            │
│  ├── Analyze project                                │
│  ├── Generate graph JSON                            │
│  └── Return issues + graph                          │
└─────────────────────────────────────────────────────┘
```

### Data Flow

```
1. User Action
   ↓
2. Command Handler (TypeScript)
   ↓
3. CLI Runner spawns Go process
   ↓
4. Go CLI analyzes code
   ↓
5. Go returns JSON (graph + issues)
   ↓
6. Extension parses JSON into models
   ↓
7. Update UI:
   - Refresh tree view
   - Add editor decorations
   - Update webview
   ↓
8. User sees results
```

---

## 4. Core Components

### 4.1 Extension Entry Point

**File:** `src/extension.ts`

```typescript
import * as vscode from 'vscode';
import { ComponentTreeProvider } from './providers/ComponentTreeProvider';
import { DiagnosticsProvider } from './providers/DiagnosticsProvider';
import { CliRunner } from './cli/CliRunner';
import { FileWatcher } from './utils/fileWatcher';

export function activate(context: vscode.ExtensionContext) {
    console.log('React Analyzer extension activating...');

    // Initialize CLI runner
    const cliRunner = new CliRunner(context.extensionPath);

    // Initialize diagnostics
    const diagnostics = vscode.languages.createDiagnosticCollection('react-analyzer');
    context.subscriptions.push(diagnostics);

    const diagnosticsProvider = new DiagnosticsProvider(diagnostics, cliRunner);

    // Register tree view provider
    const componentTreeProvider = new ComponentTreeProvider(cliRunner);
    const treeView = vscode.window.createTreeView('react-analyzer.componentTree', {
        treeDataProvider: componentTreeProvider,
        showCollapseAll: true
    });
    context.subscriptions.push(treeView);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('react-analyzer.analyzeProject', async () => {
            await analyzeProject(cliRunner, diagnosticsProvider, componentTreeProvider);
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('react-analyzer.analyzeCurrentFile', async () => {
            await analyzeCurrentFile(cliRunner, diagnosticsProvider);
        })
    );

    // Setup file watcher
    const fileWatcher = new FileWatcher(
        ['**/*.tsx', '**/*.jsx', '**/*.ts', '**/*.js'],
        async (uri: vscode.Uri) => {
            await analyzeFile(uri, cliRunner, diagnosticsProvider);
        }
    );
    context.subscriptions.push(fileWatcher);

    // Run initial analysis
    if (vscode.workspace.workspaceFolders) {
        analyzeProject(cliRunner, diagnosticsProvider, componentTreeProvider);
    }

    console.log('React Analyzer extension activated!');
}

export function deactivate() {
    console.log('React Analyzer extension deactivating...');
}

async function analyzeProject(
    cliRunner: CliRunner,
    diagnosticsProvider: DiagnosticsProvider,
    treeProvider: ComponentTreeProvider
) {
    const workspaceRoot = vscode.workspace.workspaceFolders?.[0].uri.fsPath;
    if (!workspaceRoot) {
        vscode.window.showErrorMessage('No workspace folder open');
        return;
    }

    await vscode.window.withProgress({
        location: vscode.ProgressLocation.Notification,
        title: 'Analyzing React project...',
        cancellable: false
    }, async (progress) => {
        const result = await cliRunner.analyze(workspaceRoot, { json: true, graph: true });

        // Update diagnostics
        diagnosticsProvider.update(result);

        // Update tree view
        treeProvider.update(result.graph);

        // Show completion message
        if (result.issues.length === 0) {
            vscode.window.showInformationMessage('✓ No issues found!');
        } else {
            vscode.window.showWarningMessage(
                `Found ${result.issues.length} issues in ${result.filesWithIssues} files`
            );
        }
    });
}
```

### 4.2 Package.json (Extension Manifest)

**File:** `package.json`

```json
{
  "name": "react-analyzer",
  "displayName": "React Analyzer",
  "description": "Static analysis for React state architecture and performance",
  "version": "0.1.0",
  "publisher": "react-analyzer",
  "icon": "media/icon.png",
  "engines": {
    "vscode": "^1.85.0"
  },
  "categories": [
    "Linters",
    "Programming Languages",
    "Visualization"
  ],
  "keywords": [
    "react",
    "performance",
    "state management",
    "static analysis",
    "linter"
  ],
  "activationEvents": [
    "workspaceContains:**/*.tsx",
    "workspaceContains:**/*.jsx",
    "onCommand:react-analyzer.analyzeProject",
    "onLanguage:typescriptreact",
    "onLanguage:javascriptreact"
  ],
  "main": "./out/extension.js",
  "contributes": {
    "commands": [
      {
        "command": "react-analyzer.analyzeProject",
        "title": "React Analyzer: Analyze Entire Project",
        "icon": "$(search)"
      },
      {
        "command": "react-analyzer.analyzeCurrentFile",
        "title": "React Analyzer: Analyze Current File",
        "icon": "$(file)"
      },
      {
        "command": "react-analyzer.showGraph",
        "title": "React Analyzer: Show State Dependency Graph",
        "icon": "$(graph)"
      },
      {
        "command": "react-analyzer.simulateStateChange",
        "title": "React Analyzer: Simulate State Change",
        "icon": "$(debug-start)"
      }
    ],
    "viewsContainers": {
      "activitybar": [
        {
          "id": "react-analyzer",
          "title": "React Analyzer",
          "icon": "media/icon.svg"
        }
      ]
    },
    "views": {
      "react-analyzer": [
        {
          "id": "react-analyzer.componentTree",
          "name": "Component Explorer",
          "icon": "media/icon.svg"
        },
        {
          "id": "react-analyzer.stateList",
          "name": "State Overview"
        }
      ]
    },
    "configuration": {
      "title": "React Analyzer",
      "properties": {
        "react-analyzer.autoAnalyze": {
          "type": "boolean",
          "default": true,
          "description": "Automatically analyze files on save"
        },
        "react-analyzer.cliPath": {
          "type": "string",
          "default": "",
          "description": "Path to react-analyzer CLI (leave empty for bundled version)"
        },
        "react-analyzer.severity": {
          "type": "object",
          "default": {
            "deep-prop-drilling": "warning",
            "no-object-deps": "error",
            "unstable-props-to-memo": "error"
          },
          "description": "Severity levels for each rule"
        }
      }
    }
  },
  "scripts": {
    "vscode:prepublish": "npm run compile",
    "compile": "tsc -p ./",
    "watch": "tsc -watch -p ./",
    "lint": "eslint src --ext ts",
    "test": "node ./out/test/runTest.js"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/vscode": "^1.85.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "eslint": "^8.50.0",
    "typescript": "^5.2.0"
  },
  "dependencies": {
    "d3": "^7.8.5",
    "@types/d3": "^7.4.2"
  }
}
```

---

## 5. Communication with Go CLI

### 5.1 CLI Runner

**File:** `src/cli/CliRunner.ts`

```typescript
import * as vscode from 'vscode';
import * as path from 'path';
import * as cp from 'child_process';
import { Graph } from '../models/Graph';
import { Issue } from '../models/Issue';

export interface AnalysisResult {
    success: boolean;
    issues: Issue[];
    filesWithIssues: number;
    graph?: Graph;
    error?: string;
}

export interface AnalysisOptions {
    json?: boolean;
    graph?: boolean;
    verbose?: boolean;
    quiet?: boolean;
}

export class CliRunner {
    private cliPath: string;

    constructor(extensionPath: string) {
        // Check user configuration first
        const configPath = vscode.workspace.getConfiguration('react-analyzer').get<string>('cliPath');

        if (configPath && configPath.trim() !== '') {
            this.cliPath = configPath;
        } else {
            // Use bundled CLI
            const platform = process.platform;
            const binary = platform === 'win32' ? 'react-analyzer.exe' : 'react-analyzer';
            this.cliPath = path.join(extensionPath, 'bin', platform, binary);
        }
    }

    async analyze(targetPath: string, options: AnalysisOptions = {}): Promise<AnalysisResult> {
        return new Promise((resolve, reject) => {
            const args: string[] = [];

            // Add flags
            if (options.json) args.push('--json');
            if (options.graph) args.push('--graph');
            if (options.verbose) args.push('--verbose');
            if (options.quiet) args.push('--quiet');

            // Add target
            args.push(targetPath);

            // Spawn process
            const child = cp.spawn(this.cliPath, args, {
                cwd: vscode.workspace.workspaceFolders?.[0].uri.fsPath
            });

            let stdout = '';
            let stderr = '';

            child.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            child.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            child.on('close', (code) => {
                if (code === 0 || code === 1) {
                    // 0 = no issues, 1 = issues found (both are success)
                    try {
                        const result = this.parseOutput(stdout, options.json || false);
                        resolve(result);
                    } catch (error) {
                        reject(new Error(`Failed to parse CLI output: ${error}`));
                    }
                } else {
                    // 2 = error
                    reject(new Error(`CLI error: ${stderr}`));
                }
            });

            child.on('error', (error) => {
                reject(new Error(`Failed to spawn CLI: ${error.message}`));
            });
        });
    }

    private parseOutput(output: string, isJson: boolean): AnalysisResult {
        if (isJson) {
            const data = JSON.parse(output);
            return {
                success: true,
                issues: data.issues || [],
                filesWithIssues: data.filesWithIssues || 0,
                graph: data.graph ? this.parseGraph(data.graph) : undefined
            };
        } else {
            // Parse text output (for non-JSON mode)
            // This is a fallback, we'll primarily use JSON mode
            return {
                success: true,
                issues: [],
                filesWithIssues: 0
            };
        }
    }

    private parseGraph(graphData: any): Graph {
        // Convert JSON to Graph model
        // See GraphParser.ts for full implementation
        return Graph.fromJSON(graphData);
    }

    async checkCliAvailable(): Promise<boolean> {
        try {
            await this.analyze('--version');
            return true;
        } catch {
            return false;
        }
    }
}
```

### 5.2 Go CLI Modifications Needed

The Go CLI needs to output JSON with the graph:

**New flag:** `--graph` - Output graph JSON
**New flag:** `--json` - Output JSON format

```go
// cmd/react-analyzer/main.go
type AnalysisOutput struct {
    Success        bool            `json:"success"`
    Issues         []IssueJSON     `json:"issues"`
    FilesAnalyzed  int             `json:"filesAnalyzed"`
    FilesWithIssues int            `json:"filesWithIssues"`
    Graph          *graph.Graph    `json:"graph,omitempty"`
    Elapsed        string          `json:"elapsed"`
}

func outputJSON(output AnalysisOutput) {
    data, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
        os.Exit(2)
    }
    fmt.Println(string(data))
}
```

---

## 6. UI Components

### 6.1 Component Tree View

**File:** `src/providers/ComponentTreeProvider.ts`

```typescript
import * as vscode from 'vscode';
import { Graph } from '../models/Graph';
import { ComponentNode } from '../models/ComponentNode';

export class ComponentTreeProvider implements vscode.TreeDataProvider<ComponentTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ComponentTreeItem | undefined | null> =
        new vscode.EventEmitter<ComponentTreeItem | undefined | null>();
    readonly onDidChangeTreeData: vscode.Event<ComponentTreeItem | undefined | null> =
        this._onDidChangeTreeData.event;

    private graph?: Graph;

    constructor(private cliRunner: any) {}

    refresh(): void {
        this._onDidChangeTreeData.fire(undefined);
    }

    update(graph: Graph): void {
        this.graph = graph;
        this.refresh();
    }

    getTreeItem(element: ComponentTreeItem): vscode.TreeItem {
        return element;
    }

    getChildren(element?: ComponentTreeItem): Thenable<ComponentTreeItem[]> {
        if (!this.graph) {
            return Promise.resolve([]);
        }

        if (!element) {
            // Root level - show root components
            const rootComponents = this.graph.getRootComponents();
            return Promise.resolve(
                rootComponents.map(comp => new ComponentTreeItem(comp, this.graph!))
            );
        } else {
            // Show children of this component
            const children = this.graph.getComponentChildren(element.component.id);
            return Promise.resolve(
                children.map(child => new ComponentTreeItem(child, this.graph!))
            );
        }
    }
}

class ComponentTreeItem extends vscode.TreeItem {
    constructor(
        public readonly component: ComponentNode,
        private graph: Graph
    ) {
        super(
            component.name,
            component.children.length > 0
                ? vscode.TreeItemCollapsibleState.Collapsed
                : vscode.TreeItemCollapsibleState.None
        );

        this.tooltip = this.buildTooltip();
        this.description = this.buildDescription();
        this.iconPath = this.getIcon();
        this.contextValue = 'component';

        // Command when clicked
        this.command = {
            command: 'react-analyzer.selectComponent',
            title: 'Select Component',
            arguments: [component]
        };
    }

    private buildTooltip(): string {
        let tooltip = `${this.component.name}\n`;
        tooltip += `File: ${this.component.location.filePath}\n`;

        if (this.component.isMemoized) {
            tooltip += `✓ Memoized\n`;
        }

        const state = this.graph.getComponentState(this.component.id);
        if (state.length > 0) {
            tooltip += `\nState (${state.length}):\n`;
            state.forEach(s => {
                tooltip += `  - ${s.name} (${s.type})\n`;
            });
        }

        return tooltip;
    }

    private buildDescription(): string {
        const parts: string[] = [];

        if (this.component.isMemoized) {
            parts.push('memo');
        }

        const stateCount = this.graph.getComponentState(this.component.id).length;
        if (stateCount > 0) {
            parts.push(`${stateCount} state`);
        }

        return parts.join(', ');
    }

    private getIcon(): vscode.ThemeIcon {
        if (this.component.isMemoized) {
            return new vscode.ThemeIcon('symbol-class', new vscode.ThemeColor('charts.green'));
        }
        return new vscode.ThemeIcon('symbol-class');
    }
}
```

### 6.2 Diagnostics Provider

**File:** `src/providers/DiagnosticsProvider.ts`

```typescript
import * as vscode from 'vscode';
import { AnalysisResult } from '../cli/CliRunner';
import { Issue } from '../models/Issue';

export class DiagnosticsProvider {
    constructor(
        private diagnosticCollection: vscode.DiagnosticCollection,
        private cliRunner: any
    ) {}

    update(result: AnalysisResult): void {
        // Clear all diagnostics first
        this.diagnosticCollection.clear();

        // Group issues by file
        const issuesByFile = new Map<string, Issue[]>();
        for (const issue of result.issues) {
            const uri = vscode.Uri.file(issue.filePath);
            const key = uri.toString();

            if (!issuesByFile.has(key)) {
                issuesByFile.set(key, []);
            }
            issuesByFile.get(key)!.push(issue);
        }

        // Create diagnostics for each file
        for (const [uriString, issues] of issuesByFile) {
            const uri = vscode.Uri.parse(uriString);
            const diagnostics = issues.map(issue => this.issueToDiagnostic(issue));
            this.diagnosticCollection.set(uri, diagnostics);
        }
    }

    private issueToDiagnostic(issue: Issue): vscode.Diagnostic {
        const range = new vscode.Range(
            issue.line - 1,
            issue.column - 1,
            issue.line - 1,
            issue.column + 20  // Approximate length
        );

        const diagnostic = new vscode.Diagnostic(
            range,
            issue.message,
            this.getSeverity(issue.rule)
        );

        diagnostic.source = 'react-analyzer';
        diagnostic.code = issue.rule;

        return diagnostic;
    }

    private getSeverity(rule: string): vscode.DiagnosticSeverity {
        const config = vscode.workspace.getConfiguration('react-analyzer');
        const severityMap = config.get<Record<string, string>>('severity') || {};
        const severity = severityMap[rule] || 'warning';

        switch (severity) {
            case 'error':
                return vscode.DiagnosticSeverity.Error;
            case 'warning':
                return vscode.DiagnosticSeverity.Warning;
            case 'info':
                return vscode.DiagnosticSeverity.Information;
            case 'hint':
                return vscode.DiagnosticSeverity.Hint;
            default:
                return vscode.DiagnosticSeverity.Warning;
        }
    }
}
```

---

## 7. VS Code API Usage

### 7.1 Key APIs Reference

**Window API** - UI interactions
```typescript
vscode.window.showInformationMessage('Success!');
vscode.window.showErrorMessage('Failed!');
vscode.window.showWarningMessage('Warning!');

vscode.window.withProgress({
    location: vscode.ProgressLocation.Notification,
    title: 'Analyzing...',
}, async (progress) => {
    // Long-running task
});
```

**Workspace API** - File system and configuration
```typescript
const workspaceRoot = vscode.workspace.workspaceFolders?.[0].uri.fsPath;
const config = vscode.workspace.getConfiguration('react-analyzer');
const autoAnalyze = config.get<boolean>('autoAnalyze');
```

**Languages API** - Diagnostics
```typescript
const diagnostics = vscode.languages.createDiagnosticCollection('react-analyzer');
diagnostics.set(uri, [diagnostic1, diagnostic2]);
```

**Commands API** - Register commands
```typescript
vscode.commands.registerCommand('react-analyzer.analyze', handler);
```

**File System Watcher** - Respond to file changes
```typescript
const watcher = vscode.workspace.createFileSystemWatcher('**/*.{tsx,jsx}');
watcher.onDidChange(uri => {
    // File changed
});
```

---

## 8. Development Workflow

### Setup

```bash
cd vscode-extension
npm install
npm run compile
```

### Testing Locally (F5)

1. Open `vscode-extension/` in VS Code
2. Press `F5` - launches Extension Development Host
3. Test extension in new window
4. Use Debug Console to see console.log output

### Debugging

**Set breakpoints** in TypeScript code
**Use Debug Console** for runtime inspection
**Check Output panel** → "React Analyzer" for logs

### Packaging

```bash
npm install -g vsce
vsce package
# Creates react-analyzer-0.1.0.vsix
```

### Publishing

```bash
vsce publish
```

---

## 9. Testing Strategy

### Unit Tests

Test individual components:

```typescript
// src/test/suite/cliRunner.test.ts
import * as assert from 'assert';
import { CliRunner } from '../../cli/CliRunner';

suite('CLI Runner Test Suite', () => {
    test('Parse JSON output', async () => {
        const runner = new CliRunner('/path/to/extension');
        // Test JSON parsing
    });
});
```

### Integration Tests

Test full workflows in VS Code:

```typescript
// src/test/suite/extension.test.ts
import * as vscode from 'vscode';
import * as assert from 'assert';

suite('Extension Test Suite', () => {
    test('Activate extension', async () => {
        const ext = vscode.extensions.getExtension('react-analyzer.react-analyzer');
        await ext?.activate();
        assert.ok(ext?.isActive);
    });
});
```

---

## 10. Publishing

### Prerequisites

1. Create Microsoft account
2. Create Azure DevOps organization
3. Generate Personal Access Token (PAT)
4. Create publisher: `vsce create-publisher <name>`

### Publishing Checklist

- [ ] Update version in package.json
- [ ] Update CHANGELOG.md
- [ ] Test on Windows, Mac, Linux
- [ ] Create .vsix: `vsce package`
- [ ] Test .vsix installation
- [ ] Publish: `vsce publish`

---

## Next Steps

1. **Set up project structure** (Week 5)
2. **Implement CLI runner** (Week 5)
3. **Create tree view provider** (Week 6)
4. **Add diagnostics** (Week 6)
5. **Test on real repositories** (Week 7)

---

**End of Document**
