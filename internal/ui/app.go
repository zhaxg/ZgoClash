package ui

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

type trafficSample struct {
	down int64
	up   int64
}

// UI is the window's whole state. It is built on the main thread only, except
// where a goroutine hands work back through Window.Update.
type UI struct {
	ctrl *app.Controller
	win  *mygo.Window
	mark *mu.Bitmap

	page int

	samples    []trafficSample
	lastSample time.Time
	lastRate   engine.Traffic

	systemDark    bool
	systemThemeAt time.Time

	// Dashboard extras fetched off-thread.
	exitInfo   app.ExitInfo
	exitErr    string
	exitLoaded bool
	services   []app.ServiceStatus
	servErr    string
	servLoaded bool
	dashFeed   int // 0 live, 1 recent, 2 dns
	peakDown   int64
	peakUp     int64

	// Proxies page.
	proxies       map[string]engine.Proxy
	groups        []engine.ProxyGroupView
	proxyErr      string
	proxyLoading  bool
	proxyLoadedAt time.Time
	groupIdx      int
	nodeQuery     string
	nodeSort      int
	delays        map[string]int
	testing       map[string]bool
	testingGroup  bool
	grid          mu.GridState
	visible       []string

	// Profiles page.
	busy       map[string]bool
	addOpen    bool
	addName    string
	addURL     string
	manualOpen bool
	manualName string
	manualText string
	editOpen   bool
	editID     string
	editName   string
	editURL    string
	editIsURL  bool
	deleteOpen bool
	deleteID   string
	deleteName string

	// Connections page.
	connQuery    string
	connDetail   *engine.Connection
	connCloseAll bool
	connRows     []engine.Connection
	connList     mu.ListState

	// Tools page.
	portText       string
	testURLText    string
	controllerText string
	autoUpdateIdx  int
	logFilter      int
	logFollow      bool
	logScroll      mu.ScrollState
	logReady       bool
	configOpen     bool
	configText     string

	sidebarExpanded bool
	prevPage        int

	active  atomic.Bool
	toastMu sync.Mutex
	toasts  []mu.Toast
}

func New(ctrl *app.Controller, mark []byte) *UI {
	ensureFonts()
	u := &UI{
		ctrl:            ctrl,
		busy:            map[string]bool{},
		delays:          map[string]int{},
		testing:         map[string]bool{},
		logFollow:       true,
		autoUpdateIdx:   2,
		sidebarExpanded: false,
	}
	if b, err := mu.DecodeBitmap(mark); err == nil {
		u.mark = b
	}
	u.active.Store(true)
	return u
}

func (u *UI) SetWindow(win *mygo.Window) { u.win = win }

// SetActive tells the UI whether its window is on screen. While it is away,
// the caches the pages built are dropped and the window is not asked to
// redraw; the pages fill them again when it returns.
func (u *UI) SetActive(active bool) {
	u.active.Store(active)
	if active {
		if u.win != nil && !u.win.IsDestroyed() {
			u.win.Invalidate()
		}
		return
	}
	u.proxies = nil
	u.groups = nil
	u.delays = nil
	u.visible = nil
	u.samples = nil
	u.lastSample = time.Time{}
	u.lastRate = engine.Traffic{}
	u.connRows = nil
	u.connDetail = nil
	u.services = nil
	u.servLoaded = false
	u.exitInfo = app.ExitInfo{}
	u.exitLoaded = false
	u.proxyErr = ""
	u.proxyLoadedAt = time.Time{}
}

// Notify shows a message from any goroutine: the controller's callbacks and
// the app's background work report this way.
func (u *UI) Notify(level, message string) {
	u.toastMu.Lock()
	u.toasts = append(u.toasts, mu.Toast{Type: level, Title: message})
	u.toastMu.Unlock()
	if u.active.Load() && u.win != nil && !u.win.IsDestroyed() {
		u.win.Invalidate()
	}
}

func (u *UI) drainToasts(c *mu.Context) {
	u.toastMu.Lock()
	list := u.toasts
	u.toasts = nil
	u.toastMu.Unlock()
	for _, t := range list {
		c.AddToast(t)
	}
}

// run does slow work off the main thread, then reports the outcome. Actions
// with the same key cannot overlap.
func (u *UI) run(key string, fn func() error, done func()) {
	if u.busy[key] || u.win == nil {
		return
	}
	u.busy[key] = true
	go func() {
		err := fn()
		u.win.Update(func() {
			delete(u.busy, key)
			if err != nil {
				if err.Error() != "已取消" {
					u.Notify("error", err.Error())
				}
				return
			}
			if done != nil {
				done()
			}
		})
	}()
}

