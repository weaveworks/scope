base   = github.com/alicebob/cello
parts  = app bridge probe
rev    = $(shell git rev-parse --short HEAD)
os     = darwin linux
arch   = amd64
second = $(shell echo $1 | cut -d'-' -f2)
third  = $(shell echo $1 | cut -d'-' -f3)

.PHONY: all clean

all: $(foreach os,$(os),$(foreach arch,$(arch),$(os)-$(arch)-$(rev).tar.gz))

%.tar.gz: $(foreach part,$(parts),$(part)-%)
	tar czvf $@ $^

app-%:
	GOOS=$(call second,$@) GOARCH=$(call third,$@) go build -o $@ $(base)/app

bridge-%:
	GOOS=$(call second,$@) GOARCH=$(call third,$@) go build -o $@ $(base)/bridge

probe-%:
	GOOS=$(call second,$@) GOARCH=$(call third,$@) go build -o $@ $(base)/probe

clean:
	rm *.tar.gz || true

