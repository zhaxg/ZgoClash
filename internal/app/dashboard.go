package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ExitInfo is where the proxied traffic leaves, as an IP lookup reports it.
type ExitInfo struct {
	IP           string    `json:"ip"`
	CountryCode  string    `json:"countryCode"`
	Country      string    `json:"country"`
	Organization string    `json:"organization"`
	ASN          string    `json:"asn"`
	CheckedAt    time.Time `json:"checkedAt"`
}

// ServiceStatus is one reachability probe of a well-known site.
type ServiceStatus struct {
	Name  string `json:"name"`
	OK    bool   `json:"ok"`
	Delay int    `json:"delay"`
	Err   string `json:"err,omitempty"`
}

// serviceTargets are the sites the dashboard probes through the core. The
// request goes out the mixed port, so the rules pick the exit like any other
// traffic.
var serviceTargets = []struct {
	name string
	url  string
}{
	{"Google", "https://www.google.com/generate_204"},
	{"GitHub", "https://github.com/"},
	{"OpenAI", "https://openai.com/"},
	{"Twitter", "https://x.com/"},
}

// exitSources are tried in order until one answers; they agree on the address
// but differ in how much detail they carry.
var exitSources = []string{
	"https://ipwho.is/",
	"https://ipapi.co/json/",
	"https://get.geojs.io/v1/ip/geo.json",
}

var errNoCore = errors.New("核心未运行")

// proxyClient builds an HTTP client that reaches the network through the
// core's mixed port, so a lookup reflects the route the rules choose.
func (c *Controller) proxyClient(timeout time.Duration) (*http.Client, error) {
	port := c.Settings().MixedPort
	if port == 0 {
		return nil, errNoCore
	}
	transport := &http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) {
			return url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
		},
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}

// ExitInfo looks up the address the traffic leaves by. The result is cached
// until it is a few minutes old or the core restarts.
func (c *Controller) ExitInfo(ctx context.Context, force bool) (ExitInfo, error) {
	c.exitMu.Lock()
	if !force && !c.exitInfo.CheckedAt.IsZero() && time.Since(c.exitInfo.CheckedAt) < 3*time.Minute {
		cached := c.exitInfo
		c.exitMu.Unlock()
		return cached, nil
	}
	c.exitMu.Unlock()

	client, err := c.proxyClient(12 * time.Second)
	if err != nil {
		return ExitInfo{}, err
	}
	info, err := lookupExit(ctx, client)
	if err != nil {
		return ExitInfo{}, err
	}
	info.CheckedAt = time.Now()

	c.exitMu.Lock()
	c.exitInfo = info
	c.exitMu.Unlock()
	return info, nil
}

func lookupExit(ctx context.Context, client *http.Client) (ExitInfo, error) {
	var last error
	for _, src := range exitSources {
		info, err := fetchExit(ctx, client, src)
		if err == nil && info.IP != "" {
			return info, nil
		}
		if err != nil {
			last = err
		}
	}
	if last == nil {
		last = errors.New("无法获取出口 IP")
	}
	return ExitInfo{}, last
}

func fetchExit(ctx context.Context, client *http.Client, src string) (ExitInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return ExitInfo{}, err
	}
	req.Header.Set("User-Agent", "ZgoClash/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return ExitInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ExitInfo{}, fmt.Errorf("%s: %d", src, resp.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ExitInfo{}, err
	}
	str := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k].(string); ok {
				return v
			}
		}
		return ""
	}
	info := ExitInfo{
		IP:          str("ip", "ip_address"),
		CountryCode: strings.ToUpper(str("country_code", "country_code2")),
		Country:     str("country", "country_name"),
	}
	// Each source nests its connection/ASN detail differently.
	if conn, ok := raw["connection"].(map[string]any); ok {
		if v, ok := conn["org"].(string); ok {
			info.Organization = v
		} else if v, ok := conn["isp"].(string); ok {
			info.Organization = v
		}
		if v, ok := conn["asn"].(string); ok {
			info.ASN = v
		} else if v, ok := conn["asn"].(float64); ok {
			info.ASN = fmt.Sprintf("AS%d", int64(v))
		}
	} else {
		info.Organization = str("org", "isp", "organization")
		info.ASN = str("asn")
	}
	return info, nil
}

// ProbeServices checks a couple of well-known sites through the core and
// reports how long each took.
func (c *Controller) ProbeServices(ctx context.Context) []ServiceStatus {
	client, err := c.proxyClient(8 * time.Second)
	if err != nil {
		out := make([]ServiceStatus, len(serviceTargets))
		for i, t := range serviceTargets {
			out[i] = ServiceStatus{Name: t.name, Err: err.Error()}
		}
		return out
	}
	out := make([]ServiceStatus, len(serviceTargets))
	var wg sync.WaitGroup
	for i, t := range serviceTargets {
		wg.Add(1)
		go func(i int, t struct {
			name string
			url  string
		}) {
			defer wg.Done()
			out[i] = probeService(ctx, client, t.name, t.url)
		}(i, t)
	}
	wg.Wait()
	return out
}

func probeService(ctx context.Context, client *http.Client, name, target string) ServiceStatus {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return ServiceStatus{Name: name, Err: err.Error()}
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return ServiceStatus{Name: name, Err: err.Error()}
	}
	resp.Body.Close()
	delay := int(time.Since(start).Milliseconds())
	// Any answer means the site is reachable through the core, even when it
	// refuses the request: what the card shows is the delay, and only a
	// request that never came back is a failure.
	return ServiceStatus{Name: name, OK: true, Delay: delay}
}

// LocalIP finds this machine's address on the LAN, for the intranet card. It
// is resolved once and cached: the snapshot reads it every frame.
var (
	localIPOnce sync.Once
	localIPVal  string
)

func LocalIP() string {
	localIPOnce.Do(func() {
		localIPVal = findLocalIP()
	})
	return localIPVal
}

func findLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
				continue
			}
			return ipnet.IP.String()
		}
	}
	return ""
}
