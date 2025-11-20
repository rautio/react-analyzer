import * as vscode from 'vscode';
import * as path from 'path';
import * as child_process from 'child_process';

export class GraphWebview {
    private static currentPanel: GraphWebview | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private readonly _cliPath: string;
    private _disposables: vscode.Disposable[] = [];

    public static async show(extensionUri: vscode.Uri, filePath?: string) {
        const column = vscode.window.activeTextEditor
            ? vscode.window.activeTextEditor.viewColumn
            : undefined;

        // If we already have a panel, show it
        if (GraphWebview.currentPanel) {
            GraphWebview.currentPanel._panel.reveal(vscode.ViewColumn.Two);
            if (filePath) {
                await GraphWebview.currentPanel.updateGraph(filePath);
            }
            return;
        }

        // Otherwise, create a new panel
        const panel = vscode.window.createWebviewPanel(
            'reactAnalyzerGraph',
            'React Component Graph',
            vscode.ViewColumn.Two,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
                localResourceRoots: [
                    vscode.Uri.joinPath(extensionUri, 'media')
                ]
            }
        );

        GraphWebview.currentPanel = new GraphWebview(panel, extensionUri);

        if (filePath) {
            await GraphWebview.currentPanel.updateGraph(filePath);
        }
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
        this._panel = panel;
        this._extensionUri = extensionUri;

        // Get CLI path from configuration or use bundled binary
        const configuredPath = vscode.workspace.getConfiguration('reactAnalyzer').get<string>('cliPath');
        if (configuredPath && configuredPath.length > 0) {
            this._cliPath = configuredPath;
            console.log('Using configured CLI path:', this._cliPath);
        } else {
            // Use bundled binary from parent directory during development
            this._cliPath = path.join(path.dirname(extensionUri.fsPath), 'react-analyzer');
            console.log('Using default CLI path:', this._cliPath);
        }

        // Set the webview's initial html content
        this._panel.webview.html = this._getHtmlContent();

