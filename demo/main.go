package main

import (
	"embed"
	"github.com/azurity/go-onefile"
	"io/fs"
	"net/http"
	"os"
)

//go:embed static
var static embed.FS

func main() {
	overwrite := &onefile.Overwrite{
		Fsys: os.DirFS("resources"),
		Pair: map[string]string{
			"data/home.html": "home.html",
		},
	}
	fsys, _ := fs.Sub(static, "static")
	handle := onefile.New(fsys, overwrite, "index.html")
	http.Handle("/", handle)
	_ = http.ListenAndServe(":8080", nil)
}
