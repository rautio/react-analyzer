/**
 * Type definitions for React Analyzer CLI output
 * These match the JSON structures defined in Go
 */

export interface RelatedInformation {
    filePath: string;
    line: number;
    column: number;
    message: string;
}

export interface Issue {
    rule: string;
    message: string;
    filePath: string;
    line: number;
    column: number;
    related?: RelatedInformation[];
}

export interface Stats {
    filesAnalyzed: number;
    filesWithIssues: number;
    filesClean: number;
    totalIssues: number;
    durationMs: number;
}

export interface Location {
    filePath: string;
    line: number;
    column: number;
    component: string;
}

export interface PropDefinition {
    name: string;
    type: string;
    required: boolean;
    location: Location;
}

export interface ComponentNode {
    id: string;
    name: string;
    type: 'function' | 'class' | 'memo';
    location: Location;
    isMemoized: boolean;
    stateNodes: string[];
    consumedState: string[];
    parent: string;
    children: string[];
    props: PropDefinition[];
    propsPassedTo: Record<string, string[]>;
    propsUsedLocally: string[];
}

export interface StateNode {
    id: string;
    name: string;
    type: 'useState' | 'useReducer' | 'context' | 'prop' | 'derived';
    dataType: 'primitive' | 'object' | 'array' | 'function' | 'unknown';
    location: Location;
    mutable: boolean;
    dependencies: string[];
    updateLocations: Location[];
}

export interface Edge {
    id: string;
    sourceId: string;
    targetId: string;
    type: 'defines' | 'consumes' | 'updates' | 'passes' | 'derives';
    weight: number;
    propName: string;
    location: Location;
}

export interface Graph {
    stateNodes: Record<string, StateNode>;
    componentNodes: Record<string, ComponentNode>;
    edges: Edge[];
}

export interface AnalysisResult {
    issues: Issue[];
    stats: Stats;
    graph?: Graph;
}
