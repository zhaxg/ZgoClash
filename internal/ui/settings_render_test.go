package ui

import (
	"image/png"
	"os"
	"testing"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

// TestRenderSettings draws the settings page, so the new grid and the
// external controller row can be looked at without a desktop.
func TestRenderSettings(t *testing.T) {
	u := &UI{busy: map[string]bool{}}
	snap := app.Snapshot{
		Running:        true,
		Starting:       false,
		AppVersion:     "1.0.0",
		CoreVersion:    "1.19.33",
		Uptime:         3661,
		MixedPort:      7890,
		ControllerPort: 9090,
		Home:           `C:\Users\ZHAXG\AppData\Roaming\ZgoClash`,
		Settings: app.Settings{
			Theme:         "system",
			MixedPort:     7890,
			TunStack:      "gvisor",
			TestURL:       "http://www.gstatic.com/generate_204",
			LogLevel:      "info",
			Mode:          "rule",
			AllowLan:      false,
			IPv6:          false,
			AutoStartCore: true,
			CloseToTray:   true,
		},
	}
	p := paletteFor(mu.LightTheme())
	img := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u.coreSection(c, snap, p)
		})
	}, 900, 300, 1)

	f, err := os.Create("settings_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}

	u2 := &UI{busy: map[string]bool{}, portText: "7890", controllerText: "9090"}
	img2 := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u2.networkSection(c, snap, p)
		})
	}, 900, 640, 1)

	f2, err := os.Create("network_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	if err := png.Encode(f2, img2); err != nil {
		t.Fatal(err)
	}

	u3 := &UI{busy: map[string]bool{}}
	img3 := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u3.coreSection(c, snap, p)
			u3.aboutSection(c, snap, p)
		})
	}, 900, 520, 1)

	f3, err := os.Create("about_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f3.Close()
	if err := png.Encode(f3, img3); err != nil {
		t.Fatal(err)
	}
}

// TestRenderMarks draws the subscriptions page and the node list, so the
// marks that show what is in use can be looked at.
func TestRenderMarks(t *testing.T) {
	p := paletteFor(mu.LightTheme())
	snap := app.Snapshot{
		AppVersion: "1.0.0",
		Profiles: []app.Profile{
			{ID: "a", Name: "🇸🇬 新加坡专线02|BGP|流媒体", Type: "url", URL: "https://example.com/sub.yaml", ProxyCount: 42, RuleCount: 100, GroupCount: 12, UpdatedAt: 1},
			{ID: "b", Name: "手动配置", Type: "manual", ProxyCount: 3, RuleCount: 10, GroupCount: 2},
		},
		CurrentProfileID: "a",
	}
	u := &UI{busy: map[string]bool{}}
	img := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u.profilesPage(c, snap, p)
		})
	}, 900, 320, 1)

	f, err := os.Create("marks_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}

	group := engine.ProxyGroupView{
		Name: "🚀 节点选择",
		Now:  "🇭🇰 香港家宽01",
		All:  []string{"🇭🇰 香港家宽01", "🇯🇵 日本家宽02", "🇺🇸 美国高速03"},
	}
	proxies := map[string]engine.Proxy{}
	for _, name := range group.All {
		proxies[name] = engine.Proxy{Name: name, Type: "Shadowsocks", History: []engine.DelayHistory{{Delay: 88}}}
	}
	u2 := &UI{
		busy:    map[string]bool{},
		proxies: proxies,
		groups:  []engine.ProxyGroupView{group},
		delays:  map[string]int{},
	}
	img2 := mu.Render(func(c *mu.Context) {
		mu.Column(c).FillWidth().Background(p.Page).Children(func() {
			u2.proxiesPage(c, app.Snapshot{Running: true}, p)
		})
	}, 900, 420, 1)

	f2, err := os.Create("nodes_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	if err := png.Encode(f2, img2); err != nil {
		t.Fatal(err)
	}

	img3 := mu.Render(func(c *mu.Context) {
		mu.Row(c).Gap(16).AlignItems(mu.Center).Background(mu.Hex("#ffffff")).Children(func() {
			mu.Icon(c, iconClear).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconBoltFilled).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconCheckCircle).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconUnlink).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconPlus).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconSync).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconDots).FontSize(24).TextColor(mu.Hex("#5f6ffb"))
			mu.Icon(c, iconGithub).FontSize(24)
		})
	}, 360, 60, 2)

	f3, err := os.Create("icons_test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f3.Close()
	if err := png.Encode(f3, img3); err != nil {
		t.Fatal(err)
	}
}
