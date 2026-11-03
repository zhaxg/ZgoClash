package ui

import (
	"context"
	"fmt"
	"time"

	mu "github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"

	"zgo-clash/internal/app"
)

func (u *UI) profilesPage(c *mu.Context, snap app.Snapshot, p palette) {
	mu.Column(c).Grow(1).Padding(22, 24).Gap(13).Children(func() {
		u.pageHeader(c, p, "订阅", "订阅、文件与手动配置", func() {
			if el := u.iconButton(c, p, iconSync, "更新全部", u.busy["update-all"]); el.Clicked() {
				u.updateAllProfiles()
			}
			u.iconButton(c, p, iconPlus, "添加配置").Menu(func(m *mu.Menu) {
				if m.Item("从订阅链接添加").Chosen() {
					u.addOpen = true
				}
				if m.Item("从文件导入").Chosen() {
					u.importFile()
				}
				if m.Item("手动创建").Chosen() {
					u.manualOpen = true
				}
				if m.Item("从剪贴板导入").Chosen() {
					u.importClipboard()
				}
			})
		})

		if len(snap.Profiles) == 0 {
			card(c, p).Grow(1).Center().Gap(8).Children(func() {
				mu.Text(c, "还没有配置").FontSize(16).Bold()
				mu.Text(c, "添加一个订阅链接，或导入本地 Clash 配置。").FontSize(12).TextColor(p.Muted)
				if mu.PrimaryButton(c, "添加订阅").Clicked() {
					u.addOpen = true
				}
			})
		} else {
			mu.Scroll(c).Grow(1).Children(func() {
				mu.Column(c).Gap(12).Children(func() {
					for _, prof := range snap.Profiles {
						u.profileCard(c, p, snap, prof)
					}
					mu.Text(c, fmt.Sprintf("共 %d 个配置", len(snap.Profiles))).FontSize(11).
						TextColor(p.Muted).Padding(2, 4)
				})
			})
		}
	})
	u.profileModals(c, p)
	u.configModal(c, p)
}

func (u *UI) profileCard(c *mu.Context, p palette, snap app.Snapshot, prof app.Profile) {
	current := prof.ID == snap.CurrentProfileID
	busyKey := "profile:" + prof.ID
	busy := u.busy[busyKey]
	switching := u.busy["use:"+prof.ID]

	mu.Row(c).Gap(8).AlignItems(mu.Start).Children(func() {
		// The card itself is the button that switches to the profile.
		body := mu.ButtonBase(c).Grow(1).Padding(18).Radius(14).Background(p.Card).Border(1, p.Border).
			AlignItems(mu.Start).Justify(mu.Start).Label(prof.Name).Children(func() {
			mu.Column(c).Gap(5).Grow(1).Children(func() {
				mu.Row(c).Gap(10).AlignItems(mu.Center).Children(func() {
					mu.Row(c).Gap(7).AlignItems(mu.Center).Grow(1).Children(func() {
						mu.Text(c, prof.Name).FontSize(15).Bold().SingleLine().Font(emojiFamily)
						typeName := "手动"
						switch prof.Type {
						case "url":
							typeName = "订阅"
						case "file":
							typeName = "文件"
						}
						pill(c, typeName, p.Muted, p.Soft)
					})
					// The mark of the profile in use, next to the menu.
					if current {
						mu.Icon(c, iconCheckCircle).FontSize(27).TextColor(p.Mark)
					}
					u.iconButton(c, p, iconDots, "更多").Menu(func(m *mu.Menu) {
						if prof.Type == "url" {
							if m.Item("同步").Chosen() {
								u.updateProfile(prof.ID)
							}
						}
						if m.Item("编辑").Chosen() {
							u.editID = prof.ID
							u.editName = prof.Name
							u.editURL = prof.URL
							u.editIsURL = prof.Type == "url"
							u.editOpen = true
						}
						if m.Item("查看").Chosen() {
							u.openConfig(prof.ID)
						}
						if m.Item("导出").Chosen() {
							u.exportProfile(prof.ID)
						}
						m.Separator()
						if m.Item("删除").Chosen() {
							u.deleteID = prof.ID
							u.deleteName = prof.Name
							u.deleteOpen = true
						}
					})
				})
				if prof.URL != "" {
					mu.Text(c, hostOf(prof.URL)).FontSize(11).TextColor(p.Muted).SingleLine()
				}
				mu.Row(c).Gap(14).Wrap().Children(func() {
					mu.Text(c, fmt.Sprintf("%d 节点 · %d 规则 · %d 分组", prof.ProxyCount, prof.RuleCount, prof.GroupCount)).
						FontSize(11).TextColor(p.Muted)
					mu.Text(c, "更新于 "+formatAgo(prof.UpdatedAt)).FontSize(11).TextColor(p.Muted)
					if busy {
						mu.Text(c, "同步中…").FontSize(11).TextColor(p.Accent)
					}
					if exp := formatExpire(prof.Expire); exp != "" {
						mu.Text(c, exp).FontSize(11).TextColor(p.Warn)
					}
				})
				if prof.Total > 0 {
					mu.Row(c).Gap(10).AlignItems(mu.Center).Children(func() {
						mu.Progress(c, clamp01(percent(prof.Upload+prof.Download, prof.Total))).Grow(1).Height(6)
						mu.Text(c, fmt.Sprintf("%s / %s", formatBytes(prof.Upload+prof.Download), formatBytes(prof.Total))).
							FontSize(10.5).TextColor(p.Muted)
					})
				}
			})
		})
		if !current && !switching {
			if body.Hovered() {
				body.Background(p.CardHover)
			}
			if body.Clicked() {
				u.useProfile(prof.ID)
			}
		}
	})
}

