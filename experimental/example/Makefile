CC=gcc
CFLAGS=-g -lpthread

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

all: qotd/qotd goapp/app shout/shout

qotd/qotd: qotd/qotd.o
	gcc -o $@ $< $(CFLAGS)

goapp/app: goapp/app.go
shout/shout: shout/shout.go

shout/shout goapp/app:
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

clean:
	rm -f qotd/*.o qotd/qotd goapp/app

run: all
	docker-compose kill  || true
	docker-compose rm -f || true
	docker-compose build
	docker-compose up -d
