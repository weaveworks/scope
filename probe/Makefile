.PHONY: all vet lint build test clean

all: build test vet lint

vet:
	go vet ./...

lint:
	golint .

build:
	go build

test:
	#GOOS=darwin go build
	#GOOS=linux go build
	go test

clean:
	go clean

