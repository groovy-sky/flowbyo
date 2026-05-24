package diagram

import (
	"fmt"
	"sort"
	"strings"
)

type positionedNode struct {
	ID    string
	Label string
	X     int
	Y     int
}

type boxRegion struct {
	top, bottom int
	left, right int
}

// ParseTextSchemaToDiagram parses a simple ASCII/markdown-style schema into a graph-based diagram.
func ParseTextSchemaToDiagram(name, diagramType, version, text string) (*Diagram, error) {
	lines := splitLines(text)
	if len(lines) == 0 {
		return nil, fmt.Errorf("schema is empty")
	}

	nodes, err := extractNodes(lines)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found in schema")
	}

	edges := extractEdges(lines, nodes)

	compressCoordinates(nodes)

	graph := &GraphModel{Nodes: make(map[string]*GraphNode)}
	for _, n := range nodes {
		coord := Coordinate{n.X, n.Y}
		graph.Nodes[n.ID] = &GraphNode{
			ID:    n.ID,
			Label: n.Label,
			Type:  inferNodeType(n.Label),
			Scope: "root",
			Coord: &coord,
		}
	}

	for _, e := range edges {
		graph.Edges = append(graph.Edges, GraphEdge{
			FromID: e[0],
			ToID:   e[1],
			Scope:  "root",
		})
	}

	return &Diagram{
		Meta: Meta{
			Name:        name,
			DiagramType: diagramType,
			Version:     version,
		},
		Graph: graph,
		View: View{
			Direction: "TB",
			Theme:     "light",
			CellSize:  []int{160, 90},
			Padding:   50,
			Routing:   "straight",
		},
	}, nil
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\t", "    ")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.Trim(text, "\n")
	if text == "" {
		return nil
	}
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		trim := strings.TrimSpace(line)
		if trim == "```" || trim == "~~~" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func extractNodes(lines []string) ([]*positionedNode, error) {
	var regions []boxRegion
	var nodes []*positionedNode
	usedIDs := make(map[string]int)

	for y := 0; y < len(lines); y++ {
		line := lines[y]
		left, right, ok := detectBoxBorder(line)
		if !ok {
			continue
		}

		end := -1
		for z := y + 1; z < len(lines); z++ {
			l2, r2, ok2 := detectBoxBorder(lines[z])
			if ok2 && l2 == left && r2 == right {
				end = z
				break
			}
		}
		if end == -1 || end == y+1 {
			continue
		}

		var parts []string
		for k := y + 1; k < end; k++ {
			interior, ok := extractBoxInterior(lines[k], left, right)
			if ok && strings.TrimSpace(interior) != "" {
				parts = append(parts, strings.TrimSpace(interior))
			}
		}
		if len(parts) > 0 {
			label := strings.Join(parts, " / ")
			id := makeUniqueID(label, usedIDs)
			nodes = append(nodes, &positionedNode{
				ID:    id,
				Label: label,
				X:     (left + right) / 2,
				Y:     (y + end) / 2,
			})
			regions = append(regions, boxRegion{top: y, bottom: end, left: left, right: right})
			y = end
		}
	}

	for y, line := range lines {
		segments := extractLineLabelSegments(line)
		for _, seg := range segments {
			start, end := seg.start, seg.end
			if insideBoxRegion(y, (start+end)/2, regions) {
				continue
			}
			label := strings.TrimSpace(seg.text)
			if looksLikeConnector(label) {
				continue
			}
			id := makeUniqueID(label, usedIDs)
			nodes = append(nodes, &positionedNode{
				ID:    id,
				Label: label,
				X:     (start + end) / 2,
				Y:     y,
			})
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("could not parse any labeled nodes")
	}

	return dedupeNodes(nodes), nil
}

type lineSegment struct {
	start int
	end   int
	text  string
}

// extractLineLabelSegments splits a line into label candidates using 2+ spaces as separators,
// while preserving single spaces inside labels.
func extractLineLabelSegments(line string) []lineSegment {
	var out []lineSegment
	start := -1
	spaceRun := 0

	flush := func(end int) {
		if start == -1 {
			return
		}
		text := strings.TrimSpace(line[start:end])
		if text != "" {
			leftTrim := len(line[start:end]) - len(strings.TrimLeft(line[start:end], " "))
			rightTrim := len(line[start:end]) - len(strings.TrimRight(line[start:end], " "))
			segStart := start + leftTrim
			segEnd := end - rightTrim
			out = append(out, lineSegment{start: segStart, end: segEnd, text: text})
		}
		start = -1
	}

	for i := 0; i < len(line); i++ {
		if line[i] == ' ' {
			if start != -1 {
				spaceRun++
				if spaceRun >= 2 {
					flush(i - 1)
				}
			}
			continue
		}

		if start == -1 {
			start = i
		}
		spaceRun = 0
	}

	if start != -1 {
		flush(len(line))
	}

	return out
}

func extractEdges(lines []string, nodes []*positionedNode) [][2]string {
	var edges [][2]string
	seen := make(map[[2]string]bool)

	for y, line := range lines {
		for x, ch := range line {
			if ch != '|' {
				continue
			}
			up := nearestNode(nodes, x, y, -1)
			down := nearestNode(nodes, x, y, 1)
			if up != nil && down != nil && up.ID != down.ID {
				e := [2]string{up.ID, down.ID}
				if !seen[e] {
					seen[e] = true
					edges = append(edges, e)
				}
			}
		}

		if strings.ContainsRune(line, '/') || strings.ContainsRune(line, '\\') {
			parent := nearestParentAbove(nodes, y)
			if parent == nil {
				continue
			}
			children := nextRowChildren(nodes, y)
			for _, c := range children {
				e := [2]string{parent.ID, c.ID}
				if !seen[e] && c.ID != parent.ID {
					seen[e] = true
					edges = append(edges, e)
				}
			}
		}
	}

	// Implicit continuity rule:
	// if a node has no explicit outgoing edge, connect it to the nearest node below
	// when their X coordinates are close. This recovers common stacked blocks where
	// the ASCII art omits an explicit connector between adjacent rows.
	hasOut := make(map[string]bool)
	for _, e := range edges {
		hasOut[e[0]] = true
	}

	const continuityXTolerance = 8
	for _, src := range nodes {
		if hasOut[src.ID] {
			continue
		}

		var best *positionedNode
		bestScore := 1 << 30
		for _, dst := range nodes {
			if dst.Y <= src.Y || dst.ID == src.ID {
				continue
			}
			dx := dst.X - src.X
			if dx < 0 {
				dx = -dx
			}
			if dx > continuityXTolerance {
				continue
			}
			dy := dst.Y - src.Y
			score := dy*100 + dx
			if score < bestScore {
				bestScore = score
				best = dst
			}
		}

		if best != nil {
			e := [2]string{src.ID, best.ID}
			if !seen[e] {
				seen[e] = true
				edges = append(edges, e)
				hasOut[src.ID] = true
			}
		}
	}

	return edges
}

func detectBoxBorder(line string) (int, int, bool) {
	left := strings.IndexRune(line, '+')
	right := strings.LastIndex(line, "+")
	if left == -1 || right <= left+1 {
		return 0, 0, false
	}
	mid := line[left+1 : right]
	if strings.Trim(mid, "-") != "" {
		return 0, 0, false
	}
	return left, right, true
}

func extractBoxInterior(line string, left, right int) (string, bool) {
	if left < 0 || right >= len(line) || right <= left {
		return "", false
	}
	if rune(line[left]) != '|' || rune(line[right]) != '|' {
		return "", false
	}
	return line[left+1 : right], true
}

func insideBoxRegion(y, x int, regions []boxRegion) bool {
	for _, r := range regions {
		if y >= r.top && y <= r.bottom && x >= r.left && x <= r.right {
			return true
		}
	}
	return false
}

func looksLikeConnector(label string) bool {
	trim := strings.TrimSpace(label)
	if trim == "" {
		return true
	}
	for _, ch := range trim {
		if strings.ContainsRune("-|/\\+", ch) {
			continue
		}
		return false
	}
	return true
}

func sanitizeID(label string) string {
	var b strings.Builder
	for i, ch := range strings.ToUpper(label) {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			if i == 0 && ch >= '0' && ch <= '9' {
				b.WriteRune('N')
				b.WriteRune('_')
			}
			b.WriteRune(ch)
			continue
		}
		if ch == ' ' || ch == '-' {
			if s := b.String(); s != "" && s[len(s)-1] != '_' {
				b.WriteRune('_')
			}
		}
	}
	id := strings.Trim(b.String(), "_")
	if id == "" {
		return "NODE"
	}
	if id[0] >= '0' && id[0] <= '9' {
		return "N_" + id
	}
	return id
}

