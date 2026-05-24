# Nested Matrix Diagram Storage and UI Rendering

## 1. Goal

Build a solution that stores flowcharts, block-schemas, and basic diagrams in a text-based **nested matrix** format and renders a UI directly from it.

The solution must support:

1. storing elements directly inside matrix cells
2. assigning every block its own required unique `id`
3. storing that `id` inside the element object in the matrix cell
4. storing links inside source elements
5. storing nested child matrices inside elements
6. rendering the UI directly from the same structure
7. saving and loading without losing structure, identity, or layout

Core model:

- **matrix = container + layout**
- **cell = element or null**
- **element.id = block identity**
- **element.links = outgoing edges**
- **element.matrix = nested child diagram**
- **UI = rendered projection of nested matrices**

---

## 2. Core Principles

### 2.1 Identity and Position Are Different
Each block has:

- an **id**: who the block is
- a **coordinate**: where the block is

Rules:

- `id` is stored in the element object
- coordinate is defined by the element’s position in the matrix
- moving a block changes its coordinate, not its `id`

### 2.2 Matrix Is the Source of Truth
The matrix is the canonical storage model.

It defines:

- which elements exist
- where they are placed
- which elements are nested
- which outgoing links each element has

### 2.3 Links Are Stored in Source Elements
Each element stores its own outgoing links.

Each link points to target coordinates in the same matrix scope.

### 2.4 Nesting Is Native
Any element may contain a child matrix.

This allows:

- subprocesses
- grouped blocks
- systems with internal modules
- expandable/collapsible containers

---

## 3. Scope

### 3.1 Supported in MVP
- flowcharts
- block diagrams
- dependency diagrams
- process flows
- branching and merging
- loops
- nested/grouped diagrams
- simple IT/system diagrams

### 3.2 Not in MVP
- full BPMN
- full UML
- collaborative editing
- multi-page diagrams
- advanced free-form layout
- advanced animation
- cross-scope links

---

## 4. Data Model

## 4.1 Matrix

A matrix is a 2D grid.

- rows = `y`
- columns = `x`
- coordinate = `(x, y)`

Each cell contains either:

- `null`
- or an `element`

A matrix may exist at:

- the root of the document
- inside an element as a child matrix

### 4.1.1 Matrix Role
A matrix provides:

- spatial placement
- storage for elements
- scope for local links
- hierarchy when nested

---

## 4.2 Element

Each occupied cell contains exactly one element object.

### 4.2.1 Required Fields
- `id`
- `label`
- `type`

### 4.2.2 Optional Fields
- `links`
- `matrix`
- `description`
- `style`

### 4.2.3 Element Rules
- every occupied cell must contain an element
- every element must have its own `id`
- `id` is the canonical identity of the block
- the cell location defines the element coordinate
- an element may optionally contain a child `matrix`

---

## 4.3 Link

A link is stored inside the source element.

### 4.3.1 Link Fields
- `to: [x, y]`
- `label` optional
- `condition` optional
- `edge_type` optional
- `style` optional

### 4.3.2 Link Rules
- links are local to the current matrix scope
- `to` coordinates must point to an occupied cell in the same matrix
- runtime resolves target coordinates to the target element and target `id`

---

## 4.4 View

Optional rendering settings may include:

- `direction`
- `theme`
- `cell_size`
- `padding`
- `routing`
- `collapse_nested`

The `view` section affects rendering only. It does not define identity or structure.

---

## 5. Identity Rules

### 5.1 Mandatory ID
Every occupied matrix cell must contain an element with an `id`.

Invalid:

```yaml
- label: Start
  type: start
```

Valid:

```yaml
- id: START_1
  label: Start
  type: start
```

### 5.2 Unique ID
For MVP, all element IDs must be globally unique across the entire document, including nested matrices.

This makes:

- validation simpler
- runtime lookup easier
- export more stable
- editing safer

### 5.3 Stable ID
A block `id` must remain stable when:

- the block moves to a different coordinate
- rows/columns are inserted or removed
- the parent matrix is reorganized

IDs must not be derived from coordinates.

Good:

- `AUTH_VALIDATE_PASSWORD`

Bad:

- `cell_2_1`

---

## 6. Nesting Rules

### 6.1 Child Matrix
An element may contain a child matrix in its `matrix` field.

### 6.2 Local Coordinate Scope
Coordinates are local to the matrix that contains them.

This means:

- the root matrix has its own coordinate space
- every child matrix has its own coordinate space

### 6.3 Local Link Scope
For MVP, links stay within the current matrix.

So:

- root elements link to root coordinates
- child elements link to child coordinates in their own child matrix

