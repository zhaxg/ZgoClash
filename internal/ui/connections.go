package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
	"zgo-clash/internal/engine"
)

func (u *UI) connectionsPage(c *mu.Context, snap app.Snapshot, p palette) {
	live := u.ctrl.Connections()
	rows := u.filterConnections(live.Connections)
	u.connRows = rows
	compact := compactConnections(c)

	mu.Column(c).Grow(1).Padding(22, 24).Gap(13).Children(func() {
		u.pageHeader(c, p, "连接", "核心当前处理的每一条连接", func() {
			if el := u.iconButton(c, p, iconUnlink, "关闭全部",
				len(live.Connections) == 0 || u.busy["close-all"]).Tooltip("断开所有连接"); el.Clicked() {
				u.connCloseAll = true
			}
			if el := u.iconButton(c, p, iconRefresh, "刷新"); el.Clicked() {
				u.win.Invalidate()
			}
		})

		mu.Row(c).Gap(12).AlignItems(mu.Center).Wrap().Children(func() {
			mu.SearchField(c, &u.connQuery).Width(280).Placeholder("搜索主机、进程或规则")
			mu.Text(c, fmt.Sprintf("%d 条连接", len(live.Connections))).FontSize(11.5).TextColor(p.Muted)
			mu.Text(c, "↑ "+formatBytes(live.UploadTotal)).FontSize(11.5).TextColor(p.Muted)
			mu.Text(c, "↓ "+formatBytes(live.DownloadTotal)).FontSize(11.5).TextColor(p.Teal)
			mu.Spacer(c)
			if !snap.Running {
				mu.Text(c, "核心未运行").FontSize(11.5).TextColor(p.Bad)
			} else {
				mu.Text(c, "自动刷新").FontSize(11.5).TextColor(p.Muted)
			}
		})

		card(c, p).Grow(1).Gap(0).Padding(0).Children(func() {
			mu.Row(c).Padding(10, 14).Gap(12).AlignItems(mu.Center).BorderWidth(0, 0, 1, 0).
				BorderColor(p.Border).Children(func() {
				headerText(c, p, "主机 / 目标", 0)
				if !compact {
					headerText(c, p, "代理链", 170)
					headerText(c, p, "网络", 64)
				}
				headerText(c, p, "上传", 80)
				headerText(c, p, "下载", 80)
				headerText(c, p, "持续时间", 82)
				mu.Box(c).Width(26)
			})
			if len(rows) == 0 {
				mu.Column(c).Grow(1).Center().Gap(6).Children(func() {
					mu.Text(c, "暂无活动连接").FontSize(13.5).TextColor(p.Muted)
					mu.Text(c, "有流量经过核心时，连接会出现在这里。").FontSize(11.5).TextColor(p.Muted)
				})
				return
			}
			u.connList.Key = func(row int) any {
				if row < len(u.connRows) {
					return u.connRows[row].ID
				}
				return row
			}
			mu.List(c, &u.connList, len(rows), func(i int) {
				u.connectionRow(c, p, rows[i], compact)
			}).Grow(1)
		})
	})

	u.connectionModals(c, p)
}

func compactConnections(c *mu.Context) bool {
	w, _ := c.Size()
	return w < 1080
}

func headerText(c *mu.Context, p palette, label string, width float32) {
	el := mu.Text(c, label).FontSize(10.5).TextColor(p.Muted)
	if width > 0 {
		el.Width(width)
	} else {
		el.Grow(1)
	}
}

func (u *UI) filterConnections(items []engine.Connection) []engine.Connection {
	query := strings.ToLower(strings.TrimSpace(u.connQuery))
	out := make([]engine.Connection, 0, len(items))
	for _, cn := range items {
		if query != "" {
			hay := strings.ToLower(cn.Metadata.Host + " " + cn.Metadata.DestinationIP + " " +
				cn.Metadata.Process + " " + cn.Rule + " " + strings.Join(cn.Chains, " "))
			if !strings.Contains(hay, query) {
				continue
			}
		}
		out = append(out, cn)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start > out[j].Start })
	return out
}

