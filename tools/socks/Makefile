.PHONY: all clean

IMAGE_TAR=image.tar
IMAGE_NAME=weaveworks/socksproxy
GIT_REVISION := $(shell git rev-parse HEAD)
PROXY_EXE=proxy
NETGO_CHECK=@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	rm $@; \
	echo "\nYour go standard library was built without the 'netgo' build tag."; \
	echo "To fix that, run"; \
	echo "    sudo go clean -i net"; \
	echo "    sudo go install -tags netgo std"; \
	false; \
}

all: $(IMAGE_TAR)

$(IMAGE_TAR): Dockerfile $(PROXY_EXE)
	docker build --build-arg=revision=$(GIT_REVISION) -t $(IMAGE_NAME) .
	docker save $(IMAGE_NAME):latest > $@

$(PROXY_EXE): *.go
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)
	$(NETGO_CHECK)

clean:
	-docker rmi $(IMAGE_NAME)
	rm -rf $(PROXY_EXE) $(IMAGE_TAR)
	go clean ./...