func makeUniqueID(label string, used map[string]int) string {
	base := sanitizeID(label)
	if used[base] == 0 {
		used[base] = 1
		return base
	}
	used[base]++
	return fmt.Sprintf("%s_%d", base, used[base])
}

func dedupeNodes(nodes []*positionedNode) []*positionedNode {
	type key struct{ x, y int }
	seen := make(map[key]*positionedNode)
	for _, n := range nodes {
		k := key{n.X, n.Y}
		if existing, ok := seen[k]; ok {
			if len(n.Label) > len(existing.Label) {
				seen[k] = n
			}
			continue
		}
		seen[k] = n
	}
	out := make([]*positionedNode, 0, len(seen))
	for _, n := range seen {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y == out[j].Y {
			return out[i].X < out[j].X
		}
		return out[i].Y < out[j].Y
	})
	return out
}

func nearestNode(nodes []*positionedNode, x, y, dir int) *positionedNode {
	const xTolerance = 4
	var best *positionedNode
	bestDist := 1 << 30
	for _, n := range nodes {
		if dir < 0 && n.Y >= y {
			continue
		}
		if dir > 0 && n.Y <= y {
			continue
		}
		dx := n.X - x
		if dx < 0 {
			dx = -dx
		}
		if dx > xTolerance {
			continue
		}
		dy := n.Y - y
		if dy < 0 {
			dy = -dy
		}
		if dy < bestDist {
			bestDist = dy
			best = n
		}
	}
	return best
}

