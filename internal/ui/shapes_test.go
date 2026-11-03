package ui

import (
	"image/png"
	"os"
	"testing"

	mu "github.com/egoist/mygo/ui"
)

// TestRenderShapes draws the three dashboard shapes, so that their geometry
// can be checked against the ones FlClash shows.
func TestRenderShapes(t *testing.T) {
	img := mu.Render(func(c *mu.Context) {
		mu.Column(c).Padding(20).Gap(20).Background(mu.Hex("#ffffff")).Children(func() {
			mu.Row(c).Gap(20).AlignItems(mu.Center).Children(func() {
				mu.Box(c).Size(140, 140).Draw(func(p *mu.Painter, r mu.Rect) {
					shapeCookie9.draw(p, r.X, r.Y, r.W, 0, mu.Hex("#5f6ffb"))
				})
				mu.Box(c).Size(140, 140).Draw(func(p *mu.Painter, r mu.Rect) {
					shapeClover4.draw(p, r.X, r.Y, r.W, 0, mu.Hex("#e0518a"))
				})
				mu.Box(c).Size(140, 140).Draw(func(p *mu.Painter, r mu.Rect) {
					shapeCookie12.draw(p, r.X, r.Y, r.W, 0, mu.Hex("#0ea5a5"))
				})
			})
			// The button's morph, at a few of its steps.
			mu.Row(c).Gap(20).Children(func() {
				for _, step := range []int{0, 6, 12, 18, 24} {
					mu.Box(c).Size(120, 120).Draw(func(p *mu.Painter, r mu.Rect) {
						shape := morphPolygon(step)
						shape.draw(p, r.X, r.Y, r.W, morphTurn(float32(step)/morphSteps), mu.Hex("#5f6ffb"))
					})
				}
			})
		})
	}, 700, 340, 2)

	f, err := os.Create("shapes_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
