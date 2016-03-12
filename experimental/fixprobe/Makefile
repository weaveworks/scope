.PHONY: all vet lint build test clean

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

