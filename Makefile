.PHONY: all deps static clean client-lint client-test client-sync

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=sudo
DOCKER_SQUASH=$(shell which docker-squash)
DOCKERHUB_USER=weaveworks
APP_EXE=app/scope-app
PROBE_EXE=probe/scope-probe
FIXPROBE_EXE=experimental/fixprobe/fixprobe
SCOPE_IMAGE=$(DOCKERHUB_USER)/scope
SCOPE_EXPORT=scope.tar
SCOPE_UI_BUILD_EXPORT=scope_ui_build.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
SCOPE_VERSION=$(shell git rev-parse --short HEAD)

all: $(SCOPE_EXPORT)

$(SCOPE_EXPORT): $(APP_EXE) $(PROBE_EXE) docker/*
	@if [ -z '$(DOCKER_SQUASH)' ]; then echo "Please install docker-squash by running 'make deps'." && exit 1; fi
	cp $(APP_EXE) $(PROBE_EXE) docker/
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker save $(SCOPE_IMAGE):latest | sudo $(DOCKER_SQUASH) -t $(SCOPE_IMAGE) | tee $@ | $(SUDO) docker load

$(APP_EXE): app/*.go render/*.go report/*.go xfer/*.go

$(PROBE_EXE): probe/*.go probe/docker/*.go probe/endpoint/*.go probe/host/*.go probe/process/*.go probe/overlay/*.go report/*.go xfer/*.go

$(APP_EXE) $(PROBE_EXE):
	go get -d -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version $(SCOPE_VERSION)" -tags netgo -o $@ ./$(@D)
	@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	        rm $@; \
	        echo "\nYour go standard library was built without the 'netgo' build tag."; \
	        echo "To fix that, run"; \
	        echo "    sudo go clean -i net"; \
	        echo "    sudo go install -tags netgo std"; \
	        false; \
	    }

static: client/build/app.js
	esc -o app/static.go -prefix client/build client/build

client/build/app.js: client/app/scripts/*
	mkdir -p client/build
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) gulp build --release

client-test: client/test/*
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm test

client-lint:
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm run lint

client-sync:
	docker run -ti --net=host -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) gulp sync

$(SCOPE_UI_BUILD_EXPORT): client/Dockerfile client/gulpfile.js client/package.json client/webpack.config.js client/.eslintrc
	docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	docker save $(SCOPE_UI_BUILD_IMAGE):latest > $@

clean:
	go clean ./...
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_EXPORT) $(APP_EXE) $(PROBE_EXE) client/build

deps:
	go get \
		github.com/jwilder/docker-squash \
		github.com/golang/lint/golint \
		github.com/fzipp/gocyclo \
		github.com/mattn/goveralls \
		github.com/mjibson/esc \
		github.com/kisielk/errcheck \
		github.com/aktau/github-release
