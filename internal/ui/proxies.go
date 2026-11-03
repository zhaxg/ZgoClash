package ui

import (
	"fmt"
	"sort"
	"strings"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

func (u *UI) proxiesPage(c *mu.Context, snap app.Snapshot, p palette) {
	mu.Column(c).Grow(1).Padding(22, 24).Gap(13).Children(func() {
		u.pageHeader(c, p, "代理", "查看代理组并切换节点", func() {
			if el := u.iconButton(c, p, iconBoltFilled, "测试当前组",
				u.testingGroup || !snap.Running || len(u.groups) == 0); el.Clicked() {
				u.testGroup()
			}
			if el := u.iconButton(c, p, iconRefresh, "刷新代理"); el.Clicked() {
				u.loadProxies()
			}
		})

		if !snap.Running {
			u.notRunningCard(c, p)
			return
		}
		if u.proxyLoading && len(u.groups) == 0 {
			mu.Row(c).Grow(1).Center().Gap(10).Children(func() {
				mu.Spinner(c)
				mu.Text(c, "正在读取代理组…").TextColor(p.Muted)
			})
			return
		}
		if u.proxyErr != "" && len(u.groups) == 0 {
			card(c, p).Children(func() {
				mu.Text(c, "读取代理失败").FontSize(14).Bold().TextColor(p.Bad)
				mu.Text(c, u.proxyErr).FontSize(12).TextColor(p.Muted)
				if mu.Button(c, "重试").Clicked() {
					u.loadProxies()
				}
			})
			return
		}
		if len(u.groups) == 0 {
			card(c, p).Children(func() {
				mu.Text(c, "当前配置没有代理组").FontSize(13).Bold()
				mu.Text(c, "在配置中加入 proxy-groups 后即可在这里切换节点。").FontSize(12).TextColor(p.Muted)
			})
			return
		}

		g, _ := u.currentGroup()
		mu.Row(c).Gap(7).Wrap().Children(func() {
			for i, group := range u.groups {
				idx := i
				active := idx == u.groupIdx
				el := mu.ButtonBase(c).Padding(5, 11).Radius(9).Gap(7).AlignItems(mu.Center).
					Label(group.Name).Children(func() {
					mu.Text(c, withFlag(group.Name)).FontSize(12).Font(emojiFamily)
					mu.Text(c, fmt.Sprintf("%d", len(group.All))).FontSize(10).TextColor(p.Muted)
				})
				if active {
					el.Background(p.AccentSoft)
					el.TextColor(p.Accent)
				} else {
					el.Background(p.Soft)
					if el.Hovered() {
						el.Background(p.CardHover)
					}
				}
				if el.Clicked() {
					u.groupIdx = idx
				}
			}
		})

		mu.Row(c).Gap(10).AlignItems(mu.Center).Wrap().Children(func() {
			mu.SearchField(c, &u.nodeQuery).Width(240).Placeholder("搜索节点")
			mu.MenuButton(c, "排序", func(m *mu.Menu) {
				labels := []string{"默认顺序", "按延迟", "按名称"}
				for i, label := range labels {
					if m.Item(label).Checked(u.nodeSort == i).Chosen() {
						u.nodeSort = i
					}
				}
			})
			mu.Spacer(c)
			mu.Text(c, fmt.Sprintf("%s · %d 个节点", withFlag(g.Name), len(g.All))).FontSize(11.5).TextColor(p.Muted).Font(emojiFamily)
			if g.Now != "" {
				mu.Text(c, "当前 "+withFlag(g.Now)).FontSize(11.5).TextColor(p.Accent).SingleLine().MaxWidth(260).Font(emojiFamily)
			}
		})

		names := u.visibleNodes(g)
		if len(names) == 0 {
			card(c, p).Children(func() {
				mu.Text(c, "没有匹配的节点").FontSize(13).TextColor(p.Muted)
			})
			return
		}

		u.visible = names
		u.grid.Key = func(i int) any {
			if i < len(u.visible) {
				return u.visible[i]
			}
			return i
		}
		u.grid.Label = func(i int) string {
			if i < len(u.visible) {
				return u.visible[i]
			}
			return ""
		}
		mu.GridView(c, &u.grid, len(names), 220, 68, func(i int) {
			u.nodeCard(c, p, g, names[i])
		}).Grow(1)
	})
}

func (u *UI) visibleNodes(g engine.ProxyGroupView) []string {
	names := append([]string(nil), g.All...)
	query := strings.ToLower(strings.TrimSpace(u.nodeQuery))
	if query != "" {
		filtered := names[:0]
		for _, name := range names {
			if strings.Contains(strings.ToLower(name), query) {
				filtered = append(filtered, name)
			}
		}
		names = filtered
	}
	switch u.nodeSort {
	case 1:
		sort.SliceStable(names, func(i, j int) bool {
			a, b := u.delayFor(names[i]), u.delayFor(names[j])
			if a <= 0 {
				a = 1 << 30
			}
			if b <= 0 {
				b = 1 << 30
			}
			return a < b
		})
	case 2:
		sort.SliceStable(names, func(i, j int) bool {
			return strings.ToLower(names[i]) < strings.ToLower(names[j])
		})
	}
	return names
}

func (u *UI) nodeCard(c *mu.Context, p palette, g engine.ProxyGroupView, name string) {
	proxy := u.proxies[name]
	selected := g.Now == name
	bg := p.Card
	if selected {
		bg = p.AccentSoft
	}
	delay := u.delayFor(name)
	testing := u.testing[name]

	el := mu.ButtonBase(c).Fill().Padding(9, 10).Radius(14).Background(bg).
		AlignItems(mu.Center).Label(name).Children(func() {
		mu.Row(c).Gap(8).AlignItems(mu.Center).FillWidth().Grow(1).Children(func() {
			mu.Column(c).Gap(4).Grow(1).AlignItems(mu.Start).Children(func() {
				mu.Text(c, withFlag(name)).FontSize(12.5).Font(emojiFamily).TextColor(p.Text).SingleLine()
				mu.Text(c, proxy.Type).FontSize(10).TextColor(p.Muted)
			})
			// The mark of the node in use, above the delay. It keeps its
			// place when another node is the one in use, so the cards do
			// not shift.
			mu.Column(c).Gap(2).AlignItems(mu.End).Children(func() {
				mark := p.Mark
				if !selected {
					mark = mu.RGBA(0, 0, 0, 0)
				}
				mu.Icon(c, iconCheckCircle).FontSize(23).TextColor(mark)
				// The delay is the button that tests the node.
				test := mu.ButtonBase(c).Padding(2, 6).Radius(6).Label("测试 " + name).
					Tooltip("测试延迟").Children(func() {
					if testing {
						loadingDots(c, p.Muted)
					} else {
						mu.Text(c, formatDelay(delay)).FontSize(11).Bold().FixedLineHeight(delayLineHeight).
							TextColor(delayColor(p, delay)).NoWrap()
					}
				})
				if test.Hovered() {
					test.Background(p.Soft)
				}
				if test.Clicked() {
					u.testNode(name)
				}
			})
		})
	})
	if el.Hovered() && !selected {
		el.Background(p.CardHover)
	}
	if el.Clicked() {
		u.selectNode(g.Name, name)
	}
}

func (u *UI) notRunningCard(c *mu.Context, p palette) {
	mu.Row(c).Grow(1).Center().Children(func() {
		card(c, p).Width(420).AlignItems(mu.Center).Gap(10).Children(func() {
			mu.Box(c).Size(46, 46).Radius(23).Background(p.AccentSoft).Center().Children(func() {
				mu.Icon(c, iconPower).FontSize(20).TextColor(p.Accent)
			})
			mu.Text(c, "核心尚未运行").FontSize(16).Bold()
			mu.Text(c, "启动 mihomo 核心后才能读取代理组和节点。").FontSize(12).TextColor(p.Muted).TextAlign(mu.Center)
			mu.PrimaryButton(c, "启动核心").Disabled(u.busy["core"] || u.ctrl.Snapshot().Starting).OnClick(u.toggleCore)
		})
	})
}
