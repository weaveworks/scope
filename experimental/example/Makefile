CC=gcc
CFLAGS=-g -lpthread

all: qotd.marker app.marker client.marker searchapp.marker shout.marker frontend.marker echo.marker

qotd/qotd: qotd/qotd.o
	gcc -o $@ $< $(CFLAGS)

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

searchapp/searchapp: searchapp/app.go
shout/shout: shout/shout.go
shout/shout searchapp/searchapp:
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

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
