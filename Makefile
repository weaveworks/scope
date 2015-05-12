.PHONY: all build client static dist test clean

APP=app/app
FIXPROBE=experimental/fixprobe/fixprobe

all: build

build:
	go build ./...

client:
	cd client && make build && rm -f dist/.htaccess

static:
	go get github.com/mjibson/esc
	cd app && esc -o static.go -prefix ../client/dist ../client/dist

dist: client static build

test: ${APP} ${FIXPROBE}
	# app and fixprobe needed for integration tests
	go test ./...

${APP}:
	cd app && go build

${FIXPROBE}:
	cd experimental/fixprobe && go build

clean:
	go clean ./...

