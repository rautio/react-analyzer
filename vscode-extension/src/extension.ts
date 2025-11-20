import * as vscode from 'vscode';
import { CliRunner } from './cliRunner';
import { DiagnosticsProvider } from './diagnosticsProvider';
import { ComponentTreeProvider } from './treeViewProvider';
import { GraphWebview } from './views/GraphWebview';

let cliRunner: CliRunner;
let diagnosticsProvider: DiagnosticsProvider;
let treeProvider: ComponentTreeProvider;

/**
 * Extension activation function
 * Called when the extension is activated (on first command or language activation)
 */
export async function activate(context: vscode.ExtensionContext) {
    console.log('React Analyzer extension is now active');

    // Initialize CLI runner, diagnostics provider, and tree view
    cliRunner = new CliRunner(context);
    diagnosticsProvider = new DiagnosticsProvider();
    treeProvider = new ComponentTreeProvider();

    // Register tree view
    vscode.window.registerTreeDataProvider('reactAnalyzerComponentTree', treeProvider);

    // Check if CLI is available
    const isAvailable = await cliRunner.checkCliAvailable();
    if (!isAvailable) {
        vscode.window.showWarningMessage(
            'React Analyzer CLI not found. Please check your configuration.'
        );
    } else {
        try {
            const version = await cliRunner.getVersion();
            console.log(`React Analyzer CLI version: ${version}`);
        } catch (error) {
            console.error('Failed to get CLI version:', error);
        }
    }

    // Register command: Analyze current file
    const analyzeFileCommand = vscode.commands.registerCommand(
        'reactAnalyzer.analyzeFile',
        async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showErrorMessage('No active editor');
                return;
            }

            const filePath = editor.document.uri.fsPath;
            if (!isReactFile(filePath)) {
                vscode.window.showInformationMessage('Not a React file');
                return;
            }

            await analyzeFile(filePath);
        }
    );

    // Register command: Analyze workspace
    const analyzeWorkspaceCommand = vscode.commands.registerCommand(
        'reactAnalyzer.analyzeWorkspace',
        async () => {
            const workspaceFolders = vscode.workspace.workspaceFolders;
            if (!workspaceFolders || workspaceFolders.length === 0) {
                vscode.window.showErrorMessage('No workspace folder open');
                return;
            }

            // For now, analyze the first workspace folder
            const workspacePath = workspaceFolders[0].uri.fsPath;
            await analyzeWorkspace(workspacePath);
        }
    );

    // Register command: Clear diagnostics
    const clearDiagnosticsCommand = vscode.commands.registerCommand(
        'reactAnalyzer.clearDiagnostics',
        () => {
            diagnosticsProvider.clear();
            vscode.window.showInformationMessage('React Analyzer diagnostics cleared');
        }
    );

    // Register command: Refresh component tree
    const refreshTreeCommand = vscode.commands.registerCommand(
        'reactAnalyzer.refreshComponentTree',
        async () => {
            // Re-analyze the current workspace to update the tree
            const workspaceFolders = vscode.workspace.workspaceFolders;
            if (workspaceFolders && workspaceFolders.length > 0) {
                const workspacePath = workspaceFolders[0].uri.fsPath;
                await analyzeWorkspace(workspacePath);
            }
        }
    );

    // Register command: Navigate to location
    const navigateCommand = vscode.commands.registerCommand(
        'reactAnalyzer.navigateToLocation',
        async (filePath: string, line: number, column: number) => {
            const uri = vscode.Uri.file(filePath);
            const document = await vscode.workspace.openTextDocument(uri);
            const editor = await vscode.window.showTextDocument(document);

            // Convert from 1-indexed to 0-indexed
            const position = new vscode.Position(Math.max(0, line - 1), Math.max(0, column));
            editor.selection = new vscode.Selection(position, position);
            editor.revealRange(new vscode.Range(position, position), vscode.TextEditorRevealType.InCenter);
        }
    );

    // Register command: Show dependency graph
    const showGraphCommand = vscode.commands.registerCommand(
        'reactAnalyzer.showGraph',
        async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showErrorMessage('No active editor');
                return;
            }

            const filePath = editor.document.uri.fsPath;
            if (!isReactFile(filePath)) {
                vscode.window.showInformationMessage('Please open a React file (.tsx, .jsx, .ts, .js)');
                return;
            }

            await GraphWebview.show(context.extensionUri, filePath);
        }
    );

    // Register event: Analyze on save (if enabled)
    const onSaveHandler = vscode.workspace.onDidSaveTextDocument(async (document) => {
        const config = vscode.workspace.getConfiguration('reactAnalyzer');
        const analyzeOnSave = config.get<boolean>('analyzeOnSave', true);
        const enabled = config.get<boolean>('enabled', true);

        if (enabled && analyzeOnSave && isReactFile(document.uri.fsPath)) {
            await analyzeFile(document.uri.fsPath);
        }
    });

    // Add disposables to context
    context.subscriptions.push(
        analyzeFileCommand,
        analyzeWorkspaceCommand,
        clearDiagnosticsCommand,
        refreshTreeCommand,
        navigateCommand,
        showGraphCommand,
        onSaveHandler,
        diagnosticsProvider
    );
}

