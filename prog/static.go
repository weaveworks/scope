package main

import (
	"net/http"

	"github.com/weaveworks/scope/prog/externalui"
	"github.com/weaveworks/scope/prog/staticui"
)

func GetFS(use_external bool) http.FileSystem {
	if use_external {
		return externalui.FS(false)
	} else {
		return staticui.FS(false)
	}
}
