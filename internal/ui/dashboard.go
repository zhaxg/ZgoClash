package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

// Row heights mirror FlClash's desktop grid so the cards sit in the same
// rhythm. The whole page scrolls; on a tall window there is a little room
// under the last row, which is fine.
// The row heights and the gap mirror FlClash's desktop grid, so the cards
// sit in the same rhythm.
const (
	rowTopH    = 208
	rowCtrlH   = 64
	rowMiddleH = 176
	rowTablesH = 248
	dashGap    = 16
)

// decorationTurn is how far the card's decoration turns as the core starts.
const decorationTurn = 2 * 3.14159265 / 12

// easeOutBack is the curve FlClash's decorations and button use: it goes
// past its target and comes back.
func easeOutBack(t float32) float32 {
	const c1 = 1.70158
	const c3 = c1 + 1
	x := t - 1
	return 1 + c3*x*x*x + c1*x*x
}

func (u *UI) dashboardPage(c *mu.Context, snap app.Snapshot, p palette) {
	// The rows above the tables keep their heights, and the tables take what
	// is left, as FlClash's dashboard does. The page scrolls when the
	// window is too short for them all.
	mu.Scroll(c).Grow(1).Children(func() {
		mu.Column(c).Padding(16, 24).Gap(dashGap).Children(func() {
			// Row 1 — status and the two directions of traffic. Stretch keeps
			// every card in the row the same height.
			mu.Row(c).Height(rowTopH).Gap(dashGap).AlignItems(mu.Stretch).Children(func() {
				u.outboundCard(c, snap, p)
				u.speedCard(c, snap, p, true)
				u.speedCard(c, snap, p, false)
			})

			// Row 2 — outbound mode bar and the two takeover switches.
			mu.Row(c).Height(rowCtrlH).Gap(dashGap).AlignItems(mu.Stretch).Children(func() {
				u.outboundModeBar(c, snap, p)
				u.takeoverTile(c, snap, p, true)
				u.takeoverTile(c, snap, p, false)
			})

			// Row 3 — four info cards.
			mu.Row(c).Height(rowMiddleH).Gap(dashGap).AlignItems(mu.Stretch).Children(func() {
				u.intranetCard(c, snap, p)
				u.profileDashCard(c, snap, p)
				u.serviceCard(c, snap, p)
				u.memoryCard(c, snap, p)
			})

			// Row 4 — the connections table and the proxy groups.
			mu.Row(c).Height(rowTablesH).MinHeight(120).Gap(dashGap).AlignItems(mu.Stretch).Children(func() {
				u.connectionsDashCard(c, snap, p)
				u.groupsDashCard(c, snap, p)
			})
		})
	})
}

// ---- shared bits ----------------------------------------------------------

// dashCard is one white card in the grid. Its padding and corner radius are
// the ones FlClash's dashboard cards use.
func dashCard(c *mu.Context, p palette) mu.Element {
	return mu.Column(c).Padding(20, 22).Gap(11).Radius(32).Background(p.Card).
		Border(1, p.Border).Shadow(2, 6, 18, 0, p.Shadow)
}

// dashHeader is the icon + title row at the top of a card.
func (u *UI) dashHeader(c *mu.Context, p palette, icon *mu.SVG, title string, trailing func()) {
	mu.Row(c).AlignItems(mu.Center).Gap(8).Children(func() {
		mu.Icon(c, icon).FontSize(15).TextColor(p.Muted)
		mu.Text(c, title).FontSize(13).Bold().TextColor(p.Text)
		mu.Spacer(c)
		if trailing != nil {
			trailing()
		}
	})
}

// tag is a small pill chip.
func (u *UI) tag(c *mu.Context, p palette, text string, fg, bg mu.Color) {
	mu.Box(c).Padding(2, 8).Radius(99).Background(bg).Children(func() {
		mu.Text(c, text).FontSize(10).TextColor(fg)
	})
}

// ---- Row 1 ----------------------------------------------------------------

