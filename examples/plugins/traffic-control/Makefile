.PHONY: run clean

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=traffic-control
IMAGE=weavescope-$(EXE)-plugin
UPTODATE=.$(EXE).uptodate

run: $(UPTODATE)
	# --net=host gives us the remote hostname, in case we're being launched against a non-local docker host.
	# We could also pass in the `-hostname=foo` flag, but that doesn't work against a remote docker host.
	$(SUDO) docker run --rm -it \
		--net=host \
		--pid=host \
		--privileged \
		-v /var/run:/var/run \
		--name $(IMAGE) $(IMAGE)

$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE) .
	touch $@

$(EXE): main.go
	$(SUDO) docker run --rm \
		-v "$$PWD":/go/src/hosting/org/$(EXE) \
		-v $(shell pwd)/../../../vendor:/go/src/hosting/org/$(EXE)/vendor \
		-w /go/src/hosting/org/$(EXE) \
		golang:1.6 go build -v


clean:
	- rm -rf $(UPTODATE) $(EXE)
	- $(SUDO) docker rmi $(IMAGE)
