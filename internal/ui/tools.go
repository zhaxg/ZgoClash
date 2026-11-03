package ui

import (
	"fmt"
	"strconv"
	"strings"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/app"
)

// repoURL is the project's page, linked in the About section.
const repoURL = "https://github.com/zhaxg/ZgoClash"

func (u *UI) toolsPage(c *mu.Context, snap app.Snapshot, p palette) {
	u.ensureTools(snap.Settings)
	mu.Scroll(c).Grow(1).Children(func() {
		mu.Column(c).FillWidth().Padding(22, 24).Gap(14).Children(func() {
			u.pageHeader(c, p, "设置", "应用偏好、网络参数与内核信息", nil)
			u.generalSection(c, snap, p)
			u.networkSection(c, snap, p)
			u.coreSection(c, snap, p)
			u.aboutSection(c, snap, p)
			mu.Box(c).Height(10)
		})
	})
	u.configModal(c, p)
}

func (u *UI) section(c *mu.Context, p palette, title, desc string, body func()) {
	card(c, p).Gap(9).Children(func() {
		mu.Text(c, title).FontSize(14.5).Bold()
		if desc != "" {
			mu.Text(c, desc).FontSize(11).TextColor(p.Muted)
		}
		body()
	})
}

func (u *UI) settingRow(c *mu.Context, p palette, title, desc string, control func()) {
	mu.Row(c).Padding(8, 0).Gap(14).AlignItems(mu.Center).Children(func() {
		mu.Column(c).Gap(2).Grow(1).Children(func() {
			mu.Text(c, title).FontSize(13)
			if desc != "" {
				mu.Text(c, desc).FontSize(11).TextColor(p.Muted)
			}
		})
		control()
	})
}

func (u *UI) generalSection(c *mu.Context, snap app.Snapshot, p palette) {
	u.section(c, p, "通用", "界面与启动行为", func() {
		theme := 0
		switch snap.Settings.Theme {
		case "light":
			theme = 1
		case "dark":
			theme = 2
		}
		u.settingRow(c, p, "外观", "跟随系统或固定明暗", func() {
			if mu.Segmented(c, &theme, "跟随系统", "浅色", "深色").Changed() {
				value := []string{"system", "light", "dark"}[theme]
				u.patchSettings(func(s *app.Settings) { s.Theme = value })
			}
		})

		mu.Divider(c)
		u.switchRow(c, p, "开机自动启动", "登录 Windows 后在后台启动 ZgoClash",
			snap.Settings.AutoLaunch, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.AutoLaunch = v })
			})
		u.switchRow(c, p, "静默启动", "开机启动时不显示主窗口",
			snap.Settings.SilentStart, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.SilentStart = v })
			})
		u.switchRow(c, p, "关闭窗口时最小化到托盘", "保持内核在后台运行",
			snap.Settings.CloseToTray, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.CloseToTray = v })
			})
		u.switchRow(c, p, "启动时运行核心", "打开应用后自动启动 mihomo",
			snap.Settings.AutoStartCore, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.AutoStartCore = v })
			})
	})
}

