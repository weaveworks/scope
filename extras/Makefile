.PHONY: all test clean

DIRS=$(shell find . -maxdepth 2 -name *.go | xargs -n1 dirname | sort -u)
TARGETS=$(join $(patsubst %,%/,$(DIRS)),$(patsubst ./%,%,$(DIRS)))
BUILD_IN_CONTAINER=true
RM=--rm

all: $(TARGETS)

ifeq ($(BUILD_IN_CONTAINER),true)
$(TARGETS):
	$(SUDO) docker run -ti $(RM) -v $(shell pwd)/../:/go/src/github.com/weaveworks/scope -e GOARCH -e GOOS \
		weaveworks/scope-backend-build -C extras $@
else
$(TARGETS):
	go build -ldflags "-extldflags \"-static\"" -tags 'netgo unsafe' -o $@ ./$(@D)
endif

test:
	go test ./...

clean:
	go clean ./...
