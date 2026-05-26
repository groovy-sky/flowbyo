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
		Meta:       diagram.Meta,
		Elements:   make(map[string]*RuntimeElement),
		ScopeIndex: make(map[string]map[Coordinate]string),
		View:       diagram.View,
	}

	if diagram.Graph != nil && len(diagram.Graph.Nodes) > 0 {
		model.Graph = diagram.Graph
		populateRuntimeFromGraph(model)
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
