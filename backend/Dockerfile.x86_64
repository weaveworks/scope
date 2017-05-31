FROM ubuntu:yakkety
ENV GOPATH /go
ENV GOVERSION 1.7
ENV PATH /go/bin:/usr/lib/go-${GOVERSION}/bin:/usr/bin:/bin:/usr/sbin:/sbin
RUN apt-get update && \
	apt-get install -y libpcap-dev python-requests time file shellcheck golang-${GOVERSION} git gcc-arm-linux-gnueabihf curl && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN go clean -i net && \
	go install -tags netgo std && \
	go install -race -tags netgo std
RUN curl -fsSL -o shfmt https://github.com/mvdan/sh/releases/download/v1.3.0/shfmt_v1.3.0_linux_amd64 && \
    chmod +x shfmt && \
    mv shfmt /usr/bin
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/fatih/hclfmt \
		github.com/mjibson/esc \
		github.com/client9/misspell/cmd/misspell && \
	chmod a+wr --recursive /usr/lib/go-${GOVERSION}/pkg && \
	rm -rf /go/pkg/ /go/src/
COPY build.sh /
ENTRYPOINT ["/build.sh"]
