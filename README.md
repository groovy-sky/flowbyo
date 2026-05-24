# Flowbyo

Flowbyo converts text-based diagrams into visual outputs.

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
go build -o flowbyo ./cmd/flowbyo
```

### 1) Convert Text Schema To SVG/HTML

```bash
go run cmd/flowbyo/main.go from-schema \
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
go run cmd/flowbyo/main.go load \
  --input diagram.yaml \
  --output diagram.svg \
  --html diagram.html
```

Validation-only mode:

```bash
go run cmd/flowbyo/main.go load --input diagram.yaml --validate
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
cmd/flowbyo/main.go         CLI
pkg/diagram/schema_parser.go text schema parsing
pkg/diagram/types.go         model types
pkg/diagram/parser.go        YAML load + runtime build
pkg/diagram/validator.go     validation
pkg/diagram/renderer.go      SVG + HTML rendering
pkg/diagram/graph.go         graph/matrix conversion helpers
examples/                    sample inputs
```

## Notes

- Cross-scope edges are currently rejected.
- Matrix is no longer the source of truth.
- Output HTML embeds the generated SVG directly.

## License

MIT