// outboundCard shows the status chip, the exit IP with country and quality
// chips, the run time and organization, and the big start/stop button.
//
// FlClash puts a cookie-12 shape behind the card, in its top right corner,
// which grows and turns while the core runs, and a country badge that is a
// four-leaf clover rather than a circle. The button morphs from a circle
// into a cookie-9 when the core starts.
func (u *UI) outboundCard(c *mu.Context, snap app.Snapshot, p palette) {
	running := snap.Running
	target := float32(0)
	if running {
		target = 1
	}
	card := dashCard(c, p).Grow(2).Padding(0)
	// The decoration grows and turns with the core's state.
	progress := card.AnimateTo("deco", target, mu.Animation{Duration: 450 * time.Millisecond, Ease: easeOutBack})
	// The decoration is painted behind the card's content, in its top
	// right corner. The clip keeps it inside the card's rounded corners.
	card.Draw(func(pt *mu.Painter, r mu.Rect) {
		pt.Clip(r, 32, func() {
			// FlClash's decoration: a cookie-12 in the card's top right
			// corner, half of it outside, which grows and turns a twelfth
			// of a turn while the core runs.
			size := float32(290)
			if r.W < size {
				size = r.W
			}
			x := r.X + r.W - size + 64
			y := r.Y - 87
			shapeCookie12.draw(pt, x, y, size*(1+0.12*progress), progress*decorationTurn,
				p.Card.Mix(p.Accent, 0.1))
		})
	})
	card.Children(func() {
		mu.Row(c).Padding(20, 22).Grow(1).Gap(16).AlignItems(mu.Center).Children(func() {
			mu.Column(c).Grow(1).Gap(10).Children(func() {
				mu.Row(c).AlignItems(mu.Center).Gap(8).Children(func() {
					mu.Icon(c, iconStatus).FontSize(15).TextColor(p.Muted)
					mu.Text(c, "状态").FontSize(13).Bold().TextColor(p.Text)
					statusText, statusFg, statusBg := "未运行", p.Muted, p.Soft
					switch {
					case snap.Starting:
						statusText, statusFg, statusBg = "启动中", p.Warn, p.Warn.Alpha(0.14)
					case snap.Running:
						// A running core shows a bright green chip with
						// white letters, as FlClash does.
						statusText, statusFg, statusBg = "运行中", mu.Hex("#ffffff"), mu.Hex("#1fbf75")
					}
					u.tag(c, p, statusText, statusFg, statusBg)
				})

				// The exit address. FlClash puts the country badge first, on
				// the left of the address, and the address with its chips
				// takes the room that is left.
				mu.Row(c).Gap(12).AlignItems(mu.Center).Children(func() {
					u.countryBadge(c, p)
					mu.Column(c).Grow(1).Gap(4).Children(func() {
						ip := "—"
						if u.exitInfo.IP != "" {
							ip = u.exitInfo.IP
						} else if u.exitErr != "" && !snap.Running {
							ip = "未连接"
						}
						mu.Text(c, ip).FontSize(21).Bold().SingleLine()
						// Quality chips.
						mu.Row(c).Gap(6).Children(func() {
							if org := u.exitInfo.Organization; org != "" {
								u.tag(c, p, orgType(org), p.Teal, p.Teal.Alpha(0.12))
							}
							if d := u.currentDelay(); d > 0 {
								label, fg := "质量优", p.Good
								switch {
								case d >= 600:
									label, fg = "质量一般", p.Bad
								case d >= 250:
									label, fg = "质量良", p.Warn
								}
								u.tag(c, p, label, fg, fg.Alpha(0.13))
							}
						})
					})
				})
				mu.Spacer(c)
				mu.Row(c).Gap(28).AlignItems(mu.Start).Children(func() {
					u.infoCell(c, p, "运行时长", runTimeText(snap))
					// The organization can be long, so it takes the room
					// that is left.
					mu.Column(c).Grow(1).Gap(2).Children(func() {
						mu.Text(c, "组织").FontSize(10).TextColor(p.Muted).SingleLine()
						mu.Text(c, orgName(u.exitInfo.Organization)).FontSize(12).
							TextColor(p.Text).SingleLine().MaxLines(1)
					})
				})
			})

			u.startToggle(c, snap, p)
		})
	})
}

// countryBadge is the exit's country on a four-leaf clover, as FlClash shows
// it.
func (u *UI) countryBadge(c *mu.Context, p palette) {
	code := u.exitInfo.CountryCode
	size := float32(48)
	el := mu.Box(c).Size(size, size)
	el.Draw(func(pt *mu.Painter, r mu.Rect) {
		shapeClover4.draw(pt, r.X, r.Y, r.W, 0, p.Soft)
	})
	if code == "" {
		el.Children(func() {
			mu.Icon(c, iconGlobe).FontSize(16).TextColor(p.Muted)
		})
		return
	}
	el.Children(func() {
		mu.Row(c).Grow(1).Center().Children(func() {
			mu.Text(c, strings.ToUpper(code)).FontSize(12).Bold().TextColor(p.Text)
		})
	})
}

// startToggle is the big button: a circle that morphs into a cookie-9 while
// the core runs, the way FlClash's M3 Expressive loader does. Its glyph turns
// from a play triangle into a pause bar as it goes.
func (u *UI) startToggle(c *mu.Context, snap app.Snapshot, p palette) {
	busy := u.busy["core"] || snap.Starting
	size := float32(104)
	target := float32(0)
	if snap.Running {
		target = 1
	}
	el := mu.ButtonBase(c).Size(size, size).Label("启动核心").Key("core-toggle")
	progress := el.AnimateTo("morph", target, mu.Animation{Duration: 450 * time.Millisecond, Ease: easeOutBack})
	// FlClash fills the button with the primary colour in both states, and
	// only greys it when there is no profile to run.
	fill := p.Accent
	ink := p.AccentText
	if len(snap.Profiles) == 0 {
		fill = p.Soft
		ink = p.Muted
	}
	el.Draw(func(pt *mu.Painter, r mu.Rect) {
		// The shape the button takes: a plain circle when stopped, a
		// cookie-9 when it runs.
		halo := p.Accent.Alpha(0.2)
		if progress <= 0.001 {
			radius := min(r.W, r.H) / 2
			for i := 1; i <= 3; i++ {
				pt.Stroke(r, halo.Alpha(0.07), radius, float32(i)*2.5*2)
			}
			pt.Fill(r, fill, radius)
		} else {
			shape := morphShape(progress, r.X, r.Y, r.W)
			// The shadow the button casts: the shape itself, grown a little and
			// fading away, so the halo follows the morph.
			for i := 1; i <= 3; i++ {
				grow := float32(i) * 2.5
				pt.StrokePath(shape, grow*2, halo.Alpha(0.07))
			}
			pt.FillPath(shape, fill)
		}
		// The glyph: a play triangle that becomes a pause bar.
		drawPlayPause(pt, r, progress, ink)
	})
	if busy {
		el.Children(func() {
			mu.Spinner(c)
		})
	}
	if el.Clicked() && !busy {
		u.toggleCore()
	}
}

