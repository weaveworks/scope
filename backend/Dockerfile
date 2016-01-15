FROM golang:1.5.3
ENV GO15VENDOREXPERIMENT 1
RUN apt-get update && apt-get install -y libpcap-dev python-requests time
RUN go clean -i net && go install -tags netgo std &&  go install -race -tags netgo std
RUN go get -tags netgo \
		github.com/golang/lint/golint \
		github.com/fzipp/gocyclo \
		github.com/kisielk/errcheck
COPY build.sh /
ENTRYPOINT ["/build.sh"]
