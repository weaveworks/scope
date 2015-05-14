tracer.tar: main/main main/Dockerfile
	docker build -t weaveworks/tracer main/
	docker save weaveworks/tracer:latest >$@

main/main: main/*.go main/static.go ptrace/*.go
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version $(SCOPE_VERSION)" -tags netgo -o $@ ./$(@D)

main/static.go: ui/*
	esc -o main/static.go -prefix ui ui

clean:
	go clean ./..
	rm -f main/static.go tracer.tar main/main
