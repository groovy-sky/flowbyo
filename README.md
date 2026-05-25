# Goflowyourself

Goflowyourself converts text-based diagrams into visual outputs.

Primary goal:
- text/markdown schema -> SVG
- text/markdown schema -> HTML

Secondary goal:
- load graph YAML -> validate -> render SVG/HTML

## Current Model

The canonical model is graph-based.

- `diagram.graph.nodes`: block identity and coordinates
- `diagram.graph.edges`: links between node ids
- `diagram.view`: render preferences

`matrix` input is still accepted only as a legacy compatibility path.

## CLI

Build:

```bash
go build -o goflowyourself ./cmd/goflowyourself
```

### 1) Convert Text Schema To SVG/HTML

```bash
go run cmd/goflowyourself/main.go from-schema \
  --input examples/network-schema.txt \
  --output network.svg \
  --html network.html \
  --save network.yaml
```

Flags:
- `--input`: text/markdown schema file
- `--output`: SVG output path (optional)
- `--html`: HTML output path (optional)
- `--save`: save generated graph YAML (optional)
- `--name`: diagram name (optional)
- `--diagram-type`: `flowchart` or `block-diagram` (optional)
- `--version`: metadata version (optional)

At least one of `--output`, `--html`, or `--save` is required.

### 2) Load Graph YAML And Render

```bash
go run cmd/goflowyourself/main.go load \
  --input diagram.yaml \
  --output diagram.svg \
  --html diagram.html
```

Validation-only mode:

```bash
go run cmd/goflowyourself/main.go load --input diagram.yaml --validate
```

## Input Schema Example

```text
             Internet
                |
           NAT Gateway
                |
           Security VPC
        +------------------+
        | Gateway LB       |
        | Palo Alto FW     |
        +------------------+
                |
            GWLBe
                |
        Transit Gateway
        /      |      \
     VPC-A   VPC-B   VPC-C
```

```svg
<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="510" height="1060" viewBox="0 0 510 1060">
<defs>
  <style>
    .node { fill: white; stroke: #333; stroke-width: 2; }
    .node-start, .node-end { fill: #90EE90; rx: 10; ry: 10; }
    .node-process { fill: #87CEEB; }
    .node-decision { fill: #FFD700; }
    .node-subprocess { fill: #DDA0DD; }
    .node-text { font-family: Arial, sans-serif; font-size: 12px; text-anchor: middle; dominant-baseline: middle; }
    .link { stroke: #333; stroke-width: 2; fill: none; marker-end: url(#arrowhead); }
    .link-label { font-family: Arial, sans-serif; font-size: 10px; fill: #333; }
    #arrowhead { fill: #333; }
  </style>
  <marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
    <polygon points="0 0, 10 3, 0 6" />
  </marker>
</defs>
<rect width="510" height="1060" fill="white" stroke="#ddd" stroke-width="1"/><line x1="254" y1="110" x2="254" y2="180" class="link"/><line x1="255" y1="250" x2="255" y2="320" class="link"/><line x1="255" y1="530" x2="255" y2="600" class="link"/><line x1="254" y1="670" x2="254" y2="740" class="link"/><line x1="255" y1="810" x2="255" y2="880" class="link"/><line x1="195" y1="810" x2="140" y2="880" class="link"/><line x1="314" y1="810" x2="370" y2="880" class="link"/><line x1="254" y1="390" x2="254" y2="460" class="link"/><rect x="200" y="40" width="110" height="70" class="node node-component"/><text x="255" y="75" class="node-text">Internet</text><rect x="199" y="180" width="111" height="70" class="node node-component"/><text x="254" y="215" class="node-text">NAT Gateway</text><rect x="196" y="320" width="118" height="70" class="node node-component"/><text x="255" y="355" class="node-text">Security VPC</text><rect x="150" y="460" width="209" height="70" class="node node-component"/><text x="254" y="495" class="node-text">Gateway LB / Palo Alto FW</text><rect x="200" y="600" width="110" height="70" class="node node-component"/><text x="255" y="635" class="node-text">GWLBe</text><rect x="185" y="740" width="139" height="70" class="node node-component"/><text x="254" y="775" class="node-text">Transit Gateway</text><rect x="40" y="880" width="110" height="70" class="node node-component"/><text x="95" y="915" class="node-text">VPC-A</text><rect x="200" y="880" width="110" height="70" class="node node-component"/><text x="255" y="915" class="node-text">VPC-B</text><rect x="360" y="880" width="110" height="70" class="node node-component"/><text x="415" y="915" class="node-text">VPC-C</text>
</svg>
```

## Graph YAML Example

```yaml
meta:
  name: Network Security Flow
  diagram_type: block-diagram
  version: "1"
graph:
  nodes:
    INTERNET:
      id: INTERNET
      label: Internet
      type: component
      scope: root
      coord: [4, 0]
  edges:
    - from: INTERNET
      to: NAT_GATEWAY
      scope: root
view:
  direction: TB
  padding: 50
  cell_size: [160, 90]
  routing: straight
```

## Layout Rules Implemented

- Build graph from parsed schema
- Compute node size from label text length
- Compute node coordinates with configurable X/Y spacing
- Route links source-bottom -> target-top wherever possible
- Handle branch fan-out (`/ | \`) and box blocks
- Ignore markdown code-fence lines in schema input

## Project Structure

```text
cmd/goflowyourself/main.go         CLI
pkg/diagram/schema_parser.go text schema parsing
pkg/diagram/types.go         model types
pkg/diagram/parser.go        YAML load + runtime build
pkg/diagram/validator.go     validation
pkg/diagram/renderer.go      SVG + HTML rendering
pkg/diagram/graph.go         graph/matrix conversion helpers
examples/                    sample inputs
```

## Maintenance

- See `MAINTENANCE.md` for architecture responsibilities, verification commands, and required post-change documentation updates.

## Notes

- Cross-scope edges are currently rejected.
- Matrix is no longer the source of truth.
- Output HTML embeds the generated SVG directly.

## License

MIT
