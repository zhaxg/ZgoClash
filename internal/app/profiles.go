package app

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"

	"zgo-clash/internal/engine"
)

const maxProfileBytes = 32 << 20

func (c *Controller) AddURL(ctx context.Context, rawURL string, name string) (Profile, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return Profile{}, errors.New("请输入有效的 http(s) 订阅链接")
	}
	data, info, err := fetchSubscription(ctx, rawURL)
	if err != nil {
		return Profile{}, err
	}
	if strings.TrimSpace(name) == "" {
		name = parsed.Host
	}
	return c.addProfile(Profile{
		Name:       name,
		Type:       "url",
		URL:        rawURL,
		Upload:     info.upload,
		Download:   info.download,
		Total:      info.total,
		Expire:     info.expire,
		UpdatedAt:  time.Now().UnixMilli(),
		ProxyCount: info.proxies,
		RuleCount:  info.rules,
		GroupCount: info.groups,
	}, data)
}

// AddFile asks for a local YAML file and adds it.
func (c *Controller) AddFile(parent *mygo.Window) (Profile, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: parent,
		Title:  "选择配置文件",
		Filters: []mygo.FileFilter{
			{Name: "配置文件", Extensions: []string{"yaml", "yml", "conf", "txt"}},
			{Name: "所有文件", Extensions: []string{"*"}},
		},
	})
	if err != nil {
		return Profile{}, err
	}
	if len(paths) == 0 {
		return Profile{}, errors.New("已取消")
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		return Profile{}, err
	}
	if !looksLikeClashConfig(data) {
		return Profile{}, errors.New("不是有效的 Clash/mihomo 配置")
	}
	name := strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))
	return c.addProfile(Profile{
		Name:      name,
		Type:      "file",
		UpdatedAt: time.Now().UnixMilli(),
	}, data)
}

func (c *Controller) CreateManual(name, content string) (Profile, error) {
	if strings.TrimSpace(name) == "" {
		name = "手动配置"
	}
	if !looksLikeClashConfig([]byte(content)) {
		return Profile{}, errors.New("配置内容需要包含 proxies、proxy-groups 或 rules")
	}
	return c.addProfile(Profile{
		Name:      name,
		Type:      "manual",
		UpdatedAt: time.Now().UnixMilli(),
	}, []byte(content))
}

// ImportClipboard adds whatever the clipboard holds: a subscription URL or a
// Clash configuration.
func (c *Controller) ImportClipboard() (Profile, error) {
	text := strings.TrimSpace(mygo.Clipboard.ReadText())
	if text == "" {
		return Profile{}, errors.New("剪贴板为空")
	}
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		return c.AddURL(ctx, text, "")
	}
	if looksLikeClashConfig([]byte(text)) {
		return c.CreateManual("剪贴板配置", text)
	}
	return Profile{}, errors.New("剪贴板内容既不是订阅链接，也不是 Clash 配置")
}

func (c *Controller) addProfile(p Profile, data []byte) (Profile, error) {
	if p.ID == "" {
		p.ID = newID()
	}
	p.ProxyCount, p.RuleCount, p.GroupCount = parseProfileStats(data)
	if err := c.store.WriteProfile(p.ID, data); err != nil {
		return Profile{}, err
	}
	c.mu.Lock()
	c.profiles = append(c.profiles, p)
	items := append([]Profile(nil), c.profiles...)
	current := c.currentID
	c.mu.Unlock()
	if err := c.store.SaveProfiles(current, items); err != nil {
		return Profile{}, err
	}
	c.changed()
	return p, nil
}

func (c *Controller) UpdateProfile(id string) (Profile, error) {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	return c.updateProfileLocked(id)
}

