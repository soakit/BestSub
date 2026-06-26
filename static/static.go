package static

import (
	"embed"
	"io/fs"
)

//go:embed all:out
var staticFS embed.FS

var StaticFS = func() fs.FS {
	result, err := fs.Sub(staticFS, "out")
	if err != nil {
		panic(err)
	}
	return result
}()