// drawPlayPause draws the button's glyph: FlClash's play triangle folding
// into a pause bar, turned a quarter as it goes.
func drawPlayPause(pt *mu.Painter, r mu.Rect, progress float32, ink mu.Color) {
	// The glyph is a 24-unit square, as FlClash's glyphs are.
	const unit = 24.0
	size := r.W * 0.34
	ox := r.X + (r.W-size)/2
	oy := r.Y + (r.H-size)/2
	scale := size / unit
	// The turn FlClash's glyph makes as it folds.
	turn := progress * 3.14159265 / 2
	cos, sin := float32(math.Cos(float64(turn))), float32(math.Sin(float64(turn)))
	turned := func(x, y float64) (float32, float32) {
		dx := float32(x) - 12
		dy := float32(y) - 12
		return ox + size/2 + dx*cos - dy*sin, oy + size/2 + dx*sin + dy*cos
	}

	// The play triangle, and the two pause bars it folds into.
	triangle := [][2]float64{{8.3, 5.6}, {18.8, 12}, {8.3, 18.4}}
	halves := [][4][2]float64{
		{{8.3, 5.6}, {18.8, 12}, {18.8, 12}, {8.3, 12}},
		{{8.3, 12}, {18.8, 12}, {18.8, 12}, {8.3, 18.4}},
	}
	bars := [][4][2]float64{
		{{5.9, 6.9}, {18.1, 6.9}, {18.1, 9.1}, {5.9, 9.1}},
		{{5.9, 14.9}, {18.1, 14.9}, {18.1, 17.1}, {5.9, 17.1}},
	}
	fold := progress
	if fold <= 0.001 {
		pts := make([][2]float32, 0, len(triangle))
		for _, corner := range triangle {
			x, y := turned(corner[0], corner[1])
			pts = append(pts, [2]float32{x, y})
		}
		pt.FillPath(glyphPath(pts, 1.2*scale), ink)
		return
	}
	// Each half of the triangle becomes one bar.
	for i, half := range halves {
		pts := make([][2]float32, 0, len(half))
		for j, corner := range half {
			x := corner[0] + (bars[i][j][0]-corner[0])*float64(fold)
			y := corner[1] + (bars[i][j][1]-corner[1])*float64(fold)
			px, py := turned(x, y)
			pts = append(pts, [2]float32{px, py})
		}
		pt.FillPath(glyphPath(pts, 1.2*scale), ink)
	}
}

// glyphPath is a closed polyline whose corners are rounded off, as FlClash's
// glyphs are. Points that repeat are dropped, so a corner that never moves
// does not divide by nothing.
func glyphPath(pts [][2]float32, radius float32) *mu.Path {
	// Drop the points that repeat, keeping the shape.
	uniq := pts[:0:0]
	for _, pt := range pts {
		if len(uniq) > 0 {
			last := uniq[len(uniq)-1]
			if last[0] == pt[0] && last[1] == pt[1] {
				continue
			}
		}
		uniq = append(uniq, pt)
	}
	for len(uniq) > 1 && uniq[0] == uniq[len(uniq)-1] {
		uniq = uniq[:len(uniq)-1]
	}
	n := len(uniq)
	path := &mu.Path{}
	if n < 3 {
		return path
	}
	corner := func(i int) (enter, exit [2]float32) {
		v := uniq[i]
		prev := uniq[(i-1+n)%n]
		next := uniq[(i+1)%n]
		toPrev := [2]float32{prev[0] - v[0], prev[1] - v[1]}
		toNext := [2]float32{next[0] - v[0], next[1] - v[1]}
		lp := float32(math.Hypot(float64(toPrev[0]), float64(toPrev[1])))
		ln := float32(math.Hypot(float64(toNext[0]), float64(toNext[1])))
		if lp == 0 || ln == 0 {
			return v, v
		}
		r := radius
		if lp/2 < r {
			r = lp / 2
		}
		if ln/2 < r {
			r = ln / 2
		}
		return [2]float32{v[0] + toPrev[0]/lp*r, v[1] + toPrev[1]/lp*r},
			[2]float32{v[0] + toNext[0]/ln*r, v[1] + toNext[1]/ln*r}
	}
	_, first := corner(0)
	path.MoveTo(first[0], first[1])
	for i := 1; i <= n; i++ {
		v := uniq[i%n]
		enter, exit := corner(i % n)
		path.LineTo(enter[0], enter[1])
		path.QuadTo(v[0], v[1], exit[0], exit[1])
	}
	path.Close()
	return path
}

