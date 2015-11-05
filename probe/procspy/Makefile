.PHONY: all build buildall test install bench
all: test build buildall install

build: 
	go build
	go vet
	golint .

buildall:
	GOOS=darwin go build
	GOOS=linux go build

test:
	go test

install:
	go install

bench:
	go test -bench .
