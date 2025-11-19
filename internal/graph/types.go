package graph

// Location represents a position in source code
type Location struct {
	FilePath  string `json:"filePath"`
	Line      uint32 `json:"line"`
	Column    uint32 `json:"column"`
	Component string `json:"component"` // Name of the containing component
}

// StateType represents the type of state node
type StateType string

const (
	StateTypeUseState   StateType = "useState"
	StateTypeUseReducer StateType = "useReducer"
	StateTypeContext    StateType = "context"
	StateTypeProp       StateType = "prop"
	StateTypeDerived    StateType = "derived"
)

// DataType represents the type of data stored in state
type DataType string

const (
	DataTypePrimitive DataType = "primitive"
	DataTypeObject    DataType = "object"
	DataTypeArray     DataType = "array"
	DataTypeFunction  DataType = "function"
	DataTypeUnknown   DataType = "unknown"
)

// StateNode represents a piece of state in the application
type StateNode struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            StateType  `json:"type"`
	DataType        DataType   `json:"dataType"`
	Location        Location   `json:"location"`
	Mutable         bool       `json:"mutable"`
	Dependencies    []string   `json:"dependencies"`    // IDs of state nodes this depends on
	UpdateLocations []Location `json:"updateLocations"` // Where this state is updated
}

// ComponentType represents the type of React component
type ComponentType string

const (
	ComponentTypeFunction ComponentType = "function"
	ComponentTypeClass    ComponentType = "class"
	ComponentTypeMemo     ComponentType = "memo"
)

// PropDefinition represents a prop defined on a component
type PropDefinition struct {
	Name     string   `json:"name"`
	Type     DataType `json:"type"`
	Required bool     `json:"required"`
	Location Location `json:"location"`
}

// ComponentNode represents a React component
type ComponentNode struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Type          ComponentType       `json:"type"`
	Location      Location            `json:"location"`
	IsMemoized    bool                `json:"isMemoized"`
	StateNodes    []string            `json:"stateNodes"`    // IDs of state defined in this component
	ConsumedState []string            `json:"consumedState"` // IDs of state consumed by this component
	Parent        string              `json:"parent"`        // ID of parent component
	Children      []string            `json:"children"`      // IDs of child components
	Props         []PropDefinition    `json:"props"`
	PropsPassedTo map[string][]string `json:"propsPassedTo"` // childComponentID -> []propNames
}

// EdgeType represents the type of relationship between nodes
type EdgeType string

const (
	EdgeTypeDefines  EdgeType = "defines"  // Component defines state
	EdgeTypeConsumes EdgeType = "consumes" // Component consumes state
	EdgeTypeUpdates  EdgeType = "updates"  // Component updates state
	EdgeTypePasses   EdgeType = "passes"   // Component passes prop to child
	EdgeTypeDerives  EdgeType = "derives"  // State derives from other state
)

// Edge represents a directed relationship in the graph
type Edge struct {
	ID       string   `json:"id"`
	SourceID string   `json:"sourceId"` // ID of source node
	TargetID string   `json:"targetId"` // ID of target node
	Type     EdgeType `json:"type"`
	Weight   float64  `json:"weight"`   // For future use (e.g., render frequency)
	PropName string   `json:"propName"` // For passes edges
	Location Location `json:"location"`
}

// Graph represents the complete state dependency graph
type Graph struct {
	StateNodes     map[string]*StateNode     `json:"stateNodes"`
	ComponentNodes map[string]*ComponentNode `json:"componentNodes"`
	Edges          []Edge                    `json:"edges"`
	EdgesBySource  map[string][]Edge         `json:"-"` // Index for fast lookup
	EdgesByTarget  map[string][]Edge         `json:"-"` // Index for fast lookup
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		StateNodes:     make(map[string]*StateNode),
		ComponentNodes: make(map[string]*ComponentNode),
		Edges:          []Edge{},
		EdgesBySource:  make(map[string][]Edge),
		EdgesByTarget:  make(map[string][]Edge),
	}
}
