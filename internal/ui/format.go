package ui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	mu "github.com/egoist/mygo/ui"
)

func formatBytes(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit && exp < 4; v /= unit {
		div *= unit
		exp++
	}
	value := float64(n) / float64(div)
	suffix := "KMGTPE"[exp]
	digits := 0
	if value < 10 {
		digits = 2
	} else if value < 100 {
		digits = 1
	}
	return strconv.FormatFloat(value, 'f', digits, 64) + " " + string(suffix) + "B"
}

func formatSpeed(n int64) string {
	if n <= 0 {
		return "0 B/s"
	}
	return formatBytes(n) + "/s"
}

func formatCount(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead > 0 {
		b.WriteString(s[:lead])
	}
	for i := lead; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// formatUptime shows a running time as 2d 03:04:05 or 03:04:05.
func formatUptime(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if d > 0 {
		return fmt.Sprintf("%d天 %02d:%02d:%02d", d, h, m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// formatDuration shows a length of time in words: 2 小时 13 分.
func formatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0 秒"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d 秒", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%d 分", seconds/60)
	}
	if seconds < 86400 {
		h, m := seconds/3600, (seconds%3600)/60
		if m == 0 {
			return fmt.Sprintf("%d 小时", h)
		}
		return fmt.Sprintf("%d 小时 %d 分", h, m)
	}
	d, h := seconds/86400, (seconds%86400)/3600
	if h == 0 {
		return fmt.Sprintf("%d 天", d)
	}
	return fmt.Sprintf("%d 天 %d 小时", d, h)
}

func formatAgo(ms int64) string {
	if ms <= 0 {
		return "从未"
	}
	t := time.UnixMilli(ms)
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return fmt.Sprintf("%d 分钟前", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d 小时前", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d 天前", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func formatExpire(sec int64) string {
	if sec <= 0 {
		return ""
	}
	t := time.Unix(sec, 0)
	d := time.Until(t)
	if d <= 0 {
		return "已过期"
	}
	return fmt.Sprintf("%d 天后到期", int(d.Hours()/24))
}

func formatDate(ms int64) string {
	if ms <= 0 {
		return "—"
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

func formatDelay(d int) string {
	if d <= 0 {
		return "超时"
	}
	return fmt.Sprintf("%d ms", d)
}

func delayColor(p palette, d int) mu.Color {
	if d <= 0 {
		return p.Muted
	}
	if d < 200 {
		return p.Good
	}
	if d < 600 {
		return p.Warn
	}
	return p.Bad
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	return u.Host
}

func percent(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