// openConfig shows the configuration of a profile in the editor: the
// runtime one the core loaded for the profile in use, the file it was
// fetched from for the others.
func (u *UI) openConfig(id string) {
	u.run("config", func() error {
		var text string
		var err error
		if id == u.ctrl.Snapshot().CurrentProfileID {
			text, err = u.ctrl.RuntimeConfig()
		} else {
			text, err = u.ctrl.ProfileContent(id)
		}
		if err != nil {
			return err
		}
		u.win.Update(func() {
			u.configText = text
			u.configState = newConfigEditor(text)
			u.configOpen = true
		})
		return nil
	}, nil)
}

// newConfigEditor makes the editor that shows a configuration: the YAML of
// it, with the folds of the gutter and the numbers of the lines. The font
// falls back to the colour emoji font, or a flag would be drawn as the
// letters of its country code.
func newConfigEditor(text string) *editor.State {
	buf := editor.NewBuffer(text)
	buf.SetLanguage(editor.LanguageFor("config.yaml"))
	return &editor.State{
		Buffer:     buf,
		ReadOnly:   true,
		Font:       mu.Font{Family: monoFamily, Size: 11.5},
		WrapGuides: []int{100},
	}
}

// configModal shows the configuration of a profile in a code editor, whose
// folds hide the long lists away.
func (u *UI) configModal(c *mu.Context, p palette) {
	mu.Modal(c, &u.configOpen, func() {
		mu.Column(c).Width(820).Height(600).Gap(10).Children(func() {
			mu.Text(c, "运行时配置").FontSize(17).Bold()
			mu.Text(c, "这是核心实际加载的 YAML，包含应用覆盖后的设置。").FontSize(11.5).TextColor(p.Muted)
			editor.View(c, u.configState).Grow(1).FillWidth()
			mu.Row(c).Justify(mu.End).Children(func() {
				if mu.PrimaryButton(c, "关闭").Clicked() {
					u.configOpen = false
				}
			})
		})
	})
}

func (u *UI) useProfile(id string) {
	u.run("use:"+id, func() error { return u.ctrl.SetCurrentProfile(id) }, func() {
		u.Notify("success", "已切换配置")
		u.proxyLoadedAt = time.Time{}
	})
}

func (u *UI) updateProfile(id string) {
	u.run("profile:"+id, func() error {
		_, err := u.ctrl.UpdateProfile(id)
		return err
	}, func() {
		u.Notify("success", "订阅已更新")
		u.proxyLoadedAt = time.Time{}
	})
}

func (u *UI) updateAllProfiles() {
	u.run("update-all", func() error {
		report, err := u.ctrl.UpdateAll()
		if err != nil {
			return err
		}
		if len(report.Failed) > 0 {
			return fmt.Errorf("%d 个订阅更新失败: %s", len(report.Failed), report.Failed[0])
		}
		return nil
	}, func() {
		u.Notify("success", "全部订阅已更新")
		u.proxyLoadedAt = time.Time{}
	})
}

func (u *UI) importFile() {
	u.run("add-file", func() error {
		_, err := u.ctrl.AddFile(u.win)
		return err
	}, func() {
		u.Notify("success", "配置已导入")
	})
}

func (u *UI) importClipboard() {
	u.run("clipboard", func() error {
		_, err := u.ctrl.ImportClipboard()
		return err
	}, func() {
		u.Notify("success", "已从剪贴板导入")
	})
}

