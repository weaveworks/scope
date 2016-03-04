.PHONY: run clean

EXE=iowait
IMAGE=weavescope-iowait-plugin
UPTODATE=.$(EXE).uptodate

run: $(UPTODATE)
	docker run --rm -it --privileged -v /var/run/scope/plugins:/var/run/scope/plugins --name $(IMAGE) $(IMAGE) -hostname=$(shell hostname)

$(UPTODATE): $(EXE) Dockerfile
	docker build -t $(IMAGE) .
	touch $@

$(EXE): main.go
	docker run --rm -v "$$PWD":/usr/src/$(EXE) -w /usr/src/$(EXE) golang:1.6 go build -v

clean:
	- rm -rf $(UPTODATE) $(EXE)
	- docker rmi $(IMAGE)
