package onefile

import (
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
)

type OneFile struct {
	fs         fs.FS
	fileServer http.Handler
	overwrite  *Overwrite
	fallback   string
	Index      string
}

type Overwrite struct {
	Fsys fs.FS
	Pair map[string]string
}

func New(fsys fs.FS, overwrite *Overwrite, fallback string) *OneFile {
	return &OneFile{
		fs:         fsys,
		fileServer: http.FileServer(http.FS(fsys)),
		overwrite:  overwrite,
		fallback:   fallback,
		Index:      "index.html",
	}
}

func (o *OneFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path[len(path)-1] == '/' {
		path += o.Index
	}
	path = path[1:]
	if o.overwrite != nil {
		if aim, ok := o.overwrite.Pair[path]; ok {
			if _, err := fs.Stat(o.overwrite.Fsys, aim); err == nil {
				f, err := o.overwrite.Fsys.Open(aim)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
				} else {
					w.WriteHeader(http.StatusOK)
					io.Copy(w, f)
					f.Close()
				}
				return
			}
		}
	}
	if _, err := fs.Stat(o.fs, path); err == nil {
		o.fileServer.ServeHTTP(w, r)
	} else {
		_, filename := filepath.Split(path)
		if filepath.Ext(path) != "" {
			if r.Header.Get("if-none-match") == filename {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if o.fallback == "" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			f, err := o.fs.Open(o.fallback)
			if err != nil {
				if err == fs.ErrNotExist {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
				}
			}
			w.WriteHeader(http.StatusOK)
			io.Copy(w, f)
			f.Close()
		}
	}
}
