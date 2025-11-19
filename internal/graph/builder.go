package graph

import (
	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
)

// Builder constructs a Graph from parsed modules
type Builder struct {
	graph    *Graph
	resolver *analyzer.ModuleResolver
}

// NewBuilder creates a new graph builder
func NewBuilder(resolver *analyzer.ModuleResolver) *Builder {
	return &Builder{
		graph:    NewGraph(),
		resolver: resolver,
	}
}

// Build constructs the graph from all parsed modules
func (b *Builder) Build() (*Graph, error) {
	modules := b.resolver.GetModules()

	// Phase 1: Build component nodes
	for _, module := range modules {
		if err := b.buildComponentNodes(module); err != nil {
			return nil, err
		}
	}

	// Phase 2: Build state nodes and edges within components
	for _, module := range modules {
		if err := b.buildStateNodes(module); err != nil {
			return nil, err
		}
	}

	// Phase 3: Build component hierarchy (parent-child relationships)
	for _, module := range modules {
		if err := b.buildComponentHierarchy(module); err != nil {
			return nil, err
		}
	}

	// Phase 4: Build prop passing edges
	for _, module := range modules {
		if err := b.buildPropPassingEdges(module); err != nil {
			return nil, err
		}
	}

	return b.graph, nil
}

// buildComponentNodes creates ComponentNode entries for all React components
func (b *Builder) buildComponentNodes(module *analyzer.Module) error {
	// Acquire tree-sitter lock for AST operations
	b.resolver.LockTreeSitter()
	defer b.resolver.UnlockTreeSitter()

	ast := module.AST

	// Walk AST to find function components
	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Look for function declarations
		if nodeType == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName == "" || !isReactComponent(componentName) {
				return true
			}

			line, col := node.StartPoint()
			componentID := GenerateComponentID(componentName, module.FilePath, line+1)

			// Check if component is memoized from symbol table
			isMemoized := false
			if symbol, exists := module.Symbols[componentName]; exists {
				isMemoized = symbol.IsMemoized
			}

			// Extract props
			props := b.extractProps(node)

			compNode := &ComponentNode{
				ID:            componentID,
				Name:          componentName,
				Type:          ComponentTypeFunction,
				Location:      Location{FilePath: module.FilePath, Line: line + 1, Column: col, Component: componentName},
				IsMemoized:    isMemoized,
				StateNodes:    []string{},
				ConsumedState: []string{},
				Children:      []string{},
				Props:         props,
				PropsPassedTo: make(map[string][]string),
			}

			b.graph.AddComponentNode(compNode)
		}

		// TODO: Handle arrow function components (const Foo = () => {})
		// This requires traversing variable_declarator nodes

		return true
	})

	return nil
}

// buildStateNodes creates StateNode entries for state defined in components
func (b *Builder) buildStateNodes(module *analyzer.Module) error {
	// Acquire tree-sitter lock for AST operations
	b.resolver.LockTreeSitter()
	defer b.resolver.UnlockTreeSitter()

	ast := module.AST

	// Track current component context
	var currentComponent *ComponentNode

	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Track which component we're in
		if nodeType == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName != "" && isReactComponent(componentName) {
				line, _ := node.StartPoint()
				componentID := GenerateComponentID(componentName, module.FilePath, line+1)
				if comp, exists := b.graph.ComponentNodes[componentID]; exists {
					currentComponent = comp
				}
			}
		}

		// Look for variable_declarator with useState: const [count, setCount] = useState(0)
		if nodeType == "variable_declarator" && currentComponent != nil {
			// Get the initializer (right side of =)
			initializer := node.ChildByFieldName("value")
			if initializer != nil && initializer.IsHookCall() {
				callee := initializer.ChildByFieldName("function")
				if callee != nil && callee.Text() == "useState" {
					// Get the name pattern (left side of =)
					pattern := node.ChildByFieldName("name")
					b.handleUseStateWithPattern(initializer, pattern, currentComponent, module.FilePath)
				}
			}
		}

		return true
	})

	return nil
}

// buildComponentHierarchy establishes parent-child relationships between components
func (b *Builder) buildComponentHierarchy(module *analyzer.Module) error {
	// Acquire tree-sitter lock for AST operations
	b.resolver.LockTreeSitter()
	defer b.resolver.UnlockTreeSitter()

	ast := module.AST

	// Track current component context
	var currentComponent *ComponentNode

	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Track which component we're in
		if nodeType == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName != "" && isReactComponent(componentName) {
				line, _ := node.StartPoint()
				componentID := GenerateComponentID(componentName, module.FilePath, line+1)
				if comp, exists := b.graph.ComponentNodes[componentID]; exists {
					currentComponent = comp
				}
			}
		}

		// Look for JSX elements (child component usage)
		if (nodeType == "jsx_element" || nodeType == "jsx_self_closing_element") && currentComponent != nil {
			childComponentName := b.getJSXComponentName(node)
			if childComponentName != "" && isReactComponent(childComponentName) {
				// Find the child component node
				childCompID := b.findComponentID(childComponentName, module.FilePath)
				if childCompID != "" {
					// Add to parent's children list if not already there
					if !contains(currentComponent.Children, childCompID) {
						currentComponent.Children = append(currentComponent.Children, childCompID)
					}

					// Set child's parent
					if childComp, exists := b.graph.ComponentNodes[childCompID]; exists {
						childComp.Parent = currentComponent.ID
					}
				}
			}
		}

		return true
	})

	return nil
}

