.PHONY: all clean

all: runner

runner: *.go
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -linkmode=external" -tags netgo -o $@ ./$(@D)

clean:
	rm -rf runner
	go clean ./...
