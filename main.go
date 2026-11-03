// ZgoClash is a compact mihomo client for Windows, built with MyGo's native UI.
package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	mui "github.com/egoist/mygo/ui"
	"golang.org/x/image/draw"

	"zgo-clash/internal/app"
	appui "zgo-clash/internal/ui"
)

const appVersion = "1.0.0"

//go:embed resources/icon.png
var appIcon []byte

//go:embed resources/tray-enable.png
var trayEnabled []byte

//go:embed resources/tray-disable.png
var trayDisabled []byte

// traySide is the size the notification area draws; the source art is much
// larger, so it is scaled once here instead of by the shell at every update.
const traySide = 64

var (
	trayEnabledPNG  = trayImage(trayEnabled)
	trayDisabledPNG = trayImage(trayDisabled)
)

func trayImage(src []byte) []byte {
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return src
	}
	b := img.Bounds()
	if b.Dx() <= traySide && b.Dy() <= traySide {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, traySide, traySide))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	var buf bytes.Buffer
	if png.Encode(&buf, dst) != nil {
		return src
	}
	return buf.Bytes()
}

var (
	mainWindow    *mygo.Window
	mainUI        *appui.UI
	mainCtrl      *app.Controller
	tray          *mygo.Tray
	trayCoreItem  *mygo.MenuItem
	trayProxyItem *mygo.MenuItem
	quitting      atomic.Bool
	windowAwake   atomic.Bool
	trayMu        sync.Mutex
	trayShownMu   sync.Mutex
	trayRunning   bool
	trayShownRun  bool
	trayShownProx bool
	trayShownStat string
)

func main() {
	mygo.App.SetName("ZgoClash")
	mygo.App.SetVersion(appVersion)
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}

	mygo.App.OnSecondInstance(func(args []string, workingDir string) { showMain() })
	mygo.App.OnActivate(func(hasVisibleWindows bool) {
		if !hasVisibleWindows {
			showMain()
		}
	})
	mygo.App.OnWindowAllClosed(func() { mygo.App.Quit() })

	mygo.App.WhenReady(func() {
		dir, err := mygo.App.Path(mygo.PathUserData)
		if err != nil {
			log.Println("user data:", err)
			dir = "."
		}
		ctrl := app.NewController(dir)
		if err := ctrl.Init(); err != nil {
			mygo.Dialog.Error("ZgoClash 启动失败", err.Error())
			mygo.App.Quit()
			return
		}

		appUI := appui.New(ctrl, appIcon)
		mainUI, mainCtrl = appUI, ctrl
		settings := ctrl.Settings()
		hidden := settings.SilentStart || mygo.App.WasOpenedAtLogin()
		windowAwake.Store(!hidden)

		// On small or high-DPI displays the default size can be larger than
		// the desktop, so fit it to the work area before opening. The height
		// covers the dashboard in full: 776 of rows and padding plus the
		// 44 px title bar, so the overview shows no scroll bar.
		width, height := 1200, 840
		if d := mygo.Screen.PrimaryDisplay(); d.WorkArea.Width > 0 {
			width = min(width, d.WorkArea.Width-48)
			height = min(height, d.WorkArea.Height-48)
		}
		width = max(width, 860)
		height = max(height, 560)

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "ZgoClash",
			Width:     width,
			Height:    height,
			MinWidth:  860,
			MinHeight: 560,
			// No StateKey: the window opens centered on its display every
			// time, at the size fitted to the work area above.
			Hidden:          hidden,
			TitleBarStyle:   mygo.TitleBarHidden,
			TitleBarHeight:  44,
			BackgroundColor: "light-dark(#f2f5fa, #0d1016)",
			Content:         mui.View(appUI.View),
		})
		mainWindow = win
		appUI.SetWindow(win)

		mygo.App.OnBeforeQuit(func(e *mygo.QuitEvent) {
			quitting.Store(true)
			_ = ctrl.Stop()
		})
		win.OnClose(func(e *mygo.CloseEvent) {
			if quitting.Load() {
				return
			}
			if ctrl.Settings().CloseToTray {
				e.PreventDefault()
				win.Hide()
			}
		})
		// With no window on screen, the polls and the caches go away and the
		// process keeps only the core's own memory until it is shown again.
		win.OnShow(func() { setAwake(appUI, ctrl, true) })
		win.OnHide(func() { setAwake(appUI, ctrl, false) })
		win.OnRestore(func() { setAwake(appUI, ctrl, true) })
		win.OnMinimize(func() { setAwake(appUI, ctrl, false) })
		if hidden {
			win.OnReadyToShow(func() { win.Hide() })
			appUI.SetActive(false)
			go ctrl.SetIdle(true)
		}

		createTray(ctrl)
		ctrl.SetCallbacks(func() {
			if windowAwake.Load() {
				win.Invalidate()
			}
			syncTray(ctrl)
		}, appUI.Notify)

		if settings.AutoStartCore {
			go func() {
				time.Sleep(600 * time.Millisecond)
				if err := ctrl.Start(); err != nil {
					log.Println("start core:", err)
				}
			}()
		}
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func showMain() {
	if mainWindow == nil || mainWindow.IsDestroyed() {
		return
	}
	mainWindow.Show()
	mainWindow.Focus()
}

