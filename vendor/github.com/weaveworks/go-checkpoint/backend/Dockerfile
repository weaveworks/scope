FROM golang:1.6.2
RUN apt-get update && \
	apt-get install -y python-requests time file sudo && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/client9/misspell/cmd/misspell && \
	rm -rf /go/pkg/ /go/src/
COPY build.sh /
ENTRYPOINT ["/build.sh"]
