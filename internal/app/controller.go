// Package app owns the settings, the profiles and the running mihomo core.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	mihomolog "github.com/metacubex/mihomo/log"

	"zgo-clash/internal/engine"
)

const (
	maxLogs    = 600
	idleRecent = 20
	appVersion = "1.19.33.1"
)

// Controller is the app's model. Its methods are safe for the UI goroutine and
// for background work alike.
type Controller struct {
	store *Store

	lifeMu sync.Mutex
	mu     sync.RWMutex

	initialized bool
	settings    Settings
	profiles    []Profile
	currentID   string

	engine         *engine.Engine
	running        bool
	starting       bool
	startedAt      time.Time
	lastErr        string
	controllerPort int
	groupOrder     []string

	traffic engine.Traffic
	conns   engine.ConnectionsResult

	idle    bool
	idleGen int

	memMu   sync.Mutex
	memAt   time.Time
	memCore uint64
	memApp  uint64

	exitMu   sync.Mutex
	exitInfo ExitInfo

	logs     []engine.LogEntry
	logsMu   sync.Mutex
	initOnce sync.Once

	onChange func()
	onToast  func(level, message string)
}

func NewController(dir string) *Controller {
	return &Controller{store: NewStore(dir)}
}

// SetCallbacks sets the functions run after a state change and for messages
// worth a toast. They may run on any goroutine and must return at once.
func (c *Controller) SetCallbacks(changed func(), toast func(level, message string)) {
	c.mu.Lock()
	c.onChange = changed
	c.onToast = toast
	c.mu.Unlock()
}

func (c *Controller) changed() {
	c.mu.RLock()
	fn := c.onChange
	c.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

func (c *Controller) toast(level, message string) {
	c.mu.RLock()
	fn := c.onToast
	c.mu.RUnlock()
	if fn != nil {
		fn(level, message)
	}
}

func (c *Controller) Init() error {
	var err error
	c.initOnce.Do(func() {
		err = c.init()
	})
	return err
}

func (c *Controller) init() error {
	if err := c.store.ensure(); err != nil {
		return err
	}
	settings := c.store.LoadSettings()
	current, profiles := c.store.LoadProfiles()
	if len(profiles) == 0 {
		p, e := c.createDefaultProfile()
		if e != nil {
			return e
		}
		profiles = []Profile{p}
		current = p.ID
	}
	if _, ok := profileByID(profiles, current); !ok {
		current = profiles[0].ID
	}

	c.mu.Lock()
	c.settings = settings
	c.profiles = profiles
	c.currentID = current
	c.mu.Unlock()

	c.engine = engine.New(c.store.coreHome(), engine.Callbacks{
		OnTraffic: func(t engine.Traffic) {
			c.mu.Lock()
			c.traffic = t
			c.mu.Unlock()
			c.changed()
		},
		OnConnections: func(v engine.ConnectionsResult) {
			c.mu.Lock()
			c.conns = v
			c.mu.Unlock()
			c.changed()
		},
		OnLog: c.appendLog,
	})
	mihomolog.SetLevel(engine.LogLevelOf(settings.LogLevel))
	_ = mygo.App.SetOpenAtLogin(settings.AutoLaunch)
	_ = c.store.SaveSettings(settings)

	c.mu.Lock()
	c.initialized = true
	c.mu.Unlock()
	go c.autoUpdateLoop()
	return nil
}

func (c *Controller) createDefaultProfile() (Profile, error) {
	now := time.Now().UnixMilli()
	p := Profile{
		ID:        "default",
		Name:      "默认配置",
		Type:      "manual",
		UpdatedAt: now,
	}
	content := defaultProfileContent()
	p.ProxyCount, p.RuleCount, p.GroupCount = parseProfileStats(content)
	if err := c.store.WriteProfile(p.ID, content); err != nil {
		return Profile{}, err
	}
	if err := c.store.SaveProfiles(p.ID, []Profile{p}); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func profileByID(items []Profile, id string) (Profile, bool) {
	for _, p := range items {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

// AppVersion is the version the About section shows.
func AppVersion() string { return appVersion }

func (c *Controller) Settings() Settings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}

func (c *Controller) Profiles() []Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Profile(nil), c.profiles...)
}

func (c *Controller) CurrentID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentID
}

