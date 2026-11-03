package engine

// Traffic is the rate and the totals the core reports.
type Traffic struct {
	Up        int64 `json:"up"`
	Down      int64 `json:"down"`
	UpTotal   int64 `json:"upTotal"`
	DownTotal int64 `json:"downTotal"`
}

// DelayHistory is one delay test result of a proxy.
type DelayHistory struct {
	Time  string `json:"time"`
	Delay int    `json:"delay"`
}

// Proxy is a mihomo proxy or proxy group as the controller reports it.
type Proxy struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Now      string         `json:"now,omitempty"`
	All      []string       `json:"all,omitempty"`
	History  []DelayHistory `json:"history,omitempty"`
	Alive    bool           `json:"alive"`
	UDP      bool           `json:"udp"`
	XUDP     bool           `json:"xudp"`
	TFO      bool           `json:"tfo"`
	Icon     string         `json:"icon,omitempty"`
	Hidden   bool           `json:"hidden,omitempty"`
	Provider string         `json:"provider,omitempty"`
}

// LastDelay returns the most recent tested delay, if any.
func (p Proxy) LastDelay() (int, bool) {
	if len(p.History) == 0 {
		return 0, false
	}
	return p.History[len(p.History)-1].Delay, true
}

// ProxyGroupView keeps the order of the profile's groups.
type ProxyGroupView struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Now    string   `json:"now"`
	All    []string `json:"all"`
	Icon   string   `json:"icon,omitempty"`
	Hidden bool     `json:"hidden,omitempty"`
}

// ConnectionMetadata is the metadata mihomo attaches to a connection.
type ConnectionMetadata struct {
	Network           string `json:"network"`
	Type              string `json:"type"`
	SourceIP          string `json:"sourceIP"`
	DestinationIP     string `json:"destinationIP"`
	SourcePort        string `json:"sourcePort"`
	DestinationPort   string `json:"destinationPort"`
	Host              string `json:"host"`
	DNSMode           string `json:"dnsMode"`
	ProcessPath       string `json:"processPath"`
	Process           string `json:"process"`
	SpecialProxy      string `json:"specialProxy"`
	RemoteDestination string `json:"remoteDestination"`
	InboundName       string `json:"inboundName"`
	InboundPort       string `json:"inboundPort"`
}

// Connection is one live connection.
type Connection struct {
	ID          string             `json:"id"`
	Metadata    ConnectionMetadata `json:"metadata"`
	Upload      int64              `json:"upload"`
	Download    int64              `json:"download"`
	Start       string             `json:"start"`
	Chains      []string           `json:"chains"`
	Rule        string             `json:"rule"`
	RulePayload string             `json:"rulePayload"`
}

// ConnectionsResult is the core's connections snapshot.
type ConnectionsResult struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Connections   []Connection `json:"connections"`
	Memory        uint64       `json:"memory"`
}

// LogEntry is one line of the core log.
type LogEntry struct {
	Time    int64  `json:"time"`
	Level   string `json:"level"`
	Payload string `json:"payload"`
}
