FROM alpine:latest
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories && \
    apk add --update runit && \
    rm -rf /var/cache/apk/*

COPY zombie /
RUN mkdir -p /etc/service/zombie
COPY run-zombie /etc/service/zombie/run

COPY /runsvinit /