func (u *UI) View(c *mu.Context) {
	ensureFonts()
	snap := u.ctrl.Snapshot()
	u.applyTheme(c, snap.Settings.Theme)
	t := c.Theme()
	p := paletteFor(t)
	u.drainToasts(c)
	u.sampleTraffic(snap)

	if snap.Running && (u.page == 0 || u.page == 1) && !u.proxyLoading && time.Since(u.proxyLoadedAt) > 12*time.Second {
		u.loadProxies()
	}
	if u.page == 0 && snap.Running {
		u.loadExit(false)
		if !u.servLoaded {
			u.loadServices()
		}
	}
	if u.page == 0 || u.page == 1 {
		c.After(time.Second)
	}

	bar := c.TitleBar()
	titleH := bar.Height
	if titleH < 44 {
		titleH = 44
	}

	mu.Column(c).Fill().Background(p.Page).Children(func() {
		u.titleBar(c, snap, p, titleH, bar)
		mu.Row(c).Grow(1).AlignItems(mu.Stretch).Children(func() {
			u.sidebar(c, p)
			mu.Column(c).Grow(1).FillHeight().Children(func() {
				u.pageBody(c, snap, p)
			})
		})
	})
}

func (u *UI) titleBar(c *mu.Context, snap app.Snapshot, p palette, height float32, bar mu.TitleBar) {
	mu.Row(c).Height(height).AlignItems(mu.Center).Gap(10).
		Padding(0, bar.Right+18, 0, bar.Left+18).
		BorderWidth(0, 0, 1, 0).BorderColor(p.Border).
		DragWindow().Children(func() {
		if u.mark != nil {
			mu.Image(c, u.mark).Size(26, 26).Fit(mu.Contain)
		} else {
			mu.Box(c).Size(26, 26).Radius(8).Background(p.Accent).Center().Children(func() {
				mu.Text(c, "Z").FontSize(14).Bold().TextColor(p.AccentText)
			})
		}
		mu.Text(c, "ZgoClash").FontSize(14).Bold()
		mu.Text(c, "·").FontSize(14).TextColor(p.Muted)
		mu.Text(c, pageNames[u.page]).FontSize(13).TextColor(p.Muted)
		mu.Spacer(c)
	})
}

const (
	sidebarCompact  float32 = 48
	sidebarExpanded float32 = 220
	sidebarItemH    float32 = 40
	sidebarItemGap  float32 = 4
	sidebarInset    float32 = 4
	sidebarIcon     float32 = 24
	sidebarIconBox  float32 = 24
)

func (u *UI) sidebar(c *mu.Context, p palette) {
	target := float32(0)
	if u.sidebarExpanded {
		target = 1
	}
	col := mu.Column(c).Padding(sidebarInset, sidebarInset).AlignItems(mu.Start).
		Background(p.Sidebar).BorderWidth(0, 1, 0, 0).BorderColor(p.Border)
	progress := col.AnimateWith("sidebar", target, 250*time.Millisecond, mu.EaseInOut)
	width := sidebarCompact + (sidebarExpanded-sidebarCompact)*progress
	col.Width(width)
	inner := width - sidebarInset*2

	col.Children(func() {
		u.sidebarToggle(c, p, progress)
		mu.Box(c).Height(8)
		u.sidebarNav(c, p, progress, inner)
		mu.Spacer(c)
	})
}

func (u *UI) sidebarToggle(c *mu.Context, p palette, progress float32) {
	el := mu.ButtonBase(c).Width(sidebarCompact - sidebarInset*2).Height(sidebarItemH).
		Radius(10).AlignItems(mu.Center).Justify(mu.Center).Label("显示或隐藏侧边栏")
	if el.Hovered() {
		el.Background(p.CardHover)
	}
	el.Children(func() {
		mu.Box(c).Size(sidebarIconBox, sidebarIconBox).Draw(func(pt *mu.Painter, r mu.Rect) {
			drawSidebarGlyph(pt, r, progress, p.Muted)
		})
	})
	tip := "展开侧边栏"
	if u.sidebarExpanded {
		tip = "收起侧边栏"
	}
	el.Tooltip(tip)
	if el.Clicked() {
		u.sidebarExpanded = !u.sidebarExpanded
	}
}

