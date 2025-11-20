package graph

import (
	"fmt"
)

// PropDrillingViolation represents a detected prop drilling anti-pattern
type PropDrillingViolation struct {
	PropName              string
	Origin                Location
	Consumer              Location
	PassthroughCount      int
	PassthroughComponents []ComponentReference
	Depth                 int
	Recommendation        string
}

// ComponentReference represents a reference to a component
type ComponentReference struct {
	Name     string
	FilePath string
	Line     uint32
}

// DetectPropDrilling finds all prop drilling violations in the graph
// A violation occurs when a prop is passed through multiple component levels
// with intermediate components not using the prop (passthrough only)
// minPassthroughComponents: minimum number of passthrough components to trigger a violation
func DetectPropDrilling(g *Graph, minPassthroughComponents int) []PropDrillingViolation {
	var violations []PropDrillingViolation

	// Find all state origins (useState, useReducer, etc.)
	origins := findPropOrigins(g)

	// For each origin, trace where it flows
	for stateID, origin := range origins {
		// Find all components that consume this state (including through renaming)
		consumersWithNames := findPropConsumersWithNames(stateID, g)

		// Only report violations for leaf consumers (deepest in the chain)
		// to avoid reporting multiple violations for the same prop drilling path
		leafConsumers := findLeafConsumersWithNames(consumersWithNames, g)

		// For each leaf consumer, trace the path from origin to consumer
		for _, consumer := range leafConsumers {
			path := tracePropPath(origin, consumer.comp, g)

			// Violation if there are enough passthrough components
			// A passthrough component is one that receives a prop and passes it down
			// without using it locally (not referencing it in the component body)
			// This distinguishes from components that both use AND pass props
			if len(path.PassthroughPath) >= minPassthroughComponents {
				violation := PropDrillingViolation{
					PropName:              origin.Name,
					Origin:                origin.Location,
					Consumer:              consumer.comp.Location,
					PassthroughCount:      len(path.PassthroughPath),
					PassthroughComponents: buildComponentReferences(path.PassthroughPath),
					Depth:                 path.Depth,
					Recommendation:        generateRecommendation(origin, path),
				}

				violations = append(violations, violation)
			}
		}
	}

	return violations
}

// PropPath represents the path a prop takes from origin to consumer
type PropPath struct {
	Origin          *StateNode
	Consumer        *ComponentNode
	PassthroughPath []*ComponentNode // Components that pass but don't use the prop
	Depth           int              // Number of levels (origin to consumer)
}

// findPropOrigins finds all state nodes that are origins (defined via useState, etc.)
func findPropOrigins(g *Graph) map[string]*StateNode {
	origins := make(map[string]*StateNode)

	for id, stateNode := range g.StateNodes {
		// Props that are defined as state are origins
		// Includes derived state (properties of state objects like settings.locale)
		if stateNode.Type == StateTypeUseState ||
			stateNode.Type == StateTypeUseReducer ||
			stateNode.Type == StateTypeContext ||
			stateNode.Type == StateTypeDerived {
			origins[id] = stateNode
		}
	}

	return origins
}

// consumerWithName tracks a component and the current name of the prop at that component
type consumerWithName struct {
	comp        *ComponentNode
	propNameNow string
}

// findLeafConsumersWithNames filters consumers to only include leaf nodes
// (components that don't pass this prop to any children)
func findLeafConsumersWithNames(consumers []consumerWithName, g *Graph) []consumerWithName {
	var leafConsumers []consumerWithName

	for _, consumer := range consumers {
		// Check if this consumer passes the prop to any child
		// We check if any outgoing edge passes the prop (by current name or as source ident)
		hasOutgoingProp := false
		outgoingEdges := g.GetOutgoingEdges(consumer.comp.ID)
		for _, edge := range outgoingEdges {
			if edge.Type == EdgeTypePasses {
				// Check if this edge passes our tracked prop
				if edge.PropName == consumer.propNameNow || edge.PropSourceIdent == consumer.propNameNow {
					hasOutgoingProp = true
					break
				}
			}
		}

		// Only include if it doesn't pass the prop to children (leaf node)
		if !hasOutgoingProp {
			leafConsumers = append(leafConsumers, consumer)
		}
	}

	return leafConsumers
}

