// Package engine runs mihomo in this process and talks to its controller over
// loopback.
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub"
	"github.com/metacubex/mihomo/hub/executor"
	"github.com/metacubex/mihomo/hub/route"
	mihomolog "github.com/metacubex/mihomo/log"
)

const Version = "1.19.33"

// Callbacks receive the core's live data.
type Callbacks struct {
	OnTraffic     func(Traffic)
	OnConnections func(ConnectionsResult)
	OnLog         func(LogEntry)
}

// Engine runs mihomo and exposes the parts of its controller the app uses.
type Engine struct {
	home string
	cb   Callbacks

	lifeMu     sync.Mutex
	mu         sync.RWMutex
	running    bool
	base       string
	secret     string
	cancel     context.CancelFunc
	runCtx     context.Context
	pollCtx    context.Context
	pollCancel context.CancelFunc
	polling    bool
	client     *http.Client
	streamer   *http.Client

	// Connections that have since closed, newest first. The poller moves an
	// active connection here when it stops appearing in the snapshot, so the
	// dashboard can show recent requests after they end.
	histMu     sync.Mutex
	histActive map[string]Connection
	histRecent []Connection
}

// New builds an engine. Logs it receives before Start are still delivered.
func New(home string, cb Callbacks) *Engine {
	e := &Engine{
		home:       home,
		cb:         cb,
		client:     loopbackClient(20 * time.Second),
		streamer:   loopbackClient(0),
		histActive: map[string]Connection{},
	}
	C.Version = Version
	C.BuildTime = "unknown time"
	go e.consumeLogs()
	return e
}

func loopbackClient(timeout time.Duration) *http.Client {
	// The controller is on 127.0.0.1: never send it through a proxy from the
	// environment, which the user's own proxy settings may have set.
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		MaxIdleConns:          8,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

func (e *Engine) consumeLogs() {
	sub := mihomolog.Subscribe()
	defer mihomolog.UnSubscribe(sub)
	for ev := range sub {
		if e.cb.OnLog == nil {
			continue
		}
		e.cb.OnLog(LogEntry{
			Time:    time.Now().UnixMilli(),
			Level:   ev.Type(),
			Payload: ev.Payload,
		})
	}
}

func (e *Engine) Running() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Start parses cfg and opens the listeners it asks for, then waits for the
// controller to answer. external lets the controller answer other devices
// too, not only this one.
func (e *Engine) Start(cfg []byte, controllerPort int, external bool, secret, level string) error {
	e.lifeMu.Lock()
	defer e.lifeMu.Unlock()
	e.stopLocked()

	e.histMu.Lock()
	e.histActive = map[string]Connection{}
	e.histRecent = nil
	e.histMu.Unlock()

	// SetHomeDir first: config.Init creates the starter config.yaml through
	// C.Path.Config(), which resolves against the home directory. SetConfig
	// keeps that file inside it instead of the process's working directory.
	C.SetHomeDir(e.home)
	C.SetConfig(filepath.Join(e.home, "config.yaml"))
	if err := config.Init(e.home); err != nil {
		return err
	}
	mihomolog.SetLevel(LogLevelOf(level))

	// The controller listens on the loopback address unless the user let
	// other devices reach it; the app itself always talks to it there.
	listen := fmt.Sprintf("127.0.0.1:%d", controllerPort)
	if external {
		listen = fmt.Sprintf("0.0.0.0:%d", controllerPort)
	}
	if err := parseSafely(cfg, listen, secret); err != nil {
		shutdownSafely()
		return err
	}

	base := fmt.Sprintf("http://127.0.0.1:%d", controllerPort)
	if err := waitForCore(e.client, base, secret, 10*time.Second); err != nil {
		shutdownSafely()
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pollCtx, pollCancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.running = true
	e.base = base
	e.secret = secret
	e.cancel = cancel
	e.runCtx = ctx
	e.pollCtx = pollCtx
	e.pollCancel = pollCancel
	e.polling = true
	e.mu.Unlock()

	go e.trafficLoop(pollCtx)
	go e.connectionsLoop(pollCtx)
	return nil
}

func (e *Engine) Stop() {
	e.lifeMu.Lock()
	defer e.lifeMu.Unlock()
	e.stopLocked()
}

func (e *Engine) stopLocked() {
	e.mu.Lock()
	running := e.running
	cancel := e.cancel
	base := e.base
	e.running = false
	e.cancel = nil
	e.runCtx = nil
	e.pollCtx = nil
	e.pollCancel = nil
	e.polling = false
	e.base = ""
	e.secret = ""
	e.mu.Unlock()
	if !running {
		return
	}
	if cancel != nil {
		cancel()
	}
	shutdownSafely()
	closeControlPort(base)
}

// SetPolling starts or stops the traffic and connection polls the UI reads.
// The core keeps serving while they are stopped: with no window to show, the
// two requests a second only cost memory.
func (e *Engine) SetPolling(on bool) {
	e.mu.Lock()
	if !e.running || e.polling == on {
		e.mu.Unlock()
		return
	}
	if !on {
		if e.pollCancel != nil {
			e.pollCancel()
		}
		e.pollCtx, e.pollCancel = nil, nil
		e.polling = false
		e.mu.Unlock()
		// The live set is stale after the pause, so the first snapshot
		// taken once polling resumes starts it anew instead of reporting
		// every connection as closed.
		e.histMu.Lock()
		e.histActive = map[string]Connection{}
		e.histMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(e.runCtx)
	e.pollCtx, e.pollCancel = ctx, cancel
	e.polling = true
	e.mu.Unlock()
	go e.trafficLoop(ctx)
	go e.connectionsLoop(ctx)
}

func parseSafely(cfg []byte, addr, secret string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("核心启动失败: %v", r)
		}
	}()
	return hub.Parse(cfg, hub.WithExternalController(addr), hub.WithSecret(secret))
}

func shutdownSafely() {
	defer func() { _ = recover() }()
	executor.Shutdown()
}

// closeControlPort closes the core's control port. The core's own shutdown
// closes the proxy listeners but leaves the port answering, so it would stay
// bound after the core is stopped. The route package closes it in a goroutine,
// so the wait is for the close to be done: the core may start again on the
// same port right after, and the late close would take that one down.
func closeControlPort(base string) {
	if base == "" {
		return
	}
	addr := strings.TrimPrefix(base, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	route.ReCreateServer(&route.Config{})
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return // the port is closed
		}
		_ = conn.Close()
		time.Sleep(50 * time.Millisecond)
	}
}