func (c *Controller) currentProfile() (Profile, []byte, error) {
	c.mu.RLock()
	id := c.currentID
	items := append([]Profile(nil), c.profiles...)
	c.mu.RUnlock()
	p, ok := profileByID(items, id)
	if !ok {
		return Profile{}, nil, errors.New("没有可用的配置，请先添加一个")
	}
	data, err := c.store.ReadProfile(id)
	if err != nil {
		return Profile{}, nil, fmt.Errorf("读取配置失败: %w", err)
	}
	return p, data, nil
}

// Snapshot copies everything the shell shows.
func (c *Controller) Snapshot() Snapshot {
	c.mu.RLock()
	settings := c.settings
	profiles := append([]Profile(nil), c.profiles...)
	currentID := c.currentID
	running := c.running
	starting := c.starting
	startedAt := c.startedAt
	lastErr := c.lastErr
	port := c.controllerPort
	traffic := c.traffic
	conns := c.conns
	c.mu.RUnlock()

	name := ""
	if p, ok := profileByID(profiles, currentID); ok {
		name = p.Name
	}
	uptime := int64(0)
	if running && !startedAt.IsZero() {
		uptime = int64(time.Since(startedAt).Seconds())
	}
	elevated := isElevated()
	tcp, udp := 0, 0
	for _, cn := range conns.Connections {
		if cn.Metadata.Network == "udp" {
			udp++
		} else {
			tcp++
		}
	}
	memCore, memApp := c.usage()
	recent := 0
	if running && c.engine != nil {
		recent = c.engine.RecentCount()
	}
	return Snapshot{
		Running:            running,
		Starting:           starting,
		AppVersion:         appVersion,
		CoreVersion:        engine.Version,
		StartedAt:          startedAt,
		Uptime:             uptime,
		Settings:           settings,
		Profiles:           profiles,
		CurrentProfileID:   currentID,
		CurrentProfileName: name,
		MixedPort:          settings.MixedPort,
		ControllerPort:     port,
		SystemProxy:        settings.SystemProxy,
		CanTun:             tunSupported() || elevated,
		Elevated:           elevated,
		Home:               c.store.dir,
		LastError:          lastErr,
		Traffic:            traffic,
		ConnectionsCount:   len(conns.Connections),
		TCPCount:           tcp,
		UDPCount:           udp,
		RecentCount:        recent,
		MemoryBytes:        memCore + memApp,
		MemoryCore:         memCore,
		MemoryApp:          memApp,
		LocalIP:            LocalIP(),
	}
}

// usage reports the core's heap and the app's share of the process working
// set. It refreshes at most every two seconds: reading the runtime's
// statistics and the process counters on every frame is costly, and the
// memory card does not need fresher numbers.
func (c *Controller) usage() (core, app uint64) {
	c.memMu.Lock()
	defer c.memMu.Unlock()
	if time.Since(c.memAt) < 2*time.Second {
		return c.memCore, c.memApp
	}
	core = engine.CoreMemory()
	app = processMemory()
	if app > core {
		app -= core
	} else {
		app = 0
	}
	c.memCore, c.memApp, c.memAt = core, app, time.Now()
	return core, app
}

// Status reports the flags the tray shows, without building a full snapshot.
func (c *Controller) Status() (running, starting, proxy bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running, c.starting, c.settings.SystemProxy
}

// SetIdle tells the controller that no window is on screen. While idle it
// stops the polls that feed the dashboard and gives the memory they and the
// pages held back to the system. SetIdle(false) as the window returns resumes
// them and wakes the UI so the pages fill again.
func (c *Controller) SetIdle(idle bool) {
	c.mu.Lock()
	if c.idle == idle {
		c.mu.Unlock()
		return
	}
	c.idle = idle
	c.idleGen++
	gen := c.idleGen
	c.mu.Unlock()

	if c.engine != nil {
		c.engine.SetPolling(!idle)
		if idle {
			c.engine.TrimRecent(idleRecent)
		}
	}

	if !idle {
		setIdleMemory(false)
		c.changed()
		return
	}

	c.mu.Lock()
	c.traffic = engine.Traffic{}
	c.conns = engine.ConnectionsResult{}
	c.mu.Unlock()
	setIdleMemory(true)

	// Traffic still allocates while the core forwards requests, so hand
	// the pages it grows back once more after the window has settled.
	go func() {
		time.Sleep(20 * time.Second)
		c.mu.RLock()
		still := c.idle && c.idleGen == gen
		c.mu.RUnlock()
		if still {
			setIdleMemory(true)
		}
	}()
}

