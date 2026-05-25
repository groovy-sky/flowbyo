# Goflowyourself Maintenance Guide

## Purpose
- Convert text/markdown schema or YAML diagrams into SVG/HTML outputs.
- Canonical data model is graph-based (`diagram.graph.nodes` + `diagram.graph.edges`).
- Matrix input exists for legacy compatibility and is converted to graph runtime projection.

## Core Flow
- CLI entry: `cmd/goflowyourself/main.go`
- `from-schema`: text schema -> parse -> runtime model -> validate -> render -> optional graph YAML save.
- `load`: YAML -> parse -> runtime model -> validate -> render/validate-only -> optional YAML save.

## Key Files
- `cmd/goflowyourself/main.go`: command parsing + execution flow.
- `pkg/diagram/schema_parser.go`: ASCII/markdown schema parsing into graph nodes/edges.
- `pkg/diagram/parser.go`: `BuildRuntimeModel`, graph/matrix compatibility paths, runtime indexes.
- `pkg/diagram/validator.go`: structure/identity/link/semantic validation.
- `pkg/diagram/renderer.go`: SVG and HTML rendering.
- `pkg/diagram/graph.go`: graph/matrix projection helpers.
- `pkg/diagram/types.go`: shared data structures.

## Constraints / Behavior
- Cross-scope edges are rejected in MVP.
- Flowchart requires root `start` and `end` nodes.
- HTML output embeds SVG.

## Verify After Changes
- `go build ./...`
- Run representative conversion path, e.g.:
  - `go run cmd/goflowyourself/main.go from-schema --input examples/network-schema.txt --output /tmp/network.svg --html /tmp/network.html --save /tmp/network.yaml`
- If graph/matrix conversion was touched, also test `load` command on YAML input.

## Required Update Rule
- Read this file before making code changes.
- If behavior, commands, architecture, constraints, or key file responsibilities changed, update this file in the same work session.
- Keep updates concise and factual.

## Name Basis (Working Hypothesis)
- The name `goflowyourself` is likely based on `go + flow + yourself`.
- Interpreted meaning: "define your own flow and run it", then Goflowyourself validates and renders it.
- This is a documented project hypothesis unless maintainers confirm a different official origin.