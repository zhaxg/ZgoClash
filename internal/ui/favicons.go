package ui

import (
	mu "github.com/egoist/mygo/ui"

	"zgo-clash/internal/ui/assets"
)

// The service card shows the mark of each site it watches. The marks are
// small pictures the app carries, so the card never waits for the network to
// show them.
var (
	markGoogle  = mustMark(assets.Google)
	markGithub  = mustMark(assets.Github)
	markOpenAI  = mustMark(assets.OpenAI)
	markTwitter = mustMark(assets.Twitter)
)

// mustMark decodes a mark, or gives up on it: a mark that cannot be read is
// simply not shown.
func mustMark(data []byte) *mu.Bitmap {
	b, err := mu.DecodeBitmap(data)
	if err != nil {
		return nil
	}
	return b
}

// serviceMark is the mark of the site with that name.
func serviceMark(name string) *mu.Bitmap {
	switch name {
	case "Google":
		return markGoogle
	case "GitHub":
		return markGithub
	case "OpenAI":
		return markOpenAI
	case "Twitter":
		return markTwitter
	}
	return nil
}
