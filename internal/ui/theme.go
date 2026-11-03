package ui

import (
	"time"

	mu "github.com/egoist/mygo/ui"
)

// palette is the look of the app's own surfaces: the page, the sidebar, the
// cards and the small colored details. Widgets take their own colors from the
// MyGo theme, which themeFor tunes to match.
type palette struct {
	Page       mu.Color
	Sidebar    mu.Color
	Card       mu.Color
	CardHover  mu.Color
	Soft       mu.Color
	Border     mu.Color
	Text       mu.Color
	Muted      mu.Color
	Accent     mu.Color
	AccentSoft mu.Color
	AccentText mu.Color
	Teal       mu.Color
	Good       mu.Color
	Warn       mu.Color
	Bad        mu.Color
	Track      mu.Color
	Shadow     mu.Color

	// Mark is the bright green of the check that shows what is in use.
	Mark mu.Color

	// The traffic cards' dot matrix: the tinted box the dots sit in, the
	// idle dots, and the two ends of the ramp the filled dots follow.
	ChartBg    mu.Color
	ChartIdle  mu.Color
	ChartLight mu.Color
	ChartDeep  mu.Color
	ChartHot   mu.Color
}

func darkPalette() palette {
	return palette{
		Page:       mu.Hex("#0d1016"),
		Sidebar:    mu.Hex("#11151d"),
		Card:       mu.Hex("#171c25"),
		CardHover:  mu.Hex("#1d2330"),
		Soft:       mu.Hex("#202734"),
		Border:     mu.Hex("#262e3c"),
		Text:       mu.Hex("#e9edf6"),
		Muted:      mu.Hex("#8b95a9"),
		Accent:     mu.Hex("#5f6ffb"),
		AccentSoft: mu.RGBA(95, 111, 251, 0.16),
		AccentText: mu.Hex("#ffffff"),
		Teal:       mu.Hex("#2dd4bf"),
		Good:       mu.Hex("#35d39a"),
		Warn:       mu.Hex("#f2b544"),
		Bad:        mu.Hex("#f16a6a"),
		Track:      mu.RGBA(255, 255, 255, 0.07),
		Shadow:     mu.RGBA(0, 0, 0, 0.45),
		Mark:       mu.Hex("#4ade80"),

		ChartBg:    mu.Hex("#16233a"),
		ChartIdle:  mu.Hex("#2b3d5c"),
		ChartLight: mu.Hex("#7cc4ff"),
		ChartDeep:  mu.Hex("#3b6fe0"),
		ChartHot:   mu.Hex("#f2564f"),
	}
}

func lightPalette() palette {
	return palette{
		Page:       mu.Hex("#f2f5fa"),
		Sidebar:    mu.Hex("#ffffff"),
		Card:       mu.Hex("#ffffff"),
		CardHover:  mu.Hex("#f6f8fc"),
		Soft:       mu.Hex("#eef2f8"),
		Border:     mu.Hex("#e2e7f0"),
		Text:       mu.Hex("#171c28"),
		Muted:      mu.Hex("#6e788b"),
		Accent:     mu.Hex("#4f5ef7"),
		AccentSoft: mu.RGBA(79, 94, 247, 0.10),
		AccentText: mu.Hex("#ffffff"),
		Teal:       mu.Hex("#0e9f8e"),
		Good:       mu.Hex("#0f9d76"),
		Warn:       mu.Hex("#c47a09"),
		Bad:        mu.Hex("#d94a4a"),
		Track:      mu.Hex("#e9edf4"),
		Shadow:     mu.RGBA(23, 28, 40, 0.10),
		Mark:       mu.Hex("#22c55e"),

		ChartBg:    mu.Hex("#eff5ff"),
		ChartIdle:  mu.Hex("#cfe0fb"),
		ChartLight: mu.Hex("#9ecbff"),
		ChartDeep:  mu.Hex("#3b6fe0"),
		ChartHot:   mu.Hex("#e8453c"),
	}
}

func paletteFor(t *mu.Theme) palette {
	if t != nil && t.Dark {
		return darkPalette()
	}
	return lightPalette()
}

func themeFor(dark bool) *mu.Theme {
	var t *mu.Theme
	p := lightPalette()
	if dark {
		t = mu.DarkTheme()
		p = darkPalette()
	} else {
		t = mu.LightTheme()
	}
	copy := *t
	copy.Accent = p.Accent
	copy.AccentHover = p.Accent.Mix(mu.RGBA(255, 255, 255, 1), 0.12)
	copy.AccentPressed = p.Accent.Mix(mu.RGBA(0, 0, 0, 1), 0.12)
	copy.AccentText = p.AccentText
	copy.Selection = p.Accent.Alpha(0.28)
	copy.Focus = p.Accent
	copy.Radius = 9
	copy.Spacing = 4
	copy.FontSize = 14
	return &copy
}

// applyTheme picks the app's theme. In system mode it reads the desktop's
// appearance every few seconds, so the app follows it without losing the
// app's own accent and metrics.
func (u *UI) applyTheme(c *mu.Context, mode string) {
	dark := false
	switch mode {
	case "dark":
		dark = true
	case "light":
		dark = false
	default:
		if u.systemThemeAt.IsZero() || time.Since(u.systemThemeAt) > 5*time.Second {
			u.systemDark = systemDarkTheme()
			u.systemThemeAt = time.Now()
		}
		dark = u.systemDark
	}
	c.SetTheme(themeFor(dark))
}