// findPropConsumersWithNames finds all components in the prop flow chain for a specific prop
// and tracks the current name of the prop at each component (may differ due to renaming)
func findPropConsumersWithNames(propID string, g *Graph) []consumerWithName {
	var consumers []consumerWithName

	// Get the state node to find the prop name
	stateNode, exists := g.StateNodes[propID]
	if !exists {
		return consumers
	}
	initialPropName := stateNode.Name

	// Start from the origin component
	originComp, err := g.FindStateOrigin(propID)
	if err != nil {
		return consumers
	}

	// Use BFS to find all components that receive this specific prop
	// Track both component ID and the current prop name at that component
	type queueItem struct {
		compID      string
		propNameNow string // The name of the prop at this component (may differ due to renaming)
	}

	visited := make(map[string]bool)
	queue := []queueItem{{compID: originComp.ID, propNameNow: initialPropName}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.compID] {
			continue
		}
		visited[current.compID] = true

		// Check all "passes" edges from this component
		outgoingEdges := g.GetOutgoingEdges(current.compID)
		for _, edge := range outgoingEdges {
			// Only follow edges that pass THIS specific prop (by name or source identifier for renamed props)
			if edge.Type == EdgeTypePasses {
				// Match by prop name OR source identifier (handles renaming)
				// We need to check if this edge passes the prop we're tracking
				matches := edge.PropName == current.propNameNow || edge.PropSourceIdent == current.propNameNow
				if matches {
					// This component passes our specific prop to a child
					targetComp, exists := g.ComponentNodes[edge.TargetID]
					if !exists {
						continue
					}

					// Add the child component as a consumer
					if !visited[edge.TargetID] {
						consumers = append(consumers, consumerWithName{
							comp:        targetComp,
							propNameNow: edge.PropName, // Use the new name at this component
						})
						// Continue tracking with the NEW prop name (it may have been renamed)
						queue = append(queue, queueItem{
							compID:      edge.TargetID,
							propNameNow: edge.PropName, // Use the new name going forward
						})
					}
				}
			}
		}
	}

	return consumers
}

// componentUsesProp checks if a component actually uses a prop or just passes it through
func componentUsesProp(comp *ComponentNode, propName string, g *Graph) bool {
	// Check if this prop is in PropsUsedLocally (determined during AST analysis)
	for _, usedProp := range comp.PropsUsedLocally {
		if usedProp == propName {
			return true
		}
	}

	// If the component has no children, it must be using the prop
	// (This is a fallback for components created before PropsUsedLocally was added)
	if len(comp.Children) == 0 {
		return true
	}

	// If the prop is not in PropsUsedLocally and the component has children,
	// it's likely a passthrough component
	return false
}

// tracePropPath traces the path from a state origin to a consumer component
func tracePropPath(origin *StateNode, consumer *ComponentNode, g *Graph) *PropPath {
	path := &PropPath{
		Origin:          origin,
		Consumer:        consumer,
		PassthroughPath: []*ComponentNode{},
		Depth:           0,
	}

	// Find the component that defines this state
	originComp, err := g.FindStateOrigin(origin.ID)
	if err != nil {
		return path
	}

	// Find shortest path between origin component and consumer component
	componentPath := g.FindPathBetweenComponents(originComp.ID, consumer.ID)
	if componentPath == nil || len(componentPath) < 2 {
		return path
	}

	// Build passthrough path (exclude origin and consumer)
	path.Depth = len(componentPath)

	// Extract the prop name from the state node
	propName := origin.Name

	for i := 1; i < len(componentPath)-1; i++ {
		if comp, exists := g.ComponentNodes[componentPath[i]]; exists {
			// Check if this component is a true passthrough (doesn't use the prop)
			if !componentUsesProp(comp, propName, g) {
				path.PassthroughPath = append(path.PassthroughPath, comp)
			}
		}
	}

	return path
}

// buildComponentReferences converts ComponentNodes to ComponentReferences
func buildComponentReferences(components []*ComponentNode) []ComponentReference {
	refs := make([]ComponentReference, len(components))
	for i, comp := range components {
		refs[i] = ComponentReference{
			Name:     comp.Name,
			FilePath: comp.Location.FilePath,
			Line:     comp.Location.Line,
		}
	}
	return refs
}

// generateRecommendation creates a helpful recommendation for fixing the violation
func generateRecommendation(origin *StateNode, path *PropPath) string {
	// Simple recommendation for Phase 2.1
	return fmt.Sprintf(
		"Consider using Context API to avoid passing '%s' through %d component levels. "+
			"Create a context at %s and consume it directly in %s.",
		origin.Name,
		path.Depth,
		origin.Location.Component,
		path.Consumer.Name,
	)
}
