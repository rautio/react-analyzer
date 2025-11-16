package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// wrapNode wraps a tree-sitter node in our Node type
func wrapNode(tsNode *sitter.Node, content []byte) *Node {
	if tsNode == nil {
		return nil
	}

	return &Node{
		tsNode:  tsNode,
		content: content,
	}
}

// Type returns the node type (e.g., "function_declaration", "call_expression")
func (n *Node) Type() string {
	if n == nil || n.tsNode == nil {
		return ""
	}
	return n.tsNode.Type()
}

// Text returns the source code text for this node
func (n *Node) Text() string {
	if n == nil || n.tsNode == nil {
		return ""
	}
	return n.tsNode.Content(n.content)
}

// Children returns all child nodes
func (n *Node) Children() []*Node {
	if n == nil || n.tsNode == nil {
		return nil
	}

	count := int(n.tsNode.ChildCount())
	children := make([]*Node, 0, count)

	for i := 0; i < count; i++ {
		child := n.tsNode.Child(i)
		if child != nil {
			children = append(children, wrapNode(child, n.content))
		}
	}

	return children
}

// NamedChildren returns only named child nodes (skips punctuation, etc.)
func (n *Node) NamedChildren() []*Node {
	if n == nil || n.tsNode == nil {
		return nil
	}

	count := int(n.tsNode.NamedChildCount())
	children := make([]*Node, 0, count)

	for i := 0; i < count; i++ {
		child := n.tsNode.NamedChild(i)
		if child != nil {
			children = append(children, wrapNode(child, n.content))
		}
	}

	return children
}

// ChildByFieldName returns a child node by field name
func (n *Node) ChildByFieldName(field string) *Node {
	if n == nil || n.tsNode == nil {
		return nil
	}

	child := n.tsNode.ChildByFieldName(field)
	return wrapNode(child, n.content)
}

// StartPoint returns the starting position of this node
func (n *Node) StartPoint() (row, col uint32) {
	if n == nil || n.tsNode == nil {
		return 0, 0
	}
	point := n.tsNode.StartPoint()
	return point.Row, point.Column
}

// EndPoint returns the ending position of this node
func (n *Node) EndPoint() (row, col uint32) {
	if n == nil || n.tsNode == nil {
		return 0, 0
	}
	point := n.tsNode.EndPoint()
	return point.Row, point.Column
}

// IsHookCall checks if this node is a call to a React hook (function starting with "use")
func (n *Node) IsHookCall() bool {
	if n == nil || n.Type() != "call_expression" {
		return false
	}

	// Get the function being called
	funcNode := n.ChildByFieldName("function")
	if funcNode == nil {
		return false
	}

	// Get the function name
	funcName := funcNode.Text()

	// React hooks start with "use"
	return strings.HasPrefix(funcName, "use") && len(funcName) > 3
}

// GetDependencyArray returns the dependency array for a hook call
// For useEffect/useMemo/useCallback, this is the last argument
func (n *Node) GetDependencyArray() *Node {
	if n == nil || !n.IsHookCall() {
		return nil
	}

	// Get arguments
	args := n.ChildByFieldName("arguments")
	if args == nil {
		return nil
	}

	// Get all named children (skips commas, parentheses)
	namedArgs := args.NamedChildren()
	if len(namedArgs) == 0 {
		return nil
	}

	// Dependency array is the last argument
	lastArg := namedArgs[len(namedArgs)-1]

	// Check if it's an array
	if lastArg.Type() == "array" {
		return lastArg
	}

	return nil
}

// GetArrayElements returns the elements of an array node
func (n *Node) GetArrayElements() []*Node {
	if n == nil || n.Type() != "array" {
		return nil
	}

	// Named children of array are the elements (skips brackets and commas)
	return n.NamedChildren()
}

// Walk traverses the AST depth-first, calling visitor for each node
func (n *Node) Walk(visitor func(*Node) bool) {
	if n == nil {
		return
	}

	// Call visitor, if it returns false, stop traversal
	if !visitor(n) {
		return
	}

	// Recursively visit children
	for _, child := range n.Children() {
		child.Walk(visitor)
	}
}
