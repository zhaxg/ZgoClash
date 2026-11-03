package ui

import (
	"strings"
	"sync"

	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/ui/assets"
)

// emojiFamily names the colour emoji font registered by the app, after the
// interface family: the system font draws letters, digits and CJK, and only
// the glyphs it lacks fall through to Twemoji. Putting Twemoji first would
// hide digits, which it maps to blank keycap components.
const emojiFamily = "system-ui, Twemoji Mozilla"

// monoFamily draws text in the monospace font and falls back to the colour
// emoji font behind it. Without the fallback a flag is drawn as the letters
// of its country code, because the monospace font has glyphs for the
// regional indicators.
const monoFamily = "monospace, " + emojiFamily

var fontsOnce sync.Once

// ensureFonts makes sure the colour emoji font is available before the
// first frame paints.
func ensureFonts() {
	fontsOnce.Do(func() {
		_ = mu.RegisterFont(assets.Twemoji, "Twemoji Mozilla")
	})
}

// isRegional reports whether r is a regional indicator symbol.
func isRegional(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

// isEmoji reports whether r is part of an emoji: a pictograph, a dingbat, a
// transport symbol or a regional indicator.
func isEmoji(r rune) bool {
	switch {
	case r >= 0x1F000 && r <= 0x1FAFF: // pictographs, flags and transport
		return true
	case r >= 0x2600 && r <= 0x27BF: // misc symbols and dingbats
		return true
	case r >= 0x2B00 && r <= 0x2BFF: // arrows and other marks
		return true
	case r == 0xFE0F: // the emoji variation selector
		return true
	}
	return false
}

// hasEmojiRune reports whether s holds an emoji of any kind.
func hasEmojiRune(s string) bool {
	for _, r := range s {
		if isEmoji(r) {
			return true
		}
	}
	return false
}

// flagGlyph returns the flag emoji s already holds, or "".
func flagGlyph(s string) string {
	rs := []rune(s)
	for i := 0; i+1 < len(rs); i++ {
		if isRegional(rs[i]) && isRegional(rs[i+1]) {
			return string(rs[i : i+2])
		}
	}
	return ""
}

// flagRules maps a country to the substrings that identify it in a proxy
// name. The checked string is lower-cased; the first match wins.
var flagRules = []struct {
	flag string
	keys []string
}{
	{"🇭🇰", []string{"香港", "hong kong", "hongkong", "hkg", "hk-"}},
	{"🇹🇼", []string{"台湾", "台灣", "taiwan", "twn", "tpe", "-tw-", "tw-"}},
	{"🇯🇵", []string{"日本", "东京", "大阪", "tokyo", "japan", "osa", "-jp-", "jp-"}},
	{"🇰🇷", []string{"韩国", "首尔", "seoul", "korea", "-kr-", "kr-"}},
	{"🇸🇬", []string{"新加坡", "狮城", "singapore", "sgp", "-sg-"}},
	{"🇺🇸", []string{"美国", "洛杉矶", "纽约", "西雅图", "硅谷", "圣何塞", "los angeles", "san jose", "usa", "united states", "-us-", "us-"}},
	{"🇬🇧", []string{"英国", "伦敦", "london", "uk", "gbr", "united kingdom"}},
	{"🇩🇪", []string{"德国", "法兰克福", "germany", "deutschland", "frankfurt"}},
	{"🇫🇷", []string{"法国", "巴黎", "paris", "france"}},
	{"🇳🇱", []string{"荷兰", "阿姆斯特丹", "amsterdam"}},
	{"🇸🇪", []string{"瑞典", "stockholm"}},
	{"🇮🇹", []string{"意大利", "米兰", "milan", "italy"}},
	{"🇪🇸", []string{"西班牙", "马德里", "madrid", "spain"}},
	{"🇨🇭", []string{"瑞士", "苏黎世", "zurich", "switzerland"}},
	{"🇷🇺", []string{"俄罗斯", "莫斯科", "moscow", "russia"}},
	{"🇦🇺", []string{"澳大利亚", "悉尼", "sydney", "melbourne", "australia"}},
	{"🇨🇦", []string{"加拿大", "多伦多", "温哥华", "toronto", "vancouver", "canada"}},
	{"🇮🇳", []string{"印度", "孟买", "delhi", "mumbai", "india"}},
	{"🇲🇾", []string{"马来西亚", "吉隆坡", "malaysia", "kuala lumpur"}},
	{"🇹🇭", []string{"泰国", "曼谷", "bangkok", "thailand"}},
	{"🇻🇳", []string{"越南", "河内", "胡志明", "hanoi", "ho chi minh", "vietnam"}},
	{"🇵🇱", []string{"波兰", "华沙", "warsaw", "poland"}},
	{"🇹🇷", []string{"土耳其", "伊斯坦布尔", "istanbul", "turkey"}},
	{"🇧🇷", []string{"巴西", "圣保罗", "sao paulo", "brazil"}},
	{"🇦🇷", []string{"阿根廷", "buenos aires", "argentina"}},
	{"🇮🇩", []string{"印尼", "雅加达", "jakarta", "indonesia"}},
	{"🇳🇿", []string{"新西兰", "奥克兰", "auckland", "new zealand"}},
	{"🇲🇽", []string{"墨西哥", "墨西哥城", "mexico"}},
	{"🇦🇹", []string{"奥地利", "维也纳", "vienna", "austria"}},
	{"🇧🇪", []string{"比利时", "布鲁塞尔", "brussels", "belgium"}},
	{"🇩🇰", []string{"丹麦", "哥本哈根", "copenhagen", "denmark"}},
	{"🇳🇴", []string{"挪威", "奥斯陆", "oslo", "norway"}},
	{"🇨🇿", []string{"捷克", "布拉格", "prague", "czech"}},
	{"🇵🇭", []string{"菲律宾", "马尼拉", "manila", "philippines"}},
	{"🇿🇦", []string{"南非", "johannesburg", "south africa"}},
	{"🇦🇪", []string{"阿联酋", "迪拜", "dubai", "uae"}},
	{"🇶🇦", []string{"卡塔尔", "doha", "qatar"}},
}

// flagFor returns the flag of the country a proxy name belongs to, if it can
// tell. A name that already carries an emoji is marked by whoever wrote it,
// with the flag of its country or with another mark of its kind, so the app
// leaves it alone.
func flagFor(name string) string {
	if hasEmojiRune(name) {
		return ""
	}
	// Names separate their codes with -, _ or . alike (HK-01, GB_HK_01),
	// so the rules only need to spell the hyphen.
	flat := strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(name))
	for _, rule := range flagRules {
		for _, key := range rule.keys {
			if strings.Contains(flat, key) {
				return rule.flag
			}
		}
	}
	return ""
}

// withFlag prefixes a proxy name with the flag of its country, unless the
// name already has one.
func withFlag(name string) string {
	if flag := flagFor(name); flag != "" {
		return flag + " " + name
	}
	return name
}

// nodeFlag returns the flag of a node, whether its name holds one or appears
// to belong to that country.
func nodeFlag(name string) string {
	if f := flagFor(name); f != "" {
		return f
	}
	return flagGlyph(name)
}