func waitForCore(client *http.Client, base, secret string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodGet, base+"/version", nil)
		req.Header.Set("Authorization", "Bearer "+secret)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("控制器返回 %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(120 * time.Millisecond)
	}
	if last == nil {
		last = errors.New("启动超时")
	}
	return fmt.Errorf("核心未就绪: %w", last)
}

// LogLevelOf maps a settings value to mihomo's log level.
func LogLevelOf(level string) mihomolog.LogLevel {
	switch level {
	case "silent":
		return mihomolog.SILENT
	case "error":
		return mihomolog.ERROR
	case "warning":
		return mihomolog.WARNING
	case "debug":
		return mihomolog.DEBUG
	default:
		return mihomolog.INFO
	}
}

func (e *Engine) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	e.mu.RLock()
	base, secret, running := e.base, e.secret, e.running
	e.mu.RUnlock()
	if !running {
		return nil, errors.New("核心未运行")
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(data))
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, coreError(resp.StatusCode, data)
	}
	return data, nil
}

func coreError(status int, data []byte) error {
	var v struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &v) == nil && v.Message != "" {
		return errors.New(v.Message)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		text = http.StatusText(status)
	}
	return errors.New(text)
}

func (e *Engine) get(ctx context.Context, path string, out any) error {
	data, err := e.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func proxyPath(name string) string {
	return "/proxies/" + url.PathEscape(name)
}

func (e *Engine) trafficLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := e.readTraffic(ctx); err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}
}

func (e *Engine) readTraffic(ctx context.Context) error {
	e.mu.RLock()
	base, secret, running := e.base, e.secret, e.running
	e.mu.RUnlock()
	if !running {
		return errors.New("stopped")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/traffic", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := e.streamer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("traffic: %d", resp.StatusCode)
	}
	dec := json.NewDecoder(resp.Body)
	var upTotal, downTotal int64
	for {
		var t Traffic
		if err := dec.Decode(&t); err != nil {
			return err
		}
		if t.UpTotal == 0 && t.DownTotal == 0 {
			upTotal += t.Up
			downTotal += t.Down
			t.UpTotal, t.DownTotal = upTotal, downTotal
		} else {
			upTotal, downTotal = t.UpTotal, t.DownTotal
		}
		if e.cb.OnTraffic != nil {
			e.cb.OnTraffic(t)
		}
	}
}

func (e *Engine) connectionsLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var snapshot ConnectionsResult
			if err := e.get(ctx, "/connections", &snapshot); err != nil {
				continue
			}
			e.trackHistory(snapshot.Connections)
			if e.cb.OnConnections != nil {
				e.cb.OnConnections(snapshot)
			}
		}
	}
}