// morphTurn is how far the button's shape turns as it morphs.
func morphTurn(progress float32) float32 {
	return progress * 2 * 3.14159265 / 9
}

func (u *UI) infoCell(c *mu.Context, p palette, label, value string) {
	mu.Column(c).Gap(2).Children(func() {
		mu.Text(c, label).FontSize(10).TextColor(p.Muted).SingleLine()
		mu.Text(c, value).FontSize(12).TextColor(p.Text).SingleLine().MaxLines(1)
	})
}

func (u *UI) bigPowerButton(c *mu.Context, snap app.Snapshot, p palette) {
	busy := u.busy["core"] || snap.Starting
	bg := p.Accent
	fg := p.AccentText
	if snap.Running {
		bg = p.Soft
		fg = p.Good
	}
	el := mu.ButtonBase(c).Size(62, 62).Radius(31).Background(bg).AlignItems(mu.Center).
		Justify(mu.Center).Label("核心开关").Children(func() {
		if busy {
			mu.Spinner(c)
		} else {
			mu.Icon(c, iconPower).FontSize(24).TextColor(fg)
		}
	})
	if el.Hovered() && !snap.Running {
		el.Background(p.Accent.Mix(p.Text, 0.12))
	}
	if el.Clicked() && !busy {
		u.toggleCore()
	}
}

// speedCard is the upload or download traffic card with a dot matrix that
// shows how much of the last stretch was used.
func (u *UI) speedCard(c *mu.Context, snap app.Snapshot, p palette, up bool) {
	label, icon := "下载", iconArrowDown
	value, peak, total := snap.Traffic.Down, u.peakDown, snap.Traffic.DownTotal
	accent := p.Accent
	if up {
		label, icon = "上传", iconArrowUp
		value, peak, total = snap.Traffic.Up, u.peakUp, snap.Traffic.UpTotal
		accent = p.Muted
	}
	el := dashCard(c, p).Grow(1)
	el.Children(func() {
		mu.Row(c).AlignItems(mu.Center).Gap(8).Children(func() {
			mu.Icon(c, icon).FontSize(15).TextColor(accent)
			mu.Text(c, label).FontSize(13).Bold().TextColor(p.Text)
			mu.Spacer(c)
			mu.Text(c, "峰值 "+formatSpeed(peak)).FontSize(10).TextColor(p.Muted).NoWrap()
		})
		mu.Text(c, formatSpeed(value)).FontSize(27).Bold().TextColor(p.Text)
		mu.Text(c, "累计 "+formatBytes(total)).FontSize(10.5).TextColor(p.Muted)
		mu.Spacer(c)
		// The dot matrix: the tinted box at the bottom of the card, showing
		// the traffic of the last stretch. The dots keep an even inset on
		// every side of the box, and the box keeps the same margin on both
		// sides of the card.
		mu.Row(c).FillWidth().Padding(0, 6).Children(func() {
			mu.Box(c).Height(60).FillWidth().Radius(10).Background(p.ChartBg).Clip().
				Draw(func(pt *mu.Painter, r mu.Rect) {
					const dotPad = 4
					inner := mu.Rect{
						X: r.X + dotPad,
						Y: r.Y + dotPad,
						W: r.W - 2*dotPad,
						H: r.H - 2*dotPad,
					}
					if inner.W <= 0 || inner.H <= 0 {
						return
					}
					m := newDotMatrix(inner.W, inner.H)
					m.draw(pt, inner, u.trafficWindow(up), p)
				})
		})
	})
}

// trafficWindow is the traffic of the last stretch, oldest first, which the
// traffic cards show as a dot matrix.
func (u *UI) trafficWindow(up bool) []int64 {
	out := make([]int64, 0, len(u.samples))
	for _, s := range u.samples {
		if up {
			out = append(out, s.up)
		} else {
			out = append(out, s.down)
		}
	}
	return out
}

// ---- Row 2 ----------------------------------------------------------------

// outboundModeBar is the wide pill holding the outbound mode segmented
// control, matching FlClash's control row.
func (u *UI) outboundModeBar(c *mu.Context, snap app.Snapshot, p palette) {
	mu.Row(c).Grow(2).Padding(10, 16).Radius(15).Background(p.Card).Border(1, p.Border).
		AlignItems(mu.Center).Gap(14).Children(func() {
		mu.Icon(c, iconSplit).FontSize(16).TextColor(p.Muted)
		mu.Text(c, "出站模式").FontSize(13).Bold().TextColor(p.Text)
		mu.Spacer(c)
		idx := modeIndex(snap.Settings.Mode)
		if mu.Segmented(c, &idx, "规则", "全局", "直连").Changed() {
			mode := modeString(idx)
			u.run("mode", func() error { return u.ctrl.SetMode(mode) }, nil)
		}
	})
}

