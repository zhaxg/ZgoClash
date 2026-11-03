package app

import (
	"time"

	"zgo-clash/internal/engine"
)

// Settings is what the Tools page writes and what the core is built from.
type Settings struct {
	Language           string `json:"language"`
	Theme              string `json:"theme"`
	AutoLaunch         bool   `json:"autoLaunch"`
	SilentStart        bool   `json:"silentStart"`
	CloseToTray        bool   `json:"closeToTray"`
	AutoStartCore      bool   `json:"autoStartCore"`
	SystemProxy        bool   `json:"systemProxy"`
	Tun                bool   `json:"tun"`
	TunStack           string `json:"tunStack"`
	MixedPort          int    `json:"mixedPort"`
	AllowLan           bool   `json:"allowLan"`
	IPv6               bool   `json:"ipv6"`
	Mode               string `json:"mode"`
	LogLevel           string `json:"logLevel"`
	TestURL            string `json:"testUrl"`
	AutoUpdateInterval int    `json:"autoUpdateInterval"`
	ControllerPort     int    `json:"controllerPort"`
	// ControllerExternal lets the core's REST API answer from other
	// devices, not only from this one.
	ControllerExternal bool `json:"controllerExternal"`
}

// Profile is a subscription, a local file or a hand-written configuration.
type Profile struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	UpdatedAt  int64  `json:"updatedAt"`
	Upload     int64  `json:"upload"`
	Download   int64  `json:"download"`
	Total      int64  `json:"total"`
	Expire     int64  `json:"expire"`
	ProxyCount int    `json:"proxyCount"`
	RuleCount  int    `json:"ruleCount"`
	GroupCount int    `json:"groupCount"`
}

// UpdateReport says what a subscription update did.
type UpdateReport struct {
	Updated []string
	Failed  []string
}

// Snapshot is everything the shell shows about the app at one moment.
type Snapshot struct {
	Running            bool
	Starting           bool
	AppVersion         string
	CoreVersion        string
	StartedAt          time.Time
	Uptime             int64
	Settings           Settings
	Profiles           []Profile
	CurrentProfileID   string
	CurrentProfileName string
	MixedPort          int
	ControllerPort     int
	SystemProxy        bool
	CanTun             bool
	Elevated           bool
	Home               string
	LastError          string
	Traffic            engine.Traffic
	ConnectionsCount   int
	TCPCount           int
	UDPCount           int
	RecentCount        int
	MemoryBytes        uint64
	MemoryApp          uint64
	MemoryCore         uint64
	LocalIP            string
}

// CurrentProfile returns the profile in use.
func (s Snapshot) CurrentProfile() (Profile, bool) {
	for _, p := range s.Profiles {
		if p.ID == s.CurrentProfileID {
			return p, true
		}
	}
	return Profile{}, false
}