func (u *UI) networkSection(c *mu.Context, snap app.Snapshot, p palette) {
	u.section(c, p, "网络", "代理端口、TUN 与系统集成", func() {
		u.switchRow(c, p, "系统代理", "把 Windows 的 Internet 代理指向核心",
			snap.Settings.SystemProxy, true, func(v bool) {
				u.run("sysproxy", func() error { return u.ctrl.SetSystemProxy(v) }, nil)
			})
		tunDesc := "创建虚拟网卡接管全部流量"
		if !snap.CanTun {
			tunDesc = "需要管理员权限，重启为管理员后可用"
		}
		u.switchRow(c, p, "TUN 模式", tunDesc, snap.Settings.Tun, snap.CanTun, func(v bool) {
			u.run("tun", func() error { return u.ctrl.SetTun(v) }, nil)
		}, func() {
			// The restart button sits after the switch, in the same row.
			if !snap.Elevated {
				if mu.Button(c, "以管理员身份重启").Height(30).Clicked() {
					u.run("elevate", func() error { return u.ctrl.RestartElevated() }, nil)
				}
			}
		})

		stackDesc := "TUN 使用的网络栈"
		u.settingRow(c, p, "TUN 栈", stackDesc, func() {
			idx := 0
			switch snap.Settings.TunStack {
			case "system":
				idx = 1
			case "mixed":
				idx = 2
			}
			if mu.Segmented(c, &idx, "GVisor", "系统", "混合").Changed() {
				value := []string{"gvisor", "system", "mixed"}[idx]
				u.patchSettings(func(s *app.Settings) { s.TunStack = value })
			}
		})

		mu.Divider(c)
		u.settingRow(c, p, "混合端口", "HTTP / SOCKS 共用的本地端口", func() {
			mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
				mu.TextInput(c.Key("mixed-port"), &u.portText).Width(92)
				if mu.Button(c, "应用").Height(30).Clicked() {
					u.saveMixedPort()
				}
			})
		})
		u.switchRow(c, p, "允许局域网连接", "同一网络内的设备可以通过本机代理上网",
			snap.Settings.AllowLan, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.AllowLan = v })
			})
		u.switchRow(c, p, "IPv6", "允许核心使用 IPv6 解析与连接",
			snap.Settings.IPv6, true, func(v bool) {
				u.patchSettings(func(s *app.Settings) { s.IPv6 = v })
			})

		mu.Divider(c)
		u.settingRow(c, p, "延迟测试地址", "节点测速使用的 URL", func() {
			mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
				mu.TextInput(c.Key("test-url"), &u.testURLText).Width(384)
				if mu.Button(c, "应用").Height(30).Clicked() {
					value := strings.TrimSpace(u.testURLText)
					if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
						u.Notify("error", "测试地址需要以 http:// 或 https:// 开头")
					} else {
						u.patchSettings(func(s *app.Settings) { s.TestURL = value })
					}
				}
			})
		})

		autoLabels := []string{"关闭", "每 6 小时", "每 12 小时", "每 24 小时", "每 72 小时"}
		autoHours := []int{0, 6, 12, 24, 72}
		u.settingRow(c, p, "订阅自动更新", "按间隔刷新所有订阅", func() {
			sel := autoLabels[u.autoUpdateIdx]
			if mu.Select(c, &sel, autoLabels).Changed() {
				for i, label := range autoLabels {
					if label == sel {
						u.autoUpdateIdx = i
						hours := autoHours[i]
						u.patchSettings(func(s *app.Settings) { s.AutoUpdateInterval = hours })
					}
				}
			}
		})

		u.settingRow(c, p, "外部控制端口", "核心 REST API 的端口，回车生效", func() {
			mu.Row(c).Gap(12).AlignItems(mu.Center).Children(func() {
				input := mu.TextInput(c.Key("controller-port"), &u.controllerText).Width(92)
				external := snap.Settings.ControllerExternal
				toggle := mu.Switch(c, &external)
				if toggle.Changed() || input.Submitted() {
					u.saveControllerPort(external)
				}
			})
		})
	})
}

// saveMixedPort stores the port the mixed proxy listens on.
func (u *UI) saveMixedPort() {
	mixed, err := strconv.Atoi(strings.TrimSpace(u.portText))
	if err != nil || mixed < 1 || mixed > 65535 {
		u.Notify("error", "混合端口需要是 1-65535 之间的数字")
		return
	}
	u.patchSettings(func(s *app.Settings) {
		s.MixedPort = mixed
	})
	u.Notify("info", "端口设置已保存，核心会在需要时重启")
}

// saveControllerPort stores the port the core's REST API listens on, and
// whether other devices may reach it.
func (u *UI) saveControllerPort(external bool) {
	control, err := strconv.Atoi(strings.TrimSpace(u.controllerText))
	if err != nil || control < 1 || control > 65535 {
		u.Notify("error", "外部控制端口需要是 1-65535 之间的数字")
		return
	}
	u.patchSettings(func(s *app.Settings) {
		s.ControllerPort = control
		s.ControllerExternal = external
	})
	u.Notify("info", "端口设置已保存，核心会在需要时重启")
}

