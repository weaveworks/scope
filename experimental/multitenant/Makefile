BUILD_IN_CONTAINER=true

all: .frontend.uptodate

.frontend.uptodate: frontend/*
	docker build -t weaveworks/scope-frontend frontend/
	touch $@

clean:
	go clean ./..
	rm -f .*.uptodate
