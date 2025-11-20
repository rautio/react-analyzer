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

	// Write header - use TD (top-down) for vertical layout
	sb.WriteString("flowchart TD\n")

	// Track written nodes to avoid duplicates
	writtenNodes := make(map[string]bool)

	// Write component nodes
	for id, node := range g.ComponentNodes {
		if writtenNodes[id] {
			continue
		}
		writtenNodes[id] = true

		// Sanitize ID for Mermaid (no special characters)
		nodeID := sanitizeID(id)

		// Create node label with memoization indicator
		label := node.Name
		if node.IsMemoized {
			label = "⚡ " + label
		}

		// Write node definition
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, label))
	}

	// Write state nodes
	for id, state := range g.StateNodes {
		if writtenNodes[id] {
			continue
		}
		writtenNodes[id] = true

		nodeID := sanitizeID(id)

		// Create label with state type and name
		label := fmt.Sprintf("%s: %s", state.Type, state.Name)

		// Use different shape for state nodes (rounded)
		sb.WriteString(fmt.Sprintf("    %s(\"%s\")\n", nodeID, label))
	}

	sb.WriteString("\n")

	// Write metadata as comments after all node definitions
	sb.WriteString("    %% Metadata\n")
	for id, node := range g.ComponentNodes {
		nodeID := sanitizeID(id)
		sb.WriteString(fmt.Sprintf("    %%%% meta:%s|file:%s|line:%d|type:%s|memoized:%t|nodetype:component\n",
			nodeID,
			node.Location.FilePath,
			node.Location.Line,
			getNodeType(node, g),
			node.IsMemoized,
		))
	}
	for id, state := range g.StateNodes {
		nodeID := sanitizeID(id)
		sb.WriteString(fmt.Sprintf("    %%%% meta:%s|file:%s|line:%d|type:state|memoized:false|nodetype:state|statetype:%s|datatype:%s\n",
			nodeID,
			state.Location.FilePath,
			state.Location.Line,
			state.Type,
			state.DataType,
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

	// Write styling based on node types with better colors and contrast
	for id, node := range g.ComponentNodes {
		nodeID := sanitizeID(id)
		nodeType := getNodeType(node, g)

		var fillColor, strokeColor, textColor string
		switch nodeType {
		case "origin":
			// State origin - vibrant green with dark text
			fillColor = "#10b981"
			strokeColor = "#059669"
			textColor = "#ffffff"
		case "passthrough":
			// Passthrough - vibrant orange with dark text
			fillColor = "#f59e0b"
			strokeColor = "#d97706"
			textColor = "#ffffff"
		case "consumer":
			// Consumer - vibrant blue with white text
			fillColor = "#3b82f6"
			strokeColor = "#2563eb"
			textColor = "#ffffff"
		default:
			// Regular components - neutral gray with dark text
			fillColor = "#9ca3af"
			strokeColor = "#6b7280"
			textColor = "#ffffff"
		}

		sb.WriteString(fmt.Sprintf("    style %s fill:%s,stroke:%s,stroke-width:2px,color:%s\n",
			nodeID, fillColor, strokeColor, textColor))
	}

	// Write styling for state nodes
	for id, state := range g.StateNodes {
		nodeID := sanitizeID(id)

		var fillColor, strokeColor, textColor string
		switch state.Type {
		case StateTypeUseState:
			// useState - purple
			fillColor = "#a855f7"
			strokeColor = "#9333ea"
			textColor = "#ffffff"
		case StateTypeUseReducer:
			// useReducer - deep purple
			fillColor = "#7c3aed"
			strokeColor = "#6d28d9"
			textColor = "#ffffff"
		case StateTypeContext:
			// Context - pink
			fillColor = "#ec4899"
			strokeColor = "#db2777"
			textColor = "#ffffff"
		case StateTypeProp:
			// Prop - cyan
			fillColor = "#06b6d4"
			strokeColor = "#0891b2"
			textColor = "#ffffff"
		default:
			// Derived/unknown - teal
			fillColor = "#14b8a6"
			strokeColor = "#0d9488"
			textColor = "#ffffff"
		}

		sb.WriteString(fmt.Sprintf("    style %s fill:%s,stroke:%s,stroke-width:2px,color:%s\n",
			nodeID, fillColor, strokeColor, textColor))
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
