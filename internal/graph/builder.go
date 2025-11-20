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
	// Phase 0: Discover and parse all imported files recursively
	b.discoverAndParseImports()

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

// discoverAndParseImports recursively discovers and parses all imported files
// This ensures the graph builder has all necessary modules even when analyzing a single file
// Currently only traverses downward (files imported by the current file)
// TODO: Implement reverse import index for upward traversal (files that import the current file)
// This would enable prop drilling detection when starting from leaf components
func (b *Builder) discoverAndParseImports() {
	visited := make(map[string]bool)

	// Start with all currently loaded modules
	modules := b.resolver.GetModules()

	for filePath := range modules {
		// Traverse down: files this file imports
		b.discoverImportsRecursive(filePath, visited)
	}
}

// discoverImportsRecursive recursively parses imported files (downward traversal)
func (b *Builder) discoverImportsRecursive(filePath string, visited map[string]bool) {
	// Skip if already visited
	if visited[filePath] {
		return
	}
	visited[filePath] = true

	// Get the module (should already be parsed)
	modules := b.resolver.GetModules()
	module, exists := modules[filePath]
	if !exists {
		return
	}

	// Parse all imports recursively
	for _, imp := range module.Imports {
		// Try to resolve the import
		resolvedPath, err := b.resolver.Resolve(filePath, imp.Source)
		if err != nil {
			// Skip unresolvable imports (external packages, etc.)
			continue
		}

		// Check if this module is already parsed
		if _, exists := modules[resolvedPath]; !exists {
			// Parse the imported file
			if _, err := b.resolver.GetModule(resolvedPath); err != nil {
				// Skip files that fail to parse
				continue
			}
			// Refresh modules map after parsing
			modules = b.resolver.GetModules()
		}

		// Recursively discover imports in this file
		b.discoverImportsRecursive(resolvedPath, visited)
	}
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

			// Extract prop names for local usage analysis
			propNames := make([]string, len(props))
			for i, prop := range props {
				propNames[i] = prop.Name
			}

			// Analyze which props are used locally (not just passed to children)
			propsUsedLocally := b.findPropsUsedLocally(node, propNames)

			compNode := &ComponentNode{
				ID:               componentID,
				Name:             componentName,
				Type:             ComponentTypeFunction,
				Location:         Location{FilePath: module.FilePath, Line: line + 1, Column: col, Component: componentName},
				IsMemoized:       isMemoized,
				StateNodes:       []string{},
				ConsumedState:    []string{},
				Children:         []string{},
				Props:            props,
				PropsPassedTo:    make(map[string][]string),
				PropsUsedLocally: propsUsedLocally,
			}

			b.graph.AddComponentNode(compNode)
		}

		// Handle arrow function components (const Foo = () => {})
		if nodeType == "variable_declarator" {
			componentName := b.getComponentNameFromArrowFunction(node)
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

			// Extract props from arrow function
			arrowFunc := b.getArrowFunctionNode(node)
			if arrowFunc == nil {
				return true
			}
			props := b.extractPropsFromArrowFunction(arrowFunc)

			// Extract prop names for local usage analysis
			propNames := make([]string, len(props))
			for i, prop := range props {
				propNames[i] = prop.Name
			}

			// Analyze which props are used locally (not just passed to children)
			propsUsedLocally := b.findPropsUsedLocally(node, propNames)

			compNode := &ComponentNode{
				ID:               componentID,
				Name:             componentName,
				Type:             ComponentTypeFunction,
				Location:         Location{FilePath: module.FilePath, Line: line + 1, Column: col, Component: componentName},
				IsMemoized:       isMemoized,
				StateNodes:       []string{},
				ConsumedState:    []string{},
				Children:         []string{},
				Props:            props,
				PropsPassedTo:    make(map[string][]string),
				PropsUsedLocally: propsUsedLocally,
			}

			b.graph.AddComponentNode(compNode)
		}

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

	// Use helper to process nodes with proper component context
	b.walkWithComponentContext(ast, module.FilePath, func(node *parser.Node, currentComponent *ComponentNode) bool {
		nodeType := node.Type()

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

	// Use helper to process nodes with proper component context
	b.walkWithComponentContext(ast, module.FilePath, func(node *parser.Node, currentComponent *ComponentNode) bool {
		nodeType := node.Type()

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

	// Use helper to process nodes with proper component context
	b.walkWithComponentContext(ast, module.FilePath, func(node *parser.Node, currentComponent *ComponentNode) bool {
		nodeType := node.Type()

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
		// Handle regular attributes: <Child prop={value} />
		if child.Type() == "jsx_attribute" {
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
					// Handle simple identifier: <Child theme={theme} />
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

					// Handle member expression: <Child locale={settings.locale} />
					if exprChild.Type() == "member_expression" {
						objectName, propertyName := b.extractMemberExpression(exprChild)
						if objectName != "" && propertyName != "" {
							// Check if the object is a parent variable (state or prop)
							if b.isParentVariable(objectName, parentComp) {
								// Create or find virtual state node for this property
								// This allows prop drilling detection to trace from object.property
								b.ensurePropertyStateNode(objectName, propertyName, parentComp, filePath)

								propsPassedToChild = append(propsPassedToChild, propName)

								// Create "passes" edge with the property name
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
		}

		// Handle spread attributes: <Child {...props} />
		if child.Type() == "jsx_spread_attribute" {
			b.handleSpreadAttribute(child, parentComp, childComp, filePath, &propsPassedToChild)
		}
	}

	// Update parent's PropsPassedTo map
	if len(propsPassedToChild) > 0 {
		parentComp.PropsPassedTo[childComp.ID] = propsPassedToChild
	}
}

// extractMemberExpression extracts object and property names from a member_expression node
// Example: settings.locale -> ("settings", "locale")
// For nested expressions (config.settings.theme), returns the root object and full property path
func (b *Builder) extractMemberExpression(memberExpr *parser.Node) (objectName string, propertyName string) {
	if memberExpr.Type() != "member_expression" {
		return "", ""
	}

	// Get the object (left side)
	object := memberExpr.ChildByFieldName("object")
	if object != nil && object.Type() == "identifier" {
		objectName = object.Text()
	}

	// Get the property (right side)
	property := memberExpr.ChildByFieldName("property")
	if property != nil && property.Type() == "property_identifier" {
		propertyName = property.Text()
	}

	return objectName, propertyName
}

// ensurePropertyStateNode creates a virtual state node for an object property
// This enables prop drilling detection to track properties like settings.locale as separate flows
func (b *Builder) ensurePropertyStateNode(objectName string, propertyName string, component *ComponentNode, filePath string) {
	// Generate ID for the virtual state node (using property name as the state name)
	virtualStateID := GenerateStateID(component.Name, propertyName, filePath, component.Location.Line)

	// Check if this virtual state already exists
	if _, exists := b.graph.StateNodes[virtualStateID]; exists {
		return // Already created
	}

	// Find the parent state node (the object)
	var parentStateID string
	for _, stateID := range component.StateNodes {
		if stateNode, exists := b.graph.StateNodes[stateID]; exists {
			if stateNode.Name == objectName {
				parentStateID = stateID
				break
			}
		}
	}

	if parentStateID == "" {
		// Parent state not found, can't create virtual state
		return
	}

	// Create virtual state node for the property
	virtualStateNode := &StateNode{
		ID:       virtualStateID,
		Name:     propertyName,
		Type:     StateTypeDerived, // Mark as derived from parent object
		DataType: DataTypeUnknown,
		Location: component.Location,
		Mutable:  false, // Properties accessed via member expression are typically read-only
	}

	b.graph.AddStateNode(virtualStateNode)
	component.StateNodes = append(component.StateNodes, virtualStateID)

	// Create "derives" edge from parent state to virtual state
	derivesEdge := Edge{
		ID:       GenerateEdgeID(EdgeTypeDerives, parentStateID, virtualStateID),
		SourceID: parentStateID,
		TargetID: virtualStateID,
		Type:     EdgeTypeDerives,
		Location: component.Location,
	}
	b.graph.AddEdge(derivesEdge)

	// Create "defines" edge from component to virtual state
	definesEdge := Edge{
		ID:       GenerateEdgeID(EdgeTypeDefines, component.ID, virtualStateID),
		SourceID: component.ID,
		TargetID: virtualStateID,
		Type:     EdgeTypeDefines,
		Location: component.Location,
	}
	b.graph.AddEdge(definesEdge)
}

// handleSpreadAttribute processes JSX spread attributes like <Child {...props} />
func (b *Builder) handleSpreadAttribute(spreadNode *parser.Node, parentComp *ComponentNode, childComp *ComponentNode, filePath string, propsPassedToChild *[]string) {
	// Get the expression being spread
	// JSX spread attribute structure: jsx_spread_attribute -> ... -> identifier
	var spreadIdentifier string

	// Walk children to find the identifier
	for _, child := range spreadNode.Children() {
		if child.Type() == "identifier" {
			spreadIdentifier = child.Text()
			break
		}
		// Also check nested children (for cases like {...})
		for _, grandchild := range child.Children() {
			if grandchild.Type() == "identifier" {
				spreadIdentifier = grandchild.Text()
				break
			}
		}
		if spreadIdentifier != "" {
			break
		}
	}

	if spreadIdentifier == "" {
		return
	}

	// Check if the spread identifier is the parent's props object
	// This handles cases like: function Parent(props) { return <Child {...props} /> }
	isParentPropsObject := false
	for _, prop := range parentComp.Props {
		if prop.Name == spreadIdentifier && prop.Type == DataTypeObject {
			isParentPropsObject = true
			break
		}
	}

	// Check if it's a parent's state or prop variable
	isParentVariable := b.isParentVariable(spreadIdentifier, parentComp)

	if isParentPropsObject {
		// Spreading the entire props object - we assume all props are being passed
		// Since we can't know which individual props without more analysis,
		// we create edges for all of the child's props
		for _, childProp := range childComp.Props {
			// Skip object-type props
			if childProp.Type == DataTypeObject {
				continue
			}

			*propsPassedToChild = append(*propsPassedToChild, childProp.Name)

			// Create edge for this prop
			line, col := spreadNode.StartPoint()
			edge := Edge{
				ID:       GenerateEdgeIDWithProp(EdgeTypePasses, parentComp.ID, childComp.ID, childProp.Name),
				SourceID: parentComp.ID,
				TargetID: childComp.ID,
				Type:     EdgeTypePasses,
				PropName: childProp.Name,
				Location: Location{FilePath: filePath, Line: line + 1, Column: col, Component: parentComp.Name},
			}
			b.graph.AddEdge(edge)
		}
	} else if isParentVariable {
		// Spreading a parent's state or prop (e.g., {...config} where config is state)
		// We assume all properties are being passed to all of child's props
		for _, childProp := range childComp.Props {
			// Skip object-type props
			if childProp.Type == DataTypeObject {
				continue
			}

			*propsPassedToChild = append(*propsPassedToChild, childProp.Name)

			// Create edge for this prop
			line, col := spreadNode.StartPoint()
			edge := Edge{
				ID:       GenerateEdgeIDWithProp(EdgeTypePasses, parentComp.ID, childComp.ID, childProp.Name),
				SourceID: parentComp.ID,
				TargetID: childComp.ID,
				Type:     EdgeTypePasses,
				PropName: childProp.Name,
				Location: Location{FilePath: filePath, Line: line + 1, Column: col, Component: parentComp.Name},
			}
			b.graph.AddEdge(edge)
		}
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

	// Search in imported files using resolver
	return b.findComponentInImports(componentName, filePath)
}

// findComponentInImports searches for a component in imported files
func (b *Builder) findComponentInImports(componentName, currentFile string) string {
	// Get the current module to access its imports
	modules := b.resolver.GetModules()
	currentModule, exists := modules[currentFile]
	if !exists {
		return ""
	}

	// Search through all imports
	for _, imp := range currentModule.Imports {
		// Check if this import includes the component we're looking for

		// Check default import: import MyComponent from './MyComponent'
		if imp.Default == componentName {
			resolvedPath, err := b.resolver.Resolve(currentFile, imp.Source)
			if err != nil {
				continue // Skip unresolved imports (external packages, etc.)
			}

			// Find component in the imported file
			componentID := b.findComponentInFile(componentName, resolvedPath)
			if componentID != "" {
				return componentID
			}
		}

		// Check named imports: import { MyComponent } from './components'
		for _, named := range imp.Named {
			// LocalName is what it's called in the current file
			// ImportedName is what it's called in the source file
			if named.LocalName == componentName {
				resolvedPath, err := b.resolver.Resolve(currentFile, imp.Source)
				if err != nil {
					continue
				}

				// Search for the original exported name in the imported file
				componentID := b.findComponentInFile(named.ImportedName, resolvedPath)
				if componentID != "" {
					return componentID
				}
			}
		}

		// Check namespace imports: import * as Components from './components'
		// Usage: <Components.MyComponent />
		// This would require parsing the member access, which we'll handle separately
	}

	return ""
}

// findComponentInFile searches for a component by name in a specific file
// If the file hasn't been parsed yet, it will be parsed automatically
func (b *Builder) findComponentInFile(componentName, filePath string) string {
	// First check if component already exists in graph
	for id, comp := range b.graph.ComponentNodes {
		if comp.Name == componentName && comp.Location.FilePath == filePath {
			return id
		}
	}

	// Component not found - the file might not have been parsed yet
	// Try to parse it now
	modules := b.resolver.GetModules()
	if _, exists := modules[filePath]; !exists {
		// File not yet parsed - parse it now
		module, err := b.resolver.GetModule(filePath)
		if err != nil {
			// Failed to parse, can't find component
			return ""
		}

		// Now that we've parsed the file, build components from it
		if module != nil {
			b.buildComponentsFromModule(filePath, module)

			// Search again in the newly built components
			for id, comp := range b.graph.ComponentNodes {
				if comp.Name == componentName && comp.Location.FilePath == filePath {
					return id
				}
			}
		}
	}

	return ""
}

// buildComponentsFromModule builds component nodes from a specific module
// This is used when we need to lazily load an imported file
func (b *Builder) buildComponentsFromModule(filePath string, module *analyzer.Module) {
	// Build component nodes
	b.buildComponentNodes(module)
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

// getComponentNameFromArrowFunction extracts component name from arrow function variable declarator
func (b *Builder) getComponentNameFromArrowFunction(node *parser.Node) string {
	if node.Type() != "variable_declarator" {
		return ""
	}

	// Get the name field (identifier)
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return ""
	}

	// Must be an identifier
	if nameNode.Type() != "identifier" {
		return ""
	}

	componentName := nameNode.Text()

	// Check if value is an arrow function (or React.memo wrapping arrow function)
	valueNode := node.ChildByFieldName("value")
	if valueNode == nil {
		return ""
	}

	// Check if it's directly an arrow function or wrapped in React.memo/memo
	if !b.isArrowFunctionComponent(valueNode) {
		return ""
	}

	return componentName
}

// isArrowFunctionComponent checks if a node is an arrow function or React.memo wrapping one
func (b *Builder) isArrowFunctionComponent(node *parser.Node) bool {
	if node == nil {
		return false
	}

	// Direct arrow function
	if node.Type() == "arrow_function" {
		return true
	}

	// React.memo(...) or memo(...) wrapping arrow function
	if node.Type() == "call_expression" {
		// Get the function being called
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			return false
		}

		funcText := funcNode.Text()
		// Check if it's memo or React.memo
		if funcText == "memo" || funcText == "React.memo" {
			// Get arguments
			argsNode := node.ChildByFieldName("arguments")
			if argsNode == nil {
				return false
			}

			// Check if first argument is arrow function
			for _, child := range argsNode.Children() {
				if child.Type() == "arrow_function" {
					return true
				}
			}
		}
	}

	return false
}

// getArrowFunctionNode gets the arrow function node, unwrapping React.memo if needed
func (b *Builder) getArrowFunctionNode(node *parser.Node) *parser.Node {
	if node.Type() != "variable_declarator" {
		return nil
	}

	valueNode := node.ChildByFieldName("value")
	if valueNode == nil {
		return nil
	}

	// Direct arrow function
	if valueNode.Type() == "arrow_function" {
		return valueNode
	}

	// React.memo(...) wrapping arrow function
	if valueNode.Type() == "call_expression" {
		argsNode := valueNode.ChildByFieldName("arguments")
		if argsNode == nil {
			return nil
		}

		// Return first arrow function argument
		for _, child := range argsNode.Children() {
			if child.Type() == "arrow_function" {
				return child
			}
		}
	}

	return nil
}

// extractPropsFromArrowFunction extracts prop definitions from arrow function parameters
func (b *Builder) extractPropsFromArrowFunction(arrowFunc *parser.Node) []PropDefinition {
	var props []PropDefinition

	if arrowFunc.Type() != "arrow_function" {
		return props
	}

	// Get parameters node
	params := arrowFunc.ChildByFieldName("parameters")
	if params == nil {
		return props
	}

	// Iterate through parameters (could be formal_parameters or just parenthesized expression)
	for _, param := range params.Children() {
		// Handle different parameter types
		paramType := param.Type()

		// Required or optional parameter (TypeScript)
		if paramType == "required_parameter" || paramType == "optional_parameter" {
			pattern := param.ChildByFieldName("pattern")
			if pattern != nil {
				props = append(props, b.extractPropsFromPattern(pattern)...)
			}
		}

		// Identifier (plain JavaScript: const Foo = (props) => ...)
		if paramType == "identifier" {
			propName := param.Text()
			line, col := param.StartPoint()
			props = append(props, PropDefinition{
				Name:     propName,
				Type:     DataTypeObject,
				Required: true,
				Location: Location{Line: line + 1, Column: col},
			})
		}

		// Object pattern (destructured props: const Foo = ({ prop1, prop2 }) => ...)
		if paramType == "object_pattern" {
			props = append(props, b.extractPropsFromPattern(param)...)
		}
	}

	return props
}

// extractPropsFromPattern extracts props from a pattern node (shared between function and arrow function)
func (b *Builder) extractPropsFromPattern(pattern *parser.Node) []PropDefinition {
	var props []PropDefinition

	if pattern == nil {
		return props
	}

	// Handle destructured props: { prop1, prop2 }
	if pattern.Type() == "object_pattern" {
		line, col := pattern.StartPoint()
		for _, child := range pattern.Children() {
			if child.Type() == "shorthand_property_identifier_pattern" {
				propName := child.Text()
				props = append(props, PropDefinition{
					Name:     propName,
					Type:     DataTypeUnknown,
					Required: true,
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

	// Handle non-destructured props: props
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

	return props
}

// findPropsUsedLocally analyzes a component's AST to determine which props are referenced locally
// (not just passed to children). This is used to distinguish between pure passthrough components
// and components that actually use their props.
func (b *Builder) findPropsUsedLocally(componentNode *parser.Node, propNames []string) []string {
	if len(propNames) == 0 {
		return []string{}
	}

	// Build a set of prop names for quick lookup
	propSet := make(map[string]bool)
	for _, prop := range propNames {
		propSet[prop] = true
	}

	propsUsed := make(map[string]bool)

	// Find the function body
	var body *parser.Node
	if componentNode.Type() == "function_declaration" {
		body = componentNode.ChildByFieldName("body")
	} else if componentNode.Type() == "variable_declarator" {
		// Arrow function case
		arrowFunc := b.getArrowFunctionNode(componentNode)
		if arrowFunc != nil {
			body = arrowFunc.ChildByFieldName("body")
		}
	}

	if body == nil {
		return []string{}
	}

	// Walk the body and find prop references
	var walk func(*parser.Node, bool)
	walk = func(node *parser.Node, inJSXAttributeValue bool) {
		nodeType := node.Type()

		// Check if we're entering a JSX attribute value context
		// JSX attribute values are where props are passed to children
		newInJSXAttributeValue := inJSXAttributeValue
		if nodeType == "jsx_attribute" {
			// We're in a JSX attribute, mark that child nodes are in attribute value context
			newInJSXAttributeValue = true
		}

		// If this is an identifier that matches a prop name
		if nodeType == "identifier" {
			propName := node.Text()
			if propSet[propName] {
				// Only count it as "used locally" if we're NOT in a JSX attribute value
				// where it's being passed to a child
				if !inJSXAttributeValue {
					propsUsed[propName] = true
				}
			}
		}

		// Recursively walk children
		for _, child := range node.Children() {
			walk(child, newInJSXAttributeValue)
		}
	}

	walk(body, false)

	// Convert map to slice
	result := []string{}
	for prop := range propsUsed {
		result = append(result, prop)
	}

	return result
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

// walkWithComponentContext walks the AST while tracking the current component context
// The callback receives each node and the current component (nil if not inside a component)
func (b *Builder) walkWithComponentContext(ast *parser.AST, filePath string, callback func(*parser.Node, *ComponentNode) bool) {
	var componentStack []*ComponentNode

	var walk func(*parser.Node) bool
	walk = func(node *parser.Node) bool {
		// Check if we're entering a component (function declaration)
		if node.Type() == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName != "" && isReactComponent(componentName) {
				line, _ := node.StartPoint()
				componentID := GenerateComponentID(componentName, filePath, line+1)
				if comp, exists := b.graph.ComponentNodes[componentID]; exists {
					componentStack = append(componentStack, comp)
				}
			}
		}

		// Check if we're entering an arrow function component (via variable_declarator)
		if node.Type() == "variable_declarator" {
			componentName := b.getComponentNameFromArrowFunction(node)
			if componentName != "" && isReactComponent(componentName) {
				line, _ := node.StartPoint()
				componentID := GenerateComponentID(componentName, filePath, line+1)
				if comp, exists := b.graph.ComponentNodes[componentID]; exists {
					componentStack = append(componentStack, comp)
				}
			}
		}

		// Get current component (top of stack)
		var currentComponent *ComponentNode
		if len(componentStack) > 0 {
			currentComponent = componentStack[len(componentStack)-1]
		}

		// Call the user callback
		if !callback(node, currentComponent) {
			return false
		}

		// Walk children
		for _, child := range node.Children() {
			if !walk(child) {
				return false
			}
		}

		// Pop component stack when leaving component nodes
		if node.Type() == "function_declaration" {
			componentName := b.getComponentNameFromFunction(node)
			if componentName != "" && isReactComponent(componentName) && len(componentStack) > 0 {
				componentStack = componentStack[:len(componentStack)-1]
			}
		}

		if node.Type() == "variable_declarator" {
			componentName := b.getComponentNameFromArrowFunction(node)
			if componentName != "" && isReactComponent(componentName) && len(componentStack) > 0 {
				componentStack = componentStack[:len(componentStack)-1]
			}
		}

		return true
	}

	walk(ast.Root)
}
