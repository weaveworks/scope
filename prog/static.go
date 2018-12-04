package main

import (
	"net/http"

	"github.com/weaveworks/scope/prog/externalui"
	"github.com/weaveworks/scope/prog/staticui"
)

// GetFS obtains the UI code
func GetFS(useExternal bool) http.FileSystem {
	if useExternal {
		return externalui.FS(false)
	}
	return staticui.FS(false)
}