func (u *UI) connectionRow(c *mu.Context, p palette, cn engine.Connection, compact bool) {
	host := cn.Metadata.Host
	if host == "" {
		host = cn.Metadata.DestinationIP
	}
	if host == "" {
		host = "(未知目标)"
	}
	if cn.Metadata.DestinationPort != "" {
		host += ":" + cn.Metadata.DestinationPort
	}
	chain := "直连"
	if len(cn.Chains) > 0 {
		chain = strings.Join(cn.Chains, " → ")
	}
	dur := ""
	if start := startTimeOf(cn); !start.IsZero() {
		dur = formatDuration(int64(time.Since(start).Seconds()))
	}

	el := mu.ButtonBase(c).Fill().Padding(8, 14).Gap(12).AlignItems(mu.Center).
		BorderWidth(0, 0, 1, 0).BorderColor(p.Border).Label(host).Children(func() {
		mu.Column(c).Gap(1).Grow(1).Children(func() {
			mu.Text(c, host).FontSize(12).TextColor(p.Text).SingleLine()
			if cn.Metadata.Process != "" {
				mu.Text(c, cn.Metadata.Process).FontSize(9.5).TextColor(p.Muted).SingleLine()
			}
		})
		if !compact {
			mu.Text(c, chain).FontSize(11).Font(emojiFamily).TextColor(p.Muted).Width(170).SingleLine()
			mu.Text(c, cn.Metadata.Network).FontSize(11).TextColor(p.Muted).Width(64)
		}
		mu.Text(c, formatBytes(cn.Upload)).FontSize(11).TextColor(p.Muted).Width(80)
		mu.Text(c, formatBytes(cn.Download)).FontSize(11).TextColor(p.Teal).Width(80)
		mu.Text(c, dur).FontSize(11).TextColor(p.Muted).Width(82)
		closeBtn := mu.ButtonBase(c).Size(24, 24).Radius(6).AlignItems(mu.Center).Justify(mu.Center).
			Label("关闭连接").Tooltip("关闭这条连接").Children(func() {
			mu.Icon(c, iconX).FontSize(11).TextColor(p.Muted)
		})
		if closeBtn.Hovered() {
			closeBtn.Background(p.Soft)
		}
		if closeBtn.Clicked() {
			u.closeConnection(cn.ID)
		}
	})
	if el.Hovered() {
		el.Background(p.CardHover)
	}
	if el.Clicked() {
		copy := cn
		u.connDetail = &copy
	}
}

func (u *UI) closeConnection(id string) {
	u.run("conn:"+id, func() error {
		return u.ctrl.CloseConnection(id)
	}, func() {
		if u.connDetail != nil && u.connDetail.ID == id {
			u.connDetail = nil
		}
	})
}

func (u *UI) connectionModals(c *mu.Context, p palette) {
	open := u.connDetail != nil
	mu.Modal(c, &open, func() {
		cn := u.connDetail
		if cn == nil {
			return
		}
		host := cn.Metadata.Host
		if host == "" {
			host = cn.Metadata.DestinationIP
		}
		mu.Column(c).Width(560).Gap(10).Children(func() {
			mu.Text(c, host).FontSize(16).Bold().SingleLine()
			mu.Text(c, "连接详情").FontSize(11).TextColor(p.Muted)

			mu.Column(c).Gap(6).Padding(12).Radius(10).Background(p.Soft).Children(func() {
				detailLine(c, p, "目标地址", cn.Metadata.DestinationIP+":"+cn.Metadata.DestinationPort)
				detailLine(c, p, "来源", cn.Metadata.SourceIP+":"+cn.Metadata.SourcePort)
				detailLine(c, p, "网络 / 类型", cn.Metadata.Network+" / "+cn.Metadata.Type)
				detailLine(c, p, "代理链", strings.Join(cn.Chains, " → "))
				detailLine(c, p, "规则", cn.Rule+" "+cn.RulePayload)
				detailLine(c, p, "进程", cn.Metadata.Process)
				detailLine(c, p, "上传 / 下载", formatBytes(cn.Upload)+" / "+formatBytes(cn.Download))
				detailLine(c, p, "开始时间", formatConnStart(*cn))
				detailLine(c, p, "连接 ID", cn.ID)
			})

			mu.Row(c).Gap(8).Justify(mu.End).Children(func() {
				if mu.Button(c, "关闭").Clicked() {
					u.connDetail = nil
				}
				if mu.PrimaryButton(c, "关闭此连接").Clicked() {
					u.closeConnection(cn.ID)
				}
			})
		})
	})
	if !open {
		u.connDetail = nil
	}

	if mu.AlertDialog(c, &u.connCloseAll, "关闭全部连接？",
		"所有正在经过核心的连接都会被断开，应用会重新建立连接。", "取消", "关闭全部") == 1 {
		u.run("close-all", func() error { return u.ctrl.CloseAllConnections() }, func() {
			u.Notify("success", "已关闭全部连接")
		})
	}
}

func detailLine(c *mu.Context, p palette, label, value string) {
	mu.Row(c).Gap(12).Children(func() {
		mu.Text(c, label).FontSize(11).TextColor(p.Muted).Width(88)
		mu.Text(c, value).FontSize(11.5).Font(emojiFamily).TextColor(p.Text).Grow(1).Selectable()
	})
}

func formatConnStart(cn engine.Connection) string {
	start := startTimeOf(cn)
	if start.IsZero() {
		return cn.Start
	}
	return start.Format("2006-01-02 15:04:05")
}