// Recent returns connections that have closed, newest first, for the dashboard.
func (c *Controller) Recent() []engine.Connection {
	c.mu.RLock()
	e := c.engine
	running := c.running
	c.mu.RUnlock()
	if !running || e == nil {
		return nil
	}
	return e.Recent()
}

// Connections returns the last snapshot the core's poller took.
func (c *Controller) Connections() engine.ConnectionsResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := c.conns
	out.Connections = append([]engine.Connection(nil), c.conns.Connections...)
	return out
}

func (c *Controller) appendLog(line engine.LogEntry) {
	c.logsMu.Lock()
	c.logs = append(c.logs, line)
	if len(c.logs) > maxLogs {
		// Reuse the buffer instead of copying the tail into a new slice on
		// every line once it is full.
		c.logs = append(c.logs[:0], c.logs[len(c.logs)-maxLogs:]...)
	}
	c.logsMu.Unlock()
	c.changed()
}

func (c *Controller) Logs(limit int) []engine.LogEntry {
	c.logsMu.Lock()
	defer c.logsMu.Unlock()
	if limit <= 0 || limit > len(c.logs) {
		limit = len(c.logs)
	}
	out := make([]engine.LogEntry, limit)
	copy(out, c.logs[len(c.logs)-limit:])
	return out
}

func (c *Controller) ClearLogs() {
	c.logsMu.Lock()
	c.logs = nil
	c.logsMu.Unlock()
	c.changed()
}

func (c *Controller) Start() error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	return c.startLocked()
}

func (c *Controller) startLocked() error {
	c.mu.Lock()
	if c.running || c.starting {
		c.mu.Unlock()
		return nil
	}
	c.starting = true
	c.lastErr = ""
	settings := c.settings
	c.mu.Unlock()
	c.changed()

	defer func() {
		c.mu.Lock()
		c.starting = false
		c.mu.Unlock()
		c.changed()
	}()

	_, data, err := c.currentProfile()
	if err != nil {
		return c.failStart(err)
	}
	if settings.Tun && !isElevated() {
		return c.failStart(errors.New("TUN 模式需要管理员权限，请以管理员身份重启"))
	}
	port, err := freePort(settings.ControllerPort)
	if err != nil {
		return c.failStart(fmt.Errorf("分配控制端口失败: %w", err))
	}
	secret := newSecret()
	cfg, groups, err := buildRuntimeConfig(data, settings, port, secret)
	if err != nil {
		return c.failStart(err)
	}
	if _, err := c.store.WriteRuntime(cfg); err != nil {
		return c.failStart(fmt.Errorf("写入运行时配置失败: %w", err))
	}
	if err := c.engine.Start(cfg, port, settings.ControllerExternal, secret, settings.LogLevel); err != nil {
		return c.failStart(err)
	}

	c.mu.Lock()
	c.running = true
	c.startedAt = time.Now()
	c.controllerPort = port
	c.groupOrder = groups
	c.traffic = engine.Traffic{}
	c.conns = engine.ConnectionsResult{}
	idle := c.idle
	c.mu.Unlock()
	if idle {
		// Started from the tray while the window is away: keep the polls
		// off until it shows again.
		c.engine.SetPolling(false)
	}

	if settings.SystemProxy {
		if err := applySystemProxy(true, fmt.Sprintf("127.0.0.1:%d", settings.MixedPort)); err != nil {
			c.setProxySetting(false)
			c.mu.Lock()
			c.lastErr = err.Error()
			c.mu.Unlock()
			c.toast("error", "系统代理设置失败: "+err.Error())
		}
	}
	c.changed()
	return nil
}

func (c *Controller) failStart(err error) error {
	c.mu.Lock()
	c.lastErr = err.Error()
	c.running = false
	c.mu.Unlock()
	c.changed()
	return err
}

func (c *Controller) Stop() error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	c.stopLocked(true)
	return nil
}