        // Listen for when the panel is disposed
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        // Handle messages from the webview
        this._panel.webview.onDidReceiveMessage(
            async (message) => {
                switch (message.type) {
                    case 'jumpToSource':
                        await this.jumpToSource(message.file, message.line);
                        break;
                    case 'error':
                        vscode.window.showErrorMessage(message.message);
                        break;
                }
            },
            null,
            this._disposables
        );
    }

    public async updateGraph(filePath: string) {
        try {
            console.log('=== Graph Generation Debug ===');
            console.log('File path:', filePath);

            // Run CLI to get graph data
            const analysisResult = await this.getGraphData(filePath);

            console.log('Analysis result - issues:', analysisResult.issues?.length || 0);
            console.log('Graph - component nodes:', Object.keys(analysisResult.graph?.componentNodes || {}).length);
            console.log('Graph - state nodes:', Object.keys(analysisResult.graph?.stateNodes || {}).length);
            console.log('Graph - edges:', analysisResult.graph?.edges?.length || 0);

            // Send to webview (webview will transform to Cytoscape format)
            this._panel.webview.postMessage({
                type: 'renderGraph',
                data: {
                    graph: analysisResult.graph,
                    issues: analysisResult.issues
                }
            });
        } catch (error) {
            console.error('=== Update graph error ===');
            console.error('Error type:', error instanceof Error ? error.constructor.name : typeof error);
            console.error('Error message:', error instanceof Error ? error.message : String(error));
            console.error('Full error:', error);

            const errorMsg = error instanceof Error ? error.message : String(error);
            vscode.window.showErrorMessage(`Failed to generate graph - check Output panel for details`);
            console.error('=== FULL ERROR FOR USER ===');
            console.error(errorMsg);

            // Also display error in the webview
            this._panel.webview.postMessage({
                type: 'showError',
                error: errorMsg
            });
        }
    }

    private async getGraphData(filePath: string): Promise<any> {
        return new Promise((resolve, reject) => {
            // Use --json --graph to get full graph data
            const args = ['--json', '--graph', filePath];

            console.log(`=== CLI Execution ===`);
            console.log(`Running: ${this._cliPath} ${args.join(' ')}`);

            const process = child_process.spawn(this._cliPath, args);

            let stdout = '';
            let stderr = '';

            process.stdout.on('data', (data: Buffer) => {
                const chunk = data.toString();
                stdout += chunk;
                console.log('CLI stdout chunk:', chunk.substring(0, 100));
            });

            process.stderr.on('data', (data: Buffer) => {
                const chunk = data.toString();
                stderr += chunk;
                console.error('CLI stderr chunk:', chunk);
            });

            process.on('error', (error) => {
                console.error('CLI spawn error:', error);
                reject(new Error(`Failed to spawn CLI process: ${error.message}`));
            });

            process.on('close', (code) => {
                console.log(`=== CLI Exit ===`);
                console.log(`Exit code: ${code}`);
                console.log(`Stdout length: ${stdout.length}`);
                console.log(`Stderr length: ${stderr.length}`);

                if (stderr) {
                    console.error('CLI stderr:', stderr);
                }

                // Exit code 1 means issues were found (expected for analysis with violations)
                // Exit code 0 means no issues found
                // Exit code >= 2 means actual error
                if (code !== null && code >= 2) {
                    console.error(`CLI failed with exit code ${code}`);
                    const errorDetails = stderr || stdout.substring(0, 500) || 'No error details available';
                    reject(new Error(`Parse Error:\n${errorDetails}`));
                    return;
                }

                if (!stdout || stdout.trim().length === 0) {
                    console.error('CLI returned empty output');
                    const errorDetails = stderr || 'No error details available';
                    reject(new Error(`No graph output received.\nError: ${errorDetails}`));
                    return;
                }

                // Parse JSON output
                try {
                    const data = JSON.parse(stdout);
                    console.log('CLI succeeded, parsed JSON data');
                    resolve(data);
                } catch (parseError) {
                    console.error('Failed to parse JSON output:', parseError);
                    console.error('Raw output:', stdout.substring(0, 500));
                    reject(new Error(`Failed to parse JSON: ${parseError instanceof Error ? parseError.message : String(parseError)}`));
                }
            });
        });
    }

    private async jumpToSource(file: string, line: number) {
        try {
            const document = await vscode.workspace.openTextDocument(file);
            const editor = await vscode.window.showTextDocument(document, vscode.ViewColumn.One);
            const position = new vscode.Position(Math.max(0, line - 1), 0);
            editor.selection = new vscode.Selection(position, position);
            editor.revealRange(new vscode.Range(position, position));
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to open file: ${error}`);
        }
    }

    public dispose() {
        GraphWebview.currentPanel = undefined;

        this._panel.dispose();

        while (this._disposables.length) {
            const disposable = this._disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }

    private _getHtmlContent(): string {
        const webview = this._panel.webview;
        const nonce = getNonce();

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'nonce-${nonce}' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; style-src ${webview.cspSource} 'unsafe-inline'; img-src ${webview.cspSource} https:;">
    <title>React Component Graph</title>
    <script nonce="${nonce}" src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.28.1/cytoscape.min.js"></script>
    <script nonce="${nonce}" src="https://cdn.jsdelivr.net/npm/dagre@0.8.5/dist/dagre.min.js"></script>
    <script nonce="${nonce}" src="https://cdn.jsdelivr.net/npm/cytoscape-dagre@2.5.0/cytoscape-dagre.min.js"></script>
    <style>
        body {
            background-color: var(--vscode-editor-background);
            color: var(--vscode-editor-foreground);
            font-family: var(--vscode-font-family);
            margin: 0;
            padding: 0;
            overflow: hidden;
        }

        .toolbar {
            padding: 12px;
            background: var(--vscode-sideBar-background);
            border-bottom: 1px solid var(--vscode-panel-border);
            display: flex;
            gap: 8px;
            align-items: center;
        }

        .toolbar button {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 6px 12px;
            border-radius: 2px;
            cursor: pointer;
            font-size: 11px;
        }

        .toolbar button:hover {
            background: var(--vscode-button-hoverBackground);
        }

        #cy {
            width: 100%;
            height: calc(100vh - 48px);
            background: var(--vscode-editor-background);
        }

        .detail-panel {
            position: fixed;
            right: 0;
            top: 48px;
            width: 350px;
            height: calc(100vh - 48px);
            background: var(--vscode-sideBar-background);
            border-left: 1px solid var(--vscode-panel-border);
            padding: 16px;
            overflow-y: auto;
            transform: translateX(100%);
            transition: transform 0.3s ease;
            box-shadow: -2px 0 8px rgba(0,0,0,0.1);
        }

        .detail-panel.visible {
            transform: translateX(0);
        }

        .detail-panel h3 {
            margin-top: 0;
            color: var(--vscode-editor-foreground);
            font-size: 16px;
            font-weight: bold;
        }

        .detail-panel .close-btn {
            position: absolute;
            top: 8px;
            right: 8px;
            background: transparent;
            border: none;
            color: var(--vscode-editor-foreground);
            cursor: pointer;
            font-size: 20px;
            padding: 4px 8px;
        }

        .detail-panel .close-btn:hover {
            background: var(--vscode-list-hoverBackground);
        }

        .detail-info {
            font-size: 12px;
            margin-bottom: 16px;
        }

        .detail-info p {
            margin: 4px 0;
        }

        .detail-info strong {
            color: var(--vscode-symbolIcon-keywordForeground);
        }

        .violation-list {
            margin-top: 16px;
        }

        .violation-item {
            background: var(--vscode-inputValidation-errorBackground);
            border-left: 3px solid var(--vscode-inputValidation-errorBorder);
            padding: 8px;
            margin-bottom: 8px;
            border-radius: 2px;
        }

        .violation-item .rule {
            font-weight: bold;
            font-size: 11px;
            color: var(--vscode-errorForeground);
            margin-bottom: 4px;
        }

        .violation-item .message {
            font-size: 11px;
            line-height: 1.4;
            color: var(--vscode-editor-foreground);
        }

        .violation-item .location {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
            margin-top: 4px;
        }

        .loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100%;
            font-size: 18px;
            color: var(--vscode-descriptionForeground);
        }
    </style>
</head>
<body>
    <div class="toolbar">
        <span style="font-size: 11px; font-weight: bold;">Layout:</span>
        <select id="layout-direction" style="font-size: 11px; margin-left: 4px; padding: 2px; background: var(--vscode-input-background); color: var(--vscode-input-foreground); border: 1px solid var(--vscode-input-border);">
            <option value="LR" selected>Left→Right</option>
            <option value="TB">Top→Down</option>
        </select>
        <button id="fit-screen" title="Fit to Screen">Fit</button>
        <button id="reset-zoom" title="Reset View">Reset</button>
    </div>

    <div id="cy"></div>

    <div id="detail-panel" class="detail-panel">
        <button class="close-btn" id="close-detail">&times;</button>
        <h3 id="detail-title">Component Details</h3>
        <div id="detail-content"></div>
    </div>

    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        let cy = null;
        let graphData = null;

        // Initialize cytoscape
        function initCytoscape() {
            if (!window.cytoscape) {
                console.error('Cytoscape library not loaded');
                return;
            }

            // Register dagre extension
            if (window.cytoscapeDagre) {
                cytoscape.use(cytoscapeDagre);
            }

            cy = cytoscape({
                container: document.getElementById('cy'),
                style: [
                    {
                        selector: 'node',
                        style: {
                            'label': 'data(label)',
                            'text-valign': 'center',
                            'text-halign': 'center',
                            'font-size': '12px',
                            'color': '#ffffff',
                            'background-color': '#666',
                            'border-width': 2,
                            'border-color': '#444',
                            'width': 'label',
                            'height': 'label',
                            'padding': '12px',
                            'shape': 'roundrectangle'
                        }
                    },
                    {
                        selector: 'node.origin',
                        style: {
                            'background-color': '#059669',
                            'border-color': '#047857',
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.passthrough',
                        style: {
                            'background-color': '#d97706',
                            'border-color': '#b45309',
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.consumer',
                        style: {
                            'background-color': '#2563eb',
                            'border-color': '#1d4ed8',
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.regular',
                        style: {
                            'background-color': '#6b7280',
                            'border-color': '#4b5563',
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.state',
                        style: {
                            'background-color': '#9333ea',
                            'border-color': '#7e22ce',
                            'shape': 'ellipse',
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.context',
                        style: {
                            'background-color': '#0891b2',
                            'border-color': '#0e7490',
                            'shape': 'diamond',
                            'width': 60,
                            'height': 60,
                            'color': '#ffffff'
                        }
                    },
                    {
                        selector: 'node.memoized',
                        style: {
                            'border-width': 3,
                            'border-style': 'dashed'
                        }
                    },
                    {
                        selector: 'node.has-violations',
                        style: {
                            'border-color': '#fbbf24',
                            'border-width': 3
                        }
                    },
                    {
                        selector: 'node:selected',
                        style: {
                            'overlay-color': '#60a5fa',
                            'overlay-opacity': 0.3,
                            'overlay-padding': 8
                        }
                    },
                    {
                        selector: 'edge',
                        style: {
                            'width': 3,
                            'line-color': '#64748b',
                            'target-arrow-color': '#64748b',
                            'target-arrow-shape': 'triangle',
                            'curve-style': 'bezier',
                            'label': 'data(label)',
                            'font-size': '11px',
                            'color': '#f1f5f9',
                            'text-background-color': '#334155',
                            'text-background-opacity': 0.95,
                            'text-background-padding': '4px'
                        }
                    },
                    {
                        selector: 'edge.breaks-memo',
                        style: {
                            'line-color': '#dc2626',
                            'target-arrow-color': '#dc2626',
                            'line-style': 'solid',
                            'width': 4
                        }
                    },
                    {
                        selector: 'edge.unstable',
                        style: {
                            'line-color': '#ea580c',
                            'target-arrow-color': '#ea580c',
                            'line-style': 'dashed',
                            'width': 3
                        }
                    },
                    {
                        selector: 'edge.stable-optimized',
                        style: {
                            'line-color': '#16a34a',
                            'target-arrow-color': '#16a34a',
                            'line-style': 'solid',
                            'width': 3
                        }
                    },
                    {
                        selector: 'edge.stable-primitive',
                        style: {
                            'line-color': '#64748b',
                            'target-arrow-color': '#64748b',
                            'line-style': 'solid',
                            'width': 2.5
                        }
                    },
                    {
                        selector: 'edge.consumes',
                        style: {
                            'line-color': '#0891b2',
                            'target-arrow-color': '#0891b2',
                            'line-style': 'dotted',
                            'width': 3
                        }
                    },
                    {
                        selector: '.highlighted',
                        style: {
                            'overlay-color': '#60a5fa',
                            'overlay-opacity': 0.3,
                            'overlay-padding': 8
                        }
                    },
                    {
                        selector: 'node.dimmed',
                        style: {
                            'opacity': 0.3
                        }
                    },
                    {
                        selector: 'edge.dimmed',
                        style: {
                            'opacity': 0.2
                        }
                    }
                ],
                layout: {
                    name: 'preset'
                },
                minZoom: 0.2,
                maxZoom: 3,
                wheelSensitivity: 0.2
            });

            // Click handler for nodes
            cy.on('tap', 'node', function(evt) {
                const node = evt.target;
                showNodeDetails(node);
            });

            // Click handler for edges
            cy.on('tap', 'edge', function(evt) {
                const edge = evt.target;
                showEdgeDetails(edge);
            });

            // Close detail panel when clicking background
            cy.on('tap', function(evt) {
                if (evt.target === cy) {
                    hideDetailPanel();
                    clearHighlights();
                }
            });
        }

        function renderGraph(data) {
            if (!cy) {
                initCytoscape();
            }

            graphData = data;

            // Transform regular graph JSON to Cytoscape format
            const cytoscapeData = transformGraphToCytoscape(data.graph, data.issues);

            // Group violations by node
            const violationsByNode = {};
            if (data.issues) {
                data.issues.forEach(issue => {
                    // Match violations to nodes by file path AND line number
                    Object.entries(data.graph.componentNodes || {}).forEach(([id, node]) => {
                        // Check if violation is within the component's scope
                        // Match if same file and line is at or after component definition
                        if (node.location.filePath === issue.filePath &&
                            issue.line >= node.location.line) {
                            // For components, only add if no children or issue is before first child
                            let belongsToThisNode = true;

                            // Check if this issue actually belongs to a child component
                            if (node.children && node.children.length > 0) {
                                const childNodes = node.children.map(childId => data.graph.componentNodes[childId]).filter(c => c);
                                const isInChild = childNodes.some(child =>
                                    child.location.filePath === issue.filePath &&
                                    issue.line >= child.location.line
                                );
                                if (isInChild) {
                                    belongsToThisNode = false;
                                }
                            }

                            if (belongsToThisNode) {
                                if (!violationsByNode[id]) {
                                    violationsByNode[id] = [];
                                }
                                violationsByNode[id].push(issue);
                            }
                        }
                    });
                    Object.entries(data.graph.stateNodes || {}).forEach(([id, node]) => {
                        // For state nodes, match exact line number
                        if (node.location.filePath === issue.filePath &&
                            node.location.line === issue.line) {
                            if (!violationsByNode[id]) {
                                violationsByNode[id] = [];
                            }
                            violationsByNode[id].push(issue);
                        }
                    });
                });
            }

            // Add violation classes and badges to nodes
            const elements = {
                nodes: cytoscapeData.nodes.map(node => {
                    const classes = node.classes || '';
                    const nodeViolations = violationsByNode[node.data.id] || [];

                    let label = node.data.label;
                    if (nodeViolations.length > 0) {
                        label = '⚠ ' + label + ' (' + nodeViolations.length + ')';
                    }

                    return {
                        data: {
                            ...node.data,
                            label: label,
                            violations: nodeViolations
                        },
                        classes: nodeViolations.length > 0
                            ? classes + ' has-violations'
                            : classes
                    };
                }),
                edges: cytoscapeData.edges
            };

            cy.elements().remove();
            cy.add(elements);

            runLayout();
        }

        function transformGraphToCytoscape(graph, issues) {
            const nodes = [];
            const edges = [];

            // Helper to determine node type (origin/passthrough/consumer/regular)
            function getNodeType(node, graph) {
                const hasState = node.stateNodes && node.stateNodes.length > 0;

                // Check edges to determine passthrough vs consumer
                const nodeEdges = graph.edges.filter(e => e.sourceId === node.id || e.targetId === node.id);
                const passesProps = nodeEdges.some(e => e.sourceId === node.id && e.type === 'passes');
                const consumesState = nodeEdges.some(e => e.targetId === node.id && e.type === 'consumes');

                if (hasState) return 'origin';
                if (passesProps && !consumesState) return 'passthrough';
                if (consumesState) return 'consumer';
                return 'regular';
            }

            // Transform component nodes
            Object.entries(graph.componentNodes || {}).forEach(([id, node]) => {
                const nodeType = getNodeType(node, graph);
                const classes = [nodeType];
                if (node.isMemoized) {
                    classes.push('memoized');
                }

                nodes.push({
                    data: {
                        id: id,
                        label: node.name,
                        type: nodeType,
                        nodeType: 'component',
                        file: node.location.filePath,
                        line: node.location.line,
                        memoized: node.isMemoized
                    },
                    classes: classes.join(' ')
                });
            });

            // Transform state nodes
            Object.entries(graph.stateNodes || {}).forEach(([id, node]) => {
                // Format: name (type) - cleaner, less distinctive
                let label = node.name;
                if (node.dataType && node.dataType !== 'unknown') {
                    label += ' (' + node.dataType + ')';
                }

                // Determine classes based on state type
                let classes = 'state';
                if (node.type === 'context') {
                    classes = 'state context';
                }

                nodes.push({
                    data: {
                        id: id,
                        label: label,
                        type: 'state',
                        nodeType: 'state',
                        file: node.location.filePath,
                        line: node.location.line,
                        stateType: node.type,
                        dataType: node.dataType
                    },
                    classes: classes
                });
            });

            // Transform edges
            (graph.edges || []).forEach(edge => {
                const classes = [];
                let label = edge.propName || '';

                // Add class for consumes edges (context consumption)
                if (edge.type === 'consumes') {
                    classes.push('consumes');
                    label = 'consumes';
                }

                // Use stability data from backend to determine edge styling
                // Apply to both "passes" edges (props) and "defines" edges with propName (context provider values)
                if ((edge.type === 'passes' || edge.type === 'defines') && edge.propName) {
                    console.log('Processing', edge.type, 'edge:', edge.propName, 'isStable:', edge.isStable, 'breaksMemo:', edge.breaksMemoization, 'reason:', edge.stabilityReason);

                    // Always add data type to label in parentheses
                    const dataType = edge.propDataType || 'unknown';
                    label = label + ' (' + dataType + ')';

                    // Determine class based on stability
                    if (edge.breaksMemoization) {
                        // Red: Unstable prop breaking React.memo
                        classes.push('breaks-memo');
                        label = label + ' ⚠';
                        console.log('  -> Applied class: breaks-memo (RED)');
                    } else if (edge.isStable === false) {
                        // Orange: Unstable prop (not breaking memo)
                        classes.push('unstable');
                        console.log('  -> Applied class: unstable (ORANGE)');
                    } else if (edge.isStable === true) {
                        // Check reason for different stable types
                        if (edge.stabilityReason === 'useMemo' || edge.stabilityReason === 'useCallback') {
                            // Green: Explicitly optimized
                            classes.push('stable-optimized');
                            console.log('  -> Applied class: stable-optimized (GREEN)');
                        } else if (edge.stabilityReason === 'primitive') {
                            // Gray: Primitive (inherently stable)
                            classes.push('stable-primitive');
                            console.log('  -> Applied class: stable-primitive (GRAY)');
                        } else if (edge.stabilityReason === 'identifier') {
                            // Default: identifier could be stable
                            classes.push('stable-primitive');
                            console.log('  -> Applied class: stable-primitive (GRAY)');
                        }
                    }
                }

                edges.push({
                    data: {
                        source: edge.sourceId,
                        target: edge.targetId,
                        label: label,
                        dataType: edge.propDataType || '',
                        edgeType: edge.type,
                        isStable: edge.isStable,
                        stabilityReason: edge.stabilityReason,
                        breaksMemoization: edge.breaksMemoization,
                        propSourceVar: edge.propSourceVar || ''
                    },
                    classes: classes.join(' ')
                });
            });

            return { nodes, edges };
        }

        function runLayout() {
            const layoutDir = document.getElementById('layout-direction').value;

            cy.layout({
                name: 'dagre',
                rankDir: layoutDir,
                nodeSep: 80,
                rankSep: 100,
                padding: 50,
                animate: true,
                animationDuration: 500,
                fit: true
            }).run();
        }

        function showNodeDetails(node) {
            const data = node.data();
            const panel = document.getElementById('detail-panel');
            const title = document.getElementById('detail-title');
            const content = document.getElementById('detail-content');

            // Clear previous highlights
            clearHighlights();

            title.textContent = data.label.replace(/^⚠\\\\s*/, '').replace(/\\\\s*\\\\(\\\\d+\\\\)$/, '');

            let html = '<div class="detail-info">';
            html += '<p><strong>Type:</strong> ' + (data.nodeType || 'component') + '</p>';
            html += '<p><strong>File:</strong> ' + (data.file ? data.file.split('/').pop() : 'N/A') + '</p>';
            html += '<p><strong>Line:</strong> ' + (data.line || 'N/A') + '</p>';

            if (data.nodeType === 'state' && data.dataType) {
                html += '<p><strong>Data Type:</strong> ' + data.dataType + '</p>';
            }

            if (data.nodeType === 'state' && data.stateType) {
                html += '<p><strong>State Type:</strong> ' + data.stateType + '</p>';
            }

            if (data.memoized) {
                html += '<p><strong>Memoized:</strong> ⚡ Yes</p>';
            }

            if (data.violations && data.violations.length > 0) {
                html += '</div>';
                html += '<div class="violation-list">';
                html += '<h4 style="margin: 0 0 8px 0; color: var(--vscode-errorForeground);">⚠ Issues (' + data.violations.length + ')</h4>';

                data.violations.forEach(v => {
                    html += '<div class="violation-item">';
                    html += '<div class="rule">' + v.rule + '</div>';
                    html += '<div class="message">' + v.message + '</div>';
                    html += '<div class="location">Line ' + v.line + '</div>';
                    html += '</div>';
                });

                html += '</div>';
            } else {
                html += '</div>';
            }

            content.innerHTML = html;
            panel.classList.add('visible');

            // Add click-to-jump functionality
            if (data.file && data.line) {
                content.style.cursor = 'pointer';
                content.onclick = () => {
                    vscode.postMessage({
                        type: 'jumpToSource',
                        file: data.file,
                        line: data.line
                    });
                };
            }
        }

        function showEdgeDetails(edge) {
            const data = edge.data();
            const panel = document.getElementById('detail-panel');
            const title = document.getElementById('detail-title');
            const content = document.getElementById('detail-content');

            // Clear previous highlights and highlight related nodes
            clearHighlights();
            highlightEdgeContext(edge);

            title.textContent = 'Prop: ' + data.label;

            let html = '<div class="detail-info">';
            html += '<p><strong>Type:</strong> edge (prop passing)</p>';
            html += '<p><strong>Prop Name:</strong> ' + (data.label || 'N/A') + '</p>';

            // Always show data type
            const dataTypeDisplay = data.dataType || 'unknown';
            html += '<p><strong>Data Type:</strong> ' + dataTypeDisplay + '</p>';

            if (data.isStable !== undefined) {
                const stabilityText = data.isStable ? '✓ Stable' : '⚠ Unstable';
                const stabilityColor = data.isStable ? 'var(--vscode-testing-iconPassed)' : 'var(--vscode-errorForeground)';
                html += '<p><strong>Stability:</strong> <span style="color: ' + stabilityColor + '">' + stabilityText + '</span></p>';
            }

            if (data.stabilityReason) {
                html += '<p><strong>Reason:</strong> ' + data.stabilityReason + '</p>';
                if (data.stabilityReason === 'member-expression' && data.propSourceVar) {
                    html += '<p style="font-size: 11px; color: var(--vscode-descriptionForeground);">Extracted from: <code>' + data.propSourceVar + '.' + data.label.split(' ')[0] + '</code></p>';
                    html += '<p style="font-size: 11px; color: var(--vscode-descriptionForeground);">Upstream tracing will follow the <code>' + data.propSourceVar + '</code> prop chain.</p>';
                }
            }

            if (data.breaksMemoization) {
                html += '<p style="color: var(--vscode-errorForeground); font-weight: bold;">⚠ BREAKS MEMOIZATION</p>';
                html += '<p style="font-size: 11px; line-height: 1.4;">This unstable prop is passed to a React.memo component, which will cause unnecessary re-renders.</p>';
            }

            // Show source and target info
            const sourceNode = cy.getElementById(data.source);
            const targetNode = cy.getElementById(data.target);
            if (sourceNode && targetNode) {
                html += '<hr style="margin: 12px 0; border: none; border-top: 1px solid var(--vscode-panel-border);">';
                html += '<p><strong>From:</strong> ' + sourceNode.data('label') + '</p>';
                html += '<p><strong>To:</strong> ' + targetNode.data('label') + '</p>';
            }

            // Show chain depth information
            const propName = data.label.split(' ')[0];
            const upstreamCount = traceUpstream(sourceNode, edge, propName).nodes().length;
            const downstreamCount = traceDownstream(targetNode, edge, propName).nodes().length;
            const totalChainDepth = upstreamCount + 2 + downstreamCount; // upstream + source + target + downstream

            if (upstreamCount > 0 || downstreamCount > 0) {
                html += '<hr style="margin: 12px 0; border: none; border-top: 1px solid var(--vscode-panel-border);">';
                html += '<p><strong>Chain Depth:</strong> ' + totalChainDepth + ' components</p>';
                if (upstreamCount > 0) {
                    html += '<p style="font-size: 11px;">↑ ' + upstreamCount + ' component(s) upstream to origin</p>';
                }
                if (downstreamCount > 0) {
                    html += '<p style="font-size: 11px;">↓ ' + downstreamCount + ' component(s) downstream</p>';
                }
                html += '<p style="font-size: 11px; color: var(--vscode-descriptionForeground); margin-top: 8px;">The full chain is highlighted in the graph.</p>';
            }

            html += '</div>';

            content.innerHTML = html;
            panel.classList.add('visible');
        }

        function highlightEdgeContext(edge) {
            const highlightedElements = cy.collection();
            const propName = edge.data('label').split(' ')[0]; // Extract prop name from label (before any type info)

            // Add the clicked edge and its nodes
            highlightedElements.merge(edge);
            highlightedElements.merge(edge.source());
            highlightedElements.merge(edge.target());

            // Trace upstream: Find the path from source back to the origin (state node)
            const upstream = traceUpstream(edge.source(), edge, propName);
            highlightedElements.merge(upstream);

            // Trace downstream: Find all paths from target to consumers
            const downstream = traceDownstream(edge.target(), edge, propName);
            highlightedElements.merge(downstream);

            // Highlight the collected elements
            highlightedElements.addClass('highlighted');

            // Dim everything else
            cy.elements().not(highlightedElements).addClass('dimmed');
        }

        function traceUpstream(node, originEdge, propName) {
            const collection = cy.collection();

            // If the origin edge has a propSourceVar (member expression), look for that instead
            const originData = originEdge.data();
            const lookupProp = originData.propSourceVar || propName;

            // Get incoming edges (edges pointing TO this node)
            const incomingEdges = node.incomers('edge[edgeType="passes"], edge[edgeType="defines"]');

            incomingEdges.forEach(edge => {
                // Don't trace back through the edge we came from
                if (edge.id() === originEdge.id()) {
                    return;
                }

                // For "passes" edges, check if it matches our lookup prop
                if (edge.data('edgeType') === 'passes') {
                    const edgePropName = edge.data('label').split(' ')[0];
                    if (edgePropName !== lookupProp) {
                        return; // Skip this edge, it's for a different prop
                    }

                    // Continue tracing with the matched prop name
                    collection.merge(edge);
                    const sourceNode = edge.source();
                    collection.merge(sourceNode);
                    collection.merge(traceUpstream(sourceNode, edge, edgePropName));
                }
                else if (edge.data('edgeType') === 'defines') {
                    // For "defines" edges, always follow (these connect to state nodes)
                    collection.merge(edge);
                    const sourceNode = edge.source();
                    collection.merge(sourceNode);
                    collection.merge(traceUpstream(sourceNode, edge, propName));
                }
            });

            return collection;
        }

        function traceDownstream(node, originEdge, propName) {
            const collection = cy.collection();

            // Get outgoing edges (edges starting FROM this node)
            const outgoingEdges = node.outgoers('edge[edgeType="passes"], edge[edgeType="consumes"]');

            outgoingEdges.forEach(edge => {
                // Don't trace back through the edge we came from
                if (edge.id() === originEdge.id()) {
                    return;
                }

                // For "passes" edges, only follow if it's the same prop
                if (edge.data('edgeType') === 'passes') {
                    const edgePropName = edge.data('label').split(' ')[0];
                    if (edgePropName !== propName) {
                        return; // Skip this edge, it's for a different prop
                    }
                }

                // For "consumes" edges, always follow (these connect to state consumers)

                // Add this edge and its target
                collection.merge(edge);
                const targetNode = edge.target();
                collection.merge(targetNode);

                // Recursively trace downstream from the target
                collection.merge(traceDownstream(targetNode, edge, propName));
            });

            return collection;
        }

        function clearHighlights() {
            cy.elements().removeClass('highlighted dimmed');
        }

        function hideDetailPanel() {
            document.getElementById('detail-panel').classList.remove('visible');
        }

        // Event handlers
        document.getElementById('layout-direction').addEventListener('change', runLayout);
        document.getElementById('fit-screen').addEventListener('click', () => cy.fit(50));
        document.getElementById('reset-zoom').addEventListener('click', () => cy.reset());
        document.getElementById('close-detail').addEventListener('click', hideDetailPanel);

        // Handle messages from extension
        window.addEventListener('message', event => {
            const message = event.data;

            switch (message.type) {
                case 'renderGraph':
                    console.log('Rendering graph with data:', message.data);
                    renderGraph(message.data);
                    break;
                case 'showError':
                    document.getElementById('cy').innerHTML =
                        '<div class="loading" style="color: var(--vscode-errorForeground);">Error: ' +
                        message.error + '</div>';
                    break;
            }
        });

        // Initialize
        initCytoscape();
    </script>
</body>
</html>`;
    }
}

function getNonce() {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    return text;
}
