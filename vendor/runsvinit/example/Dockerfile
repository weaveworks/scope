FROM alpine:latest
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories && \
    apk add --update runit && \
    rm -rf /var/cache/apk/*

ADD foo /
RUN mkdir -p /etc/service/foo
ADD run-foo /etc/service/foo/run

ADD bar /
RUN mkdir -p /etc/service/bar
ADD run-bar /etc/service/bar/run

ADD /runsvinit /
ENTRYPOINT ["/runsvinit"]
