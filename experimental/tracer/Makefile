BUILD_IN_CONTAINER=true

tracer.tar: main/main main/Dockerfile
	docker build -t weaveworks/tracer main/
	docker save weaveworks/tracer:latest >$@

main/main: main/*.go main/static.go ptrace/*.go

ifeq ($(BUILD_IN_CONTAINER),true)

main/main:
	docker run -ti \
		-v $(shell pwd)/../../:/go/src/github.com/weaveworks/scope \
		-e GOARCH -e GOOS -e CIRCLECI -e CIRCLE_BUILD_NUM -e CIRCLE_NODE_TOTAL \
		-e CIRCLE_NODE_INDEX -e COVERDIR -e SLOW \
		weaveworks/scope-backend-build SCOPE_VERSION=$(SCOPE_VERSION) GO_BUILD_INSTALL_DEPS=$(GO_BUILD_INSTALL_DEPS) -C experimental/tracer $@

else

main/main:
	#go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

endif

main/static.go: ui/*
	esc -o main/static.go -prefix ui ui

clean:
	go clean ./..
	rm -f main/static.go tracer.tar main/main