### 6.4 Parent/Child Relationship
If an element contains a child matrix, then all elements inside that matrix are children of that parent element.

### 6.5 UI Behavior for Nested Matrices
Nested matrices may be rendered as:

- collapsed containers
- expanded inline content
- drill-down views

MVP should support:

- collapsed
- expanded

---

## 7. File Format

Use a single human-readable structured file.

Recommended format:

- **YAML**

Reasons:

- easy to read
- easy to diff
- supports nested structures naturally
- matches the nested matrix model well

Top-level sections:

- `meta`
- `matrix`
- `view`

---

## 8. Parsing Model

The parser operates recursively.

### 8.1 Parse Steps
1. load the YAML file
2. parse `meta`
3. parse the root `matrix`
4. scan every cell
5. if the cell is occupied:
   - parse the element
   - require `id`
   - assign coordinate from matrix position
   - parse local links
   - recursively parse child `matrix` if present
6. parse `view`
7. build the runtime model

### 8.2 Runtime Model
At runtime, each element should expose:

- `id`
- `label`
- `type`
- `coord`
- `scope`
- `parentId` optional
- `links`
- `children` optional

---

## 9. Validation

## 9.1 Structural Validation
Check:

- every matrix is rectangular
- every occupied cell contains a valid element object
- every element has an `id`
- all IDs are unique
- all link targets are within matrix bounds
- all link targets point to occupied cells
- nested matrices are structurally valid

## 9.2 Semantic Validation
For flowcharts:

- at least one `start` in the root scope
- at least one `end` in the root scope
- decision nodes should have meaningful branches

For block diagrams:

- `start` and `end` are optional
- disconnected components may be allowed with warnings

## 9.3 Identity Validation
Check:

- no missing IDs
- no duplicate IDs
- IDs are strings
- IDs are stable identifiers, not coordinate-derived

## 9.4 Error Levels
Validation results should be classified as:

- `fatal`
- `warning`
- `info`

---

## 10. Rendering Strategy

## 10.1 Rendering Principle
Render directly from matrix coordinates.

For each matrix scope:

- place elements by cell position
- draw links between local coordinates
- recursively render child matrices

## 10.2 Coordinate-to-Pixel Mapping
Example:

- `pixelX = padding + x * cellWidth`
- `pixelY = padding + y * cellHeight`

For nested matrices:

- child positions are rendered relative to the parent container

## 10.3 Render Target
Primary MVP target:

- **SVG**

Reasons:

- easy to inspect
- easy to label
- good for connectors
- easy to export

## 10.4 Node Shapes
Recommended type-to-shape mapping:

- `start` -> terminal
- `end` -> terminal
- `process` -> rectangle
- `decision` -> diamond
- `input` -> parallelogram
- `subprocess` -> container/block
- `component` -> block rectangle

## 10.5 Edge Styles
Recommended edge types:

- `control`
- `branch`
- `dependency`
- `data`
- `network`

---

## 11. Editing Model

### 11.1 MVP
Support:

- load YAML
- render root matrix
- render nested matrices
- inspect elements by `id`
- inspect links
- save back without losing IDs, links, nesting, or layout

### 11.2 Later
Add:

- move elements between cells
- insert/delete rows and columns
- edit labels/types/styles
- add/remove links
- move elements between parent and child matrices
- expand/collapse containers interactively

### 11.3 Save Rules
When saving:

- each element stays in its current matrix cell
- its `id` remains unchanged
- links remain stored in the source element
- nested matrices remain inside parent elements
- null cells remain explicit where needed

---

## 12. Exports

### 12.1 MVP
- SVG
- PNG
- YAML round-trip save
- runtime JSON dump if needed internally

### 12.2 Later
- Mermaid
- Graphviz DOT
- flattened edge list
- adjacency matrix export
- PDF

---

## 13. MVP Definition

The MVP is complete when the system can:

1. load a nested matrix YAML file
2. parse elements directly from matrix cells
3. require and preserve unique block IDs
4. parse links stored in source elements
5. parse nested child matrices recursively
6. validate structure, identity, and local links
7. render a readable root diagram
8. render nested diagrams in collapsed or expanded mode
9. save and reload without losing IDs, links, nesting, or layout

---

## 14. Risks and Mitigations

### Risk 1
Deep nesting becomes hard to read.

Mitigation:
- collapse/expand support
- drill-down later

### Risk 2
Large sparse matrices waste space.

Mitigation:
- keep matrices compact
- add resize/compaction tools later

### Risk 3
Links become invalid when blocks move.

Mitigation:
- editor must update links automatically
- validator must check target occupancy

### Risk 4
Cross-scope links are complex.

Mitigation:
- MVP restricts links to the current matrix scope only

