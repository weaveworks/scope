package main

import (
	"net/http"

	"github.com/weaveworks/scope/prog/staticui"
	"github.com/weaveworks/scope/prog/externalui"
)

func GetFS(use_external bool) http.FileSystem {
	if use_external {
		return externalui.FS(false)
	} else {
		return staticui.FS(false)
	}
}
