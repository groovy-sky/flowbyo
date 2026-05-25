package diagram

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// SVGRenderer renders a diagram to SVG
type SVGRenderer struct {
	CellWidth   int
	CellHeight  int
	Padding     int
	StrokeWidth int
	NodeGapX    int
	NodeGapY    int
}

// NewSVGRenderer creates a new renderer with defaults
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{
		CellWidth:   120,
		CellHeight:  80,
		Padding:     40,
		StrokeWidth: 2,
		NodeGapX:    50,
		NodeGapY:    70,
	}
}

type nodeLayout struct {
	ID     string
	Elem   *Element
	X      int
	Y      int
	W      int
	H      int
	Layer  int
	Center int
}

// RenderToSVG generates SVG for the diagram
func (r *SVGRenderer) RenderToSVG(model *RuntimeModel, filePath string) error {
	svg := r.renderSVGDocument(model)
	return os.WriteFile(filePath, []byte(svg), 0644)
}

// RenderToHTML generates an HTML file containing embedded SVG.
func (r *SVGRenderer) RenderToHTML(model *RuntimeModel, filePath string) error {
	svg := r.renderSVGDocument(model)
	if idx := strings.Index(svg, "<svg "); idx >= 0 {
		svg = svg[idx:]
	}
	title := "Goflowyourself Diagram"
	if model != nil && strings.TrimSpace(model.Meta.Name) != "" {
		title = model.Meta.Name
	}

	html := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <style>
    body { margin: 0; font-family: Arial, sans-serif; background: #f6f7fb; color: #111; }
    header { padding: 14px 18px; border-bottom: 1px solid #dde3ee; background: #fff; }
    h1 { margin: 0; font-size: 16px; font-weight: 600; }
    .canvas { padding: 16px; overflow: auto; }
    .card { display: inline-block; background: #fff; border: 1px solid #dde3ee; border-radius: 10px; box-shadow: 0 2px 8px rgba(17,24,39,.06); }
  </style>
</head>
<body>
  <header><h1>%s</h1></header>
  <div class="canvas">
    <div class="card">%s</div>
  </div>
</body>
</html>
`, escapeXML(title), escapeXML(title), svg)

	return os.WriteFile(filePath, []byte(html), 0644)
}

func (r *SVGRenderer) renderSVGDocument(model *RuntimeModel) string {
	layouts, edges, width, height := r.buildGraphLayout(model)
	if len(layouts) == 0 {
		width, height = r.calculateDimensions(model.Root)
		width += r.Padding * 2
		height += r.Padding * 2
	}

	svg := strings.Builder{}
	svg.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<defs>
  <style>
    .node { fill: white; stroke: #333; stroke-width: %d; }
    .node-start, .node-end { fill: #90EE90; rx: 10; ry: 10; }
    .node-process { fill: #87CEEB; }
    .node-decision { fill: #FFD700; }
    .node-subprocess { fill: #DDA0DD; }
    .node-text { font-family: Arial, sans-serif; font-size: 12px; text-anchor: middle; dominant-baseline: middle; }
    .link { stroke: #333; stroke-width: %d; fill: none; marker-end: url(#arrowhead); }
    .link-label { font-family: Arial, sans-serif; font-size: 10px; fill: #333; }
    #arrowhead { fill: #333; }
  </style>
  <marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
    <polygon points="0 0, 10 3, 0 6" />
  </marker>
</defs>
`, width, height, width, height, r.StrokeWidth, r.StrokeWidth))

	// Draw background
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="white" stroke="#ddd" stroke-width="1"/>`, width, height))

	if len(layouts) > 0 {
		for _, edge := range edges {
			svg.WriteString(fmt.Sprintf(
				`<line x1="%d" y1="%d" x2="%d" y2="%d" class="link"/>`,
				edge.x1, edge.y1, edge.x2, edge.y2))
			if edge.label != "" {
				midX := (edge.x1 + edge.x2) / 2
				midY := (edge.y1 + edge.y2) / 2
				svg.WriteString(fmt.Sprintf(
					`<text x="%d" y="%d" class="link-label">%s</text>`,
					midX, midY-5, escapeXML(edge.label)))
			}
		}

		for _, l := range layouts {
			r.drawElementWithBox(l.Elem, l.X+l.W/2, l.Y+l.H/2, l.W, l.H, &svg)
		}
	} else {
		// Fallback for non-graph layouts.
		r.drawMatrix(model.Root, &svg, model, 0)
		r.drawLinks(model.Root, &svg, model, 0)
	}

	svg.WriteString("\n</svg>")
	return svg.String()
}

type routedEdge struct {
	x1, y1 int
	x2, y2 int
	label  string
}

// buildGraphLayout computes node size, coordinates with spacing, and line anchors.
func (r *SVGRenderer) buildGraphLayout(model *RuntimeModel) ([]nodeLayout, []routedEdge, int, int) {
	if model.Graph == nil || len(model.Graph.Nodes) == 0 {
		return nil, nil, 0, 0
	}

	rootNodes := make([]*GraphNode, 0)
	for _, n := range model.Graph.Nodes {
		if n.Scope == "root" {
			rootNodes = append(rootNodes, n)
		}
	}
	if len(rootNodes) == 0 {
		return nil, nil, 0, 0
	}

	layerMap := make(map[int][]*GraphNode)
	uniqueY := make(map[int]bool)
	for _, n := range rootNodes {
		if n.Coord != nil {
			uniqueY[n.Coord[1]] = true
		}
	}
	if len(uniqueY) > 0 {
		ys := make([]int, 0, len(uniqueY))
		for y := range uniqueY {
			ys = append(ys, y)
		}
		sort.Ints(ys)
		yIndex := make(map[int]int)
		for i, y := range ys {
			yIndex[y] = i
		}
		for _, n := range rootNodes {
			layer := 0
			if n.Coord != nil {
				layer = yIndex[n.Coord[1]]
			}
			layerMap[layer] = append(layerMap[layer], n)
		}
	} else {
		for _, n := range rootNodes {
			layerMap[0] = append(layerMap[0], n)
		}
	}

	layers := make([]int, 0, len(layerMap))
	for layer := range layerMap {
		layers = append(layers, layer)
	}
	sort.Ints(layers)

	layouts := make([]nodeLayout, 0, len(rootNodes))
	layoutByID := make(map[string]int)
	layerLayoutIdx := make(map[int][]int)
	layerWidths := make(map[int]int)
	layerHeights := make(map[int]int)

	for _, layer := range layers {
		nodes := layerMap[layer]
		sort.Slice(nodes, func(i, j int) bool {
			xi := graphNodeX(nodes[i])
			xj := graphNodeX(nodes[j])
			if xi == xj {
				return nodes[i].Label < nodes[j].Label
			}
			return xi < xj
		})

		cursorX := r.Padding
		maxH := 0
		for _, n := range nodes {
			elem := &Element{ID: n.ID, Label: n.Label, Type: n.Type}
			if runtimeElem, ok := model.Elements[n.ID]; ok {
				elem = &runtimeElem.Element
			}
			w, h := r.measureElement(elem)
			l := nodeLayout{
				ID:     n.ID,
				Elem:   elem,
				X:      cursorX,
				Y:      0,
				W:      w,
				H:      h,
				Layer:  layer,
				Center: cursorX + w/2,
			}
			layouts = append(layouts, l)
			idx := len(layouts) - 1
			layoutByID[n.ID] = idx
			layerLayoutIdx[layer] = append(layerLayoutIdx[layer], idx)
			cursorX += w + r.NodeGapX
			if h > maxH {
				maxH = h
			}
		}
		if cursorX > r.Padding {
			layerWidths[layer] = cursorX - r.NodeGapX + r.Padding
		}
		layerHeights[layer] = maxH
	}

	maxWidth := 0
	for _, layer := range layers {
		if layerWidths[layer] > maxWidth {
			maxWidth = layerWidths[layer]
		}
	}

	currentY := r.Padding
	for _, layer := range layers {
		layerWidth := layerWidths[layer]
		offsetX := (maxWidth - layerWidth) / 2
		for _, i := range layerLayoutIdx[layer] {
			layouts[i].X += offsetX
			layouts[i].Y = currentY
			layouts[i].Center = layouts[i].X + layouts[i].W/2
			layoutByID[layouts[i].ID] = i
		}
		currentY += layerHeights[layer] + r.NodeGapY
	}

	edges := make([]routedEdge, 0, len(model.Graph.Edges))
	for _, e := range model.Graph.Edges {
		fromIdx, okFrom := layoutByID[e.FromID]
		toIdx, okTo := layoutByID[e.ToID]
		if !okFrom || !okTo {
			continue
		}
		from := layouts[fromIdx]
		to := layouts[toIdx]

		startX := clampInt(to.Center, from.X+10, from.X+from.W-10)
		startY := from.Y + from.H
		endX := clampInt(startX, to.X+10, to.X+to.W-10)
		endY := to.Y

		if to.Layer <= from.Layer {
			startX = from.X + from.W/2
			startY = from.Y + from.H/2
			endX = to.X + to.W/2
			endY = to.Y + to.H/2
		}

		edges = append(edges, routedEdge{x1: startX, y1: startY, x2: endX, y2: endY, label: e.Label})
	}

	totalHeight := currentY + r.Padding
	if maxWidth == 0 {
		maxWidth = 800
	}
	return layouts, edges, maxWidth, totalHeight
}

func (r *SVGRenderer) measureElement(elem *Element) (int, int) {
	label := strings.TrimSpace(elem.Label)
	if label == "" {
		label = elem.ID
	}
	charW := 7
	basePadding := 34
	w := len([]rune(label))*charW + basePadding
	if w < r.CellWidth-10 {
		w = r.CellWidth - 10
	}
	h := r.CellHeight - 10
	return w, h
}

func graphNodeX(n *GraphNode) int {
	if n != nil && n.Coord != nil {
		return n.Coord[0]
	}
	return 0
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// calculateDimensions computes SVG dimensions needed for matrix
func (r *SVGRenderer) calculateDimensions(m Matrix) (width, height int) {
	if len(m) == 0 {
		return 0, 0
	}

	height = len(m) * r.CellHeight
	if len(m) > 0 {
		width = len(m[0]) * r.CellWidth
	}

	return
}

// drawMatrix recursively draws all elements in a matrix
func (r *SVGRenderer) drawMatrix(m Matrix, svg *strings.Builder, model *RuntimeModel, nestLevel int) {
	for y, row := range m {
		for x, cell := range row {
			if cell == nil {
				continue
			}

			px := r.Padding + x*r.CellWidth + r.CellWidth/2
			py := r.Padding + y*r.CellHeight + r.CellHeight/2

			r.drawElement(cell, px, py, svg)

			// Recursively draw nested matrix
			if len(cell.Matrix) > 0 && !model.View.CollapseNested {
				r.drawMatrix(cell.Matrix, svg, model, nestLevel+1)
			}
		}
	}
}

// drawElement draws a single element node
func (r *SVGRenderer) drawElement(elem *Element, centerX, centerY int, svg *strings.Builder) {
	r.drawElementWithBox(elem, centerX, centerY, r.CellWidth-10, r.CellHeight-10, svg)
}

func (r *SVGRenderer) drawElementWithBox(elem *Element, centerX, centerY, width, height int, svg *strings.Builder) {
	nodeClass := fmt.Sprintf("node node-%s", elem.Type)

	switch elem.Type {
	case "start", "end":
		// Terminal (rounded)
		svg.WriteString(fmt.Sprintf(
			`<circle cx="%d" cy="%d" r="30" class="%s"/>`,
			centerX, centerY, nodeClass))
	case "decision":
		// Diamond
		r.drawDiamond(centerX, centerY, nodeClass, svg)
	default:
		// Rectangle
		svg.WriteString(fmt.Sprintf(
			`<rect x="%d" y="%d" width="%d" height="%d" class="%s"/>`,
			centerX-width/2, centerY-height/2,
			width, height, nodeClass))
	}

	// Draw label
	svg.WriteString(fmt.Sprintf(
		`<text x="%d" y="%d" class="node-text">%s</text>`,
		centerX, centerY, escapeXML(elem.Label)))
}

// drawDiamond draws a diamond shape for decision nodes
func (r *SVGRenderer) drawDiamond(centerX, centerY int, class string, svg *strings.Builder) {
	w := r.CellWidth/2 - 10
	h := r.CellHeight/2 - 5

	points := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
		centerX, centerY-h, // top
		centerX+w, centerY, // right
		centerX, centerY+h, // bottom
		centerX-w, centerY) // left

	svg.WriteString(fmt.Sprintf(
		`<polygon points="%s" class="%s"/>`, points, class))
}

// drawLinks draws all links between elements
func (r *SVGRenderer) drawLinks(m Matrix, svg *strings.Builder, model *RuntimeModel, nestLevel int) {
	for y, row := range m {
		for x, cell := range row {
			if cell == nil {
				continue
			}

			for _, link := range cell.Links {
				fromX := r.Padding + x*r.CellWidth + r.CellWidth/2
				fromY := r.Padding + y*r.CellHeight + r.CellHeight/2

				toX := r.Padding + link.To[0]*r.CellWidth + r.CellWidth/2
				toY := r.Padding + link.To[1]*r.CellHeight + r.CellHeight/2

				svg.WriteString(fmt.Sprintf(
					`<line x1="%d" y1="%d" x2="%d" y2="%d" class="link"/>`,
					fromX, fromY, toX, toY))

				if link.Label != "" {
					midX := (fromX + toX) / 2
					midY := (fromY + toY) / 2
					svg.WriteString(fmt.Sprintf(
						`<text x="%d" y="%d" class="link-label">%s</text>`,
						midX, midY-5, escapeXML(link.Label)))
				}
			}

			// Recursively draw links in nested matrix
			if len(cell.Matrix) > 0 && !model.View.CollapseNested {
				r.drawLinks(cell.Matrix, svg, model, nestLevel+1)
			}
		}
	}
}

// escapeXML escapes XML special characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
