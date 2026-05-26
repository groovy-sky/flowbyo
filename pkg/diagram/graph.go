package diagram

import (
	"encoding/json"
	"fmt"
	"os"
)

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
