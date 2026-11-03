package app

import (
	"strings"
	"testing"
)

// TestRuntimeConfigKeepsEmoji checks that the runtime configuration keeps
// the flags of the names as they are: the YAML library writes every emoji as
// an escape, and the core and the interface would show the escape.
func TestRuntimeConfigKeepsEmoji(t *testing.T) {
	profile := []byte(`proxies:
  - name: "🇯🇵JP_B0_38.4"
    type: ss
    server: 1.2.3.4
    port: 443
proxy-groups:
  - name: "🚀 节点选择"
    type: select
    proxies:
      - "🇯🇵JP_B0_38.4"
`)
	out, _, err := buildRuntimeConfig(profile, Settings{
		MixedPort: 7890,
		Mode:      "rule",
		LogLevel:  "info",
	}, 9090, "secret")
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{"🇯🇵JP_B0_38.4", "🚀 节点选择"} {
		if !strings.Contains(text, want) {
			t.Errorf("the configuration lost %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, `\U0001`) {
		t.Errorf("the configuration holds an escape:\n%s", text)
	}
}