// drawSidebarGlyph draws FlClash's sidebar toggle: a panel whose dividing line
// moves toward the middle as the sidebar opens, as its three short lines grow.
func drawSidebarGlyph(pt *mu.Painter, r mu.Rect, progress float32, ink mu.Color) {
	scale := min(r.W, r.H) / 24
	at := func(x, y float64) (float32, float32) {
		return r.X + float32(x)*scale, r.Y + float32(y)*scale
	}
	bx, by := at(3, 4.5)
	pt.Stroke(mu.Rect{X: bx, Y: by, W: 18 * scale, H: 15 * scale}, ink, 3.5*scale, 1.8*scale)
	divider := 8.5 + 3*float64(progress)
	x0, y0 := at(divider, 4.5)
	x1, y1 := at(divider, 19.5)
	pt.Line(x0, y0, x1, y1, 1.8*scale, ink)
	for _, y := range []float64{8.5, 11, 13.5} {
		lx0, ly := at(5.5, y)
		lx1, _ := at(divider-2.5, y)
		pt.Line(lx0, ly, lx1, ly, 1.8*scale, ink)
	}
}

func (u *UI) sidebarNav(c *mu.Context, p palette, progress, width float32) {
	nav := mu.Column(c).Width(width).Gap(sidebarItemGap)
	stride := sidebarItemH + sidebarItemGap
	base := float32(u.page)*stride + (sidebarItemH-16)/2
	fast, slow := 200*time.Millisecond, 320*time.Millisecond
	topEase, botEase := mu.EaseOut, mu.EaseInOut
	if u.page < u.prevPage {
		topEase, botEase = mu.EaseInOut, mu.EaseOut
	}
	top := nav.AnimateWith("indicatorTop", base, fast, topEase)
	bottom := nav.AnimateWith("indicatorBottom", base+16, slow, botEase)
	nav.DrawOver(func(pt *mu.Painter, r mu.Rect) {
		pt.Fill(mu.Rect{X: r.X, Y: r.Y + top, W: 3, H: bottom - top}, p.Accent, 1.5)
	})
	nav.Children(func() {
		u.navItem(c, p, 0, iconDashboard, "概览", progress, width)
		u.navItem(c, p, 1, iconProxies, "代理", progress, width)
		u.navItem(c, p, 2, iconProfiles, "订阅", progress, width)
		u.navItem(c, p, 3, iconActivity, "连接", progress, width)
		u.navItem(c, p, 4, iconLogs, "日志", progress, width)
		u.navItem(c, p, 5, iconTools, "设置", progress, width)
	})
}

func (u *UI) navItem(c *mu.Context, p palette, idx int, ic *mu.SVG, label string, progress, width float32) {
	active := u.page == idx
	el := mu.ButtonBase(c).Width(width).Height(sidebarItemH).Radius(10).
		AlignItems(mu.Center).Justify(mu.Start).Gap(12).Padding(0, 8).Label(label)
	if active {
		el.Background(p.Soft)
	} else if el.Hovered() {
		el.Background(p.CardHover)
	}
	el.Children(func() {
		fg := p.Muted
		if active {
			fg = p.Text
		}
		mu.Icon(c, ic).FontSize(sidebarIcon).TextColor(fg)
		alpha := float32(clamp01(float64((progress - 0.35) / 0.45)))
		labelWidth := width - 52
		if alpha > 0.01 && labelWidth > 8 {
			mu.Text(c, label).FontSize(13).TextColor(p.Text.Alpha(alpha)).
				SingleLine().MaxWidth(labelWidth)
		}
	})
	if progress < 0.98 {
		el.Tooltip(label)
	}
	if el.Clicked() {
		u.setPage(idx)
	}
}

func (u *UI) setPage(idx int) {
	if idx == u.page {
		return
	}
	u.prevPage = u.page
	u.page = idx
	if idx == 1 {
		u.loadProxies()
	}
}

func (u *UI) toggleCore() {
	running := u.ctrl.Snapshot().Running
	u.run("core", func() error {
		if running {
			return u.ctrl.Stop()
		}
		return u.ctrl.Start()
	}, nil)
}

// copyText puts text on the clipboard and confirms with a toast.
func (u *UI) copyText(text string) {
	if text == "" {
		return
	}
	mygo.Clipboard.WriteText(text)
	u.Notify("success", "已复制到剪贴板")
}

func (u *UI) pageBody(c *mu.Context, snap app.Snapshot, p palette) {
	switch u.page {
	case 0:
		u.dashboardPage(c, snap, p)
	case 1:
		u.proxiesPage(c, snap, p)
	case 2:
		u.profilesPage(c, snap, p)
	case 3:
		u.connectionsPage(c, snap, p)
	case 4:
		u.logsPage(c, snap, p)
	case 5:
		u.toolsPage(c, snap, p)
	}
}

