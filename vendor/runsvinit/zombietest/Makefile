GO?=go
SUDO?=
RM?=--rm

.PHONY: test clean

test: .test.uptodate
	./test.bash

.test.uptodate: runsvinit zombie run-zombie Dockerfile
	$(SUDO) docker build -t zombietest .
	touch $@

runsvinit: ../*.go
	env GOOS=linux GOARCH=amd64 $(GO) build -o $@ github.com/peterbourgon/runsvinit

zombie: .build.uptodate
	$(SUDO) docker run $(RM) -v $(shell pwd):/mount zombietest-build cc -Wall -Werror -o /mount/zombie /zombie.c

.build.uptodate: build/zombie.c build/Dockerfile
	$(SUDO) docker build -t zombietest-build build/
	touch $@

clean:
	rm -rf .test.uptodate .build.uptodate runsvinit zombie
	$(SUDO) docker stop zombietest zombietest-build >/dev/null 2>&1 || true
	$(SUDO) docker rm zombietest zombietest-build >/dev/null 2>&1 || true
