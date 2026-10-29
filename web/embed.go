// Package web holds the built UI, which `make web` drops into dist
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// UI is the built app, rooted at its index.html
func UI() fs.FS {
	sub, _ := fs.Sub(dist, "dist")
	return sub
}