func (c *Controller) updateProfileLocked(id string) (Profile, error) {
	c.mu.RLock()
	items := append([]Profile(nil), c.profiles...)
	current := c.currentID
	running := c.running
	c.mu.RUnlock()
	p, ok := profileByID(items, id)
	if !ok {
		return Profile{}, errors.New("配置不存在")
	}
	if p.Type != "url" || p.URL == "" {
		return Profile{}, errors.New("只有订阅配置可以更新")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	data, info, err := fetchSubscription(ctx, p.URL)
	if err != nil {
		return Profile{}, err
	}
	p.UpdatedAt = time.Now().UnixMilli()
	p.Upload = info.upload
	p.Download = info.download
	p.Total = info.total
	p.Expire = info.expire
	p.ProxyCount, p.RuleCount, p.GroupCount = parseProfileStats(data)
	if err := c.store.WriteProfile(p.ID, data); err != nil {
		return Profile{}, err
	}
	c.mu.Lock()
	for i := range c.profiles {
		if c.profiles[i].ID == p.ID {
			c.profiles[i] = p
		}
	}
	c.mu.Unlock()
	if err := c.store.SaveProfiles(current, c.Profiles()); err != nil {
		return Profile{}, err
	}
	if id == current && running {
		_ = c.restartLocked()
	}
	c.changed()
	return p, nil
}

func (c *Controller) UpdateAll() (UpdateReport, error) {
	items := c.Profiles()
	report := UpdateReport{}
	for _, p := range items {
		if p.Type != "url" {
			continue
		}
		if _, err := c.UpdateProfile(p.ID); err != nil {
			report.Failed = append(report.Failed, p.Name+": "+err.Error())
		} else {
			report.Updated = append(report.Updated, p.Name)
		}
	}
	c.changed()
	return report, nil
}

func (c *Controller) DeleteProfile(id string) error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	c.mu.Lock()
	items := append([]Profile(nil), c.profiles...)
	current := c.currentID
	running := c.running
	c.mu.Unlock()
	p, ok := profileByID(items, id)
	if !ok {
		return errors.New("配置不存在")
	}
	_ = c.store.DeleteProfile(id)
	remaining := make([]Profile, 0, len(items))
	for _, item := range items {
		if item.ID != id {
			remaining = append(remaining, item)
		}
	}
	newCurrent := current
	if current == id {
		newCurrent = ""
		if len(remaining) > 0 {
			newCurrent = remaining[0].ID
		}
	}
	c.mu.Lock()
	c.profiles = remaining
	c.currentID = newCurrent
	c.mu.Unlock()
	if err := c.store.SaveProfiles(newCurrent, remaining); err != nil {
		return err
	}
	if current == id {
		if running {
			if len(remaining) > 0 {
				_ = c.restartLocked()
			} else {
				c.stopLocked(true)
			}
		}
	}
	c.toast("success", "已删除 "+p.Name)
	c.changed()
	return nil
}

func (c *Controller) SetCurrentProfile(id string) error {
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	c.mu.Lock()
	items := append([]Profile(nil), c.profiles...)
	running := c.running
	c.mu.Unlock()
	if _, ok := profileByID(items, id); !ok {
		return errors.New("配置不存在")
	}
	c.mu.Lock()
	c.currentID = id
	c.mu.Unlock()
	if err := c.store.SaveProfiles(id, items); err != nil {
		return err
	}
	var err error
	if running {
		err = c.restartLocked()
	}
	c.changed()
	return err
}

// EditProfile renames a profile and, for a subscription, changes the link it
// fetches.
func (c *Controller) EditProfile(id, name, url string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("名称不能为空")
	}
	url = strings.TrimSpace(url)
	c.mu.Lock()
	found := false
	isURL := false
	for i := range c.profiles {
		if c.profiles[i].ID == id {
			found = true
			isURL = c.profiles[i].Type == "url"
		}
	}
	if !found {
		c.mu.Unlock()
		return errors.New("配置不存在")
	}
	if isURL && url == "" {
		c.mu.Unlock()
		return errors.New("订阅链接不能为空")
	}
	for i := range c.profiles {
		if c.profiles[i].ID == id {
			c.profiles[i].Name = name
			if isURL {
				c.profiles[i].URL = url
			}
		}
	}
	items := append([]Profile(nil), c.profiles...)
	current := c.currentID
	c.mu.Unlock()
	if err := c.store.SaveProfiles(current, items); err != nil {
		return err
	}
	c.changed()
	return nil
}

