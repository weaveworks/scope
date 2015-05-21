.PHONY: all static clean

# If you can use Docker without being root, you can `make SUDO= <target>`
SUDO=sudo

DOCKERHUB_USER=weaveworks
APP_EXE=app/app
PROBE_EXE=probe/probe
FIXPROBE_EXE=experimental/fixprobe/fixprobe
SCOPE_IMAGE=$(DOCKERHUB_USER)/scope
SCOPE_EXPORT=scope.tar
SCOPE_UI_BUILD_EXPORT=scope_ui_build.tar
SCOPE_UI_BUILD_IMAGE=weave/scope-ui-build

all: $(SCOPE_EXPORT)

$(SCOPE_EXPORT):  $(APP_EXE) $(PROBE_EXE) docker/*
	cp $(APP_EXE) $(PROBE_EXE) docker/
	$(SUDO) docker build -t $(SCOPE_IMAGE) docker/
	$(SUDO) docker save $(SCOPE_IMAGE):latest > $@

$(APP_EXE): app/*.go report/*.go xfer/*.go

$(PROBE_EXE): probe/*.go report/*.go xfer/*.go

$(APP_EXE) $(PROBE_EXE):
	go get -tags netgo ./$(@D)
	go build -ldflags "-extldflags \"-static\"" -tags netgo -o $@ ./$(@D)

static: client/dist/scripts/bundle.js
	esc -o app/static.go -prefix client/dist client/dist

$(SCOPE_UI_BUILD_EXPORT): client/Dockerfile client/gulpfile.js client/package.json
	docker build -t $(SCOPE_UI_BUILD_IMAGE) client
	docker save $(SCOPE_UI_BUILD_IMAGE):latest > $@

client/dist/scripts/bundle.js: client/app/scripts/*
	mkdir -p client/dist
	docker run -ti -v $(shell pwd)/client/app:/home/weave/app \
		-v $(shell pwd)/client/dist:/home/weave/dist \
		$(SCOPE_UI_BUILD_IMAGE)

clean:
	go clean ./...
	rm -rf $(SCOPE_EXPORT) $(SCOPE_UI_BUILD_EXPORT) client/dist
