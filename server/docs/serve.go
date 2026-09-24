package docs

import (
	"embed"
	"io/fs"
	"net/http"
)

// API specs and the Swagger page. User manual and architecture pages are
// copied here by docs/pack.py before the binary is built; they live outside
// the Go module, so go:embed cannot see the originals.

//go:embed *.yaml
//go:embed index.html
var apiDocs embed.FS

//go:embed all:manual
var manualDocs embed.FS

//go:embed all:architecture
var architectureDocs embed.FS

func GetFS() http.FileSystem {
	return http.FS(apiDocs)
}

func ManualFS() http.FileSystem {
	sub, err := fs.Sub(manualDocs, "manual")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func ArchitectureFS() http.FileSystem {
	sub, err := fs.Sub(architectureDocs, "architecture")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
