CC=gcc
CFLAGS=-g -lpthread

%.o: %.c
	$(CC) -c -o $@ $< $(CFLAGS)

all: qotd/qotd goapp/app

qotd/qotd: qotd/qotd.o
	gcc -o $@ $< $(CFLAGS)

goapp/app: goapp/app.go
	go build -o goapp/app ./goapp

clean:
	rm -f qotd/*.o qotd/qotd goapp/app