func (c *Controller) stopLocked(clearProxy bool) {
	c.mu.Lock()
	wasRunning := c.running
	c.mu.Unlock()
	if wasRunning && c.engine != nil {
		c.engine.Stop()
	}
	c.mu.Lock()
	c.running = false
	c.starting = false
	c.startedAt = time.Time{}
	c.controllerPort = 0
	c.conns = engine.ConnectionsResult{}
	c.traffic = engine.Traffic{}
	c.mu.Unlock()
	if clearProxy {
		// The proxy in Windows goes away with the core, but the choice the
		// user made stays: it comes back with the core, and it is not
		// forgotten by a restart.
		_ = applySystemProxy(false, "")
	}
	c.changed()
}

func (c *Controller) Restart() error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	return c.restartLocked()
}

func (c *Controller) restartLocked() error {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()
	if running {
		c.stopLocked(false)
	}
	return c.startLocked()
}

func (c *Controller) setProxySetting(enabled bool) {
	c.mu.Lock()
	c.settings.SystemProxy = enabled
	settings := c.settings
	c.mu.Unlock()
	_ = c.store.SaveSettings(settings)
}

// applySystemProxy writes the proxy of Windows. The test replaces it, to
// keep the registry of the machine out of the way.
var applySystemProxy = setSystemProxy

// SetSystemProxy applies or clears the Windows Internet proxy.
func (c *Controller) SetSystemProxy(enabled bool) error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	if enabled {
		c.mu.RLock()
		running := c.running
		starting := c.starting
		port := c.settings.MixedPort
		c.mu.RUnlock()
		if starting {
			return errors.New("核心正在启动，请稍候")
		}
		if !running {
			if err := c.startLocked(); err != nil {
				return err
			}
		}
		if err := applySystemProxy(true, fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
			c.setProxySetting(false)
			c.changed()
			return err
		}
		c.setProxySetting(true)
	} else {
		if err := applySystemProxy(false, ""); err != nil {
			return err
		}
		c.setProxySetting(false)
	}
	c.changed()
	return nil
}

func (c *Controller) SetTun(enabled bool) error {
	if enabled && !isElevated() {
		return errors.New("TUN 模式需要管理员权限，请以管理员身份重启")
	}
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	c.mu.Lock()
	prev := c.settings.Tun
	c.settings.Tun = enabled
	settings := c.settings
	running := c.running
	c.mu.Unlock()
	if err := c.store.SaveSettings(settings); err != nil {
		return err
	}
	var restartErr error
	if running && prev != enabled {
		restartErr = c.restartLocked()
	}
	c.changed()
	return restartErr
}

func (c *Controller) SetMode(mode string) error {
	if mode != "rule" && mode != "global" && mode != "direct" {
		return errors.New("未知的模式")
	}
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	c.mu.Lock()
	c.settings.Mode = mode
	settings := c.settings
	running := c.running
	c.mu.Unlock()
	_ = c.store.SaveSettings(settings)
	if running {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_ = c.engine.SetMode(ctx, mode)
		cancel()
	}
	c.changed()
	return nil
}

// UpdateSettings applies the whole settings form, restarting the core for the
// values it only reads at startup.
func (c *Controller) UpdateSettings(next Settings) error {
	next = normalizeSettings(next)
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()

	c.mu.RLock()
	prev := c.settings
	running := c.running
	c.mu.RUnlock()

	if next.Tun && !isElevated() {
		next.Tun = prev.Tun
	}
	if next.AutoLaunch != prev.AutoLaunch {
		_ = mygo.App.SetOpenAtLogin(next.AutoLaunch)
	}

	c.mu.Lock()
	c.settings = next
	c.mu.Unlock()
	if err := c.store.SaveSettings(next); err != nil {
		return err
	}

	if running && next.Mode != prev.Mode {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_ = c.engine.SetMode(ctx, next.Mode)
		cancel()
	}
	if running && next.LogLevel != prev.LogLevel {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_ = c.engine.SetLogLevel(ctx, next.LogLevel)
		cancel()
		mihomolog.SetLevel(engine.LogLevelOf(next.LogLevel))
	}

	restartNeeded := next.MixedPort != prev.MixedPort ||
		next.AllowLan != prev.AllowLan ||
		next.IPv6 != prev.IPv6 ||
		next.Tun != prev.Tun ||
		next.TunStack != prev.TunStack ||
		next.ControllerPort != prev.ControllerPort ||
		next.ControllerExternal != prev.ControllerExternal

	var applyErr error
	if restartNeeded && running {
		applyErr = c.restartLocked()
	} else if next.SystemProxy != prev.SystemProxy {
		if next.SystemProxy {
			if running {
				if err := applySystemProxy(true, fmt.Sprintf("127.0.0.1:%d", next.MixedPort)); err != nil {
					applyErr = err
					c.setProxySetting(false)
				}
			} else if err := c.startLocked(); err != nil {
				applyErr = err
			}
		} else {
			if err := applySystemProxy(false, ""); err != nil {
				applyErr = err
			}
		}
	}
	c.changed()
	return applyErr
}