---

## 15. Deliverables

- nested matrix YAML format
- YAML schema
- recursive parser
- validator
- runtime hierarchical model
- SVG renderer
- example files
- save/load support

---

## 16. Acceptance Criteria

The solution is accepted when:

1. elements are stored directly inside matrix cells
2. every occupied cell contains an element with a unique `id`
3. IDs are preserved independently of coordinates
4. elements may contain nested child matrices
5. links are stored inside source elements
6. the parser reconstructs hierarchy and identity correctly
7. the renderer displays hierarchy correctly
8. save/load round-trip preserves IDs, links, nesting, and layout

---

## 17. Summary

Final model:

- **matrix = container + layout**
- **cell = element**
- **element.id = required block identity**
- **coordinate = placement**
- **element.links = local edges**
- **element.matrix = child diagram**
- **UI = rendered nested matrix hierarchy**
```

---

## 2) `diagram.schema.yaml`

This is a strict YAML schema, using JSON Schema syntax expressed in YAML.

It validates the document shape, required fields, recursive nesting, link structure, and required IDs.

```yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "diagram.schema.yaml"
title: "Nested Matrix Diagram Schema"
description: >
  Schema for nested matrix diagrams where each occupied matrix cell contains
  an element object with a required unique id, optional local links, and
  optional nested child matrix.
type: object
additionalProperties: false
required:
  - meta
  - matrix

properties:
  meta:
    $ref: "#/$defs/meta"
  matrix:
    $ref: "#/$defs/matrix"
  view:
    $ref: "#/$defs/view"

$defs:
  id:
    type: string
    minLength: 1
    pattern: "^[A-Za-z][A-Za-z0-9_-]*$"
    description: "Stable block identifier. Must be globally unique at runtime."

  coordinate:
    type: array
    description: "[x, y] local coordinate within the current matrix scope."
    minItems: 2
    maxItems: 2
    prefixItems:
      - type: integer
        minimum: 0
      - type: integer
        minimum: 0
    items: false

  style:
    type: object
    description: "Optional flexible style object."
    additionalProperties: true

  link:
    type: object
    additionalProperties: false
    required:
      - to
    properties:
      to:
        $ref: "#/$defs/coordinate"
      label:
        type: string
      condition:
        type: string
      edge_type:
        type: string
        enum:
          - control
          - branch
          - dependency
          - data
          - network
          - reference
          - custom
      style:
        $ref: "#/$defs/style"

  element:
    type: object
    additionalProperties: false
    required:
      - id
      - label
      - type
    properties:
      id:
        $ref: "#/$defs/id"
      label:
        type: string
        minLength: 1
      type:
        type: string
        enum:
          - start
          - end
          - process
          - decision
          - input
          - output
          - subprocess
          - component
          - group
          - custom
      links:
        type: array
        items:
          $ref: "#/$defs/link"
        default: []
      matrix:
        $ref: "#/$defs/matrix"
      description:
        type: string
      style:
        $ref: "#/$defs/style"

  cell:
    description: "A matrix cell is either null or an element."
    oneOf:
      - type: "null"
      - $ref: "#/$defs/element"

  row:
    type: array
    minItems: 1
    items:
      $ref: "#/$defs/cell"

  matrix:
    type: array
    minItems: 1
    items:
      $ref: "#/$defs/row"
    description: >
      2D matrix of cells. Rectangularity must be validated at runtime.

  meta:
    type: object
    additionalProperties: false
    required:
      - name
      - diagram_type
      - version
    properties:
      name:
        type: string
        minLength: 1
      diagram_type:
        type: string
        enum:
          - flowchart
          - block-schema
          - block-diagram
          - dependency-diagram
          - process-flow
          - custom
      version:
        oneOf:
          - type: integer
            minimum: 1
          - type: string
            minLength: 1

  view:
    type: object
    additionalProperties: false
    properties:
      direction:
        type: string
        enum:
          - TB
          - LR
          - BT
          - RL
      theme:
        type: string
      cell_size:
        type: array
        minItems: 2
        maxItems: 2
        prefixItems:
          - type: integer
            minimum: 1
          - type: integer
            minimum: 1
        items: false
      padding:
        type: integer
        minimum: 0
      routing:
        type: string
        enum:
          - straight
          - orthogonal
          - curved
      collapse_nested:
        type: boolean

x-runtime-validation:
  - "All element.id values must be globally unique across the entire document."
  - "Each matrix must be rectangular: all rows in the same matrix must have equal length."
  - "Every link.to coordinate must point to an occupied cell in the same matrix scope."
  - "Link scope is local to the current matrix only."
  - "Element ids must remain stable when coordinates change."