// trackHistory folds the live snapshot into the active map, moving anything
// that disappeared into the recent list.
func (e *Engine) trackHistory(live []Connection) {
	e.histMu.Lock()
	defer e.histMu.Unlock()
	seen := make(map[string]bool, len(live))
	for _, cn := range live {
		seen[cn.ID] = true
	}
	for id, cn := range e.histActive {
		if !seen[id] {
			// Prepend without rebuilding the slice each time: on the
			// way to maxRecent it grows, after that it slides.
			if len(e.histRecent) < maxRecent {
				e.histRecent = append(e.histRecent, Connection{})
			}
			if len(e.histRecent) > 1 {
				copy(e.histRecent[1:], e.histRecent[:len(e.histRecent)-1])
			}
			if len(e.histRecent) > 0 {
				e.histRecent[0] = cn
			}
			delete(e.histActive, id)
		}
	}
	for _, cn := range live {
		e.histActive[cn.ID] = cn
	}
}

// Recent returns connections that closed, newest first.
func (e *Engine) Recent() []Connection {
	e.histMu.Lock()
	defer e.histMu.Unlock()
	out := make([]Connection, len(e.histRecent))
	copy(out, e.histRecent)
	return out
}

// RecentCount reports how many closed connections are kept, without copying
// them: the dashboard only needs the number on every frame.
func (e *Engine) RecentCount() int {
	e.histMu.Lock()
	defer e.histMu.Unlock()
	return len(e.histRecent)
}

// TrimRecent keeps at most n closed connections, releasing the rest.
func (e *Engine) TrimRecent(n int) {
	if n < 0 {
		n = 0
	}
	e.histMu.Lock()
	if len(e.histRecent) > n {
		kept := make([]Connection, n)
		copy(kept, e.histRecent[:n])
		e.histRecent = kept
	}
	e.histMu.Unlock()
}

const maxRecent = 300

// SetMode patches the running core without re-reading the config.
func (e *Engine) SetMode(ctx context.Context, mode string) error {
	_, err := e.do(ctx, http.MethodPatch, "/configs", map[string]any{"mode": mode})
	return err
}

// SetLogLevel patches the running core's log level.
func (e *Engine) SetLogLevel(ctx context.Context, level string) error {
	_, err := e.do(ctx, http.MethodPatch, "/configs", map[string]any{"log-level": level})
	if err == nil {
		mihomolog.SetLevel(LogLevelOf(level))
	}
	return err
}

// Proxies reads every proxy and group the core knows.
func (e *Engine) Proxies(ctx context.Context) (map[string]Proxy, error) {
	var raw struct {
		Proxies map[string]Proxy `json:"proxies"`
	}
	if err := e.get(ctx, "/proxies", &raw); err != nil {
		return nil, err
	}
	return raw.Proxies, nil
}

// Select chooses name in the group.
func (e *Engine) Select(ctx context.Context, group, name string) error {
	_, err := e.do(ctx, http.MethodPut, proxyPath(group), map[string]any{"name": name})
	return err
}

// Delay tests one proxy against testURL.
func (e *Engine) Delay(ctx context.Context, name string, timeout int, testURL string) (int, error) {
	path := fmt.Sprintf("%s/delay?timeout=%d&url=%s", proxyPath(name), timeout, url.QueryEscape(testURL))
	data, err := e.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}
	var v struct {
		Delay int `json:"delay"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return 0, err
	}
	return v.Delay, nil
}

// GroupDelay tests every proxy of a group.
func (e *Engine) GroupDelay(ctx context.Context, group string, timeout int, testURL string) (map[string]int, error) {
	path := fmt.Sprintf("/group/%s/delay?timeout=%d&url=%s", url.PathEscape(group), timeout, url.QueryEscape(testURL))
	data, err := e.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var v map[string]int
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// CloseConnection closes one live connection.
func (e *Engine) CloseConnection(ctx context.Context, id string) error {
	_, err := e.do(ctx, http.MethodDelete, "/connections/"+url.PathEscape(id), nil)
	return err
}

// CloseAllConnections closes every live connection.
func (e *Engine) CloseAllConnections(ctx context.Context) error {
	_, err := e.do(ctx, http.MethodDelete, "/connections", nil)
	return err
}

// Memory reports the Go heap and runtime memory of the process.
func Memory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Sys
}

// CoreMemory is the heap the core is holding in this in-process runtime. The
// dashboard uses it as the "内核" share of the process's memory.
func CoreMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}
