package web

import "embed"

//go:embed views/*
var viewFiles embed.FS

//go:embed static/*
var staticFiles embed.FS

func GetViewFiles() embed.FS {
	return viewFiles
}

func GetStaticFiles() embed.FS {
	return staticFiles
}
