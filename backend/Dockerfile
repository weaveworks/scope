FROM golang:1.5.2
ENV GO15VENDOREXPERIMENT 1
RUN apt-get update && apt-get install -y libpcap-dev python-requests time
RUN go clean -i net && go install -tags netgo std &&  go install -race -tags netgo std
COPY build.sh /
ENTRYPOINT ["/build.sh"]
