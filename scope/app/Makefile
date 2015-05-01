.PHONY: all vet lint build test clean static

all: build test vet lint

vet:
	go vet ./...

lint:
	golint .

build:
	go build

test:
	go test

clean:
	go clean

static:
	go get github.com/mjibson/esc
	cd ../client && make build && rm -f dist/.htaccess
	esc -o static.go -prefix ../client/dist ../client/dist
	${MAKE} build
