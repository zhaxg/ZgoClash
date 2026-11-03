package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultTestURL = "http://www.gstatic.com/generate_204"

type profilesFile struct {
	Current string    `json:"current"`
	Items   []Profile `json:"items"`
}

// Store owns the files below the user data directory.
type Store struct {
	dir string
	mu  sync.Mutex
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) ensure() error {
	for _, dir := range []string{
		s.dir,
		filepath.Join(s.dir, "profiles"),
		filepath.Join(s.dir, "runtime"),
		filepath.Join(s.dir, "core"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) settingsPath() string {
	return filepath.Join(s.dir, "settings.json")
}

func (s *Store) profilesPath() string {
	return filepath.Join(s.dir, "profiles.json")
}

func (s *Store) profilePath(id string) string {
	return filepath.Join(s.dir, "profiles", id+".yaml")
}

func (s *Store) runtimePath() string {
	return filepath.Join(s.dir, "runtime", "config.yaml")
}

func (s *Store) coreHome() string {
	return filepath.Join(s.dir, "core")
}

func defaultSettings() Settings {
	return Settings{
		Language:           "zh",
		Theme:              "system",
		AutoLaunch:         false,
		SilentStart:        false,
		CloseToTray:        true,
		AutoStartCore:      true,
		SystemProxy:        false,
		Tun:                false,
		TunStack:           "gvisor",
		MixedPort:          7890,
		AllowLan:           false,
		IPv6:               false,
		Mode:               "rule",
		LogLevel:           "info",
		TestURL:            defaultTestURL,
		AutoUpdateInterval: 24,
		ControllerPort:     9090,
		ControllerExternal: false,
	}
}

func normalizeSettings(v Settings) Settings {
	d := defaultSettings()
	if v.Language != "zh" && v.Language != "en" {
		v.Language = d.Language
	}
	if v.Theme != "system" && v.Theme != "light" && v.Theme != "dark" {
		v.Theme = d.Theme
	}
	if v.MixedPort < 1 || v.MixedPort > 65535 {
		v.MixedPort = d.MixedPort
	}
	if v.ControllerPort < 1 || v.ControllerPort > 65535 {
		v.ControllerPort = d.ControllerPort
	}
	switch v.Mode {
	case "rule", "global", "direct":
	default:
		v.Mode = d.Mode
	}
	switch v.LogLevel {
	case "silent", "error", "warning", "info", "debug":
	default:
		v.LogLevel = d.LogLevel
	}
	switch v.TunStack {
	case "gvisor", "system", "mixed":
	default:
		v.TunStack = d.TunStack
	}
	if strings.TrimSpace(v.TestURL) == "" {
		v.TestURL = d.TestURL
	}
	if v.AutoUpdateInterval < 0 || v.AutoUpdateInterval > 24*30 {
		v.AutoUpdateInterval = d.AutoUpdateInterval
	}
	return v
}

func (s *Store) LoadSettings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.settingsPath())
	if err != nil {
		return defaultSettings()
	}
	var v Settings
	if json.Unmarshal(data, &v) != nil {
		return defaultSettings()
	}
	return normalizeSettings(v)
}

func (s *Store) SaveSettings(v Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeJSONAtomic(s.settingsPath(), normalizeSettings(v))
}

func (s *Store) LoadProfiles() (string, []Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.profilesPath())
	if err != nil {
		return "", nil
	}
	var f profilesFile
	if json.Unmarshal(data, &f) != nil {
		return "", nil
	}
	items := f.Items[:0]
	for _, p := range f.Items {
		if _, err := os.Stat(s.profilePath(p.ID)); err == nil {
			items = append(items, p)
		}
	}
	return f.Current, items
}

func (s *Store) SaveProfiles(current string, items []Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeJSONAtomic(s.profilesPath(), profilesFile{Current: current, Items: items})
}

func (s *Store) WriteProfile(id string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeFileAtomic(s.profilePath(id), data)
}

func (s *Store) ReadProfile(id string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.ReadFile(s.profilePath(id))
}

func (s *Store) DeleteProfile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.Remove(s.profilePath(id))
}

func (s *Store) WriteRuntime(data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.runtimePath()
	return path, writeFileAtomic(path, data)
}

func writeFileAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeJSONAtomic(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data)
}

func newID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func newSecret() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func defaultProfileContent() []byte {
	return []byte(`proxies: []
proxy-groups: []
rules:
  - MATCH,DIRECT
`)
}
