FROM golang:1.6.0
ENV GO15VENDOREXPERIMENT 1
RUN apt-get update && apt-get install -y libpcap-dev python-requests time
RUN go clean -i net && go install -tags netgo std &&  go install -race -tags netgo std
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/mjibson/esc
COPY build.sh /
ENTRYPOINT ["/build.sh"]
