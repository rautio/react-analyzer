import * as vscode from 'vscode';
import { Graph, ComponentNode, StateNode, Edge } from './types';

export class ComponentTreeItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly nodeType: 'component' | 'state' | 'group',
        public readonly nodeId?: string,
        public readonly filePath?: string,
        public readonly line?: number,
        public readonly column?: number
    ) {
        super(label, collapsibleState);

        // Set context value for conditional commands
        this.contextValue = nodeType;

        // Set icons based on node type
        if (nodeType === 'component') {
            this.iconPath = new vscode.ThemeIcon('symbol-class');
        } else if (nodeType === 'state') {
            this.iconPath = new vscode.ThemeIcon('symbol-variable');
        } else if (nodeType === 'group') {
            this.iconPath = new vscode.ThemeIcon('folder');
        }

        // Set description (shows on the right)
        if (nodeType === 'component' && nodeId) {
            const comp = nodeId.split(':')[0]; // Extract component type
            this.description = comp;
        }

        // Make item clickable if it has a location
        if (filePath && line !== undefined && column !== undefined) {
            this.command = {
                command: 'reactAnalyzer.navigateToLocation',
                title: 'Navigate to Location',
                arguments: [filePath, line, column]
            };
        }
    }
}

export class ComponentTreeProvider implements vscode.TreeDataProvider<ComponentTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ComponentTreeItem | undefined | null | void> = new vscode.EventEmitter<ComponentTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ComponentTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    private graph: Graph | null = null;

    constructor() {}

    /**
     * Update the tree with new graph data
     */
    updateGraph(graph: Graph | null): void {
        this.graph = graph;
        this._onDidChangeTreeData.fire();
    }

    /**
     * Get tree item representation
     */
    getTreeItem(element: ComponentTreeItem): vscode.TreeItem {
        return element;
    }

    /**
     * Get children for a tree item
     */
    getChildren(element?: ComponentTreeItem): Thenable<ComponentTreeItem[]> {
        if (!this.graph) {
            return Promise.resolve([]);
        }

        if (!element) {
            // Root level: show Components and State groups
            return Promise.resolve([
                new ComponentTreeItem('Components', vscode.TreeItemCollapsibleState.Expanded, 'group'),
                new ComponentTreeItem('State', vscode.TreeItemCollapsibleState.Collapsed, 'group')
            ]);
        }

        // Handle group items
        if (element.label === 'Components') {
            return Promise.resolve(this.getComponentNodes());
        }

        if (element.label === 'State') {
            return Promise.resolve(this.getStateNodes());
        }

        // Handle component items - show children and state
        if (element.nodeType === 'component' && element.nodeId) {
            return Promise.resolve(this.getComponentChildren(element.nodeId));
        }

        // Handle state items - show consumers
        if (element.nodeType === 'state' && element.nodeId) {
            return Promise.resolve(this.getStateConsumers(element.nodeId));
        }

        return Promise.resolve([]);
    }

    /**
     * Get all component nodes from the graph
     */
    private getComponentNodes(): ComponentTreeItem[] {
        if (!this.graph) {
            return [];
        }

        const items: ComponentTreeItem[] = [];

        // Find root components (no parent)
        for (const [id, component] of Object.entries(this.graph.componentNodes)) {
            if (!component.parent || component.parent === '') {
                items.push(this.createComponentItem(id, component));
            }
        }

        return items.sort((a, b) => a.label.localeCompare(b.label));
    }

    /**
     * Create a tree item for a component
     */
    private createComponentItem(id: string, component: ComponentNode): ComponentTreeItem {
        const hasChildren = component.children && component.children.length > 0;
        const hasState = component.stateNodes && component.stateNodes.length > 0;

        const collapsibleState = (hasChildren || hasState)
            ? vscode.TreeItemCollapsibleState.Collapsed
            : vscode.TreeItemCollapsibleState.None;

        let label = component.name;

        // Add indicators for memoization and state
        if (component.isMemoized) {
            label += ' [memo]';
        }
        if (hasState) {
            label += ` (${component.stateNodes.length} state)`;
        }

        return new ComponentTreeItem(
            label,
            collapsibleState,
            'component',
            id,
            component.location.filePath,
            component.location.line,
            component.location.column
        );
    }

    /**
     * Get children for a component (child components + state)
     */
    private getComponentChildren(componentId: string): ComponentTreeItem[] {
        if (!this.graph) {
            return [];
        }

        const component = this.graph.componentNodes[componentId];
        if (!component) {
            return [];
        }

        const items: ComponentTreeItem[] = [];

        // Add child components
        if (component.children && component.children.length > 0) {
            for (const childId of component.children) {
                const childComponent = this.graph.componentNodes[childId];
                if (childComponent) {
                    items.push(this.createComponentItem(childId, childComponent));
                }
            }
        }

        // Add state nodes defined in this component
        if (component.stateNodes && component.stateNodes.length > 0) {
            for (const stateId of component.stateNodes) {
                const stateNode = this.graph.stateNodes[stateId];
                if (stateNode) {
                    items.push(this.createStateItem(stateId, stateNode, true));
                }
            }
        }

        return items;
    }

    /**
     * Get all state nodes from the graph
     */
    private getStateNodes(): ComponentTreeItem[] {
        if (!this.graph) {
            return [];
        }

        const items: ComponentTreeItem[] = [];

        for (const [id, stateNode] of Object.entries(this.graph.stateNodes)) {
            items.push(this.createStateItem(id, stateNode, false));
        }

        return items.sort((a, b) => a.label.localeCompare(b.label));
    }

    /**
     * Create a tree item for a state node
     */
    private createStateItem(id: string, stateNode: StateNode, showType: boolean): ComponentTreeItem {
        let label = stateNode.name;

        if (showType) {
            label += ` [${stateNode.type}]`;
        }

        return new ComponentTreeItem(
            label,
            vscode.TreeItemCollapsibleState.None,
            'state',
            id,
            stateNode.location.filePath,
            stateNode.location.line,
            stateNode.location.column
        );
    }

    /**
     * Get components that consume a state node
     */
    private getStateConsumers(stateId: string): ComponentTreeItem[] {
        if (!this.graph) {
            return [];
        }

        const items: ComponentTreeItem[] = [];

        // Find all edges that target this state (components consuming it)
        for (const edge of this.graph.edges) {
            if (edge.targetId === stateId && edge.type === 'consumes') {
                const component = this.graph.componentNodes[edge.sourceId];
                if (component) {
                    items.push(this.createComponentItem(edge.sourceId, component));
                }
            }
        }

        return items;
    }
}
