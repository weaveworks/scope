.PHONY: all deps static clean client-lint client-test client-sync backend frontend shell lint ui-upload

# This specifies the architecture we're building for
# Use GOARCH, note this name differs from Linux ARCH
ARCH?=$(shell go env GOARCH)

# A list of all supported architectures here. Should be named as Go is naming platforms
# All supported architectures must have an "ifeq" block below that customizes the parameters
ALL_ARCHITECTURES=amd64 arm64
ML_PLATFORMS=linux/amd64,linux/arm64

ifeq ($(ARCH),amd64)
# The architecture to use when downloading the docker binary
	WEAVEEXEC_DOCKER_ARCH?=x86_64
# The Docker version to use when downloading the docker binary
	DOCKER_VERSION=1.13.1	
# Where to place/find Docker binary
	DOCKER_DISTRIB=.pkg/docker-$(DOCKER_VERSION).tgz
# Where to het Docker binary if arch supported by Docker
	DOCKER_DISTRIB_URL=https://get.docker.com/builds/Linux/x86_64/docker-$(DOCKER_VERSION).tgz
	
# The name of the alpine baseimage to use as the base for images:
#	docker/Dockerfile-cloud-agent
	ALPINE_BASEIMAGE?=alpine:3.4

# The name of the node baseimage to use as the base for images:
#	client/Dockerfile
	NODE_BASEIMAGE?=node:6.9.0

# The name of the Ubuntu baseimage to use as the base for images:
#	backend/Dockerfile
	UBUNTU_BASEIMAGE?=ubuntu:yakkety
endif

ifeq ($(ARCH),arm64)
# Use the host installed Docker binary for now
	WEAVEEXEC_DOCKER_ARCH?=aarch64
# The Docker version to use when downloading the docker binary
	DOCKER_VERSION=1.12.6
# Where to place/find Docker binary
	DOCKER_DISTRIB=.pkg/docker
# Where to het Docker binary if arch supported by Docker
	DOCKER_DISTRIB_URL=
	
# Using the (semi-)official alpine image
#	docker/Dockerfile-cloud-agent
	ALPINE_BASEIMAGE?=aarch64/alpine:3.5
	
# The name of the node baseimage to use as the base for images:
#	client/Dockerfile
	NODE_BASEIMAGE?=aarch64/node:6.9

# The name of the Ubuntu baseimage to use as the base for images:
#	backend/Dockerfile
	UBUNTU_BASEIMAGE?=aarch64/ubuntu:yakkety
endif




# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
# The name of the user that this Makefile should produce image artifacts for. Can/should be overridden
DOCKERHUB_USER?=weaveworks

SCOPE_EXE=prog/scope
SCOPE_EXPORT=scope.tar
CLOUD_AGENT_EXPORT=cloud-agent.tar
SCOPE_UI_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-ui-build
SCOPE_UI_BUILD_UPTODATE=.scope_ui_build.uptodate
SCOPE_BACKEND_BUILD_IMAGE=$(DOCKERHUB_USER)/scope-backend-build
SCOPE_BACKEND_BUILD_UPTODATE=.scope_backend_build.uptodate
SCOPE_VERSION=$(shell git rev-parse --short HEAD)
WEAVENET_VERSION=1.9.0
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

GO_HOST=env $(GO_ENV) go
WITH_GO_HOST_ENV=$(GO_ENV)
IMAGE_TAG=$(shell ./tools/image-tag)

GO=env $(GO_ENV) go

all: $(SCOPE_EXPORT)


# Creates the Dockerfile.your-user-here file from the templates
# Also replaces all placeholders with real values
docker/Dockerfile.cloud-agent.$(DOCKERHUB_USER): docker/Dockerfile.cloud-agent.template
	echo "DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|ALPINE_BASEIMAGE|$(ALPINE_BASEIMAGE)|g"
	sed -e "s|DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|ALPINE_BASEIMAGE|$(ALPINE_BASEIMAGE)|g" $^ > $@

docker/Dockerfile.scope.$(DOCKERHUB_USER): docker/Dockerfile.scope.template
	echo "DOCKERHUB_USER|$(DOCKERHUB_USER)|g"
	sed -e "s|DOCKERHUB_USER|$(DOCKERHUB_USER)|g" $^ > $@

client/Dockerfile.$(DOCKERHUB_USER): client/Dockerfile.template
	echo "DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|NODE_BASEIMAGE|$(NODE_BASEIMAGE)|g"
	sed -e "s|DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|NODE_BASEIMAGE|$(NODE_BASEIMAGE)|g" $^ > $@
	
backend/Dockerfile.$(DOCKERHUB_USER): backend/Dockerfile.template
	echo "DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|UBUNTU_BASEIMAGE|$(UBUNTU_BASEIMAGE)|g"
	sed -e "s|DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|UBUNTU_BASEIMAGE|$(UBUNTU_BASEIMAGE)|g" $^ > $@
