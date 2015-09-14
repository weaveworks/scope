.PHONY: all deps static clean client-lint client-test client-sync backend frontend

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=sudo
DOCKER_SQUASH=$(shell which docker-squash 2>/dev/null)
DOCKERHUB_USER=weaveworks
APP_EXE=app/scope-app
PROBE_EXE=probe/scope-probe
FIXPROBE_EXE=experimental/fixprobe/fixprobe
SCOPE_IMAGE=$(DOCKERHUB_USER)/scope
SCOPE_EXPORT=scope.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
SCOPE_UI_BUILD_UPTODATE=.scope_ui_build.uptodate
SCOPE_BACKEND_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-backend-build
SCOPE_BACKEND_BUILD_UPTODATE=.scope_backend_build.uptodate
SCOPE_VERSION=$(shell git rev-parse --short HEAD)
DOCKER_VERSION=1.3.1
DOCKER_DISTRIB=docker/docker-$(DOCKER_VERSION).tgz
DOCKER_DISTRIB_URL=https://get.docker.com/builds/Linux/x86_64/docker-$(DOCKER_VERSION).tgz
RM=--rm

all: $(SCOPE_EXPORT)

$(DOCKER_DISTRIB):
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)

docker/weave:
	curl -L git.io/weave -o docker/weave
	chmod u+x docker/weave

$(SCOPE_EXPORT): $(APP_EXE) $(PROBE_EXE) $(DOCKER_DISTRIB) docker/weave docker/entrypoint.sh docker/Dockerfile docker/run-app docker/run-probe
	@if [ -z '$(DOCKER_SQUASH)' ] ; then echo "Please install docker-squash by running 'make deps' (and make sure GOPATH/bin is in your PATH)." && exit 1 ; fi
	cp $(APP_EXE) $(PROBE_EXE) docker/
	cp $(DOCKER_DISTRIB) docker/docker.tgz
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker save $(SCOPE_IMAGE):latest | sudo $(DOCKER_SQUASH) -t $(SCOPE_IMAGE) | tee $@ | $(SUDO) docker load

$(APP_EXE): app/*.go render/*.go report/*.go xfer/*.go

$(PROBE_EXE): probe/*.go probe/docker/*.go probe/endpoint/*.go probe/host/*.go probe/process/*.go probe/overlay/*.go report/*.go xfer/*.go

$(APP_EXE) $(PROBE_EXE):
	go get -d -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\" -X main.version=$(SCOPE_VERSION)" -tags netgo -o $@ ./$(@D)
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
	docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) npm run build

client-test: client/test/*
	docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm test

client-lint:
	docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm run lint

client-start:
	docker run -ti $(RM) --net=host -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) npm start

$(SCOPE_UI_BUILD_UPTODATE): client/Dockerfile client/package.json client/webpack.local.config.js client/webpack.production.config.js client/server.js client/.eslintrc
	docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	touch $@

$(SCOPE_BACKEND_BUILD_UPTODATE): backend/*
	docker build -t $(SCOPE_BACKEND_BUILD_IMAGE) backend
	touch $@

backend: $(SCOPE_BACKEND_BUILD_UPTODATE)
	docker run -ti $(RM) -v $(shell pwd):/go/src/github.com/weaveworks/scope $(SCOPE_BACKEND_BUILD_IMAGE) /build.bash

frontend: $(SCOPE_UI_BUILD_UPTODATE)

clean:
	go clean ./...
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_EXPORT) $(APP_EXE) $(PROBE_EXE) client/build/app.js

deps:
	go get -tags netgo \
		github.com/jwilder/docker-squash \
		github.com/golang/lint/golint \
		github.com/fzipp/gocyclo \
		github.com/mattn/goveralls \
		github.com/mjibson/esc \
		github.com/kisielk/errcheck \
		github.com/aktau/github-release
