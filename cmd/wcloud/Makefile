.PHONY: all clean

all: wcloud

wcloud: *.go
	go get ./$(@D)
	go build -o $@ ./$(@D)

clean:
	rm -rf wcloud
	go clean ./...
