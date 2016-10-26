.PHONY: all deps static clean client-lint client-test client-sync backend frontend shell lint ui-upload

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
DOCKERHUB_USER=weaveworks
SCOPE_EXE=prog/scope
SCOPE_IMAGE=$(DOCKERHUB_USER)/scope
SCOPE_EXPORT=scope.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
SCOPE_UI_BUILD_UPTODATE=.scope_ui_build.uptodate
SCOPE_BACKEND_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-backend-build
SCOPE_BACKEND_BUILD_UPTODATE=.scope_backend_build.uptodate
SCOPE_VERSION=$(shell git rev-parse --short HEAD)
DOCKER_VERSION=1.6.2
DOCKER_DISTRIB=.pkg/docker-$(DOCKER_VERSION).tgz
DOCKER_DISTRIB_URL=https://get.docker.com/builds/Linux/x86_64/docker-$(DOCKER_VERSION).tgz
RUNSVINIT=vendor/runsvinit/runsvinit
CODECGEN_DIR=vendor/github.com/ugorji/go/codec/codecgen
CODECGEN_EXE=$(CODECGEN_DIR)/bin/codecgen_$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)
GET_CODECGEN_DEPS=$(shell find $(1) -maxdepth 1 -type f -name '*.go' -not -name '*_test.go' -not -name '*.codecgen.go' -not -name '*.generated.go')
CODECGEN_TARGETS=report/report.codecgen.go render/detailed/detailed.codecgen.go
RM=--rm
RUN_FLAGS=-ti
BUILD_IN_CONTAINER=true
GO_ENV=GOGC=off
GO=env $(GO_ENV) go
NO_CROSS_COMP=unset GOOS GOARCH
GO_HOST=$(NO_CROSS_COMP); $(GO)
WITH_GO_HOST_ENV=$(NO_CROSS_COMP); $(GO_ENV)
GO_BUILD_INSTALL_DEPS=-i
GO_BUILD_TAGS='netgo unsafe'
GO_BUILD_FLAGS=$(GO_BUILD_INSTALL_DEPS) -ldflags "-extldflags \"-static\" -X main.version=$(SCOPE_VERSION) -s -w" -tags $(GO_BUILD_TAGS)
IMAGE_TAG=$(shell ./tools/image-tag)

all: $(SCOPE_EXPORT)

$(DOCKER_DISTRIB):
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)

docker/weave:
	curl -L git.io/weave -o docker/weave
	chmod u+x docker/weave

$(SCOPE_EXPORT): $(SCOPE_EXE) $(DOCKER_DISTRIB) docker/weave $(RUNSVINIT) docker/Dockerfile docker/demo.json docker/run-app docker/run-probe docker/entrypoint.sh
	cp $(SCOPE_EXE) $(RUNSVINIT) docker/
	cp $(DOCKER_DISTRIB) docker/docker.tgz
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker tag $(SCOPE_IMAGE) $(SCOPE_IMAGE):$(IMAGE_TAG)
	$(SUDO) docker save $(SCOPE_IMAGE):latest > $@

