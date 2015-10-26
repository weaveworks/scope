FROM golang:1.5.1
ENV GO15VENDOREXPERIMENT 1
RUN apt-get update && apt-get install -y libpcap-dev
COPY build.sh /
ENTRYPOINT ["/build.sh"]