// buildPropPassingEdges creates edges for props passed between components
func (b *Builder) buildPropPassingEdges(module *analyzer.Module) error {
	// Acquire tree-sitter lock for AST operations
	b.resolver.LockTreeSitter()
	defer b.resolver.UnlockTreeSitter()

	ast := module.AST

	// Track current component context
	var currentComponent *ComponentNode

	ast.Root.Walk(func(node *parser.Node) bool {
		nodeType := node.Type()

		// Track which component we're in
		if nodeType == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName != "" && isReactComponent(componentName) {
				line, _ := node.StartPoint()
				componentID := GenerateComponentID(componentName, module.FilePath, line+1)
				if comp, exists := b.graph.ComponentNodes[componentID]; exists {
					currentComponent = comp
				}
			}
		}

		// Look for JSX elements (child component usage)
		if (nodeType == "jsx_element" || nodeType == "jsx_self_closing_element") && currentComponent != nil {
			b.processJSXElement(node, currentComponent, module.FilePath)
		}

		return true
	})

	return nil
}

// processJSXElement extracts prop passing from a JSX element
func (b *Builder) processJSXElement(jsxNode *parser.Node, parentComp *ComponentNode, filePath string) {
	// Get component name being rendered
	childComponentName := b.getJSXComponentName(jsxNode)
	if childComponentName == "" || !isReactComponent(childComponentName) {
		return
	}

	// Find the child component node
	childCompID := b.findComponentID(childComponentName, filePath)
	if childCompID == "" {
		return
	}

	childComp, exists := b.graph.ComponentNodes[childCompID]
	if !exists {
		return
	}

	// Get opening element to extract attributes
	var openingElement *parser.Node
	if jsxNode.Type() == "jsx_self_closing_element" {
		openingElement = jsxNode
	} else if jsxNode.Type() == "jsx_element" {
		for _, child := range jsxNode.Children() {
			if child.Type() == "jsx_opening_element" {
				openingElement = child
				break
			}
		}
	}

	if openingElement == nil {
		return
	}

	// Extract props being passed
	propsPassedToChild := []string{}

	// Iterate through direct children of opening element for attributes
	for _, child := range openingElement.Children() {
		if child.Type() != "jsx_attribute" {
			continue
		}

		// Get attribute name (it's a property_identifier child)
		var propName string
		var valueNode *parser.Node
		for _, attrChild := range child.Children() {
			if attrChild.Type() == "property_identifier" {
				propName = attrChild.Text()
			} else if attrChild.Type() == "jsx_expression" {
				valueNode = attrChild
			}
		}

		if propName == "" || valueNode == nil {
			continue
		}

		if valueNode.Type() == "jsx_expression" {
			// Check if value is a reference to parent's prop
			for _, exprChild := range valueNode.Children() {
				if exprChild.Type() == "identifier" {
					varName := exprChild.Text()

					// Check if this identifier matches a prop or state from parent
					if b.isParentVariable(varName, parentComp) {
						propsPassedToChild = append(propsPassedToChild, propName)

						// Create "passes" edge (include propName in ID to allow multiple props)
						line, col := child.StartPoint()
						edge := Edge{
							ID:       GenerateEdgeIDWithProp(EdgeTypePasses, parentComp.ID, childComp.ID, propName),
							SourceID: parentComp.ID,
							TargetID: childComp.ID,
							Type:     EdgeTypePasses,
							PropName: propName,
							Location: Location{FilePath: filePath, Line: line + 1, Column: col, Component: parentComp.Name},
						}
						b.graph.AddEdge(edge)
					}
				}
			}
		}
	}

	// Update parent's PropsPassedTo map
	if len(propsPassedToChild) > 0 {
		parentComp.PropsPassedTo[childComp.ID] = propsPassedToChild
	}
}

// isParentProp checks if a variable name is a prop of the parent component
func (b *Builder) isParentProp(varName string, parentComp *ComponentNode) bool {
	for _, prop := range parentComp.Props {
		if prop.Name == varName {
			return true
		}
	}
	return false
}

// isParentState checks if a variable name is a state variable defined by the parent component
func (b *Builder) isParentState(varName string, parentComp *ComponentNode) bool {
	// Check if any state nodes defined by this component have this name
	for _, stateID := range parentComp.StateNodes {
		if stateNode, exists := b.graph.StateNodes[stateID]; exists {
			if stateNode.Name == varName {
				return true
			}
		}
	}
	return false
}

