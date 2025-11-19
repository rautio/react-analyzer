package graph

import (
	"crypto/sha256"
	"fmt"
)

// GenerateStateID generates a unique ID for a state node
// Format: state:<componentName>:<stateName>:<hash>
func GenerateStateID(componentName, stateName, filePath string, line uint32) string {
	// Include file path and line number for uniqueness
	data := fmt.Sprintf("%s:%s:%s:%d", componentName, stateName, filePath, line)
	hash := sha256.Sum256([]byte(data))
	shortHash := fmt.Sprintf("%x", hash[:4]) // Use first 8 hex chars

	return fmt.Sprintf("state:%s:%s:%s", componentName, stateName, shortHash)
}

// GenerateComponentID generates a unique ID for a component node
// Format: component:<componentName>:<hash>
func GenerateComponentID(componentName, filePath string, line uint32) string {
	// Include file path and line number for uniqueness
	// This handles cases where same component name appears in different files
	data := fmt.Sprintf("%s:%s:%d", componentName, filePath, line)
	hash := sha256.Sum256([]byte(data))
	shortHash := fmt.Sprintf("%x", hash[:4]) // Use first 8 hex chars

	return fmt.Sprintf("component:%s:%s", componentName, shortHash)
}

// GenerateEdgeID generates a unique ID for an edge
// Format: edge:<type>:<sourceID>:<targetID>:<hash>
func GenerateEdgeID(edgeType EdgeType, sourceID, targetID string) string {
	data := fmt.Sprintf("%s:%s:%s", edgeType, sourceID, targetID)
	hash := sha256.Sum256([]byte(data))
	shortHash := fmt.Sprintf("%x", hash[:4]) // Use first 8 hex chars

	return fmt.Sprintf("edge:%s:%s", edgeType, shortHash)
}

// GenerateEdgeIDWithProp generates a unique ID for an edge that includes prop name
// Format: edge:<type>:<sourceID>:<targetID>:<propName>:<hash>
// This is used for "passes" edges where we want one edge per prop passed
func GenerateEdgeIDWithProp(edgeType EdgeType, sourceID, targetID, propName string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", edgeType, sourceID, targetID, propName)
	hash := sha256.Sum256([]byte(data))
	shortHash := fmt.Sprintf("%x", hash[:4]) // Use first 8 hex chars

	return fmt.Sprintf("edge:%s:%s", edgeType, shortHash)
}

// ExtractComponentName extracts the component name from a component ID
func ExtractComponentName(componentID string) string {
	// component:ComponentName:hash
	var name string
	fmt.Sscanf(componentID, "component:%s", &name)

	// Remove the hash suffix
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ':' {
			return name[:i]
		}
	}

	return name
}

// ExtractStateName extracts the state name from a state ID
func ExtractStateName(stateID string) string {
	// state:ComponentName:stateName:hash
	var component, name string
	fmt.Sscanf(stateID, "state:%s:%s", &component, &name)

	// Remove the hash suffix from name
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ':' {
			return name[:i]
		}
	}

	return name
}
