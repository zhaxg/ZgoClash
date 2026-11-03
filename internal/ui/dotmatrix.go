package ui

import (
	"math"

	mu "github.com/egoist/mygo/ui"
)

// The traffic cards show their traffic as a matrix of dots rather than a
// line: the box is tinted, every dot that carries no traffic is pale, and
// the ones that do are filled from the left, in a ramp that runs from pale
// blue at the left to hot at the right.

// dotMatrix is the grid a traffic card's traffic is drawn on.
type dotMatrix struct {
	// cols and rows are how many dots the box holds.
	cols, rows int
	// stepX and stepY are the pitch of the dots, in DIPs.
	stepX, stepY float32
	// dot is the radius of one dot, in DIPs.
	dot float32
}

// newDotMatrix lays a matrix out in a box of the given size. The traffic
// cards fix it at twenty columns and five rows, and the dots fill the box
// over the inset the card leaves.
func newDotMatrix(w, h float32) dotMatrix {
	const cols = 20
	const rows = 5
	stepX := w / cols
	stepY := h / rows
	dot := min(stepX, stepY) * 0.36 / 2
	return dotMatrix{cols: cols, rows: rows, stepX: stepX, stepY: stepY, dot: dot}
}

// draw paints the matrix. The values are the traffic of the last stretch,
// oldest first; each one lights the bottom of its column, so the newest
// column at the right is the current one.
func (m dotMatrix) draw(pt *mu.Painter, r mu.Rect, values []int64, p palette) {
	if m.cols <= 0 || m.rows <= 0 {
		return
	}
	// The dots are centred in their cells, so the first and last leave half
	// a cell at the box's edge.
	ox := r.X + m.stepX/2
	oy := r.Y + m.stepY/2

	// The largest value in the window is the top of the scale.
	var max int64
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		max = 1
	}

	// The window fills the matrix: the newest samples are the rightmost
	// columns, and the older ones are dropped.
	if len(values) > m.cols {
		values = values[len(values)-m.cols:]
	}
	offset := m.cols - len(values)
	if offset < 0 {
		offset = 0
	}
	for col := 0; col < m.cols; col++ {
		idx := col - offset
		lit := 0
		if idx >= 0 && idx < len(values) {
			lit = int(math.Round(float64(values[idx]) / float64(max) * float64(m.rows)))
			if lit > m.rows {
				lit = m.rows
			}
			if lit < 0 {
				lit = 0
			}
		}
		// The ramp runs across the matrix, so the newest dots are the hot
		// ones.
		at := float32(0)
		if m.cols > 1 {
			at = float32(col) / float32(m.cols-1)
		}
		for row := 0; row < m.rows; row++ {
			x := ox + float32(col)*m.stepX
			y := oy + float32(row)*m.stepY
			color := p.ChartIdle
			if row >= m.rows-lit {
				color = dotRamp(p, at)
			}
			pt.Fill(mu.Rect{X: x - m.dot, Y: y - m.dot, W: m.dot * 2, H: m.dot * 2}, color, m.dot)
		}
	}
}

// dotRamp is the colour a filled dot takes: pale blue at the start of the
// run, deep blue in the middle, and hot at the end.
func dotRamp(p palette, at float32) mu.Color {
	switch {
	case at < 0.55:
		return p.ChartLight.Mix(p.ChartDeep, at/0.55)
	case at < 0.82:
		return p.ChartDeep.Mix(p.ChartHot, (at-0.55)/0.27)
	default:
		return p.ChartHot
	}
}
