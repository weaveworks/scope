.PHONY: run clean

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=volume-count.py
IMAGE=weavescope-volume-count-plugin
UPTODATE=.$(EXE).uptodate

run: $(UPTODATE)
	$(SUDO) docker run --rm -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /var/run/scope/plugins:/var/run/scope/plugins \
		--name $(IMAGE) $(IMAGE)

$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE) .
	touch $@

clean:
	- rm -rf $(UPTODATE)
	- $(SUDO) docker rmi $(IMAGE)
