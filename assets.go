// Package cementer exists only to embed the built web client into the binary, so
// the entire application — server plus UI — ships as a single file. The web/dist
// directory is produced by `npm run build` in ./web before `go build`.
package cementer

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var distFS embed.FS

// WebDist returns the built client as a filesystem rooted at the dist directory.
func WebDist() (fs.FS, error) {
	return fs.Sub(distFS, "web/dist")
}
