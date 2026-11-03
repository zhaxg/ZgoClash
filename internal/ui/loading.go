package ui

import (
	"math"
	"time"

	mu "github.com/egoist/mygo/ui"
)

// delayLineHeight is the height of the delay text of a node, in DIPs. The
// indicator that replaces it while the delay is measured is as tall, so that
// the card does not change its height.
const delayLineHeight = 16

// loadingDots is the small indicator that shows a delay is being measured:
// three dots that swell in turn, as the icon of the app does. It is drawn by
// hand, because the SVG the design uses animates itself with SMIL, which
// the renderer does not run. Its height is the height of the delay text it
// replaces, so that the mark above it does not move.
func loadingDots(c *mu.Context, color mu.Color) mu.Element {
	const cycle = 1500 * time.Millisecond
	return mu.Box(c).Size(15, delayLineHeight).Draw(func(p *mu.Painter, r mu.Rect) {
		ms := p.Now().UnixMilli() % int64(cycle/time.Millisecond)
		p.After(time.Duration(int64(cycle/time.Millisecond)-ms) * time.Millisecond)
		cy := r.Y + r.H/2
		// The leftmost dot leads, the others follow a third of a turn
		// later.
		for _, dot := range []struct {
			x     float32
			delay float64
		}{{0.25, 0}, {0.5, 1.0 / 3}, {0.75, 2.0 / 3}} {
			phase := math.Mod(float64(ms)/float64(cycle/time.Millisecond)-dot.delay, 1)
			if phase >= 2.0/3 {
				continue // the dot rests in the last third of its turn
			}
			rad := float32(1.5 * math.Sin(phase*1.5*math.Pi))
			if rad <= 0.1 {
				continue
			}
			cx := r.X + r.W*dot.x
			p.Fill(mu.Rect{X: cx - rad, Y: cy - rad, W: rad * 2, H: rad * 2}, color, rad)
		}
	})
}