func (u *UI) coreSection(c *mu.Context, snap app.Snapshot, p palette) {
	u.section(c, p, "内核", "mihomo 的运行状态与数据", func() {
		state := "已停止"
		stateColor := p.Bad
		if snap.Starting {
			state, stateColor = "启动中", p.Warn
		} else if snap.Running {
			state, stateColor = "运行中", p.Good
		}
		info := func(label, value string, color mu.Color) {
			mu.Column(c).Gap(2).Children(func() {
				mu.Text(c, label).FontSize(11).TextColor(p.Muted)
				mu.Text(c, value).FontSize(12).TextColor(color).Selectable().SingleLine()
			})
		}
		// The facts sit in a grid four columns wide; the data directory
		// has a button of its own below.
		mu.Grid(c).Columns(4).Gap(14).Children(func() {
			info("状态", state, stateColor)
			info("运行时长", formatUptime(snap.Uptime), p.Text)
			info("混合端口", fmt.Sprintf("%d", snap.MixedPort), p.Text)
			info("API 端口", fmt.Sprintf("%d", snap.ControllerPort), p.Text)
			info("内存占用", formatBytes(int64(snap.MemoryBytes)), p.Text)
			info("内核版本", "mihomo "+snap.CoreVersion, p.Text)
		})
		if snap.LastError != "" {
			mu.Text(c, snap.LastError).FontSize(11).TextColor(p.Bad)
		}

		mu.Divider(c)
		logLevels := []string{"静默", "错误", "警告", "信息", "调试"}
		logLevelStrings := []string{"silent", "error", "warning", "info", "debug"}
		u.settingRow(c, p, "核心日志级别", "", func() {
			sel := logLevels[logLevelIndex(snap.Settings.LogLevel)]
			if mu.Select(c, &sel, logLevels).Changed() {
				for i, label := range logLevels {
					if label == sel {
						value := logLevelStrings[i]
						u.patchSettings(func(s *app.Settings) { s.LogLevel = value })
					}
				}
			}
		})

		mu.Row(c).Gap(8).Wrap().Children(func() {
			if snap.Running && mu.Button(c, "重启核心").Disabled(u.busy["restart"]).Clicked() {
				u.run("restart", func() error { return u.ctrl.Restart() }, nil)
			}
			if mu.Button(c, "打开数据目录").Clicked() {
				u.run("open-dir", func() error { return u.ctrl.OpenDataDir() }, nil)
			}
			if mu.Button(c, "运行时配置").Clicked() {
				u.openConfig()
			}
		})
	})
}

func (u *UI) openConfig() {
	u.run("config", func() error {
		text, err := u.ctrl.RuntimeConfig()
		if err != nil {
			return err
		}
		u.win.Update(func() {
			u.configText = text
			u.configOpen = true
		})
		return nil
	}, nil)
}

func (u *UI) configModal(c *mu.Context, p palette) {
	mu.Modal(c, &u.configOpen, func() {
		mu.Column(c).Width(720).Gap(10).Children(func() {
			mu.Text(c, "运行时配置").FontSize(17).Bold()
			mu.Text(c, "这是核心实际加载的 YAML，包含应用覆盖后的设置。").FontSize(11.5).TextColor(p.Muted)
			mu.TextArea(c.Key("runtime-config"), &u.configText).Height(430).ReadOnly(true).Font("monospace")
			mu.Row(c).Justify(mu.End).Children(func() {
				if mu.PrimaryButton(c, "关闭").Clicked() {
					u.configOpen = false
				}
			})
		})
	})
}

func (u *UI) aboutSection(c *mu.Context, snap app.Snapshot, p palette) {
	u.section(c, p, "关于", "", func() {
		mu.Row(c).Gap(12).AlignItems(mu.Center).Children(func() {
			if u.mark != nil {
				mu.Image(c, u.mark).Size(42, 42).Fit(mu.Contain)
			} else {
				mu.Box(c).Size(42, 42).Radius(12).Background(p.Accent).Center().Children(func() {
					mu.Text(c, "Z").FontSize(22).Bold().TextColor(p.AccentText)
				})
			}
			mu.Column(c).Gap(2).Children(func() {
				mu.Text(c, "ZgoClash "+snap.AppVersion).FontSize(14).Bold()
				mu.Text(c, "基于 mygo Native UI 与 mihomo 构建").FontSize(11).TextColor(p.Muted)
			})
		})
		mu.Text(c, "一个紧凑的 Windows mihomo 客户端：系统代理、TUN、订阅管理、节点选择与实时连接都在一个原生界面里。").
			FontSize(11.5).TextColor(p.Muted)
		mu.Row(c).Gap(8).AlignItems(mu.Center).Children(func() {
			mu.Icon(c, iconGithub).FontSize(15)
			mu.Link(c, repoURL, repoURL).FontSize(11.5)
		})
	})
}
