.PHONY: run clean

IMAGE=weavescope-http-requests-plugin
UPTODATE=.$(EXE).uptodate

run: $(UPTODATE)
	docker run --rm -it \
	  --privileged --net=host \
	  -v /lib/modules:/lib/modules \
	  -v /usr/src:/usr/src \
	  -v /sys/kernel/debug/:/sys/kernel/debug/ \
	  -v /var/run/scope/plugins:/var/run/scope/plugins \
	  --name $(IMAGE) \
	  $(IMAGE)

$(UPTODATE): Dockerfile
	docker build -t $(IMAGE) .
	touch $@

clean:
	- rm -rf $(UPTODATE)
	- docker rmi $(IMAGE)