func nearestParentAbove(nodes []*positionedNode, y int) *positionedNode {
	var best *positionedNode
	bestDist := 1 << 30
	for _, n := range nodes {
		if n.Y >= y {
			continue
		}
		dy := y - n.Y
		if dy < bestDist {
			bestDist = dy
			best = n
		}
	}
	return best
}

func nextRowChildren(nodes []*positionedNode, y int) []*positionedNode {
	nextY := -1
	for _, n := range nodes {
		if n.Y > y && (nextY == -1 || n.Y < nextY) {
			nextY = n.Y
		}
	}
	if nextY == -1 {
		return nil
	}
	out := make([]*positionedNode, 0)
	for _, n := range nodes {
		if n.Y == nextY {
			out = append(out, n)
		}
	}
	return out
}

func compressCoordinates(nodes []*positionedNode) {
	xs := make([]int, 0, len(nodes))
	ys := make([]int, 0, len(nodes))
	seenX := make(map[int]bool)
	seenY := make(map[int]bool)

	for _, n := range nodes {
		if !seenX[n.X] {
			seenX[n.X] = true
			xs = append(xs, n.X)
		}
		if !seenY[n.Y] {
			seenY[n.Y] = true
			ys = append(ys, n.Y)
		}
	}

	sort.Ints(xs)
	sort.Ints(ys)
	xIndex := make(map[int]int)
	yIndex := make(map[int]int)
	for i, x := range xs {
		xIndex[x] = i
	}
	for i, y := range ys {
		yIndex[y] = i
	}

	for _, n := range nodes {
		n.X = xIndex[n.X]
		n.Y = yIndex[n.Y]
	}
}

func inferNodeType(label string) string {
	l := strings.ToLower(strings.TrimSpace(label))
	switch {
	case strings.Contains(l, "start"):
		return "start"
	case strings.Contains(l, "end"):
		return "end"
	case strings.Contains(l, "decision"):
		return "decision"
	default:
		return "component"
	}
}
