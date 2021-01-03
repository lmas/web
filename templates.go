package web

import (
	"html/template"
	"path/filepath"
)

// LoadTemplates is a helper for quickly loading template files from a dir (using
// a filepath.Glob pattern) and an optional FuncMap. The returned map can be used
// straight away in the Options{} struct for the web handler.
//
// NOTE: it will cause a panic on any errors (cuz I think it's bad enough, while
// trying to start up the web server).
//
// TODO: rewrite this helper to use embed.FS instead, after go1.16 has landed
// in feb 2021 (see https://tip.golang.org/pkg/embed/)
func LoadTemplates(globDir string, funcs template.FuncMap) map[string]*template.Template {
	files, err := filepath.Glob(globDir)
	if err != nil {
		panic(err)
	}

	layout := files[0]
	layoutName := filepath.Base(layout)
	list := make(map[string]*template.Template)
	for _, f := range files[1:] {
		t, err := template.New(layoutName).Funcs(funcs).ParseFiles(layout, f)
		if err != nil {
			panic(err)
		}
		list[filepath.Base(f)] = t
	}
	return list
}
