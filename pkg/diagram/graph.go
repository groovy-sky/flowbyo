package diagram

import (
	"encoding/json"
	"fmt"
	"os"
)

// GenerateMatrixFromGraph reconstructs nested matrices from a graph projection.
func GenerateMatrixFromGraph(graph *GraphModel) (Matrix, error) {
	matrixByScope, err := buildMatricesByScope(graph)
	if err != nil {
		return nil, err
	}

	root, ok := matrixByScope["root"]
	if !ok {
		return nil, fmt.Errorf("graph has no root scope")
	}

	return root, nil
}

// GenerateMatricesByScopeFromGraph reconstructs all scope matrices from graph.
func GenerateMatricesByScopeFromGraph(graph *GraphModel) (Matrix, map[string]Matrix, error) {
	matrixByScope, err := buildMatricesByScope(graph)
	if err != nil {
		return nil, nil, err
	}

	root, ok := matrixByScope["root"]
	if !ok {
		return nil, nil, fmt.Errorf("graph has no root scope")
	}

	return root, matrixByScope, nil
}

// BuildDiagramFromGraph creates a diagram from graph, preserving meta and view.
func BuildDiagramFromGraph(meta Meta, view View, graph *GraphModel) (*Diagram, error) {
	return &Diagram{
		Meta:  meta,
		Graph: graph,
		View:  view,
	}, nil
}

// SaveGraphJSON writes a graph model as indented JSON.
func SaveGraphJSON(graph *GraphModel, filePath string) error {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write graph JSON: %w", err)
	}

	return nil
}

// LoadGraphJSON loads a graph model from JSON.
func LoadGraphJSON(filePath string) (*GraphModel, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read graph JSON: %w", err)
	}

	var graph GraphModel
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graph JSON: %w", err)
	}

	if graph.Nodes == nil {
		graph.Nodes = make(map[string]*GraphNode)
	}

	return &graph, nil
}

func buildMatricesByScope(graph *GraphModel) (map[string]Matrix, error) {
	nodesByScope := make(map[string][]*GraphNode)
	elementByID := make(map[string]*Element)
	matrixByScope := make(map[string]Matrix)

	for id, node := range graph.Nodes {
		if node.Scope == "" {
			return nil, fmt.Errorf("node %s has empty scope", id)
		}
		if node.Coord == nil {
			return nil, fmt.Errorf("node %s is missing coordinates", id)
		}
		nodesByScope[node.Scope] = append(nodesByScope[node.Scope], node)
	}

	for scope, nodes := range nodesByScope {
		maxX := 0
		maxY := 0
		for _, n := range nodes {
			if n.Coord[0] < 0 || n.Coord[1] < 0 {
				return nil, fmt.Errorf("node %s has negative coordinate [%d, %d]", n.ID, n.Coord[0], n.Coord[1])
			}
			if n.Coord[0] > maxX {
				maxX = n.Coord[0]
			}
			if n.Coord[1] > maxY {
				maxY = n.Coord[1]
			}
		}

		m := make(Matrix, maxY+1)
		for y := range m {
			m[y] = make(Row, maxX+1)
		}

		for _, n := range nodes {
			x, y := n.Coord[0], n.Coord[1]
			if m[y][x] != nil {
				return nil, fmt.Errorf("scope %s has overlapping coordinates at [%d,%d]", scope, x, y)
			}

			elem := &Element{
				ID:          n.ID,
				Label:       n.Label,
				Type:        n.Type,
				Description: n.Description,
				Style:       n.Style,
			}

			m[y][x] = elem
			elementByID[n.ID] = elem
		}

		matrixByScope[scope] = m
	}

	for i, edge := range graph.Edges {
		fromNode, ok := graph.Nodes[edge.FromID]
		if !ok {
			return nil, fmt.Errorf("edge %d references unknown source node %s", i, edge.FromID)
		}
		toNode, ok := graph.Nodes[edge.ToID]
		if !ok {
			return nil, fmt.Errorf("edge %d references unknown target node %s", i, edge.ToID)
		}
		if fromNode.Scope != toNode.Scope {
			return nil, fmt.Errorf("edge %d crosses scopes (%s -> %s), which is not supported in MVP", i, fromNode.Scope, toNode.Scope)
		}

		fromElem := elementByID[edge.FromID]
		toCoord := *toNode.Coord
		fromElem.Links = append(fromElem.Links, Link{
			To:        toCoord,
			Label:     edge.Label,
			Condition: edge.Condition,
			EdgeType:  edge.EdgeType,
			Style:     edge.Style,
		})
	}

	for scope, m := range matrixByScope {
		if scope == "root" {
			continue
		}

		parentElem, ok := elementByID[scope]
		if !ok {
			return nil, fmt.Errorf("scope %s does not have matching parent element id", scope)
		}
		parentElem.Matrix = m
	}

	return matrixByScope, nil
}
