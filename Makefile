.PHONY: all build static test clean

all: build

build: static
	go build ./...

static:
	go get github.com/mjibson/esc
	cd client && make build && rm -f dist/.htaccess
	cd app && esc -o static.go -prefix ../client/dist ../client/dist

test:
	go test ./...

clean:
	go clean ./...