$(RUNSVINIT): vendor/runsvinit/*.go

$(SCOPE_EXE): $(shell find ./ -path ./vendor -prune -o -type f -name *.go) prog/staticui/staticui.go prog/externalui/externalui.go $(CODECGEN_TARGETS)

report/report.codecgen.go: $(call GET_CODECGEN_DEPS,report/)
render/render.codecgen.go: $(call GET_CODECGEN_DEPS,render/)
render/detailed/detailed.codecgen.go: $(call GET_CODECGEN_DEPS,render/detailed/)
static: prog/staticui/staticui.go prog/externalui/externalui.go
prog/staticui/staticui.go: client/build/index.html
prog/externalui/externalui.go: client/build-external/index.html

ifeq ($(BUILD_IN_CONTAINER),true)

$(SCOPE_EXE) $(RUNSVINIT) lint tests shell prog/staticui/staticui.go prog/externalui/externalui.go: $(SCOPE_BACKEND_BUILD_UPTODATE)
	@mkdir -p $(shell pwd)/.pkg
	$(SUDO) docker run $(RM) $(RUN_FLAGS) \
		-v $(shell pwd):/go/src/github.com/weaveworks/scope \
		-v $(shell pwd)/.pkg:/go/pkg \
		--net=host \
		-e GOARCH -e GOOS -e CIRCLECI -e CIRCLE_BUILD_NUM -e CIRCLE_NODE_TOTAL \
		-e CIRCLE_NODE_INDEX -e COVERDIR -e SLOW -e TESTDIRS \
		$(SCOPE_BACKEND_BUILD_IMAGE) SCOPE_VERSION=$(SCOPE_VERSION) GO_BUILD_INSTALL_DEPS=$(GO_BUILD_INSTALL_DEPS) $@

else

$(SCOPE_EXE): $(SCOPE_BACKEND_BUILD_UPTODATE)
	time $(GO) build $(GO_BUILD_FLAGS) -o $@ ./$(@D)
	@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	        rm $@; \
	        echo "\nYour go standard library was built without the 'netgo' build tag."; \
	        echo "To fix that, run"; \
	        echo "    sudo go clean -i net"; \
	        echo "    sudo go install -tags netgo std"; \
	        false; \
	    }

%.codecgen.go: $(CODECGEN_EXE)
	rm -f $@; $(GO_HOST) build $(GO_BUILD_FLAGS) ./$(@D) # workaround for https://github.com/ugorji/go/issues/145
	cd $(@D) && $(WITH_GO_HOST_ENV) $(shell pwd)/$(CODECGEN_EXE) -rt $(GO_BUILD_TAGS) -u -o $(@F) $(notdir $(call GET_CODECGEN_DEPS,$(@D)))

$(CODECGEN_EXE): $(CODECGEN_DIR)/*.go
	mkdir -p $(@D)
	$(GO_HOST) build $(GO_BUILD_FLAGS) -o $@ ./$(CODECGEN_DIR)

$(RUNSVINIT): $(SCOPE_BACKEND_BUILD_UPTODATE)
	time $(GO) build $(GO_BUILD_FLAGS) -o $@ ./$(@D)

shell: $(SCOPE_BACKEND_BUILD_UPTODATE)
	/bin/bash

tests: $(SCOPE_BACKEND_BUILD_UPTODATE)
	./tools/test -no-go-get

lint: $(SCOPE_BACKEND_BUILD_UPTODATE)
	./tools/lint -ignorespelling "agre " -ignorespelling "AGRE " .

prog/staticui/staticui.go: $(SCOPE_BACKEND_BUILD_UPTODATE)
	mkdir -p prog/staticui
	esc -o $@ -pkg staticui -prefix client/build client/build

prog/externalui/externalui.go: $(SCOPE_BACKEND_BUILD_UPTODATE)
	mkdir -p prog/externalui
	esc -o $@ -pkg externalui -prefix client/build-external -include '\.html$$' client/build-external

endif

ifeq ($(BUILD_IN_CONTAINER),true)

client/build/index.html: $(shell find client/app -type f) $(SCOPE_UI_BUILD_UPTODATE)
	mkdir -p client/build
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build \
		$(SCOPE_UI_BUILD_IMAGE) npm run build

client/build-external/index.html: $(shell find client/app -type f) $(SCOPE_UI_BUILD_UPTODATE)
	mkdir -p client/build
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build-external:/home/weave/build-external \
		$(SCOPE_UI_BUILD_IMAGE) npm run build-external

client-test: $(shell find client/app/scripts -type f) $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm test

client-lint: $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) npm run lint

client-start: $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) --net=host -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build -e WEBPACK_SERVER_HOST \
		$(SCOPE_UI_BUILD_IMAGE) npm start

else

client/build/index.html:
	cd client && npm run build

client/build-external/index.html:
	cd client && npm run build-external

endif

$(SCOPE_UI_BUILD_UPTODATE): client/Dockerfile client/package.json client/webpack.local.config.js client/webpack.production.config.js client/webpack.external.config.js client/server.js client/.eslintrc
	$(SUDO) docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	touch $@

$(SCOPE_BACKEND_BUILD_UPTODATE): backend/*
	$(SUDO) docker build -t $(SCOPE_BACKEND_BUILD_IMAGE) backend
	touch $@

ui-upload: client/build-external/index.html
	AWS_ACCESS_KEY_ID=$$UI_BUCKET_KEY_ID \
	AWS_SECRET_ACCESS_KEY=$$UI_BUCKET_KEY_SECRET \
	aws s3 cp client/build-external/ s3://static.weave.works/scope-ui/ --recursive --exclude '*.html'

clean:
	$(GO) clean ./...
	# Don't actually rmi the build images - rm'ing the .uptodate files is enough to ensure
	# we rebuild the images, and rmi'ing the images causes us to have to redownload a lot of stuff.
	# $(SUDO) docker rmi $(SCOPE_UI_BUILD_IMAGE) $(SCOPE_BACKEND_BUILD_IMAGE) >/dev/null 2>&1 || true
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_UPTODATE) $(SCOPE_BACKEND_BUILD_UPTODATE) \
		$(SCOPE_EXE) $(RUNSVINIT) prog/staticui/staticui.go prog/externalui/externalui.go client/build/*.js client/build-external/*.js docker/weave .pkg \
		$(CODECGEN_TARGETS) $(CODECGEN_DIR)/bin

clean-codecgen:
	rm -rf $(CODECGEN_TARGETS) $(CODECGEN_DIR)/bin

# Dependencies are intentionally build without enforcing any tags
# since they are build on the host
deps:
	$(GO) get -u -f \
		github.com/FiloSottile/gvt \
		github.com/mattn/goveralls \
		github.com/weaveworks/github-release \
		github.com/2opremio/trifles/wscat
