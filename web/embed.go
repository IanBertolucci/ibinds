package web

import "embed"

// Assets holds the built Svelte SPA (web/dist), embedded into the binary.
//
//go:embed all:dist
var Assets embed.FS
