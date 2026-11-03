package ui

import "testing"

// TestFlagFor checks that the flag of a node is read from its name, and
// that a name which already carries an emoji is left as it is.
func TestFlagFor(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"JP-家宽-01", "🇯🇵"},
		{"JP_家宽_01", "🇯🇵"},
		{"GB_HK_01", "🇭🇰"},
		{"HK-01", "🇭🇰"},
		{"美国 洛杉矶", "🇺🇸"},
		{"US_DROCK_RTY", "🇺🇸"},
		// A name that already holds an emoji keeps it, and gets no second
		// mark.
		{"🏠 JP-家宽-01", ""},
		{"🇯🇵JP_B0_38.4", ""},
		{"🚀 节点选择", ""},
		{"♻️ 自动选择", ""},
		{"家宽前置", ""},
	}
	for _, c := range cases {
		if got := flagFor(c.name); got != c.want {
			t.Errorf("flagFor(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestWithFlag checks that the flag is put in front of the name, and that a
// name which holds one is not marked twice.
func TestWithFlag(t *testing.T) {
	if got := withFlag("JP-家宽-01"); got != "🇯🇵 JP-家宽-01" {
		t.Errorf("withFlag = %q", got)
	}
	if got := withFlag("🏠 JP-家宽-01"); got != "🏠 JP-家宽-01" {
		t.Errorf("withFlag = %q", got)
	}
}
