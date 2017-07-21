.PHONY: all deps static clean client-lint client-test client-sync backend frontend shell lint ui-upload

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
DOCKERHUB_USER=weaveworks
SCOPE_EXE=prog/scope
SCOPE_EXPORT=scope.tar
CLOUD_AGENT_EXPORT=cloud-agent.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
SCOPE_UI_BUILD_UPTODATE=.scope_ui_build.uptodate
SCOPE_BACKEND_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-backend-build
SCOPE_BACKEND_BUILD_UPTODATE=.scope_backend_build.uptodate
SCOPE_VERSION=$(shell git rev-parse --short HEAD)
WEAVENET_VERSION=1.9.0
DOCKER_VERSION=1.13.1
DOCKER_DISTRIB=.pkg/docker-$(DOCKER_VERSION).tgz
DOCKER_DISTRIB_URL=https://get.docker.com/builds/Linux/x86_64/docker-$(DOCKER_VERSION).tgz
RUNSVINIT=vendor/runsvinit/runsvinit
CODECGEN_DIR=vendor/github.com/ugorji/go/codec/codecgen
CODECGEN_EXE=$(CODECGEN_DIR)/bin/codecgen_$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)
CODECGEN_UID=0
GET_CODECGEN_DEPS=$(shell find $(1) -maxdepth 1 -type f -name '*.go' -not -name '*_test.go' -not -name '*.codecgen.go' -not -name '*.generated.go')
CODECGEN_TARGETS=report/report.codecgen.go render/detailed/detailed.codecgen.go
RM=--rm
RUN_FLAGS=-ti
BUILD_IN_CONTAINER=true
GO_ENV=GOGC=off
GO_BUILD_INSTALL_DEPS=-i
GO_BUILD_TAGS='netgo unsafe'
GO_BUILD_FLAGS=$(GO_BUILD_INSTALL_DEPS) -ldflags "-extldflags \"-static\" -X main.version=$(SCOPE_VERSION) -s -w" -tags $(GO_BUILD_TAGS)
GOOS=$(shell go tool dist env | grep GOOS | sed -e 's/GOOS="\(.*\)"/\1/')

ifeq ($(GOOS),linux)
GO_ENV+=CGO_ENABLED=1
endif

ifeq ($(GOARCH),arm)
ARM_CC=CC=/usr/bin/arm-linux-gnueabihf-gcc
endif

GO=env $(GO_ENV) $(ARM_CC) go

NO_CROSS_COMP=unset GOOS GOARCH
GO_HOST=$(NO_CROSS_COMP); env $(GO_ENV) go
WITH_GO_HOST_ENV=$(NO_CROSS_COMP); $(GO_ENV)
IMAGE_TAG=$(shell ./tools/image-tag)

all: $(SCOPE_EXPORT)

$(DOCKER_DISTRIB):
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)

docker/weave:
	curl -L https://github.com/weaveworks/weave/releases/download/v$(WEAVENET_VERSION)/weave -o docker/weave
	chmod u+x docker/weave

docker/weaveutil:
	$(SUDO) docker run --rm  --entrypoint=cat weaveworks/weaveexec:$(WEAVENET_VERSION) /usr/bin/weaveutil > $@
	chmod +x $@

docker/%: %
	cp $* docker/

docker/docker: $(DOCKER_DISTRIB)
	tar -xvzf $(DOCKER_DISTRIB) docker/docker

%.tar: docker/Dockerfile.%
	$(SUDO) docker build -t $(DOCKERHUB_USER)/$* -f $< docker/
	$(SUDO) docker tag $(DOCKERHUB_USER)/$* $(DOCKERHUB_USER)/$*:$(IMAGE_TAG)
	$(SUDO) docker save $(DOCKERHUB_USER)/$*:latest > $@

$(CLOUD_AGENT_EXPORT): docker/Dockerfile.cloud-agent docker/$(SCOPE_EXE) docker/docker docker/weave docker/weaveutil

$(SCOPE_EXPORT): docker/Dockerfile.scope $(CLOUD_AGENT_EXPORT) docker/$(RUNSVINIT) docker/demo.json docker/run-app docker/run-probe docker/entrypoint.sh