func (u *UI) exportProfile(id string) {
	u.run("export:"+id, func() error {
		_, err := u.ctrl.ExportProfile(u.win, id)
		return err
	}, func() {
		u.Notify("success", "配置已导出")
	})
}

func (u *UI) profileModals(c *mu.Context, p palette) {
	mu.Modal(c, &u.addOpen, func() {
		mu.Column(c).Width(470).Gap(10).Children(func() {
			mu.Text(c, "添加订阅").FontSize(17).Bold()
			mu.Text(c, "粘贴 Clash / mihomo 订阅链接，名称可以留空。").FontSize(11.5).TextColor(p.Muted)
			mu.Text(c, "名称").FontSize(11).TextColor(p.Muted).Padding(4, 0, 0, 0)
			mu.TextInput(c.Key("add-name"), &u.addName).Placeholder("例如：机场订阅")
			mu.Text(c, "订阅链接").FontSize(11).TextColor(p.Muted)
			mu.TextInput(c.Key("add-url"), &u.addURL).Placeholder("https://example.com/subscribe")
			mu.Row(c).Gap(8).Justify(mu.End).Children(func() {
				if mu.Button(c, "取消").Clicked() {
					u.addOpen = false
				}
				if mu.PrimaryButton(c, "添加").Disabled(u.busy["add-url"]).Clicked() {
					u.submitAddURL()
				}
			})
		})
	})

	mu.Modal(c, &u.manualOpen, func() {
		mu.Column(c).Width(620).Gap(10).Children(func() {
			mu.Text(c, "手动创建配置").FontSize(17).Bold()
			mu.Text(c, "直接粘贴 YAML 配置，至少包含 proxies、proxy-groups 或 rules。").FontSize(11.5).TextColor(p.Muted)
			mu.TextInput(c.Key("manual-name"), &u.manualName).Placeholder("配置名称")
			mu.TextArea(c.Key("manual-body"), &u.manualText).Height(320).Font("monospace")
			mu.Row(c).Gap(8).Justify(mu.End).Children(func() {
				if mu.Button(c, "取消").Clicked() {
					u.manualOpen = false
				}
				if mu.PrimaryButton(c, "创建").Disabled(u.busy["manual"]).Clicked() {
					u.submitManual()
				}
			})
		})
	})

	mu.Modal(c, &u.editOpen, func() {
		mu.Column(c).Width(420).Gap(10).Children(func() {
			mu.Text(c, "编辑配置").FontSize(17).Bold()
			mu.Text(c, "名称").FontSize(11.5).TextColor(p.Muted)
			mu.TextInput(c.Key("edit-name"), &u.editName)
			if u.editIsURL {
				mu.Text(c, "订阅链接").FontSize(11.5).TextColor(p.Muted)
				// The link is long, so the field keeps room for three
				// lines of it.
				mu.TextArea(c.Key("edit-url"), &u.editURL).Height(66)
			}
			mu.Row(c).Gap(8).Justify(mu.End).Children(func() {
				if mu.Button(c, "取消").Clicked() {
					u.editOpen = false
				}
				if mu.PrimaryButton(c, "保存").Clicked() {
					u.submitEdit()
				}
			})
		})
	})

	if mu.AlertDialog(c, &u.deleteOpen, "删除配置？",
		"“"+u.deleteName+"” 将从列表中移除，这个操作无法撤销。", "取消", "删除") == 1 {
		id := u.deleteID
		u.run("delete:"+id, func() error { return u.ctrl.DeleteProfile(id) }, func() {
			u.Notify("success", "配置已删除")
		})
	}
}

func (u *UI) submitAddURL() {
	url := u.addURL
	name := u.addName
	u.run("add-url", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		_, err := u.ctrl.AddURL(ctx, url, name)
		return err
	}, func() {
		u.addOpen = false
		u.addURL = ""
		u.addName = ""
		u.Notify("success", "订阅已添加")
	})
}

func (u *UI) submitManual() {
	name := u.manualName
	text := u.manualText
	u.run("manual", func() error {
		_, err := u.ctrl.CreateManual(name, text)
		return err
	}, func() {
		u.manualOpen = false
		u.manualName = ""
		u.manualText = ""
		u.Notify("success", "配置已创建")
	})
}

// submitEdit saves the name and the link of a profile.
func (u *UI) submitEdit() {
	id := u.editID
	name := u.editName
	url := u.editURL
	u.run("edit", func() error {
		return u.ctrl.EditProfile(id, name, url)
	}, func() {
		u.editOpen = false
		u.Notify("success", "配置已更新")
	})
}
