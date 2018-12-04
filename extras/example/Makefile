CC=gcc
CFLAGS=-g -lpthread
BUILD_IN_CONTAINER=true
GIT_REVISION := $(shell git rev-parse HEAD)

all: .qotd.marker .app.marker .client.marker .searchapp.marker .shout.marker .frontend.marker .frontend-k8s.marker .echo.marker .trace_app.marker

searchapp/searchapp: searchapp/app.go
shout/shout: shout/shout.go
qotd/qotd: qotd/qotd.o

ifeq ($(BUILD_IN_CONTAINER),true)
qotd/qotd qotd/qotd.o searchapp/searchapp shout/shout:
	$(SUDO) docker run -ti --rm -v $(shell pwd)/../../:/go/src/github.com/weaveworks/scope -e GOARCH -e GOOS \
		weaveworks/scope-backend-build -C extras/example $@
else
qotd/qotd:
	gcc -o $@ $< $(CFLAGS)

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

shout/shout searchapp/searchapp:
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)
endif

.qotd.marker: qotd/* qotd/qotd
.app.marker: app/*
.client.marker: client/*
.searchapp.marker: searchapp/* searchapp/searchapp
.shout.marker: shout/* shout/shout
.frontend.marker: frontend/*
.frontend-k8s.marker: frontend-k8s/*
.echo.marker: echo/*
.trace_app.marker: trace_app/*
.%.marker:
	docker build --build-arg=revision=$(GIT_REVISION) -t tomwilkie/$(<D) $(<D)/
	touch $@

clean:
	rm -f qotd/*.o qotd/qotd searchapp/searchapp shout/shout .*.marker