$(RUNSVINIT): vendor/runsvinit/*.go

$(SCOPE_EXE): $(shell find ./ -path ./vendor -prune -o -type f -name '*.go') prog/staticui/staticui.go prog/externalui/externalui.go $(CODECGEN_TARGETS)

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
		$(SCOPE_BACKEND_BUILD_IMAGE) SCOPE_VERSION=$(SCOPE_VERSION) GO_BUILD_INSTALL_DEPS=$(GO_BUILD_INSTALL_DEPS) CODECGEN_UID=$(CODECGEN_UID) $@

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
	cd $(@D) && $(WITH_GO_HOST_ENV) $(shell pwd)/$(CODECGEN_EXE) -d $(CODECGEN_UID) -rt $(GO_BUILD_TAGS) -u -o $(@F) $(notdir $(call GET_CODECGEN_DEPS,$(@D)))

$(CODECGEN_EXE): $(CODECGEN_DIR)/*.go
	mkdir -p $(@D)
	$(GO_HOST) build $(GO_BUILD_FLAGS) -o $@ ./$(CODECGEN_DIR)

$(RUNSVINIT): $(SCOPE_BACKEND_BUILD_UPTODATE)
	time $(GO) build $(GO_BUILD_FLAGS) -o $@ ./$(@D)

shell: $(SCOPE_BACKEND_BUILD_UPTODATE)
	/bin/bash

tests: $(SCOPE_BACKEND_BUILD_UPTODATE) $(CODECGEN_TARGETS) prog/staticui/staticui.go prog/externalui/externalui.go
	./tools/test -no-go-get

lint: $(SCOPE_BACKEND_BUILD_UPTODATE)
	./tools/lint
	./tools/shell-lint tools

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
	if test "true" != "$(SCOPE_SKIP_UI_ASSETS)"; then \
		$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
			-v $(shell pwd)/client/build:/home/weave/build \
			$(SCOPE_UI_BUILD_IMAGE) yarn run build; \
	fi

client/build-external/index.html: $(shell find client/app -type f) $(SCOPE_UI_BUILD_UPTODATE)
	mkdir -p client/build-external
	if test "true" != "$(SCOPE_SKIP_UI_ASSETS)"; then \
		$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
			-v $(shell pwd)/client/build-external:/home/weave/build-external \
			$(SCOPE_UI_BUILD_IMAGE) yarn run build-external; \
	fi

client-test: $(shell find client/app/scripts -type f) $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) yarn test

client-lint: $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/test:/home/weave/test \
		$(SCOPE_UI_BUILD_IMAGE) yarn run lint

client-start: $(SCOPE_UI_BUILD_UPTODATE)
	$(SUDO) docker run $(RM) $(RUN_FLAGS) --net=host -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/build:/home/weave/build -e WEBPACK_SERVER_HOST \
		$(SCOPE_UI_BUILD_IMAGE) yarn start

tmp/weave-scope.tgz: $(shell find client/app -type f) $(SCOPE_UI_BUILD_UPTODATE)
	$(sudo) docker run $(RUN_FLAGS) \
	-v $(shell pwd)/client/app:/home/weave/app \
	-v $(shell pwd)/tmp:/home/weave/tmp \
	$(SCOPE_UI_BUILD_IMAGE) \
	yarn run bundle

else

client/build/index.html:
	mkdir -p client/build
	if test "true" != "$(SCOPE_SKIP_UI_ASSETS)"; then cd client && yarn run build; fi

client/build-external/index.html:
	mkdir -p client/build-external
	if test "true" != "$(SCOPE_SKIP_UI_ASSETS)"; then cd client && yarn run build-external; fi

endif

$(SCOPE_UI_BUILD_UPTODATE): client/Dockerfile client/package.json client/webpack.local.config.js client/webpack.production.config.js client/server.js client/.eslintrc
	$(SUDO) docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	touch $@

$(SCOPE_BACKEND_BUILD_UPTODATE): backend/*
	$(SUDO) docker build -t $(SCOPE_BACKEND_BUILD_IMAGE) backend
	touch $@

ui-upload: client/build-external/index.html
	AWS_ACCESS_KEY_ID=$$UI_BUCKET_KEY_ID \
	AWS_SECRET_ACCESS_KEY=$$UI_BUCKET_KEY_SECRET \
	aws s3 cp client/build-external/ s3://static.weave.works/scope-ui/ --recursive --exclude '*.html'

ui-pkg-upload: tmp/weave-scope.tgz
	AWS_ACCESS_KEY_ID=$$UI_BUCKET_KEY_ID \
	AWS_SECRET_ACCESS_KEY=$$UI_BUCKET_KEY_SECRET \
	aws s3 cp tmp/weave-scope.tgz s3://weaveworks-js-modules/weave-scope/$(shell echo $(SCOPE_VERSION))/weave-scope.tgz

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

# This target is only intended for use in Netlify CI environment for generating preview pages on feature branches and pull requests.
# We need to obtain website repository (checked out under `site-build`) and place `site` directory into the context (`site-build/_weave_net_docs`).
# We then run make in `site-build` and Netlify will publish the output (`site-build/_site`).
netlify-site-preview:
	@mkdir -p site-build
	@curl --user $(WEBSITE_GITHUB_USER) --silent 'https://codeload.github.com/weaveworks/website-next/tar.gz/$(WEBSITE_BRANCH)' \
	  | tar --extract --gunzip --directory site-build --strip 1
	@cp -r site site-build/_weave_scope_docs
	@$(MAKE) -C site-build netlify_ensure_install
	@$(MAKE) -C site-build BUILD_ENV=netlify
