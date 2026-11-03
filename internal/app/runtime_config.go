package app

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func yamlUnmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func mapOf(parent map[string]any, key string) map[string]any {
	if v, ok := parent[key].(map[string]any); ok {
		return v
	}
	v := map[string]any{}
	parent[key] = v
	return v
}

// buildRuntimeConfig lays the app's settings over the profile the user chose.
func buildRuntimeConfig(profile []byte, st Settings, controllerPort int, secret string) ([]byte, []string, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(profile, &doc); err != nil {
		return nil, nil, fmt.Errorf("配置文件解析失败: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}

	doc["mixed-port"] = st.MixedPort
	doc["allow-lan"] = st.AllowLan
	doc["ipv6"] = st.IPv6
	doc["mode"] = st.Mode
	doc["log-level"] = st.LogLevel
	// The controller answers on the loopback address unless the user let
	// other devices reach it.
	host := "127.0.0.1"
	if st.ControllerExternal {
		host = "0.0.0.0"
	}
	doc["external-controller"] = fmt.Sprintf("%s:%d", host, controllerPort)
	doc["secret"] = secret
	delete(doc, "external-controller-tls")
	delete(doc, "external-controller-unix")
	delete(doc, "external-controller-pipe")

	tun := mapOf(doc, "tun")
	tun["enable"] = st.Tun
	if st.Tun {
		tun["stack"] = st.TunStack
		tun["auto-route"] = true
		tun["auto-detect-interface"] = true
		tun["dns-hijack"] = []any{"any:53"}
		tun["device"] = "ZgoClash"
		dns := mapOf(doc, "dns")
		if _, ok := dns["enable"]; !ok {
			dns["enable"] = true
		}
		if _, ok := dns["enhanced-mode"]; !ok {
			dns["enhanced-mode"] = "fake-ip"
		}
		if _, ok := dns["fake-ip-range"]; !ok {
			dns["fake-ip-range"] = "198.18.0.1/16"
		}
		if _, ok := dns["default-nameserver"]; !ok {
			dns["default-nameserver"] = []any{"223.5.5.5", "119.29.29.29"}
		}
		if _, ok := dns["nameserver"]; !ok {
			dns["nameserver"] = []any{"https://doh.pub/dns-query", "https://dns.alidns.com/dns-query"}
		}
	}

	profileSection := mapOf(doc, "profile")
	if _, ok := profileSection["store-selected"]; !ok {
		profileSection["store-selected"] = true
	}
	if _, ok := profileSection["store-fake-ip"]; !ok {
		profileSection["store-fake-ip"] = true
	}
	if _, ok := doc["unified-delay"]; !ok {
		doc["unified-delay"] = true
	}
	if _, ok := doc["tcp-concurrent"]; !ok {
		doc["tcp-concurrent"] = true
	}
	if _, ok := doc["find-process-mode"]; !ok {
		doc["find-process-mode"] = "strict"
	}

	var groups []string
	if list, ok := doc["proxy-groups"].([]any); ok {
		for _, item := range list {
			if entry, ok := item.(map[string]any); ok {
				if name, ok := entry["name"].(string); ok && name != "" {
					groups = append(groups, name)
				}
			}
		}
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, nil, err
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return nil, nil, errors.New("配置文件为空")
	}
	return out, groups, nil
}
