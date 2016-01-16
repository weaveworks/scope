CC=gcc
CFLAGS=-g -lpthread
BUILD_IN_CONTAINER=true

all: qotd.marker app.marker client.marker searchapp.marker shout.marker frontend.marker echo.marker

searchapp/searchapp: searchapp/app.go
shout/shout: shout/shout.go
qotd/qotd: qotd/qotd.o

ifeq ($(BUILD_IN_CONTAINER),true)
qotd/qotd qotd/qotd.o searchapp/searchapp shout/shout:
	$(SUDO) docker run -ti --rm -v $(shell pwd)/../../:/go/src/github.com/weaveworks/scope -e GOARCH -e GOOS \
		weaveworks/scope-backend-build -C experimental/example $@
else
qotd/qotd:
	gcc -o $@ $< $(CFLAGS)

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

shout/shout searchapp/searchapp:
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)
endif

qotd.marker: qotd/* qotd/qotd
app.marker: app/*
client.marker: client/*
searchapp.marker: searchapp/* searchapp/searchapp
shout.marker: shout/* shout/shout
frontend.marker: frontend/*
echo.marker: echo/*
%.marker:
	docker build -t tomwilkie/$(<D) $(<D)/
	touch $@

clean:
	rm -f qotd/*.o qotd/qotd searchapp/searchapp shout/shout *.marker
