package graph

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ToMermaid converts the graph to Mermaid flowchart syntax
// Returns a complete Mermaid diagram with metadata embedded as comments
func (g *Graph) ToMermaid() string {
	var sb strings.Builder

	// Write header
	sb.WriteString("flowchart TD\n")

	// Track written nodes to avoid duplicates
	writtenNodes := make(map[string]bool)

	// Write component nodes with metadata
	for id, node := range g.ComponentNodes {
		if writtenNodes[id] {
			continue
		}
		writtenNodes[id] = true

		// Sanitize ID for Mermaid (no special characters)
		nodeID := sanitizeID(id)

		// Create node label (just the component name for now)
		label := node.Name

		// Write node definition
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, label))

		// Write metadata as comment for custom parsing
		sb.WriteString(fmt.Sprintf("    %%%%{meta: %s, file: \"%s\", line: %d, type: \"%s\", memoized: %t}%%%%\n",
			nodeID,
			node.Location.FilePath,
			node.Location.Line,
			getNodeType(node, g),
			node.IsMemoized,
		))
	}

	sb.WriteString("\n")

	// Write edges
	for _, edge := range g.Edges {
		fromID := sanitizeID(edge.SourceID)
		toID := sanitizeID(edge.TargetID)

		// Create edge label (prop name if available)
		label := ""
		if edge.PropName != "" {
			label = edge.PropName
		}

		if label != "" {
			sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", fromID, label, toID))
		} else {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
		}
	}

	sb.WriteString("\n")

	// Write styling based on node types
	for id, node := range g.ComponentNodes {
		nodeID := sanitizeID(id)
		nodeType := getNodeType(node, g)

		var color string
		switch nodeType {
		case "origin":
			color = "#e1f5e1" // Green for state origin
		case "passthrough":
			color = "#fff4e1" // Yellow for passthrough
		case "consumer":
			color = "#e1f0ff" // Blue for consumer
		default:
			color = "#f5f5f5" // Gray for regular components
		}

		sb.WriteString(fmt.Sprintf("    style %s fill:%s\n", nodeID, color))
	}

	return sb.String()
}

// sanitizeID converts a graph node ID to a valid Mermaid node ID
// Replaces special characters with underscores
func sanitizeID(id string) string {
	// Replace problematic characters
	sanitized := strings.ReplaceAll(id, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Ensure it starts with a letter
	if len(sanitized) > 0 && (sanitized[0] < 'A' || sanitized[0] > 'z') {
		sanitized = "node_" + sanitized
	}

	return sanitized
}

// getNodeType determines the type of a component node for styling
// Returns: "origin", "passthrough", "consumer", or "regular"
func getNodeType(node *ComponentNode, g *Graph) string {
	// Check if this component creates state
	hasState := len(node.StateNodes) > 0

	// Check if this component is a passthrough (receives props but doesn't use them)
	isPassthrough := false
	consumesState := false

	// Look at edges to determine passthrough vs consumer
	for _, edge := range g.Edges {
		// Check if this node passes props to children
		if edge.SourceID == node.ID && edge.Type == EdgeTypePasses {
			isPassthrough = true
		}

		// Check if this node consumes state
		if edge.TargetID == node.ID && edge.Type == EdgeTypeConsumes {
			consumesState = true
		}
	}

	// Determine type
	if hasState {
		return "origin"
	} else if isPassthrough && !consumesState {
		return "passthrough"
	} else if consumesState {
		return "consumer"
	}

	return "regular"
}

// ToMermaidWithTitle returns a Mermaid diagram wrapped in a markdown code block
// Useful for outputting to .md files
func (g *Graph) ToMermaidWithTitle(title string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString("```mermaid\n")
	sb.WriteString(g.ToMermaid())
	sb.WriteString("```\n")

	return sb.String()
}

// ExtractFileBasename returns just the filename without path for display
func extractFileBasename(path string) string {
	return filepath.Base(path)
}