// takeoverTile is the pill card for the system proxy or TUN switch.
func (u *UI) takeoverTile(c *mu.Context, snap app.Snapshot, p palette, systemProxy bool) {
	title, icon := "虚拟网卡", iconVPN
	on := snap.Settings.Tun
	enabled := snap.CanTun
	sub := "已关闭"
	cb := func(v bool) { u.run("tun", func() error { return u.ctrl.SetTun(v) }, nil) }
	if systemProxy {
		title, icon = "系统代理", iconShuffle
		on = snap.Settings.SystemProxy
		enabled = true
		sub = fmt.Sprintf("端口 %d", snap.MixedPort)
		cb = func(v bool) { u.run("sysproxy", func() error { return u.ctrl.SetSystemProxy(v) }, nil) }
	}
	accent := p.Accent
	if !enabled {
		accent = p.Muted
	}
	mu.Row(c).Grow(1).Padding(10, 16).Radius(15).Background(p.Card).Border(1, p.Border).
		Shadow(2, 6, 18, 0, p.Shadow).
		AlignItems(mu.Center).Gap(11).Children(func() {
		mu.Box(c).Size(34, 34).Radius(10).Background(accent.Alpha(0.12)).Center().Children(func() {
			mu.Icon(c, icon).FontSize(16).TextColor(accent)
		})
		mu.Column(c).Grow(1).Gap(1).Children(func() {
			mu.Text(c, title).FontSize(12.5).TextColor(p.Text)
			mu.Text(c, sub).FontSize(10).TextColor(p.Muted).SingleLine()
		})
		v := on
		if mu.Switch(c, &v).Label(title).Disabled(!enabled).Changed() && enabled {
			cb(v)
		}
	})
}

// ---- Row 3 ----------------------------------------------------------------

// intranetCard shows the LAN IP and the allow-LAN switch.
func (u *UI) intranetCard(c *mu.Context, snap app.Snapshot, p palette) {
	dashCard(c, p).Grow(1).Children(func() {
		u.dashHeader(c, p, iconDevices, "内网 IP", func() {
			if el := u.iconButton(c, p, iconCopy, "复制").Key("copy-ip"); el.Clicked() {
				u.copyText(snap.LocalIP)
			}
		})
		ip := snap.LocalIP
		if ip == "" {
			ip = "无网络"
		}
		mu.Text(c, ip).FontSize(20).Bold().SingleLine()
		mu.Spacer(c)
		mu.Row(c).Padding(7, 11).Radius(11).Background(p.Soft).AlignItems(mu.Center).Gap(9).Children(func() {
			mu.Icon(c, iconWifi).FontSize(14).TextColor(p.Muted)
			mu.Text(c, "局域网代理").FontSize(11.5).Grow(1).TextColor(p.Text)
			v := snap.Settings.AllowLan
			if mu.Switch(c, &v).Label("局域网代理").Changed() {
				u.patchSettings(func(s *app.Settings) { s.AllowLan = v })
			}
		})
	})
}

// profileDashCard shows the current profile and its usage.
func (u *UI) profileDashCard(c *mu.Context, snap app.Snapshot, p palette) {
	dashCard(c, p).Grow(1).Children(func() {
		u.dashHeader(c, p, iconProfiles, "订阅", func() {
			if el := u.iconButton(c, p, iconChevron, "管理").Key("open-profiles"); el.Clicked() {
				u.setPage(2)
			}
		})
		name := snap.CurrentProfileName
		if name == "" {
			name = "未选择配置"
		}
		kind := "远程"
		if prof, ok := snap.CurrentProfile(); ok {
			switch prof.Type {
			case "manual":
				kind = "手动"
			case "file":
				kind = "本地"
			}
		}
		mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
			mu.Text(c, name).FontSize(14).Bold().SingleLine().Grow(1)
			u.tag(c, p, kind, p.Muted, p.Soft)
		})
		mu.Spacer(c)
		if prof, ok := snap.CurrentProfile(); ok && prof.Total > 0 {
			used := prof.Upload + prof.Download
			share := percent(used, prof.Total)
			mu.Row(c).AlignItems(mu.Center).Gap(8).Children(func() {
				mu.Text(c, fmt.Sprintf("%.0f%% 已用", share*100)).FontSize(17).Bold()
				mu.Spacer(c)
				mu.Text(c, fmt.Sprintf("%s / %s", formatBytes(used), formatBytes(prof.Total))).
					FontSize(10.5).TextColor(p.Muted)
			})
			// The usage as a grid of dots: ten across and two down, each dot
			// worth five percent. The dots in use are in the accent colour,
			// the rest stay pale.
			usageDots(c, p, share)
		} else {
			mu.Text(c, "暂无用量信息").FontSize(11).TextColor(p.Muted)
		}
	})
}

