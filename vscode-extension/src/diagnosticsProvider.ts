import * as vscode from 'vscode';
import { Issue } from './types';

export class DiagnosticsProvider {
    private diagnosticCollection: vscode.DiagnosticCollection;

    constructor() {
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('react-analyzer');
    }

    /**
     * Convert CLI issues to VS Code diagnostics and update the editor
     */
    updateDiagnostics(issues: Issue[]): void {
        // Clear existing diagnostics
        this.diagnosticCollection.clear();

        console.log(`[React Analyzer] Updating diagnostics for ${issues.length} issues`);

        // Group issues by file
        const issuesByFile = new Map<string, Issue[]>();
        for (const issue of issues) {
            console.log(`[React Analyzer] Issue in ${issue.filePath} at line ${issue.line}: ${issue.rule}`);
            const fileIssues = issuesByFile.get(issue.filePath) || [];
            fileIssues.push(issue);
            issuesByFile.set(issue.filePath, fileIssues);
        }

        // Create diagnostics for each file
        for (const [filePath, fileIssues] of issuesByFile) {
            const uri = vscode.Uri.file(filePath);
            console.log(`[React Analyzer] Setting ${fileIssues.length} diagnostics for ${uri.toString()}`);
            const diagnostics = fileIssues.map(issue => this.issueToDiagnostic(issue));
            this.diagnosticCollection.set(uri, diagnostics);
        }
    }

    /**
     * Convert a single issue to a VS Code diagnostic
     */
    private issueToDiagnostic(issue: Issue): vscode.Diagnostic {
        // Create range (line is 1-indexed in CLI, 0-indexed in VS Code)
        const line = Math.max(0, issue.line - 1);
        const column = Math.max(0, issue.column);

        // Create a range spanning multiple characters to make it more visible
        // Highlight at least 10 characters or to end of line
        const range = new vscode.Range(
            new vscode.Position(line, column),
            new vscode.Position(line, column + 10)
        );

        // Map rule names to severity
        const severity = this.getSeverity(issue.rule);

        console.log(`[React Analyzer] Creating diagnostic at line ${line + 1}, col ${column} with severity ${severity}: ${issue.rule}`);

        // Create diagnostic
        const diagnostic = new vscode.Diagnostic(
            range,
            issue.message,
            severity
        );

        diagnostic.source = 'React Analyzer';
        diagnostic.code = issue.rule;

        // Add related information if available
        if (issue.related && issue.related.length > 0) {
            diagnostic.relatedInformation = issue.related.map(related => {
                const relatedLine = Math.max(0, related.line - 1);
                const relatedColumn = Math.max(0, related.column);

                return new vscode.DiagnosticRelatedInformation(
                    new vscode.Location(
                        vscode.Uri.file(related.filePath),
                        new vscode.Position(relatedLine, relatedColumn)
                    ),
                    related.message
                );
            });
        }

        // Add related information or tags based on rule type
        if (issue.rule.includes('prop-drilling') || issue.rule.includes('inline')) {
            diagnostic.tags = [vscode.DiagnosticTag.Unnecessary];
        }

        return diagnostic;
    }

    /**
     * Map rule names to diagnostic severity
     */
    private getSeverity(rule: string): vscode.DiagnosticSeverity {
        // High priority rules that can cause bugs
        if (rule === 'no-stale-state' || rule === 'no-derived-state') {
            return vscode.DiagnosticSeverity.Error;
        }

        // Performance issues and code quality
        // Using Warning for all rules to make them more visible
        return vscode.DiagnosticSeverity.Warning;
    }

    /**
     * Clear all diagnostics
     */
    clear(): void {
        this.diagnosticCollection.clear();
    }

    /**
     * Dispose of the diagnostic collection
     */
    dispose(): void {
        this.diagnosticCollection.dispose();
    }
}