// isParentVariable checks if a variable is either a prop or state of the parent
func (b *Builder) isParentVariable(varName string, parentComp *ComponentNode) bool {
	return b.isParentProp(varName, parentComp) || b.isParentState(varName, parentComp)
}

// Helper methods

func (b *Builder) getComponentNameFromFunction(node *parser.Node) string {
	if node.Type() == "function_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			return nameNode.Text()
		}
	}
	return ""
}

func (b *Builder) extractProps(funcNode *parser.Node) []PropDefinition {
	var props []PropDefinition

	// Get parameters node
	params := funcNode.ChildByFieldName("parameters")
	if params == nil {
		return props
	}

	// Iterate through formal parameters
	for _, param := range params.Children() {
		if param.Type() == "required_parameter" || param.Type() == "optional_parameter" {
			// Get the pattern (could be identifier or object_pattern for destructuring)
			pattern := param.ChildByFieldName("pattern")
			if pattern == nil {
				continue
			}

			// Handle destructured props: function Foo({ prop1, prop2 }: Props)
			if pattern.Type() == "object_pattern" {
				line, col := pattern.StartPoint()
				for _, child := range pattern.Children() {
					if child.Type() == "shorthand_property_identifier_pattern" {
						propName := child.Text()
						props = append(props, PropDefinition{
							Name:     propName,
							Type:     DataTypeUnknown, // TODO: Infer from type annotation
							Required: true,            // TODO: Determine from type annotation
							Location: Location{Line: line + 1, Column: col},
						})
					} else if child.Type() == "pair_pattern" {
						// Handle renamed props: { prop: localName }
						keyNode := child.ChildByFieldName("key")
						if keyNode != nil {
							propName := keyNode.Text()
							props = append(props, PropDefinition{
								Name:     propName,
								Type:     DataTypeUnknown,
								Required: true,
								Location: Location{Line: line + 1, Column: col},
							})
						}
					}
				}
			}

			// Handle non-destructured props: function Foo(props: Props)
			// We'll track this as a single "props" object
			if pattern.Type() == "identifier" {
				propName := pattern.Text()
				line, col := pattern.StartPoint()
				props = append(props, PropDefinition{
					Name:     propName,
					Type:     DataTypeObject,
					Required: true,
					Location: Location{Line: line + 1, Column: col},
				})
			}
		}
	}

	return props
}

func (b *Builder) getJSXComponentName(node *parser.Node) string {
	// Get opening element
	var openingElement *parser.Node
	if node.Type() == "jsx_self_closing_element" {
		openingElement = node
	} else if node.Type() == "jsx_element" {
		for _, child := range node.Children() {
			if child.Type() == "jsx_opening_element" {
				openingElement = child
				break
			}
		}
	}

	if openingElement == nil {
		return ""
	}

	// Find identifier
	for _, child := range openingElement.Children() {
		if child.Type() == "identifier" || child.Type() == "jsx_identifier" {
			return child.Text()
		}
	}

	return ""
}

func (b *Builder) findComponentID(componentName, filePath string) string {
	// Search for component in the same file first
	for id, comp := range b.graph.ComponentNodes {
		if comp.Name == componentName && comp.Location.FilePath == filePath {
			return id
		}
	}

	// TODO: Search in imported files using resolver
	// This requires tracking imports and resolving component definitions

	return ""
}

func (b *Builder) handleUseStateWithPattern(useStateNode *parser.Node, pattern *parser.Node, component *ComponentNode, filePath string) {
	// Extract state variable name from pattern: [count, setCount]

	line, col := useStateNode.StartPoint()

	stateName := "state" // Default placeholder

	if pattern != nil && pattern.Type() == "array_pattern" {
		// Get first element (the state variable, not the setter)
		for _, child := range pattern.Children() {
			if child.Type() == "identifier" {
				stateName = child.Text()
				break // Use first identifier
			}
		}
	}

	stateID := GenerateStateID(component.Name, stateName, filePath, line+1)

	stateNode := &StateNode{
		ID:       stateID,
		Name:     stateName,
		Type:     StateTypeUseState,
		DataType: DataTypeUnknown, // TODO: Infer from initial value
		Location: Location{FilePath: filePath, Line: line + 1, Column: col, Component: component.Name},
		Mutable:  true,
	}

	b.graph.AddStateNode(stateNode)
	component.StateNodes = append(component.StateNodes, stateID)

	// Add "defines" edge
	edge := Edge{
		ID:       GenerateEdgeID(EdgeTypeDefines, component.ID, stateID),
		SourceID: component.ID,
		TargetID: stateID,
		Type:     EdgeTypeDefines,
		Location: Location{FilePath: filePath, Line: line + 1, Column: col, Component: component.Name},
	}
	b.graph.AddEdge(edge)
}

// isReactComponent checks if a name follows React component naming (PascalCase)
func isReactComponent(name string) bool {
	if len(name) == 0 {
		return false
	}
	// React components must start with uppercase letter
	return name[0] >= 'A' && name[0] <= 'Z'
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
