.PHONY: all deps static clean

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=sudo
DOCKER_SQUASH=$(shell which docker-squash)
DOCKERHUB_USER=weaveworks
APP_EXE=app/app
PROBE_EXE=probe/probe
FIXPROBE_EXE=experimental/fixprobe/fixprobe
SCOPE_IMAGE=$(DOCKERHUB_USER)/scope
SCOPE_EXPORT=scope.tar
SCOPE_UI_BUILD_EXPORT=scope_ui_build.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
GIT_REVISION=$(shell git rev-parse HEAD)

all: $(SCOPE_EXPORT)

$(SCOPE_EXPORT): $(APP_EXE) $(PROBE_EXE) docker/*
	@if [ -z '$(DOCKER_SQUASH)' ]; then echo "Please install docker-squash by running 'make dep'." && exit 1; fi
	cp $(APP_EXE) $(PROBE_EXE) docker/
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker save $(SCOPE_IMAGE):latest | sudo $(DOCKER_SQUASH) -t $(SCOPE_IMAGE) | tee $@ | $(SUDO) docker load
	@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	        rm $@; \
	        echo "\nYour go standard library was built without the 'netgo' build tag."; \
	        echo "To fix that, run"; \
	        echo "    sudo go clean -i net"; \
	        echo "    sudo go install -tags netgo std"; \
	        false; \
	    }

$(APP_EXE): app/*.go report/*.go xfer/*.go

$(PROBE_EXE): probe/*.go report/*.go xfer/*.go

$(APP_EXE) $(PROBE_EXE):
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version $(GIT_REVISION)" -tags netgo -o $@ ./$(@D)

static: client/dist/scripts/bundle.js
	esc -o app/static.go -prefix client/dist client/dist

client/dist/scripts/bundle.js: client/app/scripts/*
	mkdir -p client/dist
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/dist:/home/weave/dist \
		$(SCOPE_UI_BUILD_IMAGE) gulp build

client-test: client/test/*
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm test

$(SCOPE_UI_BUILD_EXPORT): client/Dockerfile client/gulpfile.js client/package.json
	docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	docker save $(SCOPE_UI_BUILD_IMAGE):latest > $@

clean:
	go clean ./...
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_EXPORT) client/dist

deps:
	go get \
		github.com/jwilder/docker-squash \
		github.com/golang/lint/golint \
		github.com/fzipp/gocyclo \
		github.com/mattn/goveralls \
		github.com/mjibson/esc \
		github.com/kisielk/errcheck