// usageDots shows how much of a subscription's traffic is used as a grid of
// dots: twenty-five across and two down, so each dot is worth two percent.
// The dots in use are in the accent colour, the rest stay pale.
func usageDots(c *mu.Context, p palette, share float64) {
	const cols = 25
	const rows = 2
	const gap = 2.5
	const maxSize = 5.0
	height := float32(rows)*maxSize + float32(rows-1)*gap
	lit := int(math.Round(clamp01(share) * float64(cols*rows)))
	mu.Box(c).FillWidth().Height(height).Draw(func(pt *mu.Painter, r mu.Rect) {
		size := (r.W - float32(cols-1)*gap) / cols
		if size > maxSize {
			size = maxSize
		}
		if size < 1 {
			size = 1
		}
		width := float32(cols)*size + float32(cols-1)*gap
		tall := float32(rows)*size + float32(rows-1)*gap
		ox := r.X + (r.W-width)/2
		oy := r.Y + (r.H-tall)/2
		for i := 0; i < cols*rows; i++ {
			x := ox + float32(i%cols)*(size+gap)
			y := oy + float32(i/cols)*(size+gap)
			color := p.Track
			if i < lit {
				color = p.Accent
			}
			pt.Fill(mu.Rect{X: x, Y: y, W: size, H: size}, color, size/2)
		}
	})
}

// serviceCard probes the four well-known sites: each one is a pill with the
// site's mark and how long it took to answer.
func (u *UI) serviceCard(c *mu.Context, snap app.Snapshot, p palette) {
	dashCard(c, p).Grow(1).Children(func() {
		u.dashHeader(c, p, iconVerified, "服务状态", func() {
			if el := u.iconButton(c, p, iconRefresh, "重测").Key("reprobe"); el.Clicked() {
				u.loadServices()
			}
		})
		if !snap.Running {
			mu.Text(c, "核心未运行").FontSize(11).TextColor(p.Muted)
			return
		}
		if !u.servLoaded {
			mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
				mu.Spinner(c)
				mu.Text(c, "正在检测…").FontSize(11).TextColor(p.Muted)
			})
			return
		}
		mu.Spacer(c)
		// The pills sit two on a row, as FlClash's service grid does.
		mu.Column(c).Gap(8).Children(func() {
			for i := 0; i < len(u.services); i += 2 {
				row := u.services[i:]
				if len(row) > 2 {
					row = row[:2]
				}
				mu.Row(c).Gap(8).Children(func() {
					for _, s := range row {
						u.servicePill(c, p, s)
					}
				})
			}
		})
	})
}

// servicePill is one site's pill: its round mark and how long it took.
func (u *UI) servicePill(c *mu.Context, p palette, s app.ServiceStatus) {
	mark := serviceMark(s.Name)
	delay, fg := "超时", p.Bad
	if s.OK {
		delay, fg = fmt.Sprintf("%d ms", s.Delay), p.Text
	}
	mu.Row(c).Grow(1).Padding(5, 10).Radius(11).Background(p.Soft).AlignItems(mu.Center).Gap(7).Children(func() {
		mu.Box(c).Size(20, 20).Radius(10).Background(mu.Hex("#ffffff")).Clip().Center().Children(func() {
			if mark != nil {
				mu.Image(c, mark).Size(18, 18).Fit(mu.Contain)
			}
		})
		mu.Text(c, delay).FontSize(11).TextColor(fg).NoWrap()
	})
}

// memoryCard shows the process memory split into app and core.
func (u *UI) memoryCard(c *mu.Context, snap app.Snapshot, p palette) {
	dashCard(c, p).Grow(1).Children(func() {
		u.dashHeader(c, p, iconMemory, "内存信息", nil)
		mu.Row(c).AlignItems(mu.Center).Gap(4).Children(func() {
			mu.Text(c, formatBytes(int64(snap.MemoryBytes))).FontSize(22).Bold()
			mu.Text(c, "总占用").FontSize(10).TextColor(p.Muted)
		})
		mu.Spacer(c)
		// Two-tone bar, rounded at both ends.
		total := float64(snap.MemoryApp + snap.MemoryCore)
		appShare := 0.0
		if total > 0 {
			appShare = float64(snap.MemoryApp) / total
		}
		share := float32(clamp01(appShare))
		// The bar is drawn by the painter, so its ends are rounded and the
		// two parts meet without a seam.
		mu.Box(c).Height(8).FillWidth().Draw(func(pt *mu.Painter, r mu.Rect) {
			pt.Fill(r, p.Track, 4)
			pt.Clip(r, 4, func() {
				app := mu.Rect{X: r.X, Y: r.Y, W: r.W * share, H: r.H}
				pt.Fill(app, p.Accent, 0)
				pt.Fill(mu.Rect{X: app.X + app.W, Y: r.Y, W: r.W - app.W, H: r.H}, p.Teal, 0)
			})
		})
		mu.Row(c).Gap(16).Wrap().Children(func() {
			u.memLegend(c, p, "应用", snap.MemoryApp, p.Accent)
			u.memLegend(c, p, "内核", snap.MemoryCore, p.Teal)
		})
	})
}

func (u *UI) memLegend(c *mu.Context, p palette, label string, bytes uint64, color mu.Color) {
	mu.Row(c).Gap(6).AlignItems(mu.Center).Children(func() {
		mu.Box(c).Size(8, 8).Radius(4).Background(color)
		mu.Text(c, label).FontSize(10.5).TextColor(p.Muted)
		mu.Text(c, formatBytes(int64(bytes))).FontSize(10.5).Bold().TextColor(p.Text)
	})
}

// ---- Row 4 ----------------------------------------------------------------