// ---- shared bits -----------------------------------------------------------

func (u *UI) pageHeader(c *mu.Context, p palette, title, subtitle string, actions func()) {
	mu.Row(c).AlignItems(mu.Center).Gap(12).Children(func() {
		mu.Column(c).Gap(3).Grow(1).Children(func() {
			mu.Text(c, title).FontSize(21).Bold()
			mu.Text(c, subtitle).FontSize(12.5).TextColor(p.Muted)
		})
		if actions != nil {
			actions()
		}
	})
}

func card(c *mu.Context, p palette) mu.Element {
	return mu.Column(c).Padding(18).Gap(12).Radius(14).Background(p.Card).Border(1, p.Border)
}

func cardTitle(c *mu.Context, p palette, title string) {
	mu.Text(c, title).FontSize(14.5).Bold()
}

func pill(c *mu.Context, text string, fg, bg mu.Color) mu.Element {
	return mu.Box(c).Padding(2, 8).Radius(99).Background(bg).Children(func() {
		mu.Text(c, text).FontSize(10.5).TextColor(fg)
	})
}

// iconButton is a square button with an icon in it. disabled turns it off,
// as the test button is while the core is busy.
func (u *UI) iconButton(c *mu.Context, p palette, ic *mu.SVG, label string, disabled ...bool) mu.Element {
	off := len(disabled) > 0 && disabled[0]
	el := mu.ButtonBase(c).Size(34, 34).Radius(9).AlignItems(mu.Center).Justify(mu.Center).Label(label).
		Disabled(off).Children(func() {
		mu.Icon(c, ic).FontSize(16).TextColor(p.Muted)
	})
	if el.Hovered() && !off {
		el.Background(p.Soft)
	}
	return el
}

// switchRow is a setting row: the title and description on the left, the
// switch on the right, and trailing controls after it when the setting needs
// one, as the TUN row's button to restart elevated.
func (u *UI) switchRow(c *mu.Context, p palette, title, desc string, on, enabled bool, cb func(bool), trailing ...func()) {
	mu.Row(c).Padding(8, 0).Gap(14).AlignItems(mu.Center).Children(func() {
		mu.Column(c).Gap(2).Grow(1).Children(func() {
			mu.Text(c, title).FontSize(13)
			if desc != "" {
				mu.Text(c, desc).FontSize(11).TextColor(p.Muted)
			}
		})
		v := on
		if mu.Switch(c, &v).Disabled(!enabled).Changed() && enabled {
			cb(v)
		}
		for _, t := range trailing {
			t()
		}
	})
}

func (u *UI) patchSettings(mut func(*app.Settings)) {
	u.run("settings", func() error {
		s := u.ctrl.Settings()
		mut(&s)
		return u.ctrl.UpdateSettings(s)
	}, nil)
}

func modeIndex(mode string) int {
	switch mode {
	case "global":
		return 1
	case "direct":
		return 2
	default:
		return 0
	}
}

func modeString(idx int) string {
	switch idx {
	case 1:
		return "global"
	case 2:
		return "direct"
	default:
		return "rule"
	}
}

func modeName(mode string) string {
	switch mode {
	case "global":
		return "全局"
	case "direct":
		return "直连"
	default:
		return "规则"
	}
}

func logLevelIndex(level string) int {
	switch level {
	case "silent":
		return 0
	case "error":
		return 1
	case "warning":
		return 2
	case "debug":
		return 4
	default:
		return 3
	}
}

func logLevelString(idx int) string {
	switch idx {
	case 0:
		return "silent"
	case 1:
		return "error"
	case 2:
		return "warning"
	case 4:
		return "debug"
	default:
		return "info"
	}
}

// ---- proxies loading -------------------------------------------------------

func (u *UI) loadProxies() {
	if u.proxyLoading || u.win == nil {
		return
	}
	if !u.ctrl.Snapshot().Running {
		u.proxyErr = "核心未运行，先启动核心后再试"
		u.proxies = nil
		u.groups = nil
		return
	}
	u.proxyLoading = true
	u.proxyErr = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		proxies, groups, err := u.ctrl.ProxyGroups(ctx)
		cancel()
		u.win.Update(func() {
			u.proxyLoading = false
			if !u.active.Load() {
				// The window went away while this ran: keep the result
				// from holding memory, and let the next show reload.
				u.proxyLoadedAt = time.Time{}
				return
			}
			u.proxyLoadedAt = time.Now()
			if err != nil {
				u.proxyErr = err.Error()
				return
			}
			u.proxies = proxies
			u.groups = groups
			if u.groupIdx >= len(groups) {
				u.groupIdx = 0
			}
		})
	}()
}

