package diagram

// Coordinate represents [x, y] position in matrix
type Coordinate [2]int

// Element represents a block/node in the diagram
type Element struct {
	ID          string         `yaml:"id"`
	Label       string         `yaml:"label"`
	Type        string         `yaml:"type"`
	Links       []Link         `yaml:"links,omitempty"`
	Matrix      Matrix         `yaml:"matrix,omitempty"`
	Description string         `yaml:"description,omitempty"`
	Style       map[string]any `yaml:"style,omitempty"`
}

// Link represents an edge from source element to target coordinate
type Link struct {
	To        Coordinate     `yaml:"to"`
	Label     string         `yaml:"label,omitempty"`
	Condition string         `yaml:"condition,omitempty"`
	EdgeType  string         `yaml:"edge_type,omitempty"`
	Style     map[string]any `yaml:"style,omitempty"`
}

// Cell is either nil or an Element
type Cell *Element

// Row is a slice of cells
type Row []Cell

// Matrix is a 2D grid of cells
type Matrix []Row

// Meta contains document metadata
type Meta struct {
	Name        string `yaml:"name"`
	DiagramType string `yaml:"diagram_type"`
	Version     string `yaml:"version"`
}

// View contains rendering settings
type View struct {
	Direction      string `yaml:"direction,omitempty"`
	Theme          string `yaml:"theme,omitempty"`
	CellSize       []int  `yaml:"cell_size,omitempty"`
	Padding        int    `yaml:"padding,omitempty"`
	Routing        string `yaml:"routing,omitempty"`
	CollapseNested bool   `yaml:"collapse_nested,omitempty"`
}

// Diagram is the top-level document
type Diagram struct {
	Meta   Meta        `yaml:"meta"`
	Graph  *GraphModel `yaml:"graph,omitempty" json:"graph,omitempty"`
	Matrix Matrix      `yaml:"matrix,omitempty"` // legacy input compatibility only
	View   View        `yaml:"view,omitempty"`
}

// RuntimeElement extends Element with computed fields
type RuntimeElement struct {
	Element
	Coord    Coordinate
	Scope    string // scope identifier for the matrix containing this element
	ParentID string // parent element id if nested
	Children []RuntimeElement
}

// RuntimeModel represents parsed and validated diagram at runtime
type RuntimeModel struct {
	Meta          Meta
	Root          Matrix
	Elements      map[string]*RuntimeElement // id -> element
	Graph         *GraphModel
	ScopeIndex    map[string]map[Coordinate]string // scope -> coord -> element id
	ScopeMatrices map[string]Matrix                // scope -> matrix data
	View          View
	Errors        []ValidationError
	Warnings      []ValidationError
}

// GraphNode represents a node in the graph projection.
type GraphNode struct {
	ID          string         `yaml:"id" json:"id"`
	Label       string         `yaml:"label" json:"label"`
	Type        string         `yaml:"type" json:"type"`
	Scope       string         `yaml:"scope" json:"scope"`
	ParentID    string         `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
	Coord       *Coordinate    `yaml:"coord,omitempty" json:"coord,omitempty"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Style       map[string]any `yaml:"style,omitempty" json:"style,omitempty"`
}

// GraphEdge represents an edge in the graph projection.
type GraphEdge struct {
	FromID    string         `yaml:"from" json:"from"`
	ToID      string         `yaml:"to" json:"to"`
	Scope     string         `yaml:"scope" json:"scope"`
	Label     string         `yaml:"label,omitempty" json:"label,omitempty"`
	Condition string         `yaml:"condition,omitempty" json:"condition,omitempty"`
	EdgeType  string         `yaml:"edge_type,omitempty" json:"edge_type,omitempty"`
	Style     map[string]any `yaml:"style,omitempty" json:"style,omitempty"`
}

// GraphModel is a hierarchical graph representation of the diagram.
type GraphModel struct {
	Nodes map[string]*GraphNode `yaml:"nodes" json:"nodes"`
	Edges []GraphEdge           `yaml:"edges,omitempty" json:"edges,omitempty"`
}

// ValidationError represents a validation issue
type ValidationError struct {
	Level   string // "fatal", "warning", "info"
	Message string
	Element string // element id if applicable
	Coord   *Coordinate
}
