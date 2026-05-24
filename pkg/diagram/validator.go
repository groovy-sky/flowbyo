package diagram

import (
	"fmt"
)

// Validate performs comprehensive validation of the runtime model
func Validate(model *RuntimeModel) {
	validateStructure(model)
	validateIdentity(model)
	validateLinks(model)
	validateSemantics(model)
}

// validateStructure checks matrix rectangularity and required fields
func validateStructure(model *RuntimeModel) {
	// Check root matrix rectangularity
	validateMatrixRectangularity(model.Root, "root", model)

	// Check each element has required fields
	for id, elem := range model.Elements {
		if elem.ID == "" {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: "Element missing required field: id",
				Element: id,
			})
		}
		if elem.Label == "" {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: "Element missing required field: label",
				Element: id,
			})
		}
		if elem.Type == "" {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: "Element missing required field: type",
				Element: id,
			})
		}

		// Validate nested matrix if present
		if len(elem.Matrix) > 0 {
			validateMatrixRectangularity(elem.Matrix, fmt.Sprintf("nested in %s", id), model)
		}
	}
}

// validateMatrixRectangularity checks all rows have equal length
func validateMatrixRectangularity(m Matrix, scope string, model *RuntimeModel) {
	if len(m) == 0 {
		model.Warnings = append(model.Warnings, ValidationError{
			Level:   "warning",
			Message: fmt.Sprintf("Empty matrix: %s", scope),
		})
		return
	}

	expectedLen := len(m[0])
	for y, row := range m {
		if len(row) != expectedLen {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Non-rectangular matrix at %s row %d: expected %d columns, got %d", scope, y, expectedLen, len(row)),
			})
		}
	}
}

// validateIdentity checks all IDs are unique and present
func validateIdentity(model *RuntimeModel) {
	for id := range model.Elements {
		// Validate ID format
		if !isValidID(id) {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Invalid ID format: %s (must start with letter, contain only alphanumeric, dash, underscore)", id),
				Element: id,
			})
		}

		// Check ID is not coordinate-derived
		if len(id) >= 5 && id[0:5] == "cell_" {
			model.Warnings = append(model.Warnings, ValidationError{
				Level:   "warning",
				Message: fmt.Sprintf("ID %s appears to be coordinate-derived; should use stable semantic names", id),
				Element: id,
			})
		}
	}
}

// isValidID checks if ID matches pattern: [A-Za-z][A-Za-z0-9_-]*
func isValidID(id string) bool {
	if len(id) == 0 {
		return false
	}

	// First character must be letter
	if !((id[0] >= 'A' && id[0] <= 'Z') || (id[0] >= 'a' && id[0] <= 'z')) {
		return false
	}

	// Remaining characters must be alphanumeric, dash, or underscore
	for i := 1; i < len(id); i++ {
		c := id[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}

	return true
}

// validateLinks checks all links point to valid occupied cells
func validateLinks(model *RuntimeModel) {
	if model.Graph == nil {
		return
	}

	for idx, edge := range model.Graph.Edges {
		fromNode, fromOK := model.Graph.Nodes[edge.FromID]
		toNode, toOK := model.Graph.Nodes[edge.ToID]

		if !fromOK {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Edge %d has missing source node: %s", idx, edge.FromID),
			})
			continue
		}

		if !toOK {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Edge %d has missing target node: %s", idx, edge.ToID),
				Element: edge.FromID,
			})
			continue
		}

		if fromNode.Scope != toNode.Scope {
			model.Errors = append(model.Errors, ValidationError{
				Level:   "fatal",
				Message: fmt.Sprintf("Edge %s -> %s crosses scopes (%s -> %s)", edge.FromID, edge.ToID, fromNode.Scope, toNode.Scope),
				Element: edge.FromID,
			})
		}

		if edge.Scope != fromNode.Scope {
			model.Warnings = append(model.Warnings, ValidationError{
				Level:   "warning",
				Message: fmt.Sprintf("Edge %s -> %s scope (%s) differs from source scope (%s)", edge.FromID, edge.ToID, edge.Scope, fromNode.Scope),
				Element: edge.FromID,
			})
		}
	}
}

// validateSemantics performs diagram-type-specific validation
func validateSemantics(model *RuntimeModel) {
	switch model.Meta.DiagramType {
	case "flowchart":
		validateFlowchart(model)
	case "block-diagram":
		validateBlockDiagram(model)
	}
}

// validateFlowchart checks flowchart-specific requirements
func validateFlowchart(model *RuntimeModel) {
	hasStart := false
	hasEnd := false

	for _, elem := range model.Elements {
		if elem.Scope != "root" {
			continue
		}

		if elem.Type == "start" {
			hasStart = true
		}
		if elem.Type == "end" {
			hasEnd = true
		}
	}

	if !hasStart {
		model.Errors = append(model.Errors, ValidationError{
			Level:   "fatal",
			Message: "Flowchart must have at least one start node",
		})
	}

	if !hasEnd {
		model.Errors = append(model.Errors, ValidationError{
			Level:   "fatal",
			Message: "Flowchart must have at least one end node",
		})
	}
}

// validateBlockDiagram checks block diagram requirements (more lenient)
func validateBlockDiagram(model *RuntimeModel) {
	// Block diagrams don't require start/end
	if len(model.Elements) == 0 {
		model.Warnings = append(model.Warnings, ValidationError{
			Level:   "warning",
			Message: "Block diagram is empty",
		})
	}
}

// HasErrors returns true if there are fatal validation errors
func (m *RuntimeModel) HasErrors() bool {
	for _, err := range m.Errors {
		if err.Level == "fatal" {
			return true
		}
	}
	return false
}

// PrintValidation outputs validation results
func (m *RuntimeModel) PrintValidation() {
	if len(m.Errors) > 0 {
		fmt.Println("=== ERRORS ===")
		for _, err := range m.Errors {
			if err.Level == "fatal" {
				fmt.Printf("[%s] %s\n", err.Level, err.Message)
			}
		}
	}

	if len(m.Warnings) > 0 {
		fmt.Println("=== WARNINGS ===")
		for _, warn := range m.Warnings {
			fmt.Printf("[%s] %s\n", warn.Level, warn.Message)
		}
	}

	if len(m.Errors) == 0 && len(m.Warnings) == 0 {
		fmt.Println("✓ Validation passed")
	}
}
