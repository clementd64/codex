package main

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:packages/front/dist
var frontendDist embed.FS

// frontend serves the frontend with index.html as the SPA entry point.
func frontend() http.Handler {
	files, err := fs.Sub(frontendDist, "packages/front/dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(files, name); errors.Is(err, fs.ErrNotExist) {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
