package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/mjpitz/myago/vue"
)

//go:generate npm run build

//go:embed dist/*
var assets embed.FS

func Handler() http.Handler {
	root, _ := fs.Sub(assets, "dist")
	return http.FileServer(vue.Wrap(http.FS(root)))
}