// connectionsDashCard lists live and recent connections behind pill tabs.
func (u *UI) connectionsDashCard(c *mu.Context, snap app.Snapshot, p palette) {
	dashCard(c, p).Grow(3).Gap(8).Children(func() {
		mu.Row(c).AlignItems(mu.Center).Gap(10).Children(func() {
			feeds := []string{
				fmt.Sprintf("%d 实时连接", snap.ConnectionsCount),
				fmt.Sprintf("%d 最近请求", snap.RecentCount),
			}
			for i, label := range feeds {
				idx := i
				el := mu.ButtonBase(c).Padding(5, 11).Radius(8).Label(label).Children(func() {
					mu.Text(c, label).FontSize(11.5)
				})
				if u.dashFeed == idx {
					el.Background(p.Accent)
					el.TextColor(p.AccentText)
				} else {
					el.Background(p.Soft)
					el.TextColor(p.Text)
				}
				if el.Clicked() {
					u.dashFeed = idx
				}
			}
			mu.Spacer(c)
			u.tag(c, p, fmt.Sprintf("TCP %d", snap.TCPCount), p.Muted, p.Soft)
			u.tag(c, p, fmt.Sprintf("UDP %d", snap.UDPCount), p.Muted, p.Soft)
		})

		// Table header.
		mu.Row(c).Gap(8).Padding(5, 8).Radius(6).Background(p.Soft).Children(func() {
			u.cell(c, p, "来源", 1.0, false)
			u.cell(c, p, "目标", 1.4, false)
			u.cell(c, p, "协议", 0.6, false)
			u.cell(c, p, "出口", 1.1, false)
			u.cell(c, p, "下载", 0.8, true)
			u.cell(c, p, "上传", 0.8, true)
			u.cell(c, p, "时长", 0.7, true)
		})

		var rows []engine.Connection
		switch u.dashFeed {
		case 0:
			rows = append([]engine.Connection(nil), u.ctrl.Connections().Connections...)
			sort.Slice(rows, func(i, j int) bool { return rows[i].Start > rows[j].Start })
		default:
			rows = u.ctrl.Recent()
		}
		if len(rows) == 0 {
			mu.Row(c).Grow(1).Center().Children(func() {
				mu.Text(c, "暂无记录").FontSize(11).TextColor(p.Muted)
			})
			return
		}
		// Rows pack tightly and are capped so the fixed card height holds.
		if len(rows) > 6 {
			rows = rows[:6]
		}
		mu.Column(c).Gap(1).Children(func() {
			for _, cn := range rows {
				u.connDashRow(c, p, cn)
			}
		})
	})
}

func (u *UI) cell(c *mu.Context, p palette, text string, grow float32, end bool) {
	el := mu.Text(c, text).FontSize(10).TextColor(p.Muted).Grow(grow).SingleLine()
	if end {
		el.TextAlign(mu.End)
	}
}