// setAwake tells the UI and the controller whether the window is on screen.
// Minimizing counts as away: Windows paints nothing while it is iconic.
func setAwake(ui *appui.UI, ctrl *app.Controller, awake bool) {
	if windowAwake.Swap(awake) != awake {
		ui.SetActive(awake)
	}
	go ctrl.SetIdle(!awake)
}

func createTray(ctrl *app.Controller) {
	trayCoreItem = &mygo.MenuItem{
		Label: "启动内核",
		Type:  mygo.MenuItemCheckbox,
		Click: func(item *mygo.MenuItem, win *mygo.Window) {
			enabled := item.IsChecked()
			go func() {
				var err error
				if enabled {
					err = ctrl.Start()
				} else {
					err = ctrl.Stop()
				}
				if err != nil {
					log.Println("tray core:", err)
				}
			}()
		},
	}
	trayProxyItem = &mygo.MenuItem{
		Label: "系统代理",
		Type:  mygo.MenuItemCheckbox,
		Click: func(item *mygo.MenuItem, win *mygo.Window) {
			enabled := item.IsChecked()
			go func() {
				if err := ctrl.SetSystemProxy(enabled); err != nil {
					log.Println("tray proxy:", err)
				}
			}()
		},
	}
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Label: "显示主界面", Click: func(*mygo.MenuItem, *mygo.Window) { showMain() }},
		mygo.Separator(),
		trayCoreItem,
		trayProxyItem,
		mygo.Separator(),
		{Label: "退出 ZgoClash", Role: mygo.RoleQuit},
	})
	t, err := mygo.NewTray(mygo.TrayOptions{
		ID:      "main",
		Icon:    trayDisabledPNG,
		ToolTip: "ZgoClash",
		Menu:    menu,
	})
	if err != nil {
		log.Println("tray:", err)
		return
	}
	tray = t
	tray.OnClick(showMain)
	syncTray(ctrl)
}

// syncTray updates the tray from the main thread, and only when something it
// shows actually changed. It runs on every core event, so it reads the two
// flags it shows rather than building a whole snapshot.
func syncTray(ctrl *app.Controller) {
	if mainWindow == nil || mainWindow.IsDestroyed() {
		return
	}
	running, starting, proxy := ctrl.Status()
	status := "未连接"
	switch {
	case starting:
		status = "启动中"
	case running:
		status = "已连接"
	}
	trayShownMu.Lock()
	if running == trayShownRun && proxy == trayShownProx && status == trayShownStat {
		trayShownMu.Unlock()
		return
	}
	trayShownRun, trayShownProx, trayShownStat = running, proxy, status
	trayShownMu.Unlock()

	mainWindow.Update(func() {
		if tray == nil {
			return
		}
		trayMu.Lock()
		defer trayMu.Unlock()
		if running != trayRunning {
			icon := trayDisabledPNG
			if running {
				icon = trayEnabledPNG
			}
			if err := tray.SetIcon(icon, false); err == nil {
				trayRunning = running
			}
		}
		if trayCoreItem != nil {
			trayCoreItem.SetChecked(running)
			if running {
				trayCoreItem.SetLabel("内核运行中")
			} else {
				trayCoreItem.SetLabel("启动内核")
			}
		}
		if trayProxyItem != nil {
			trayProxyItem.SetChecked(proxy)
		}
		tray.SetToolTip("ZgoClash · " + status)
	})
}
