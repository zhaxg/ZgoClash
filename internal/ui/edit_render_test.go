package ui

import (
	"image/png"
	"os"
	"testing"

	mu "github.com/egoist/mygo/ui"
)

// TestRenderEditModal draws the edit dialog of a profile, so the fields it
// holds can be looked at.
func TestRenderEditModal(t *testing.T) {
	p := paletteFor(mu.LightTheme())
	u := &UI{
		busy:      map[string]bool{},
		editOpen:  true,
		editName:  "SG 新加坡专线02|BGP|流媒体",
		editURL:   "https://example.com/subscribe?token=abcdefghijklmnopqrstuvwxyz0123456789&flag=clash",
		editIsURL: true,
	}
	img := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u.profileModals(c, p)
		})
	}, 520, 320, 1)

	f, err := os.Create("edit_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
