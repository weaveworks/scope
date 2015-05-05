.PHONY: default get all build test release clean

default: all

get:
	cd app && go get

all:
	$(MAKE) -C report
	$(MAKE) -C xfer
	$(MAKE) -C app
	$(MAKE) -C bridge
	$(MAKE) -C demoprobe
	$(MAKE) -C oneshot
	$(MAKE) -C fixprobe
	$(MAKE) -C genreport
	$(MAKE) -C probe
	$(MAKE) -C integration
	$(MAKE) -C web

build:
	$(MAKE) -C report build
	$(MAKE) -C xfer build
	$(MAKE) -C app build
	$(MAKE) -C bridge build
	$(MAKE) -C demoprobe build
	$(MAKE) -C oneshot build
	$(MAKE) -C fixprobe build
	$(MAKE) -C probe build
	$(MAKE) -C genreport build
	$(MAKE) -C integration build

release:
	$(MAKE) -C release

clean:
	$(MAKE) -C app clean
	$(MAKE) -C bridge clean
	$(MAKE) -C demoprobe clean
	$(MAKE) -C oneshot clean
	$(MAKE) -C fixprobe clean
	$(MAKE) -C genreport clean
	$(MAKE) -C integration clean
	$(MAKE) -C probe clean
	$(MAKE) -C report clean
	$(MAKE) -C web clean
	$(MAKE) -C xfer clean