func (c *Controller) ProfileContent(id string) (string, error) {
	data, err := c.store.ReadProfile(id)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Controller) SaveProfileContent(id, content string) error {
	if !looksLikeClashConfig([]byte(content)) {
		return errors.New("配置无效：需要包含 proxies、proxy-groups 或 rules")
	}
	c.lifeMu.Lock()
	defer c.lifeMu.Unlock()
	if err := c.store.WriteProfile(id, []byte(content)); err != nil {
		return err
	}
	proxies, rules, groups := parseProfileStats([]byte(content))
	c.mu.Lock()
	for i := range c.profiles {
		if c.profiles[i].ID == id {
			c.profiles[i].ProxyCount = proxies
			c.profiles[i].RuleCount = rules
			c.profiles[i].GroupCount = groups
			c.profiles[i].UpdatedAt = time.Now().UnixMilli()
		}
	}
	current := c.currentID
	running := c.running
	c.mu.Unlock()
	if err := c.store.SaveProfiles(current, c.Profiles()); err != nil {
		return err
	}
	if id == current && running {
		_ = c.restartLocked()
	}
	c.changed()
	return nil
}

// ExportProfile writes a profile where the user chooses.
func (c *Controller) ExportProfile(parent *mygo.Window, id string) (string, error) {
	data, err := c.store.ReadProfile(id)
	if err != nil {
		return "", err
	}
	c.mu.RLock()
	p, _ := profileByID(c.profiles, id)
	c.mu.RUnlock()
	name := p.Name
	if name == "" {
		name = "config"
	}
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      parent,
		Title:       "导出配置",
		DefaultPath: name + ".yaml",
		Filters:     []mygo.FileFilter{{Name: "YAML", Extensions: []string{"yaml", "yml"}}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

type subscriptionInfo struct {
	upload   int64
	download int64
	total    int64
	expire   int64
	proxies  int
	rules    int
	groups   int
}

func fetchSubscription(ctx context.Context, rawURL string) ([]byte, subscriptionInfo, error) {
	client := &http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, subscriptionInfo{}, err
	}
	req.Header.Set("User-Agent", "clash.meta/"+engine.Version+" zgo-clash")
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	if err != nil {
		return nil, subscriptionInfo{}, fmt.Errorf("下载订阅失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, subscriptionInfo{}, fmt.Errorf("下载订阅失败: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxProfileBytes+1))
	if err != nil {
		return nil, subscriptionInfo{}, err
	}
	if len(body) > maxProfileBytes {
		return nil, subscriptionInfo{}, errors.New("订阅内容过大")
	}
	config, err := decodeSubscription(body)
	if err != nil {
		return nil, subscriptionInfo{}, err
	}
	info := subscriptionInfo{}
	if raw := resp.Header.Get("subscription-userinfo"); raw != "" {
		info = parseUserInfo(raw)
	}
	info.proxies, info.rules, info.groups = parseProfileStats(config)
	return config, info, nil
}

func decodeSubscription(body []byte) ([]byte, error) {
	body = []byte(strings.TrimPrefix(string(body), string([]byte{0xEF, 0xBB, 0xBF})))
	if looksLikeClashConfig(body) {
		return body, nil
	}
	trimmed := strings.TrimSpace(string(body))
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := enc.DecodeString(trimmed)
		if err != nil {
			continue
		}
		if looksLikeClashConfig(decoded) {
			return decoded, nil
		}
	}
	return nil, errors.New("无法识别订阅内容，请确认是 Clash/mihomo 订阅")
}

func looksLikeClashConfig(data []byte) bool {
	text := strings.ToLower(string(data))
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, key := range []string{"proxies:", "proxy-groups:", "proxy-providers:", "rules:", "rule-providers:"} {
			if strings.HasPrefix(line, key) {
				return true
			}
		}
	}
	return false
}

func parseUserInfo(raw string) subscriptionInfo {
	info := subscriptionInfo{}
	for _, part := range strings.Split(raw, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		value, _ := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
		switch strings.ToLower(strings.TrimSpace(kv[0])) {
		case "upload":
			info.upload = value
		case "download":
			info.download = value
		case "total":
			info.total = value
		case "expire":
			info.expire = value
		}
	}
	return info
}

// parseProfileStats reads just the counts of the bulk sections.
func parseProfileStats(data []byte) (proxies, rules, groups int) {
	var doc struct {
		Proxies     []any `yaml:"proxies"`
		ProxyGroups []any `yaml:"proxy-groups"`
		Rules       []any `yaml:"rules"`
	}
	if err := yamlUnmarshal(data, &doc); err != nil {
		return 0, 0, 0
	}
	return len(doc.Proxies), len(doc.Rules), len(doc.ProxyGroups)
}
