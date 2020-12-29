/*

esc embeds files into go programs and provides http.FileSystem interfaces
to them.

It adds all named files or files recursively under named directories at the
path specified. The output file provides an http.FileSystem interface with
zero dependencies on packages outside the standard library.

Usage:
	esc [flag] [name ...]

The flags are:
	-o=""
		output filename, defaults to stdout
	-pkg="main"
		package name of output file, defaults to main
	-prefix=""
		strip given prefix from filenames
	-ignore=""
		regular expression for files to ignore
	-include=""
		regular expression for files to include
	-modtime=""
		Unix timestamp to override as modification time for all files
	-private
		unexport functions by prefixing them with esc, e.g. FS -> escFS
	-no-compress
		do not compress files

Accessing Embedded Files

After producing an output file, the assets may be accessed with the FS()
function, which takes a flag to use local assets instead (for local
development).

FS(Must)?(Byte|String) returns an asset as a (byte slice|string).
FSMust(Byte|String) panics if the asset is not found.

Go Generate

esc can be invoked by go generate:

	//go:generate esc -o static.go -pkg server static

Example

Embedded assets can be served with HTTP using the http.FileServer.
Assuming you have a directory structure similar to the following:

	.
	├── main.go
	└── static
	    ├── css
	    │   └── style.css
	    └── index.html

Where main.go contains:

	package main

	import (
		"log"
		"net/http"
	)

	func main() {
		// FS() is created by esc and returns a http.Filesystem.
		http.Handle("/static/", http.FileServer(FS(false)))
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

1. Generate the embedded data:
	esc -o static.go static
2. Start the server:
	go run main.go static.go
3. Access http://localhost:8080/static/index.html to view the files.

*/
package main
