.PHONY: all test clean

DIRS=$(shell find . -maxdepth 2 -name *.go -printf "%h\n" | uniq)
TARGETS=$(join $(patsubst %,%/,$(DIRS)),$(DIRS))
BUILD_IN_CONTAINER=true
RM=--rm

all: $(TARGETS)

ifeq ($(BUILD_IN_CONTAINER),true)
$(TARGETS):
	$(SUDO) docker run -ti $(RM) -v $(shell pwd)/../:/go/src/github.com/weaveworks/scope -e GOARCH -e GOOS \
		weaveworks/scope-backend-build -C experimental $@
else
$(TARGETS):
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)
endif

test:
	go test ./...

clean:
	go clean ./...
