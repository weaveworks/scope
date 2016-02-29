BUILD_IN_CONTAINER ?= true
RM=--rm
BUILD_IMAGE=golang:1.5.3

ifeq ($(BUILD_IN_CONTAINER),true)

all test:
	$(SUDO) docker run $(RM) -ti \
		-v $(shell pwd):/go/src/github.com/weaveworks/go-checkpoint \
		-e GOARCH -e GOOS -e BUILD_IN_CONTAINER=false \
		$(BUILD_IMAGE) make -C /go/src/github.com/weaveworks/go-checkpoint $@

else

all:
	go get .
	go build .

test:
	go get .
	go test

endif

