package main

import (
	"embed"
	"io/fs"
	"path"
)

//go:embed templates
var tmplEmbed embed.FS

//go:embed static
var staticEmbedFS embed.FS

type staticFS struct {
	fs fs.FS
}

func (sfs *staticFS) Open(name string) (fs.File, error) {
	return sfs.fs.Open(path.Join("static", name))
}

var staticEmbed = &staticFS{staticEmbedFS}