func (u *UI) connDashRow(c *mu.Context, p palette, cn engine.Connection) {
	src := cn.Metadata.Process
	if src == "" {
		if idx := strings.LastIndexAny(cn.Metadata.ProcessPath, `/\`); idx >= 0 {
			src = cn.Metadata.ProcessPath[idx+1:]
		} else {
			src = cn.Metadata.ProcessPath
		}
	}
	if src == "" && cn.Metadata.SourceIP != "" {
		src = cn.Metadata.SourceIP + ":" + cn.Metadata.SourcePort
	}
	if src == "" {
		src = "—"
	}
	target := cn.Metadata.Host
	if target == "" {
		target = cn.Metadata.DestinationIP
	}
	exit := "DIRECT"
	if len(cn.Chains) > 0 {
		exit = cn.Chains[0]
	}
	dur := connDuration(cn.Start)
	mu.Row(c).Gap(8).Padding(6, 8).AlignItems(mu.Center).Children(func() {
		mu.Text(c, src).FontSize(10.5).TextColor(p.Muted).Grow(1).SingleLine()
		mu.Text(c, target).FontSize(10.5).TextColor(p.Text).Grow(1.4).SingleLine()
		mu.Text(c, strings.ToUpper(cn.Metadata.Network)).FontSize(10.5).TextColor(p.Muted).Grow(0.6).SingleLine()
		mu.Text(c, withFlag(exit)).FontSize(10.5).Font(emojiFamily).TextColor(p.Accent).Grow(1.1).SingleLine()
		mu.Text(c, formatBytes(cn.Download)).FontSize(10.5).TextColor(p.Text).Grow(0.8).SingleLine().TextAlign(mu.End)
		mu.Text(c, formatBytes(cn.Upload)).FontSize(10.5).TextColor(p.Muted).Grow(0.8).SingleLine().TextAlign(mu.End)
		mu.Text(c, dur).FontSize(10.5).TextColor(p.Muted).Grow(0.7).SingleLine().TextAlign(mu.End)
	})
}

// groupsDashCard shows the proxy groups as tiles.
func (u *UI) groupsDashCard(c *mu.Context, snap app.Snapshot, p palette) {
	title := "策略组"
	if snap.Running && len(u.groups) > 0 {
		title = fmt.Sprintf("策略组 %d", len(u.groups))
	}
	dashCard(c, p).Grow(2).Gap(10).Children(func() {
		u.dashHeader(c, p, iconLayers, title, func() {
			if el := u.iconButton(c, p, iconChevron, "全部").Key("open-groups"); el.Clicked() {
				u.setPage(1)
			}
		})
		if !snap.Running {
			mu.Text(c, "核心启动后这里会显示策略组").FontSize(11).TextColor(p.Muted)
			return
		}
		if len(u.groups) == 0 {
			mu.Text(c, "当前配置没有策略组").FontSize(11).TextColor(p.Muted)
			return
		}
		// Two-column grid of group tiles, capped to the 2x3 FlClash shows;
		// the chevron in the header opens the full list on the 代理 page.
		shown := u.groups
		if len(shown) > 6 {
			shown = shown[:6]
		}
		mu.Column(c).Gap(9).Children(func() {
			for i := 0; i < len(shown); i += 2 {
				i := i
				mu.Row(c).Gap(9).Children(func() {
					u.groupTile(c, p, i)
					if i+1 < len(shown) {
						u.groupTile(c, p, i+1)
					}
				})
			}
		})
	})
}

func (u *UI) groupTile(c *mu.Context, p palette, idx int) {
	g := u.groups[idx]
	// The tile shows the flag of the region in its glyph box and keeps the
	// label plain, so the pair is not read twice.
	name := stripEmojiPrefix(g.Name)
	mono := ""
	if r := []rune(stripEmojiPrefix(g.Name)); len(r) > 0 {
		mono = string(r[0])
	}
	// The tile shows the flag of the node the group sits on, so the pair
	// reads as a region, not a letter.
	flag := nodeFlag(g.Now)
	if flag == "" {
		flag = nodeFlag(g.Name)
	}
	el := mu.ButtonBase(c).Grow(1).Padding(9, 11).Radius(11).Background(p.Soft).Gap(10).
		AlignItems(mu.Center).Label(name).Children(func() {
		mu.Box(c).Size(30, 30).Radius(9).Background(p.CardHover).Center().Children(func() {
			if flag != "" {
				mu.Text(c, flag).FontSize(17).Font(emojiFamily)
			} else {
				mu.Text(c, mono).FontSize(12).Bold().TextColor(p.Muted)
			}
		})
		mu.Column(c).Grow(1).Gap(1).Children(func() {
			mu.Text(c, name).FontSize(11.5).Font(emojiFamily).TextColor(p.Text).SingleLine()
			mu.Row(c).Gap(5).AlignItems(mu.Center).Children(func() {
				// FlClash marks groups parked on a blocking node with a red dot.
				if strings.EqualFold(g.Now, "REJECT") || strings.EqualFold(g.Now, "PASS") {
					mu.Box(c).Size(6, 6).Radius(3).Background(p.Bad)
				}
				mu.Text(c, stripEmojiPrefix(g.Now)).FontSize(10).Font(emojiFamily).TextColor(p.Muted).SingleLine()
			})
		})
	})
	if el.Hovered() {
		el.Background(p.CardHover)
	}
	if el.Clicked() {
		u.groupIdx = idx
		u.setPage(1)
	}
}

// stripEmojiPrefix drops the decorative emoji many subscription groups put in
// front of their names, so the tile shows "节" rather than "🚀".
func stripEmojiPrefix(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		switch {
		case r == ' ' || r == ' ' || r == '　':
			return true
		case r >= 0x1F000: // emoji blocks
			return true
		case r >= 0x2600 && r <= 0x27BF: // dingbats, misc symbols
			return true
		case r >= 0x2B00 && r <= 0x2BFF: // arrows, shapes
			return true
		case r == 0xFE0F || r == 0x200D: // variation selector, ZWJ
			return true
		}
		return false
	})
}

// ---- helpers -------------------------------------------------------------

func (u *UI) currentDelay() int {
	if g, ok := u.currentGroup(); ok {
		return u.delayFor(g.Now)
	}
	return 0
}

func runTimeText(snap app.Snapshot) string {
	if !snap.Running {
		return "—"
	}
	return formatUptime(snap.Uptime)
}

// orgType buckets an organization name into a rough quality label.
func orgType(org string) string {
	l := strings.ToLower(org)
	for _, k := range []string{"hosting", "cloud", "datacenter", "amazon", "google", "microsoft",
		"digitalocean", "alibaba", "tencent", "ovh", "hetzner", "vultr", "linode"} {
		if strings.Contains(l, k) {
			return "机房"
		}
	}
	return "商业"
}

// orgName is the organization as the card shows it, or a dash when unknown.
func orgName(org string) string {
	if org == "" {
		return "—"
	}
	return org
}

// startTimeOf parses a connection's start stamp, which the core reports as an
// RFC3339 string.
func startTimeOf(cn engine.Connection) time.Time {
	t, err := time.Parse(time.RFC3339Nano, cn.Start)
	if err != nil {
		t, err = time.Parse(time.RFC3339, cn.Start)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

// connDuration formats a connection's age as min/h.
func connDuration(start string) string {
	t, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
