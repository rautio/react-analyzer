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

            console.log('Parsed metadata:', Object.keys(metadata).length, 'nodes');
            console.log('Metadata:', metadata);

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
            vscode.window.showErrorMessage(`Failed to generate graph: ${errorMsg}`);
        }
    }

    private async getMermaidDiagram(filePath: string): Promise<string> {
        return new Promise((resolve, reject) => {
            // For large directories, just analyze the single file to avoid Mermaid size limits
            // This may cause some cross-file references to be missing, but prevents "text size exceeded" errors
            const args = ['-mermaid', filePath];

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
                    reject(new Error(`CLI exited with code ${code}. stderr: ${stderr || 'none'}. stdout: ${stdout.substring(0, 200)}`));
                    return;
                }

                if (!stdout || stdout.trim().length === 0) {
                    console.error('CLI returned empty output');
                    reject(new Error(`CLI returned no output. stderr: ${stderr || 'none'}`));
                    return;
                }

                console.log('CLI succeeded, returning output');
                resolve(stdout);
            });
        });
    }

    private parseMermaidMetadata(mermaidDiagram: string): Record<string, any> {
        const metadata: Record<string, any> = {};

        // Parse metadata from comments like:
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

            metadata[nodeId.trim()] = {
                file: file,
                line: parseInt(line),
                type: attrs.type || 'unknown',
                memoized: attrs.memoized === 'true',
                nodeType: attrs.nodetype || 'component',
                stateType: attrs.statetype,
                dataType: attrs.datatype
            };
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

        /* Highlighted state */
        .mermaid .node.highlighted rect {
            stroke: #ef4444 !important;
            stroke-width: 4px !important;
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
        <input id="search" type="text" placeholder="Search components..." />
        <span style="margin-left: 8px; font-size: 11px; font-weight: bold;">Layout:</span>
        <select id="layout-direction" style="font-size: 11px; margin-left: 4px; padding: 2px;">
            <option value="LR" selected>Left→Right (LR)</option>
            <option value="TD">Top↓Down (TD)</option>
        </select>
        <span style="margin-left: 12px; font-size: 11px; font-weight: bold;">Show:</span>
        <label style="font-size: 11px; margin-left: 4px;">
            <input type="checkbox" id="filter-state" checked /> State
        </label>
        <label style="font-size: 11px;">
            <input type="checkbox" id="filter-passthrough" checked /> Passthrough
        </label>
        <label style="font-size: 11px;">
            <input type="checkbox" id="filter-regular" checked /> Regular
        </label>
        <label style="font-size: 11px; margin-left: 8px;">
            <input type="checkbox" id="show-edges" checked /> Edges
        </label>
        <button id="zoom-in" title="Zoom In">➕</button>
        <button id="zoom-out" title="Zoom Out">➖</button>
        <button id="fit-screen" title="Fit to Screen">⛶</button>
        <button id="reset-zoom" title="Reset View">↺</button>
        <button id="reset-highlight" title="Clear Highlights">Clear</button>
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
        let currentHighlights = [];
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
            }
        });

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

            // Add click handlers to all nodes
            svg.querySelectorAll('.node').forEach(node => {
                const nodeId = node.getAttribute('id');
                const meta = metadata[nodeId];

                if (!meta) return;

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

                // Click handler
                node.addEventListener('click', () => {
                    // Jump to source
                    vscode.postMessage({
                        type: 'jumpToSource',
                        file: meta.file,
                        line: meta.line
                    });

                    // Show detail panel
                    showDetailPanel(nodeId, meta);

                    // Highlight connected nodes
                    highlightConnectedNodes(nodeId);
                });

                // Store metadata for search
                node.setAttribute('data-name', nodeId);
                node.setAttribute('data-nodetype', meta.nodeType || 'component');
                node.setAttribute('data-type', meta.type);
            });

            // Enable zoom/pan
            setupZoomPan();
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

        function highlightConnectedNodes(nodeId) {
            // Clear previous highlights
            clearHighlights();

            const svg = document.querySelector('.mermaid svg');
            if (!svg) return;

            // Find edges connected to this node
            svg.querySelectorAll('.edgePath').forEach(edge => {
                const edgeId = edge.getAttribute('id');
                if (edgeId && edgeId.includes(nodeId)) {
                    // Extract connected node ID
                    const match = edgeId.match(/flowchart-([^-]+)-([^-]+)/);
                    if (match) {
                        const [, from, to] = match;
                        const connectedId = from === nodeId ? to : from;
                        const connectedNode = document.getElementById(connectedId);
                        if (connectedNode) {
                            connectedNode.classList.add('highlighted');
                            currentHighlights.push(connectedNode);
                        }
                    }
                }
            });
        }

        function clearHighlights() {
            currentHighlights.forEach(node => node.classList.remove('highlighted'));
            currentHighlights = [];
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
        document.getElementById('search').addEventListener('input', (e) => {
            const query = e.target.value.toLowerCase();
            document.querySelectorAll('.node').forEach(node => {
                const name = node.getAttribute('data-name');
                if (name && query && !name.toLowerCase().includes(query)) {
                    node.style.opacity = '0.3';
                } else {
                    node.style.opacity = '1';
                }
            });
        });

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

        document.getElementById('reset-highlight').addEventListener('click', () => {
            clearHighlights();
        });

        document.getElementById('close-detail').addEventListener('click', () => {
            hideDetailPanel();
        });

        // Filter controls
        function applyFilters() {
            const showState = document.getElementById('filter-state').checked;
            const showPassthrough = document.getElementById('filter-passthrough').checked;
            const showRegular = document.getElementById('filter-regular').checked;

            document.querySelectorAll('.node').forEach(node => {
                const nodeType = node.getAttribute('data-nodetype');
                const type = node.getAttribute('data-type');

                let visible = true;

                if (nodeType === 'state' && !showState) {
                    visible = false;
                } else if (type === 'passthrough' && !showPassthrough) {
                    visible = false;
                } else if (type === 'regular' && !showRegular) {
                    visible = false;
                }

                node.style.display = visible ? '' : 'none';
            });

            // Also hide/show connected edges
            document.querySelectorAll('.edgePath').forEach(edge => {
                const edgeId = edge.getAttribute('id');
                if (!edgeId) return;

                // Check if any connected node is hidden
                const match = edgeId.match(/flowchart-([^-]+)-([^-]+)/);
                if (match) {
                    const [, fromId, toId] = match;
                    const fromNode = document.getElementById(fromId);
                    const toNode = document.getElementById(toId);

                    const fromHidden = fromNode && fromNode.style.display === 'none';
                    const toHidden = toNode && toNode.style.display === 'none';

                    edge.style.display = (fromHidden || toHidden) ? 'none' : '';
                }
            });

            // Hide edge labels for hidden edges
            document.querySelectorAll('.edgeLabel').forEach(label => {
                const parentEdge = label.closest('.edgePath');
                if (parentEdge && parentEdge.style.display === 'none') {
                    label.style.display = 'none';
                } else {
                    label.style.display = '';
                }
            });
        }

        document.getElementById('filter-state').addEventListener('change', applyFilters);
        document.getElementById('filter-passthrough').addEventListener('change', applyFilters);
        document.getElementById('filter-regular').addEventListener('change', applyFilters);

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
