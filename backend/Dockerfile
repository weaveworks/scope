FROM ubuntu:yakkety
ENV GOPATH /go
ENV GOVERSION 1.8
ENV PATH /go/bin:/usr/local/go/bin:/usr/bin:/bin:/usr/sbin:/sbin
ENV SCOPE_SKIP_UI_ASSETS true
RUN apt-get update && \
	apt-get install -y libpcap-dev python-requests time file shellcheck git gcc-arm-linux-gnueabihf curl build-essential && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN curl -Ls https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz | tar xz -C /usr/local
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
	chmod a+wr --recursive /usr/local/go && \
	rm -rf /go/pkg/ /go/src/
COPY build.sh /
ENTRYPOINT ["/build.sh"]
