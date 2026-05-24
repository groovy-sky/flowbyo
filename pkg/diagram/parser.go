package diagram

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAML loads and unmarshals a YAML file into a Diagram
func ParseYAML(filePath string) (*Diagram, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var diagram Diagram
	err = yaml.Unmarshal(data, &diagram)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &diagram, nil
}

// BuildRuntimeModel parses diagram and builds runtime representation
func BuildRuntimeModel(diagram *Diagram) *RuntimeModel {
	model := &RuntimeModel{
		Meta:          diagram.Meta,
		Elements:      make(map[string]*RuntimeElement),
		ScopeIndex:    make(map[string]map[Coordinate]string),
		ScopeMatrices: make(map[string]Matrix),
		View:          diagram.View,
	}

	if diagram.Graph != nil && len(diagram.Graph.Nodes) > 0 {
		model.Graph = diagram.Graph

		root, matricesByScope, err := GenerateMatricesByScopeFromGraph(diagram.Graph)
		if err != nil {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("failed to derive matrix projection from graph: %v", err),
			})
			return model
		}

		model.Root = root
		model.ScopeMatrices = matricesByScope
		populateRuntimeFromGraph(model)
		return model
	}

	if len(diagram.Matrix) > 0 {
		model.Root = diagram.Matrix
		model.ScopeMatrices["root"] = diagram.Matrix

		// Legacy compatibility path: matrix input is converted to graph.
		parseMatrix(diagram.Matrix, "root", "", model)
		buildGraphProjection(model)
		model.Warnings = append(model.Warnings, ValidationError{
			Level:   "warning",
			Message: "matrix is legacy input; graph should be used as canonical source of truth",
		})
		return model
	}

	model.Errors = append(model.Errors, ValidationError{
		Level:   "fatal",
		Message: "diagram must define graph as the source of truth",
	})

	return model
}

// populateRuntimeFromGraph fills runtime indexes directly from graph nodes.
// This avoids scanning sparse matrices and keeps graph as the canonical source.
func populateRuntimeFromGraph(model *RuntimeModel) {
	for mapID, node := range model.Graph.Nodes {
		id := node.ID
		if id == "" {
			id = mapID
		}

		if _, exists := model.Elements[id]; exists {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Duplicate element ID detected in graph: %s", id),
				Element: id,
			})
			continue
		}

		scopeID := node.Scope
		if scopeID == "" {
			scopeID = "root"
		}

		coord := Coordinate{0, 0}
		if node.Coord != nil {
			coord = *node.Coord
		}

		if _, ok := model.ScopeIndex[scopeID]; !ok {
			model.ScopeIndex[scopeID] = make(map[Coordinate]string)
		}
		if existingID, occupied := model.ScopeIndex[scopeID][coord]; occupied {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Coordinate collision in scope %s at [%d, %d] between %s and %s", scopeID, coord[0], coord[1], existingID, id),
				Element: id,
			})
			continue
		}

		runtimeElem := &RuntimeElement{
			Element: Element{
				ID:          id,
				Label:       node.Label,
				Type:        node.Type,
				Description: node.Description,
				Style:       node.Style,
			},
			Coord:    coord,
			Scope:    scopeID,
			ParentID: node.ParentID,
		}

		model.Elements[id] = runtimeElem
		model.ScopeIndex[scopeID][coord] = id
	}
}

// parseMatrix recursively processes a matrix and collects elements
func parseMatrix(m Matrix, scopeID, parentID string, model *RuntimeModel) {
	if _, ok := model.ScopeIndex[scopeID]; !ok {
		model.ScopeIndex[scopeID] = make(map[Coordinate]string)
	}

	for y, row := range m {
		for x, cell := range row {
			if cell == nil {
				continue
			}

			coord := Coordinate{x, y}
			runtimeElem := &RuntimeElement{
				Element:  *cell,
				Coord:    coord,
				Scope:    scopeID,
				ParentID: parentID,
			}

			if existing, exists := model.Elements[cell.ID]; exists {
				model.Errors = append(model.Errors, ValidationError{
					Level:   "fatal",
					Message: fmt.Sprintf("Duplicate element ID detected during parse: %s", cell.ID),
					Element: existing.ID,
				})
				continue
			}

			// Store in map
			model.Elements[cell.ID] = runtimeElem
			model.ScopeIndex[scopeID][coord] = cell.ID

			// Recursively parse nested matrix if present
			if len(cell.Matrix) > 0 {
				model.ScopeMatrices[cell.ID] = cell.Matrix
				parseMatrix(cell.Matrix, cell.ID, cell.ID, model)
			}
		}
	}
}

// buildGraphProjection derives a hierarchical graph from parsed runtime elements.
func buildGraphProjection(model *RuntimeModel) {
	graph := &GraphModel{
		Nodes: make(map[string]*GraphNode),
	}

	for id, elem := range model.Elements {
		coord := elem.Coord
		graph.Nodes[id] = &GraphNode{
			ID:          elem.ID,
			Label:       elem.Label,
			Type:        elem.Type,
			Scope:       elem.Scope,
			ParentID:    elem.ParentID,
			Coord:       &coord,
			Description: elem.Description,
			Style:       elem.Style,
		}
	}

	for id, elem := range model.Elements {
		scopeCells := model.ScopeIndex[elem.Scope]
		for idx, link := range elem.Links {
			targetID, ok := scopeCells[link.To]
			if !ok {
				model.Errors = append(model.Errors, ValidationError{
					Level:   "fatal",
					Message: fmt.Sprintf("Link target [%d, %d] does not exist in scope %s", link.To[0], link.To[1], elem.Scope),
					Element: fmt.Sprintf("%s.links[%d]", id, idx),
					Coord:   &link.To,
				})
				continue
			}

			graph.Edges = append(graph.Edges, GraphEdge{
				FromID:    id,
				ToID:      targetID,
				Scope:     elem.Scope,
				Label:     link.Label,
				Condition: link.Condition,
				EdgeType:  link.EdgeType,
				Style:     link.Style,
			})
		}
	}

	model.Graph = graph
}

// SaveYAML marshals diagram back to YAML
func SaveYAML(diagram *Diagram, filePath string) error {
	data, err := yaml.Marshal(diagram)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
