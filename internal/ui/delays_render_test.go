package ui

import (
	"image"
	"image/png"
	"os"
	"testing"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/engine"
)

// TestRenderDelayHeights draws the same node card twice, once while its
// delay is measured and once not, and checks that the mark of the node in
// use stays in the same place.
func TestRenderDelayHeights(t *testing.T) {
	p := paletteFor(mu.LightTheme())
	name := "🇭🇰 香港家宽01"
	group := engine.ProxyGroupView{Name: "节点选择", Now: name, All: []string{name}}
	proxies := map[string]engine.Proxy{
		name: {Name: name, Type: "Shadowsocks", History: []engine.DelayHistory{{Delay: 88}}},
	}

	render := func(testing bool) *image.RGBA {
		u := &UI{
			busy:    map[string]bool{},
			proxies: proxies,
			groups:  []engine.ProxyGroupView{group},
			delays:  map[string]int{},
			testing: map[string]bool{},
		}
		if testing {
			u.testing[name] = true
		}
		return mu.Render(func(c *mu.Context) {
			mu.Column(c).FillWidth().Background(p.Page).Children(func() {
				u.nodeCard(c, p, group, name)
			})
		}, 300, 80, 2)
	}

	// The mark is the only green thing on the card: its place tells whether
	// the card changed its height.
	mark := func(img *image.RGBA) (int, int) {
		top, left := -1, -1
		for y := range img.Bounds().Max.Y {
			for x := range img.Bounds().Max.X {
				r, g, b, _ := img.At(x, y).RGBA()
				if g > 0x8000 && r < 0x8000 && b < 0x8000 {
					if top < 0 {
						top, left = y, x
					}
				}
			}
		}
		return top, left
	}

	quietTop, quietLeft := mark(render(false))
	busyTop, busyLeft := mark(render(true))
	if quietTop < 0 {
		t.Fatal("the mark is not on the card")
	}
	if quietTop != busyTop || quietLeft != busyLeft {
		t.Errorf("the mark moved: %d,%d became %d,%d", quietLeft, quietTop, busyLeft, busyTop)
	}

	f, err := os.Create("delays_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, render(true)); err != nil {
		t.Fatal(err)
	}
}