/**
 * Analyze a single file
 * Note: To enable cross-file analysis (like prop drilling detection),
 * we analyze the directory containing the file, not just the file itself
 */
async function analyzeFile(filePath: string): Promise<void> {
    try {
        await vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: 'React Analyzer',
                cancellable: false,
            },
            async (progress) => {
                progress.report({ message: 'Analyzing project...' });

                // Get the directory containing the file
                // For cross-file analysis, we need to analyze all files in the directory
                const path = require('path');
                const dirPath = path.dirname(filePath);

                // Analyze the entire directory to enable cross-file graph analysis
                const result = await cliRunner.analyze(dirPath, true);
                diagnosticsProvider.updateDiagnostics(result.issues);

                // Update component tree if graph data is available
                if (result.graph) {
                    treeProvider.updateGraph(result.graph);
                }

                // Filter issues to only show those for files in the same directory
                const issuesInDir = result.issues.filter(issue =>
                    issue.filePath.startsWith(dirPath)
                );

                const issueCount = issuesInDir.length;
                if (issueCount === 0) {
                    vscode.window.showInformationMessage(`✓ No issues found in directory (${result.stats.filesAnalyzed} files analyzed)`);
                } else {
                    const plural = issueCount === 1 ? 'issue' : 'issues';
                    vscode.window.showWarningMessage(`Found ${issueCount} ${plural} in ${result.stats.filesWithIssues} files`);
                }
            }
        );
    } catch (error) {
        vscode.window.showErrorMessage(
            `React Analyzer error: ${error instanceof Error ? error.message : String(error)}`
        );
    }
}

/**
 * Analyze entire workspace
 */
async function analyzeWorkspace(workspacePath: string): Promise<void> {
    try {
        await vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: 'React Analyzer',
                cancellable: false,
            },
            async (progress) => {
                progress.report({ message: 'Analyzing workspace...' });

                const result = await cliRunner.analyze(workspacePath, true);
                diagnosticsProvider.updateDiagnostics(result.issues);

                // Update component tree if graph data is available
                if (result.graph) {
                    treeProvider.updateGraph(result.graph);
                }

                const stats = result.stats;
                const issueCount = stats.totalIssues;

                if (issueCount === 0) {
                    vscode.window.showInformationMessage(
                        `✓ No issues found in ${stats.filesAnalyzed} files`
                    );
                } else {
                    const plural = issueCount === 1 ? 'issue' : 'issues';
                    vscode.window.showWarningMessage(
                        `Found ${issueCount} ${plural} in ${stats.filesWithIssues} files`
                    );
                }
            }
        );
    } catch (error) {
        vscode.window.showErrorMessage(
            `React Analyzer error: ${error instanceof Error ? error.message : String(error)}`
        );
    }
}

/**
 * Check if a file is a React file
 */
function isReactFile(filePath: string): boolean {
    return /\.(tsx|jsx|ts|js)$/.test(filePath);
}

/**
 * Extension deactivation function
 * Called when the extension is deactivated
 */
export function deactivate() {
    console.log('React Analyzer extension is now deactivated');
}
