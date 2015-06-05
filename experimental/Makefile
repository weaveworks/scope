.PHONY: all test clean

DIRS=$(shell find . -maxdepth 2 -name *.go -printf "%h\n" | uniq)
TARGETS=$(join $(patsubst %,%/,$(DIRS)),$(DIRS))

all: $(TARGETS)

$(TARGETS):
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

test:
	go test ./...

clean:
	go clean ./...
