.PHONY: all vet lint build test clean
GIT_REVISION := $(shell git rev-parse HEAD)
GO_BUILD_FLAGS=$(GO_BUILD_INSTALL_DEPS) -ldflags "-extldflags \"-static\" -s -w" -tags netgo

all: build test vet lint

vet:
	go vet ./...

lint:
	golint .

build:
	go build $(GO_BUILD_FLAGS)

test:
	go test

docker:
	docker build --build-arg=revision=$(GIT_REVISION) -t quay.io/weaveworks/scope-deleter .

clean:
	go clean

