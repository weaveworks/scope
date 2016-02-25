.PHONY: all vet lint build test clean

GO ?= env GO15VENDOREXPERIMENT=1 GOGC=off go

all: build test vet lint

vet:
	$(GO) vet ./...

lint:
	golint .

build:
	$(GO) build

test:
	env -u GOGC $(GO) test

clean:
	$(GO) clean