func (c *Controller) RestartElevated() error {
	if isElevated() {
		return nil
	}
	if err := relaunchElevated(); err != nil {
		return err
	}
	go func() {
		time.Sleep(400 * time.Millisecond)
		mygo.App.Quit()
	}()
	return nil
}

func (c *Controller) OpenDataDir() error {
	return mygo.Shell.OpenPath(c.store.dir)
}

func (c *Controller) RuntimeConfig() (string, error) {
	data, err := os.ReadFile(c.store.runtimePath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProxyGroups reads the core's proxies and orders the groups the profile
// chose first, then the rest by name.
func (c *Controller) ProxyGroups(ctx context.Context) (map[string]engine.Proxy, []engine.ProxyGroupView, error) {
	if !c.engine.Running() {
		return nil, nil, errors.New("核心未运行")
	}
	proxies, err := c.engine.Proxies(ctx)
	if err != nil {
		return nil, nil, err
	}
	c.mu.RLock()
	order := append([]string(nil), c.groupOrder...)
	c.mu.RUnlock()

	seen := map[string]bool{}
	groups := make([]engine.ProxyGroupView, 0, 8)
	add := func(name string) {
		if seen[name] {
			return
		}
		p, ok := proxies[name]
		if !ok || len(p.All) == 0 || p.Hidden || name == "GLOBAL" {
			return
		}
		seen[name] = true
		groups = append(groups, engine.ProxyGroupView{
			Name:   p.Name,
			Type:   p.Type,
			Now:    p.Now,
			All:    append([]string(nil), p.All...),
			Icon:   p.Icon,
			Hidden: p.Hidden,
		})
	}
	for _, name := range order {
		add(name)
	}
	extras := make([]string, 0, len(proxies))
	for name, p := range proxies {
		if len(p.All) > 0 && !seen[name] && !p.Hidden && name != "GLOBAL" {
			extras = append(extras, name)
		}
	}
	sort.Strings(extras)
	for _, name := range extras {
		add(name)
	}
	return proxies, groups, nil
}

func (c *Controller) SelectProxy(group, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return c.engine.Select(ctx, group, name)
}

func (c *Controller) TestDelay(name string, timeoutMs int) (int, error) {
	settings := c.Settings()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return c.engine.Delay(ctx, name, timeoutMs, settings.TestURL)
}

func (c *Controller) TestGroupDelay(group string, timeoutMs int) (map[string]int, error) {
	settings := c.Settings()
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	return c.engine.GroupDelay(ctx, group, timeoutMs, settings.TestURL)
}

func (c *Controller) CloseConnection(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.engine.CloseConnection(ctx, id)
}

func (c *Controller) CloseAllConnections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.engine.CloseAllConnections(ctx)
}

func (c *Controller) autoUpdateLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	time.Sleep(20 * time.Second)
	c.autoUpdateOnce()
	for range ticker.C {
		c.autoUpdateOnce()
	}
}

func (c *Controller) autoUpdateOnce() {
	settings := c.Settings()
	items := c.Profiles()
	if settings.AutoUpdateInterval <= 0 {
		return
	}
	due := time.Now().Add(-time.Duration(settings.AutoUpdateInterval) * time.Hour).UnixMilli()
	var names []string
	for _, p := range items {
		if p.Type != "url" || p.URL == "" || p.UpdatedAt > due {
			continue
		}
		if _, err := c.UpdateProfile(p.ID); err == nil {
			names = append(names, p.Name)
		}
	}
	if len(names) > 0 {
		c.toast("success", "已自动更新订阅: "+strings.Join(names, "、"))
	}
}

func freePort(preferred int) (int, error) {
	if preferred > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", preferred))
		if err == nil {
			port := ln.Addr().(*net.TCPAddr).Port
			_ = ln.Close()
			return port, nil
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}
