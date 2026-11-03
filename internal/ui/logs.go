package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

// logsPage shows the core's log as a full page: the filter, the follow
// switch and the lines, which take the room the header leaves.
func (u *UI) logsPage(c *mu.Context, _ app.Snapshot, p palette) {
	filters := []string{"全部", "错误", "警告", "信息", "调试"}
	levels := []string{"", "error", "warning", "info", "debug"}
	entries := u.ctrl.Logs(0)
	filter := levels[u.logFilter]
	shown := make([]engine.LogEntry, 0, len(entries))
	for _, entry := range entries {
		if filter != "" && entry.Level != filter {
			continue
		}
		shown = append(shown, entry)
	}

	mu.Column(c).FillWidth().Grow(1).Padding(22, 24).Gap(14).Children(func() {
		u.pageHeader(c, p, "日志", "mihomo 内核日志", nil)
		card(c, p).FillWidth().Grow(1).Gap(10).Children(func() {
			mu.Row(c).Gap(10).AlignItems(mu.Center).Wrap().Children(func() {
				mu.Text(c, "显示").FontSize(11.5).TextColor(p.Muted)
				sel := filters[u.logFilter]
				if mu.Select(c, &sel, filters).Changed() {
					for i, f := range filters {
						if f == sel {
							u.logFilter = i
						}
					}
				}
				follow := u.logFollow
				if mu.Switch(c, &follow).Label("自动滚动").Changed() {
					u.logFollow = follow
				}
				mu.Text(c, "自动滚动").FontSize(11.5).TextColor(p.Muted)
				mu.Spacer(c)
				mu.Text(c, fmt.Sprintf("%d 条", len(shown))).FontSize(11).TextColor(p.Muted)
				if el := u.iconButton(c, p, iconClear, "清空日志"); el.Clicked() {
					u.ctrl.ClearLogs()
				}
			})

			if u.logFollow {
				u.logScroll.Y = math.MaxFloat32
			}
			mu.Scroll(c).Grow(1).TrackScroll(&u.logScroll).Padding(10).Radius(10).
				Background(p.Soft).Children(func() {
				mu.Column(c).Gap(3).Children(func() {
					if len(shown) == 0 {
						mu.Text(c, "暂无日志").FontSize(11).TextColor(p.Muted)
					}
					for _, line := range shown {
						timeText := ""
						if line.Time > 0 {
							timeText = time.UnixMilli(line.Time).Format("15:04:05")
						}
						color := p.Muted
						switch line.Level {
						case "error":
							color = p.Bad
						case "warning":
							color = p.Warn
						case "debug":
							color = p.Teal
						}
						mu.Row(c).Gap(8).AlignItems(mu.Start).Children(func() {
							mu.Text(c, timeText).FontSize(10).TextColor(p.Muted).Width(56)
							mu.Text(c, strings.ToUpper(line.Level)).FontSize(9.5).TextColor(color).Width(52)
							mu.Text(c, line.Payload).FontSize(10.5).Font(monoFamily).TextColor(p.Text).Grow(1)
						})
					}
				})
			})
		})
	})
}
