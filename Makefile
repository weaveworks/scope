.PHONY: all deps static clean client-lint client-test client-sync backend frontend

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=sudo
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
RUNSVINIT=vendor/runsvinit/runsvinit
RM=--rm
BUILD_IN_CONTAINER=true

all: $(SCOPE_EXPORT)

$(DOCKER_DISTRIB):
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)

docker/weave:
	curl -L git.io/weave -o docker/weave
	chmod u+x docker/weave

$(SCOPE_EXPORT): $(APP_EXE) $(PROBE_EXE) $(DOCKER_DISTRIB) docker/weave $(RUNSVINIT) docker/Dockerfile docker/run-app docker/run-probe docker/entrypoint.sh
	cp $(APP_EXE) $(PROBE_EXE) $(RUNSVINIT) docker/
	cp $(DOCKER_DISTRIB) docker/docker.tgz
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker save $(SCOPE_IMAGE):latest > $@

$(RUNSVINIT): vendor/runsvinit/*.go

$(APP_EXE): app/*.go render/*.go report/*.go xfer/*.go common/sanitize/*.go

$(PROBE_EXE): probe/*.go probe/docker/*.go probe/kubernetes/*.go probe/endpoint/*.go probe/host/*.go probe/process/*.go probe/overlay/*.go report/*.go xfer/*.go common/sanitize/*.go common/exec/*.go

ifeq ($(BUILD_IN_CONTAINER),true)
$(APP_EXE) $(PROBE_EXE) $(RUNSVINIT): $(SCOPE_BACKEND_BUILD_UPTODATE)
	$(SUDO) docker run -ti $(RM) -v $(shell pwd):/go/src/github.com/weaveworks/scope -e GOARCH -e GOOS \
		$(SCOPE_BACKEND_BUILD_IMAGE) $@
else
$(APP_EXE) $(PROBE_EXE):
	go build -ldflags "-extldflags \"-static\" -X main.version $(SCOPE_VERSION)" -tags netgo -o $@ ./$(@D)
	@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	        rm $@; \
	        echo "\nYour go standard library was built without the 'netgo' build tag."; \
	        echo "To fix that, run"; \
	        echo "    sudo go clean -i net"; \
	        echo "    sudo go install -tags netgo std"; \
	        false; \
	    }

$(RUNSVINIT):
	go build -ldflags "-extldflags \"-static\"" -o $@ ./$(@D)
endif

static: client/build/app.js
	esc -o app/static.go -prefix client/build client/build

ifeq ($(BUILD_IN_CONTAINER),true)
client/build/app.js: client/app/scripts/*
	mkdir -p client/build
	$(SUDO) docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) npm run build

client-test: client/test/*
	$(SUDO) docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm test

client-lint:
	$(SUDO) docker run -ti $(RM) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm run lint

client-start:
	$(SUDO) docker run -ti $(RM) --net=host -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) npm start
endif

$(SCOPE_UI_BUILD_UPTODATE): client/Dockerfile client/package.json client/webpack.local.config.js client/webpack.production.config.js client/server.js client/.eslintrc
	$(SUDO) docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	touch $@

$(SCOPE_BACKEND_BUILD_UPTODATE): backend/*
	$(SUDO) docker build -t $(SCOPE_BACKEND_BUILD_IMAGE) backend
	touch $@

frontend: $(SCOPE_UI_BUILD_UPTODATE)

clean:
	go clean ./...
	$(SUDO) docker rmi $(SCOPE_UI_BUILD_IMAGE) $(SCOPE_BACKEND_BUILD_IMAGE) >/dev/null 2>&1 || true
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_UPTODATE) $(SCOPE_BACKEND_BUILD_UPTODATE) \
		$(APP_EXE) $(PROBE_EXE) $(RUNSVINIT) client/build/app.js docker/weave

deps:
	go get -u -f -tags netgo \
		github.com/golang/lint/golint \
		github.com/fzipp/gocyclo \
		github.com/mattn/goveralls \
		github.com/mjibson/esc \
		github.com/kisielk/errcheck \
		github.com/weaveworks/github-release
