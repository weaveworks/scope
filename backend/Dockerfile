FROM ubuntu:yakkety
ENV GOPATH /go
ENV PATH /go/bin:/usr/lib/go-1.7/bin:/usr/bin:/bin:/usr/sbin:/sbin
RUN echo "deb [trusted=yes] http://52.8.15.63/apt/xenial xenial-nightly main" > /etc/apt/sources.list.d/iovisor.list
RUN apt-get update && \
	apt-get install -y libpcap-dev python-requests time file shellcheck libelf1 bcc-tools libbcc golang-1.7 git && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN go clean -i net && \
	go install -tags netgo std && \
	go install -race -tags netgo std
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/mjibson/esc \
		github.com/client9/misspell/cmd/misspell && \
	rm -rf /go/pkg/ /go/src/

COPY build.sh /
ENTRYPOINT ["/build.sh"]