ifeq ($(ARCH),amd64)
# only the placeholder "RACE_SUPPORT" should be removed
	sed -i "s/RACE_SUPPORT//g" $@
# only the placeholder "CURL_SHFM" should be removed
	sed -i "s/CURL_SHFMT//g" $@
	sed -i "s/CURL_SHFMT1//g" $@	
# Whole line with the placeholder "BUILD_SHFMT" should be removed
	sed -i "/BUILD_SHFMT/d" $@
	sed -i "/BUILD_SHFMT1/d" $@
endif
ifeq ($(ARCH),arm64)
# Whole line with the placeholder "RACE_SUPPORT" should be removed
	sed -i "/RACE_SUPPORT/d" $@
# only the placeholder "BUILD_SHFMT" should be removed
	sed -i "s/BUILD_SHFMT//g" $@
	sed -i "s/BUILD_SHFMT1//g" $@	
# Whole line with the placeholder "CURL_SHFMT RUN" should be removed
	sed -i "/CURL_SHFMT/d" $@
	sed -i "/CURL_SHFMT1/d" $@
endif


$(DOCKER_DISTRIB):
ifeq ($(ARCH),arm64)
# When Docker supports aarch64, fix up
	cp /usr/bin/docker $(DOCKER_DISTRIB)
endif
ifeq ($(ARCH),amd64)
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)
endif


docker/weave:
	curl -L https://github.com/weaveworks/weave/releases/download/v$(WEAVENET_VERSION)/weave -o docker/weave
	chmod u+x docker/weave

docker/weaveutil:
	$(SUDO) docker run --rm  --entrypoint=cat weaveworks/weaveexec:$(WEAVENET_VERSION) /usr/bin/weaveutil > $@
	chmod +x $@

docker/%: %
	cp $* docker/

docker/docker: $(DOCKER_DISTRIB)
ifeq ($(ARCH),arm64)
# When Docker supports aarch64, fix up
	cp $(DOCKER_DISTRIB) docker/docker 
endif
ifeq ($(ARCH),amd64)
	tar -xvzf $(DOCKER_DISTRIB) docker/docker
endif

%.tar: docker/Dockerfile.%.$(DOCKERHUB_USER)
	$(SUDO) docker build -t $(DOCKERHUB_USER)/$* -f $< docker/
	$(SUDO) docker tag $(DOCKERHUB_USER)/$* $(DOCKERHUB_USER)/$*:$(IMAGE_TAG)
	$(SUDO) docker save $(DOCKERHUB_USER)/$*:latest > $@

$(CLOUD_AGENT_EXPORT): docker/Dockerfile.cloud-agent.$(DOCKERHUB_USER) docker/$(SCOPE_EXE) docker/docker docker/weave docker/weaveutil

$(SCOPE_EXPORT): docker/Dockerfile.scope.$(DOCKERHUB_USER) $(CLOUD_AGENT_EXPORT) docker/$(RUNSVINIT) docker/demo.json docker/run-app docker/run-probe docker/entrypoint.sh

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

tests: $(SCOPE_BACKEND_BUILD_UPTODATE) $(CODECGEN_TARGETS)
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

tmp/weave-scope.tgz: $(shell find client/app -type f) $(SCOPE_UI_BUILD_UPTODATE)
	$(sudo) docker run $(RUN_FLAGS) \
	-v $(shell pwd)/client/app:/home/weave/app \
	-v $(shell pwd)/tmp:/home/weave/tmp \
	$(SCOPE_UI_BUILD_IMAGE) \
	npm run bundle

else

client/build/index.html:
	cd client && npm run build

client/build-external/index.html:
	cd client && npm run build-external

endif

$(SCOPE_UI_BUILD_UPTODATE): client/Dockerfile.$(DOCKERHUB_USER) client/package.json client/webpack.local.config.js client/webpack.production.config.js client/server.js client/.eslintrc
	$(SUDO) docker build -t $(SCOPE_UI_BUILD_IMAGE) -f client/Dockerfile.$(DOCKERHUB_USER) client
	touch $@

$(SCOPE_BACKEND_BUILD_UPTODATE): client/Dockerfile.$(DOCKERHUB_USER)
	$(SUDO) docker build -t $(SCOPE_BACKEND_BUILD_IMAGE) -f backend/Dockerfile.$(DOCKERHUB_USER) backend
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
		$(CODECGEN_TARGETS) $(CODECGEN_DIR)/bin \
		client/Dockerfile.$(DOCKERHUB_USER) backend/Dockerfile.$(DOCKERHUB_USER) \
		docker/Dockerfile.*.$(DOCKERHUB_USER) 

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
