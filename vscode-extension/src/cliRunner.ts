import * as child_process from 'child_process';
import * as path from 'path';
import * as vscode from 'vscode';
import { AnalysisResult } from './types';

export class CliRunner {
    private cliPath: string;

    constructor(context: vscode.ExtensionContext) {
        // Get CLI path from configuration or use bundled binary
        const configuredPath = vscode.workspace.getConfiguration('reactAnalyzer').get<string>('cliPath');

        if (configuredPath && configuredPath.length > 0) {
            this.cliPath = configuredPath;
        } else {
            // Use bundled binary (will be in extension root after packaging)
            // For now, use the binary from the parent directory during development
            this.cliPath = path.join(context.extensionPath, '..', 'react-analyzer');
        }
    }

    /**
     * Run analysis on a file or directory
     * @param targetPath Path to file or directory to analyze
     * @param includeGraph Whether to include the dependency graph in the results
     * @returns Promise with analysis results
     */
    async analyze(targetPath: string, includeGraph: boolean = false): Promise<AnalysisResult> {
        return new Promise((resolve, reject) => {
            const args = ['-json'];
            if (includeGraph) {
                args.push('-graph');
            }
            args.push(targetPath);

            console.log(`Running: ${this.cliPath} ${args.join(' ')}`);

            const process = child_process.spawn(this.cliPath, args);

            let stdout = '';
            let stderr = '';

            process.stdout.on('data', (data: Buffer) => {
                stdout += data.toString();
            });

            process.stderr.on('data', (data: Buffer) => {
                stderr += data.toString();
            });

            process.on('error', (error) => {
                reject(new Error(`Failed to spawn CLI process: ${error.message}`));
            });

            process.on('close', (code) => {
                // Exit codes: 0 = no issues, 1 = issues found, 2+ = error
                if (code !== null && code >= 2) {
                    reject(new Error(`CLI exited with error code ${code}: ${stderr}`));
                    return;
                }

                try {
                    const result = JSON.parse(stdout) as AnalysisResult;
                    resolve(result);
                } catch (error) {
                    reject(new Error(`Failed to parse CLI output: ${error instanceof Error ? error.message : String(error)}\nOutput: ${stdout}`));
                }
            });
        });
    }

    /**
     * Check if the CLI binary exists and is executable
     */
    async checkCliAvailable(): Promise<boolean> {
        return new Promise((resolve) => {
            child_process.exec(`"${this.cliPath}" -version`, (error, stdout) => {
                if (error) {
                    resolve(false);
                    return;
                }

                // Check if output contains version info
                resolve(stdout.includes('react-analyzer'));
            });
        });
    }

    /**
     * Get the CLI version
     */
    async getVersion(): Promise<string> {
        return new Promise((resolve, reject) => {
            child_process.exec(`"${this.cliPath}" -version`, (error, stdout) => {
                if (error) {
                    reject(new Error(`Failed to get CLI version: ${error.message}`));
                    return;
                }

                resolve(stdout.trim());
            });
        });
    }
}
