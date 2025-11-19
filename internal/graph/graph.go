package graph

import (
	"fmt"
)

// AddStateNode adds a state node to the graph
func (g *Graph) AddStateNode(node *StateNode) {
	g.StateNodes[node.ID] = node
}

// AddComponentNode adds a component node to the graph
func (g *Graph) AddComponentNode(node *ComponentNode) {
	g.ComponentNodes[node.ID] = node
}

// AddEdge adds an edge to the graph and updates indexes
func (g *Graph) AddEdge(edge Edge) {
	g.Edges = append(g.Edges, edge)

	// Update source index
	g.EdgesBySource[edge.SourceID] = append(g.EdgesBySource[edge.SourceID], edge)

	// Update target index
	g.EdgesByTarget[edge.TargetID] = append(g.EdgesByTarget[edge.TargetID], edge)
}

// GetOutgoingEdges returns all edges originating from a node
func (g *Graph) GetOutgoingEdges(nodeID string) []Edge {
	return g.EdgesBySource[nodeID]
}

// GetIncomingEdges returns all edges pointing to a node
func (g *Graph) GetIncomingEdges(nodeID string) []Edge {
	return g.EdgesByTarget[nodeID]
}

// GetEdgesByType returns all edges of a specific type
func (g *Graph) GetEdgesByType(edgeType EdgeType) []Edge {
	var result []Edge
	for _, edge := range g.Edges {
		if edge.Type == edgeType {
			result = append(result, edge)
		}
	}
	return result
}

// FindStateOrigin finds the component that defines a given state
func (g *Graph) FindStateOrigin(stateID string) (*ComponentNode, error) {
	// Look for "defines" edges pointing to this state
	edges := g.GetIncomingEdges(stateID)
	for _, edge := range edges {
		if edge.Type == EdgeTypeDefines {
			if comp, exists := g.ComponentNodes[edge.SourceID]; exists {
				return comp, nil
			}
		}
	}
	return nil, fmt.Errorf("no origin found for state %s", stateID)
}

// FindStateConsumers finds all components that consume a given state
func (g *Graph) FindStateConsumers(stateID string) []*ComponentNode {
	var consumers []*ComponentNode

	// Look for "consumes" edges originating from components to this state
	edges := g.GetIncomingEdges(stateID)
	for _, edge := range edges {
		if edge.Type == EdgeTypeConsumes {
			if comp, exists := g.ComponentNodes[edge.SourceID]; exists {
				consumers = append(consumers, comp)
			}
		}
	}

	return consumers
}

// TraceStatePropagation traces how state flows through the component tree
// Returns a list of components in the propagation path
func (g *Graph) TraceStatePropagation(stateID string) []string {
	visited := make(map[string]bool)
	var path []string

	var traverse func(nodeID string)
	traverse = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true

		// Add to path if it's a component
		if _, isComponent := g.ComponentNodes[nodeID]; isComponent {
			path = append(path, nodeID)
		}

		// Follow outgoing edges
		for _, edge := range g.GetOutgoingEdges(nodeID) {
			if edge.Type == EdgeTypePasses || edge.Type == EdgeTypeConsumes {
				traverse(edge.TargetID)
			}
		}
	}

	// Start from the state node
	traverse(stateID)

	return path
}

// GetComponentChildren returns all child components of a given component
func (g *Graph) GetComponentChildren(componentID string) []*ComponentNode {
	comp, exists := g.ComponentNodes[componentID]
	if !exists {
		return nil
	}

	var children []*ComponentNode
	for _, childID := range comp.Children {
		if child, exists := g.ComponentNodes[childID]; exists {
			children = append(children, child)
		}
	}

	return children
}

// GetComponentParent returns the parent component of a given component
func (g *Graph) GetComponentParent(componentID string) (*ComponentNode, error) {
	comp, exists := g.ComponentNodes[componentID]
	if !exists {
		return nil, fmt.Errorf("component %s not found", componentID)
	}

	if comp.Parent == "" {
		return nil, nil // No parent (root component)
	}

	parent, exists := g.ComponentNodes[comp.Parent]
	if !exists {
		return nil, fmt.Errorf("parent %s not found", comp.Parent)
	}

	return parent, nil
}

// GetComponentDepth calculates the depth of a component in the tree
// Root components have depth 0
func (g *Graph) GetComponentDepth(componentID string) int {
	depth := 0
	current := componentID

	for {
		comp, exists := g.ComponentNodes[current]
		if !exists || comp.Parent == "" {
			break
		}
		depth++
		current = comp.Parent
	}

	return depth
}

// FindPathBetweenComponents finds the shortest path between two components
// Returns the component IDs in order from start to end
func (g *Graph) FindPathBetweenComponents(startID, endID string) []string {
	// BFS to find shortest path
	type queueItem struct {
		nodeID string
		path   []string
	}

	queue := []queueItem{{nodeID: startID, path: []string{startID}}}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.nodeID == endID {
			return current.path
		}

		if visited[current.nodeID] {
			continue
		}
		visited[current.nodeID] = true

		// Explore children and parent
		if comp, exists := g.ComponentNodes[current.nodeID]; exists {
			// Check children
			for _, childID := range comp.Children {
				if !visited[childID] {
					newPath := append([]string{}, current.path...)
					newPath = append(newPath, childID)
					queue = append(queue, queueItem{nodeID: childID, path: newPath})
				}
			}

			// Check parent
			if comp.Parent != "" && !visited[comp.Parent] {
				newPath := append([]string{}, current.path...)
				newPath = append(newPath, comp.Parent)
				queue = append(queue, queueItem{nodeID: comp.Parent, path: newPath})
			}
		}
	}

	return nil // No path found
}
