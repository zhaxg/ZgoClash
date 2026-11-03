// Package assets holds the small pictures the app shows without fetching
// them: the icons of the sites the dashboard probes.
package assets

import _ "embed"

// The four sites the service card watches. Google, GitHub, OpenAI and
// Twitter, each with its own mark.
//
//go:embed favicon_google.png
var Google []byte

//go:embed favicon_github.png
var Github []byte

//go:embed favicon_openai.png
var OpenAI []byte

//go:embed favicon_twitter.png
var Twitter []byte

// Twemoji is the colour emoji font, registered under "Twemoji Mozilla" so
// text that may hold a flag or another emoji can draw it.
//
//go:embed Twemoji.Mozilla.ttf
var Twemoji []byte
