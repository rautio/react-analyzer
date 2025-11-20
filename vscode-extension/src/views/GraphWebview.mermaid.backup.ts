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
        } else {
            // Use bundled binary from parent directory during development
            this._cliPath = path.join(path.dirname(extensionUri.fsPath), 'react-analyzer');
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

            // Run CLI to get Mermaid output
            const mermaidOutput = await this.getMermaidDiagram(filePath);

            console.log('Mermaid output length:', mermaidOutput.length);
            console.log('Mermaid output preview:', mermaidOutput.substring(0, 200));
            console.log('Mermaid output (full):', mermaidOutput);

            // Check if output is actually Mermaid
            if (!mermaidOutput.includes('flowchart')) {
                console.error('ERROR: Output does not contain flowchart syntax');
                console.error('Full output:', mermaidOutput);
                throw new Error('CLI did not return valid Mermaid diagram. Output: ' + mermaidOutput.substring(0, 200));
            }

            // Parse metadata from Mermaid comments
            const metadata = this.parseMermaidMetadata(mermaidOutput);

            console.log('Parsed metadata - nodes:', Object.keys(metadata.nodes || {}).length);
            console.log('Parsed metadata - edges:', Object.keys(metadata.edges || {}).length);
            console.log('Parsed metadata - violations:', (metadata.violations || []).length);
            console.log('Full metadata:', metadata);

            // Send to webview
            this._panel.webview.postMessage({
                type: 'renderGraph',
                mermaid: mermaidOutput,
                metadata: metadata
            });
        } catch (error) {
            console.error('=== Update graph error ===');
            console.error('Error type:', error instanceof Error ? error.constructor.name : typeof error);
            console.error('Error message:', error instanceof Error ? error.message : String(error));
            console.error('Full error:', error);

            const errorMsg = error instanceof Error ? error.message : String(error);
            // Show error in both notification and output channel for better visibility
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

    private async getMermaidDiagram(filePath: string): Promise<string> {
        return new Promise((resolve, reject) => {
            // For large directories, just analyze the single file to avoid Mermaid size limits
            // This may cause some cross-file references to be missing, but prevents "text size exceeded" errors
            // NOTE: Flag must come BEFORE the file path for Go's flag package
            const args = ['--mermaid', filePath];

            console.log(`=== CLI Execution ===`);
            console.log(`Running: ${this._cliPath} ${args.join(' ')}`);
            console.log(`Note: Analyzing single file to avoid Mermaid size limits`);

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

                if (code !== null && code !== 0) {
                    console.error(`CLI failed with exit code ${code}`);
                    // Extract the most relevant error message from stderr
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

                console.log('CLI succeeded, returning output');
                resolve(stdout);
            });
        });
    }

    private parseMermaidMetadata(mermaidDiagram: string): Record<string, any> {
        const metadata: Record<string, any> = {
            nodes: {},
            edges: {},
            violations: []
        };

        // Parse node metadata from comments like:
        // %% meta:nodeId|file:path|line:10|type:origin|memoized:true|nodetype:component
        // %% meta:nodeId|file:path|line:10|type:state|memoized:false|nodetype:state|statetype:useState|datatype:primitive
        const metaRegex = /%%\s*meta:([^|\n]+)\|file:([^|\n]+)\|line:(\d+)\|([^\n]+)/g;

        let match;
        while ((match = metaRegex.exec(mermaidDiagram)) !== null) {
            const [, nodeId, file, line, rest] = match;

            // Parse remaining key:value pairs
            const attrs: Record<string, string> = {};
            const pairs = rest.split('|');
            for (const pair of pairs) {
                const [key, value] = pair.split(':');
                if (key && value) {
                    attrs[key.trim()] = value.trim();
                }
            }

            metadata.nodes[nodeId.trim()] = {
                file: file,
                line: parseInt(line),
                type: attrs.type || 'unknown',
                memoized: attrs.memoized === 'true',
                nodeType: attrs.nodetype || 'component',
                stateType: attrs.statetype,
                dataType: attrs.datatype
            };
        }

        // Parse edge metadata from comments like:
        // %% edge:Parent-Child|prop:config|datatype:object
        const edgeRegex = /%%\s*edge:([^|\n]+)\|([^\n]+)/g;
        while ((match = edgeRegex.exec(mermaidDiagram)) !== null) {
            const [, edgeId, rest] = match;

            // Parse remaining key:value pairs
            const attrs: Record<string, string> = {};
            const pairs = rest.split('|');
            for (const pair of pairs) {
                const [key, value] = pair.split(':');
                if (key && value) {
                    attrs[key.trim()] = value.trim();
                }
            }

            metadata.edges[edgeId.trim()] = {
                propName: attrs.prop || '',
                dataType: attrs.datatype || 'unknown'
            };
        }

        // Parse violation metadata from comments like:
        // %% violation|rule:deep-prop-drilling,severity:warning,line:8,file:App.tsx,message:...
        const violationRegex = /%%\s*violation\|([^\n]+)/g;
        while ((match = violationRegex.exec(mermaidDiagram)) !== null) {
            const [, rest] = match;

            // Parse key:value pairs (handle commas in message carefully)
            const attrs: Record<string, string> = {};

            // Split by comma, but be careful with message field which may contain commas
            const parts = rest.split(',');
            let currentKey = '';
            let currentValue = '';

            for (let i = 0; i < parts.length; i++) {
                const part = parts[i];
                const colonIndex = part.indexOf(':');

                if (colonIndex > 0 && ['rule', 'severity', 'line', 'file', 'message'].includes(part.substring(0, colonIndex))) {
                    // This is a new key:value pair
                    if (currentKey) {
                        attrs[currentKey] = currentValue;
                    }
                    currentKey = part.substring(0, colonIndex);
                    currentValue = part.substring(colonIndex + 1);
                } else {
                    // This is a continuation of the previous value (part of message)
                    if (currentValue) {
                        currentValue += ',' + part;
                    }
                }
            }

            // Don't forget the last key:value pair
            if (currentKey) {
                attrs[currentKey] = currentValue;
            }

            metadata.violations.push({
                rule: attrs.rule || '',
                severity: attrs.severity || 'warning',
                line: parseInt(attrs.line) || 0,
                file: attrs.file || '',
                message: (attrs.message || '').replace(/\\|/g, '|') // Unescape pipes
            });
        }

        return metadata;
    }

    private async jumpToSource(filePath: string, line: number) {
        try {
            const uri = vscode.Uri.file(filePath);
            const document = await vscode.workspace.openTextDocument(uri);
            const editor = await vscode.window.showTextDocument(document, vscode.ViewColumn.One);

            // Jump to the line
            const position = new vscode.Position(Math.max(0, line - 1), 0);
            editor.selection = new vscode.Selection(position, position);
            editor.revealRange(new vscode.Range(position, position), vscode.TextEditorRevealType.InCenter);
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to open file: ${error}`);
        }
    }

    private _getHtmlContent(): string {
        const webview = this._panel.webview;

        // Use a nonce to only allow specific scripts to run
        const nonce = getNonce();

        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'nonce-${nonce}' https://cdn.jsdelivr.net; style-src ${webview.cspSource} 'unsafe-inline' https://cdn.jsdelivr.net; img-src ${webview.cspSource} https:;">
    <title>React Component Graph</title>
    <script nonce="${nonce}" src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
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

        .toolbar input {
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            padding: 4px 8px;
            border-radius: 2px;
            flex: 1;
            max-width: 300px;
        }

        .toolbar button {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 6px 12px;
            border-radius: 2px;
            cursor: pointer;
        }

        .toolbar button:hover {
            background: var(--vscode-button-hoverBackground);
        }

        .graph-container {
            width: 100%;
            height: calc(100vh - 48px);
            overflow: hidden;
            position: relative;
            background: var(--vscode-editor-background);
        }

        #mermaid-graph {
            transform-origin: 0 0;
            transition: transform 0.1s ease-out;
            cursor: grab;
        }

        #mermaid-graph.panning {
            cursor: grabbing;
        }

        #mermaid-graph svg {
            display: block;
        }

        /* Custom node styling */
        .mermaid .node.state-origin rect {
            fill: #10b981 !important;
            stroke: #047857 !important;
            stroke-width: 2px;
        }

        .mermaid .node.passthrough rect {
            fill: #fbbf24 !important;
            stroke: #f59e0b !important;
        }

        .mermaid .node.consumer rect {
            fill: #3b82f6 !important;
            stroke: #2563eb !important;
        }

        .mermaid .node.memoized rect {
            stroke: #a855f7 !important;
            stroke-width: 3px !important;
            stroke-dasharray: 5, 5;
        }

        /* Hover effects */
        .mermaid .node:hover rect {
            filter: brightness(1.2) drop-shadow(0 0 8px currentColor);
            cursor: pointer;
            transition: all 0.2s ease;
        }

        /* Violation indicators */
        .mermaid .node.has-violations rect,
        .mermaid .node.has-violations circle,
        .mermaid .node.has-violations polygon {
            stroke: #fbbf24 !important;
            stroke-width: 3px !important;
            filter: drop-shadow(0 0 6px #fbbf24);
        }

        .mermaid .node.has-critical rect,
        .mermaid .node.has-critical circle,
        .mermaid .node.has-critical polygon {
            stroke: #ef4444 !important;
            stroke-width: 3px !important;
            filter: drop-shadow(0 0 8px #ef4444);
        }

        /* Violation badge */
        .mermaid .violation-badge {
            pointer-events: none;
        }

        /* Unstable prop edges */
        .mermaid .edgePath.unstable-prop path {
            stroke: #ef4444 !important;
            stroke-width: 2.5px !important;
            stroke-dasharray: 5, 5;
        }

        .mermaid .edgePath.unstable-prop .edgeLabel {
            background: #fee2e2 !important;
            color: #991b1b !important;
            border: 1px solid #ef4444 !important;
            border-radius: 3px;
            padding: 2px 4px;
        }

        .detail-panel {
            position: fixed;
            right: 0;
            top: 48px;
            width: 300px;
            height: calc(100vh - 48px);
            background: var(--vscode-sideBar-background);
            border-left: 1px solid var(--vscode-panel-border);
            padding: 16px;
            overflow-y: auto;
            transform: translateX(100%);
            transition: transform 0.3s ease;
        }

        .detail-panel.visible {
            transform: translateX(0);
        }

        .detail-panel h3 {
            margin-top: 0;
            color: var(--vscode-editor-foreground);
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
        <select id="layout-direction" style="font-size: 11px; margin-left: 4px; padding: 2px;">
            <option value="LR" selected>Left→Right (LR)</option>
            <option value="TD">Top↓Down (TD)</option>
        </select>
        <label style="font-size: 11px; margin-left: 12px;">
            <input type="checkbox" id="show-edges" checked /> Show Edges
        </label>
        <button id="zoom-in" title="Zoom In" style="margin-left: 12px;">➕</button>
        <button id="zoom-out" title="Zoom Out">➖</button>
        <button id="fit-screen" title="Fit to Screen">⛶</button>
        <button id="reset-zoom" title="Reset View">↺</button>
    </div>

    <div class="graph-container">
        <div id="mermaid-graph" class="mermaid">
            <div class="loading">Loading graph...</div>
        </div>
    </div>

    <div id="detail-panel" class="detail-panel">
        <button class="close-btn" id="close-detail">&times;</button>
        <h3 id="detail-title">Component Details</h3>
        <div id="detail-content"></div>
    </div>

    <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        let metadata = {};
        let originalMermaidSyntax = '';  // Store original syntax for layout switching

        // Initialize Mermaid
        mermaid.initialize({
            startOnLoad: false,
            theme: 'base',
            themeVariables: {
                primaryColor: '#1e293b',
                primaryTextColor: '#e2e8f0',
                primaryBorderColor: '#475569',
                lineColor: '#64748b'
            }
        });

        // Handle messages from extension
        window.addEventListener('message', async event => {
            const message = event.data;
            console.log('Received message:', message.type);

            switch (message.type) {
                case 'renderGraph':
                    console.log('Rendering graph, mermaid length:', message.mermaid?.length);
                    console.log('Metadata keys:', Object.keys(message.metadata || {}).length);
                    await renderGraph(message.mermaid, message.metadata);
                    break;
                case 'showError':
                    console.error('Graph generation error:', message.error);
                    showError(message.error);
                    break;
            }
        });

        function showError(errorMessage) {
            const container = document.getElementById('mermaid-graph');
            container.innerHTML = '<div style="padding: 20px; color: #ef4444; background: #fef2f2; border: 1px solid #fca5a5; border-radius: 4px; white-space: pre-wrap; font-family: monospace; font-size: 12px;">' +
                '<strong style="display: block; margin-bottom: 10px; font-size: 14px;">Error generating graph:</strong>' +
                errorMessage +
                '</div>';
        }

        async function renderGraph(mermaidSyntax, meta) {
            metadata = meta;
            originalMermaidSyntax = mermaidSyntax;  // Store for layout switching
            const container = document.getElementById('mermaid-graph');

            try {
                // Render Mermaid using the v10 API
                const { svg } = await mermaid.render('mermaid-diagram', mermaidSyntax);
                container.innerHTML = svg;

                // Enhance with custom interactions
                enhanceGraph();
            } catch (error) {
                console.error('Mermaid rendering error:', error);
                container.innerHTML = '<div class="loading">Failed to render graph: ' + error.message + '</div>';
                vscode.postMessage({ type: 'error', message: 'Failed to render graph: ' + error.message });
            }
        }

        async function changeLayout(direction) {
            if (!originalMermaidSyntax) return;

            // Replace the flowchart direction in the syntax
            let newSyntax = originalMermaidSyntax.replace(/^flowchart (TD|LR)/m, 'flowchart ' + direction);

            // Re-render with new layout
            await renderGraph(newSyntax, metadata);
        }

        function enhanceGraph() {
            const container = document.getElementById('mermaid-graph');
            const svg = container.querySelector('svg');
            if (!svg) {
                console.error('SVG not found in container');
                return;
            }

            console.log('Enhancing graph with metadata:', metadata);

            // Group violations by node (file + line matching)
            const violationsByNode = groupViolationsByNode();

            // Enhance nodes with violations and metadata
            svg.querySelectorAll('.node').forEach(node => {
                const nodeId = node.getAttribute('id');

                // Mermaid adds prefixes like "flowchart-" and suffixes like "-0", "-1", etc.
                // Strip these to match our metadata keys
                let cleanId = nodeId || '';

                // Remove "flowchart-" prefix
                if (cleanId.startsWith('flowchart-')) {
                    cleanId = cleanId.substring('flowchart-'.length);
                }

                // Remove trailing "-{digit}" suffix that Mermaid adds
                // Use a more specific pattern to only remove single trailing digits
                const match = cleanId.match(/^(.+)-(\d+)$/);
                if (match) {
                    cleanId = match[1];
                }

                console.log('Node ID mapping:', nodeId, '->', cleanId);

                const meta = metadata.nodes && metadata.nodes[cleanId];

                if (!meta) {
                    console.log('No metadata for node:', nodeId, '(cleaned:', cleanId + ')');
                    return;
                }

                // Add CSS class for styling and filtering
                node.classList.add(meta.type);
                if (meta.memoized) node.classList.add('memoized');

                // Add node type class for filtering
                if (meta.nodeType === 'state') {
                    node.classList.add('filter-state-node');
                } else if (meta.type === 'passthrough') {
                    node.classList.add('filter-passthrough-node');
                } else if (meta.type === 'regular') {
                    node.classList.add('filter-regular-node');
                }

                // Check if this node has violations
                const nodeViolations = violationsByNode[nodeId] || [];
                if (nodeViolations.length > 0) {
                    // Add visual indicator for violations
                    node.classList.add('has-violations');

                    // Count violations by severity
                    const criticalCount = nodeViolations.filter(v => v.severity === 'error').length;
                    const warningCount = nodeViolations.filter(v => v.severity === 'warning').length;

                    if (criticalCount > 0) {
                        node.classList.add('has-critical');
                    } else if (warningCount > 0) {
                        node.classList.add('has-warnings');
                    }

                    // Add badge with violation count
                    addViolationBadge(node, nodeViolations.length);
                }

                // Store metadata for search and click handling
                node.setAttribute('data-name', nodeId);
                node.setAttribute('data-nodetype', meta.nodeType || 'component');
                node.setAttribute('data-type', meta.type);
                node.setAttribute('data-file', meta.file);
                node.setAttribute('data-line', meta.line);

                // Add click handler to show details
                node.style.cursor = 'pointer';
                node.addEventListener('click', (e) => {
                    showNodeDetails(cleanId, meta, nodeViolations);
                    e.stopPropagation();
                });
            });

            // Enhance edges with prop type labels
            enhanceEdges(svg);

            // Enable zoom/pan
            setupZoomPan();

            // Log summary
            console.log('Enhanced graph:',
                svg.querySelectorAll('.node').length, 'nodes,',
                metadata.violations ? metadata.violations.length : 0, 'violations');
        }

        function groupViolationsByNode() {
            const violationsByNode = {};

            if (!metadata.violations || !metadata.nodes) {
                return violationsByNode;
            }

            console.log('Grouping violations:', metadata.violations.length, 'violations across', Object.keys(metadata.nodes).length, 'nodes');

            // Match violations to nodes by file path
            for (const violation of metadata.violations) {
                for (const [nodeId, nodeMeta] of Object.entries(metadata.nodes)) {
                    if (nodeMeta.file && nodeMeta.file.endsWith(violation.file)) {
                        if (!violationsByNode[nodeId]) {
                            violationsByNode[nodeId] = [];
                        }
                        violationsByNode[nodeId].push(violation);
                    }
                }
            }

            console.log('Violations grouped:', Object.keys(violationsByNode).map(k => k + ':' + violationsByNode[k].length).join(', '));
            return violationsByNode;
        }

        function addViolationBadge(node, count) {
            // Find the rect or shape element
            const shape = node.querySelector('rect, circle, polygon');
            if (!shape) return;

            // Get bounding box
            const bbox = shape.getBBox();

            // Create badge group
            const badge = document.createElementNS('http://www.w3.org/2000/svg', 'g');
            badge.setAttribute('class', 'violation-badge');

            // Badge background circle
            const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
            circle.setAttribute('cx', bbox.x + bbox.width + 8);
            circle.setAttribute('cy', bbox.y - 8);
            circle.setAttribute('r', '12');
            circle.setAttribute('fill', '#ef4444');
            circle.setAttribute('stroke', '#fff');
            circle.setAttribute('stroke-width', '2');

            // Badge text
            const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            text.setAttribute('x', bbox.x + bbox.width + 8);
            text.setAttribute('y', bbox.y - 8);
            text.setAttribute('text-anchor', 'middle');
            text.setAttribute('dominant-baseline', 'central');
            text.setAttribute('fill', '#fff');
            text.setAttribute('font-size', '10');
            text.setAttribute('font-weight', 'bold');
            text.textContent = count;

            badge.appendChild(circle);
            badge.appendChild(text);
            node.appendChild(badge);
        }

        function enhanceEdges(svg) {
            if (!metadata.edges) return;

            svg.querySelectorAll('.edgePath').forEach(edge => {
                const edgeId = edge.getAttribute('id');
                if (!edgeId) return;

                // Extract from and to node IDs from Mermaid's edge ID format
                // Format is typically "L-{fromId}-{toId}-0" or similar
                const match = edgeId.match(/L-([^-]+)-([^-]+)/);
                if (!match) return;

                const [, fromId, toId] = match;
                const edgeKey = fromId + '-' + toId;
                const edgeMeta = metadata.edges[edgeKey];

                if (!edgeMeta) return;

                // Add data type label to edge if it's not unknown
                if (edgeMeta.dataType && edgeMeta.dataType !== 'unknown') {
                    addEdgeLabel(edge, edgeMeta.propName, edgeMeta.dataType);

                    // Add class for unstable types (object, array, function)
                    if (['object', 'array', 'function'].includes(edgeMeta.dataType)) {
                        edge.classList.add('unstable-prop');
                    }
                }
            });
        }

        function addEdgeLabel(edge, propName, dataType) {
            // Find the edge label element (Mermaid creates these)
            const edgeLabel = edge.querySelector('.edgeLabel');
            if (edgeLabel) {
                // Append data type to existing label
                const span = document.createElement('span');
                span.style.fontSize = '10px';
                span.style.color = '#ef4444';
                span.style.fontWeight = 'bold';
                span.textContent = ' (' + dataType + ')';
                edgeLabel.appendChild(span);
            }
        }

        function showNodeDetails(nodeId, nodeMeta, violations) {
            const panel = document.getElementById('detail-panel');
            const title = document.getElementById('detail-title');
            const content = document.getElementById('detail-content');

            // Set title
            title.textContent = nodeId.replace(/_/g, ' ').replace(/component |state /gi, '');

            // Build content HTML
            let html = '<div style="font-size: 12px;">';

            // Basic info
            html += '<p><strong>File:</strong> ' + nodeMeta.file.split('/').pop() + '</p>';
            html += '<p><strong>Line:</strong> ' + nodeMeta.line + '</p>';
            html += '<p><strong>Type:</strong> ' + (nodeMeta.nodeType || 'component') + '</p>';

            if (nodeMeta.memoized) {
                html += '<p><strong>Memoized:</strong> ⚡ Yes</p>';
            }

            // Violations section
            if (violations && violations.length > 0) {
                html += '<hr style="margin: 16px 0; border: none; border-top: 1px solid var(--vscode-panel-border);">';
                html += '<h4 style="margin: 8px 0; color: #ef4444;">⚠️ Issues (' + violations.length + ')</h4>';

                violations.forEach((violation, idx) => {
                    html += '<div style="margin: 12px 0; padding: 8px; background: var(--vscode-input-background); border-left: 3px solid #ef4444; border-radius: 2px;">';
                    html += '<div style="font-weight: bold; margin-bottom: 4px;">' + violation.rule + '</div>';
                    html += '<div style="font-size: 11px; line-height: 1.4;">' + violation.message + '</div>';
                    html += '<div style="margin-top: 4px; font-size: 10px; color: var(--vscode-descriptionForeground);">Line ' + violation.line + '</div>';
                    html += '</div>';
                });
            }

            html += '<hr style="margin: 16px 0; border: none; border-top: 1px solid var(--vscode-panel-border);">';
            html += '<p style="font-size: 10px; color: var(--vscode-descriptionForeground);"><strong>Full Path:</strong><br/>' + nodeMeta.file + '</p>';
            html += '</div>';

            content.innerHTML = html;
            panel.classList.add('visible');
        }

        // Zoom and pan functionality
        let currentZoom = 1;
        let panX = 0;
        let panY = 0;
        let isPanning = false;
        let startX = 0;
        let startY = 0;

        function setupZoomPan() {
            const container = document.getElementById('mermaid-graph');
            const graphContainer = document.querySelector('.graph-container');

            // Mouse wheel zoom
            graphContainer.addEventListener('wheel', (e) => {
                e.preventDefault();

                const delta = e.deltaY > 0 ? 0.9 : 1.1;
                const newZoom = Math.min(Math.max(0.1, currentZoom * delta), 5);

                // Zoom towards mouse position
                const rect = graphContainer.getBoundingClientRect();
                const mouseX = e.clientX - rect.left;
                const mouseY = e.clientY - rect.top;

                // Adjust pan to zoom towards mouse
                panX = mouseX - (mouseX - panX) * (newZoom / currentZoom);
                panY = mouseY - (mouseY - panY) * (newZoom / currentZoom);

                currentZoom = newZoom;
                updateTransform();
            });

            // Pan with mouse drag
            graphContainer.addEventListener('mousedown', (e) => {
                // Only pan if clicking on background or SVG, not on nodes
                const target = e.target;
                if (target.classList.contains('graph-container') ||
                    target.tagName === 'svg' ||
                    target.tagName === 'g' ||
                    target.classList.contains('mermaid')) {
                    isPanning = true;
                    startX = e.clientX - panX;
                    startY = e.clientY - panY;
                    container.classList.add('panning');
                    e.preventDefault();
                }
            });

            graphContainer.addEventListener('mousemove', (e) => {
                if (isPanning) {
                    panX = e.clientX - startX;
                    panY = e.clientY - startY;
                    updateTransform();
                }
            });

            graphContainer.addEventListener('mouseup', () => {
                isPanning = false;
                container.classList.remove('panning');
            });

            graphContainer.addEventListener('mouseleave', () => {
                isPanning = false;
                container.classList.remove('panning');
            });
        }

        function updateTransform() {
            const container = document.getElementById('mermaid-graph');
            container.style.transform = 'translate(' + panX + 'px, ' + panY + 'px) scale(' + currentZoom + ')';
        }

        function zoomIn() {
            currentZoom = Math.min(currentZoom * 1.2, 5);
            updateTransform();
        }

        function zoomOut() {
            currentZoom = Math.max(currentZoom / 1.2, 0.1);
            updateTransform();
        }

        function resetZoom() {
            currentZoom = 1;
            panX = 0;
            panY = 0;
            updateTransform();
        }

        function fitToScreen() {
            const container = document.getElementById('mermaid-graph');
            const graphContainer = document.querySelector('.graph-container');
            const svg = container.querySelector('svg');

            if (!svg) return;

            const svgRect = svg.getBBox();
            const containerRect = graphContainer.getBoundingClientRect();

            // Calculate zoom to fit
            const scaleX = (containerRect.width - 40) / svgRect.width;
            const scaleY = (containerRect.height - 40) / svgRect.height;
            currentZoom = Math.min(scaleX, scaleY, 1); // Don't zoom in beyond 100%

            // Center the graph
            panX = (containerRect.width - svgRect.width * currentZoom) / 2;
            panY = (containerRect.height - svgRect.height * currentZoom) / 2;

            updateTransform();
        }

        function showDetailPanel(nodeId, meta) {
            const panel = document.getElementById('detail-panel');
            const title = document.getElementById('detail-title');
            const content = document.getElementById('detail-content');

            title.textContent = nodeId.replace(/_/g, ' ').replace(/component /g, '').replace(/state /g, '');

            let detailsHtml = \`
                <p><strong>File:</strong> \${meta.file.split('/').pop()}</p>
                <p><strong>Line:</strong> \${meta.line}</p>
            \`;

            if (meta.nodeType === 'state') {
                // State node details
                detailsHtml += \`
                    <p><strong>Node Type:</strong> State</p>
                    <p><strong>State Type:</strong> \${meta.stateType}</p>
                    <p><strong>Data Type:</strong> \${meta.dataType}</p>
                \`;
            } else {
                // Component node details
                detailsHtml += \`
                    <p><strong>Node Type:</strong> Component</p>
                    <p><strong>Role:</strong> \${meta.type}</p>
                    <p><strong>Memoized:</strong> \${meta.memoized ? '⚡ Yes' : 'No'}</p>
                \`;
            }

            detailsHtml += \`
                <p><strong>Full Path:</strong><br/><small>\${meta.file}</small></p>
            \`;

            content.innerHTML = detailsHtml;
            panel.classList.add('visible');
        }

        function hideDetailPanel() {
            document.getElementById('detail-panel').classList.remove('visible');
        }

        // Toolbar interactions
        document.getElementById('zoom-in').addEventListener('click', () => {
            zoomIn();
        });

        document.getElementById('zoom-out').addEventListener('click', () => {
            zoomOut();
        });

        document.getElementById('fit-screen').addEventListener('click', () => {
            fitToScreen();
        });

        document.getElementById('reset-zoom').addEventListener('click', () => {
            resetZoom();
        });

        document.getElementById('close-detail').addEventListener('click', () => {
            hideDetailPanel();
        });

        // Edge visibility toggle
        document.getElementById('show-edges').addEventListener('change', (e) => {
            const showEdges = e.target.checked;
            document.querySelectorAll('.edgePath, .edgeLabel').forEach(el => {
                el.style.display = showEdges ? '' : 'none';
            });
        });

        // Layout direction toggle
        document.getElementById('layout-direction').addEventListener('change', async (e) => {
            const direction = e.target.value;
            console.log('Changing layout to:', direction);
            await changeLayout(direction);
        });
    </script>
</body>
</html>`;
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
}

function getNonce() {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    return text;
}
