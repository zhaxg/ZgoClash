package ui

import (
	"testing"

	mu "github.com/egoist/mygo/ui"
)

// TestNestedMenuButton checks that clicking a menu button inside a button
// does not also click the button around it.
func TestNestedMenuButton(t *testing.T) {
	outerClicks := 0
	menuItems := 0
	tester := mu.NewTester(func(c *mu.Context) {
		mu.Column(c).Background(mu.Hex("#ffffff")).Children(func() {
			card := mu.ButtonBase(c).Fill().Padding(18).Radius(14).Background(mu.Hex("#eef")).Label("卡片").Children(func() {
				mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
					mu.Text(c, "订阅名")
					more := mu.ButtonBase(c).Size(34, 34).Radius(9).Label("更多").Children(func() {
						mu.Text(c, "更多")
					})
					more.Menu(func(m *mu.Menu) {
						if m.Item("同步").Chosen() {
							menuItems++
						}
					})
				})
			})
			if card.Clicked() {
				outerClicks++
			}
		})
	}, 400, 200)

	r, ok := tester.Find("更多")
	if !ok {
		t.Fatal("the menu button is not on screen")
	}
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	if outerClicks != 0 {
		t.Errorf("the card was clicked too: %d", outerClicks)
	}

	// The card itself still switches the profile.
	name, ok := tester.Find("订阅名")
	if !ok {
		t.Fatal("the card is not on screen")
	}
	tester.ClickAt(name.X+name.W/2, name.Y+name.H/2)
	if outerClicks != 1 {
		t.Errorf("the card was not clicked: %d", outerClicks)
	}
}