func (u *UI) currentGroup() (engine.ProxyGroupView, bool) {
	if u.groupIdx < 0 || u.groupIdx >= len(u.groups) {
		return engine.ProxyGroupView{}, false
	}
	return u.groups[u.groupIdx], true
}

func (u *UI) delayFor(name string) int {
	if d, ok := u.delays[name]; ok {
		return d
	}
	if p, ok := u.proxies[name]; ok {
		if d, ok := p.LastDelay(); ok {
			return d
		}
	}
	return 0
}

func (u *UI) selectNode(group, name string) {
	u.run("select:"+group, func() error {
		return u.ctrl.SelectProxy(group, name)
	}, func() {
		for i := range u.groups {
			if u.groups[i].Name == group {
				u.groups[i].Now = name
			}
		}
		if p, ok := u.proxies[group]; ok {
			p.Now = name
			u.proxies[group] = p
		}
	})
}

func (u *UI) testNode(name string) {
	if u.testing[name] {
		return
	}
	u.testing[name] = true
	go func() {
		d, err := u.ctrl.TestDelay(name, 5000)
		u.win.Update(func() {
			delete(u.testing, name)
			if err != nil {
				d = 0
			}
			u.delays[name] = d
		})
	}()
}

func (u *UI) testGroup() {
	g, ok := u.currentGroup()
	if !ok || u.testingGroup {
		return
	}
	u.testingGroup = true
	go func() {
		results, err := u.ctrl.TestGroupDelay(g.Name, 5000)
		u.win.Update(func() {
			u.testingGroup = false
			if err != nil {
				u.Notify("error", err.Error())
				return
			}
			for name, d := range results {
				u.delays[name] = d
			}
			u.proxyLoadedAt = time.Time{}
		})
	}()
}

func (u *UI) sampleTraffic(snap app.Snapshot) {
	if !u.lastSample.IsZero() && time.Since(u.lastSample) < 900*time.Millisecond {
		return
	}
	if !snap.Running && len(u.samples) == 0 {
		return
	}
	// The core reports the traffic of the last second, so the samples are
	// rates and the peak is the largest of them.
	u.lastSample = time.Now()
	u.lastRate = snap.Traffic
	u.samples = append(u.samples, trafficSample{down: snap.Traffic.Down, up: snap.Traffic.Up})
	if snap.Traffic.Down > u.peakDown {
		u.peakDown = snap.Traffic.Down
	}
	if snap.Traffic.Up > u.peakUp {
		u.peakUp = snap.Traffic.Up
	}
	if len(u.samples) > 60 {
		u.samples = append(u.samples[:0], u.samples[len(u.samples)-60:]...)
	}
}

// loadExit fetches where the traffic leaves by, and loadServices probes a
// couple of well-known sites. Both run off-thread and report back through
// Window.Update.
func (u *UI) loadExit(force bool) {
	if u.win == nil || (!force && u.exitLoaded) {
		return
	}
	u.exitLoaded = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		info, err := u.ctrl.ExitInfo(ctx, force)
		cancel()
		u.win.Update(func() {
			if !u.active.Load() {
				u.exitLoaded = false
				return
			}
			if err != nil {
				u.exitErr = err.Error()
				return
			}
			u.exitErr = ""
			u.exitInfo = info
		})
	}()
}

func (u *UI) loadServices() {
	if u.win == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		results := u.ctrl.ProbeServices(ctx)
		cancel()
		u.win.Update(func() {
			if !u.active.Load() {
				u.servLoaded = false
				return
			}
			u.services = results
			u.servLoaded = true
		})
	}()
}

// ensureTools fills the Tools page's text fields from settings once.
func (u *UI) ensureTools(s app.Settings) {
	if u.logReady {
		return
	}
	u.logReady = true
	u.portText = fmt.Sprintf("%d", s.MixedPort)
	u.testURLText = s.TestURL
	u.controllerText = fmt.Sprintf("%d", s.ControllerPort)
	switch s.AutoUpdateInterval {
	case 0:
		u.autoUpdateIdx = 0
	case 6:
		u.autoUpdateIdx = 1
	case 12:
		u.autoUpdateIdx = 2
	case 24:
		u.autoUpdateIdx = 3
	case 72:
		u.autoUpdateIdx = 4
	default:
		u.autoUpdateIdx = 3
	}
}

var pageNames = []string{"概览", "代理", "订阅", "连接", "日志", "设置"}
