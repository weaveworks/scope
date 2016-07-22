FROM alpine:3.3
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
COPY ./traffic-control /usr/bin/traffic-control
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk add --update iproute2 && rm -rf /var/cache/apk/*
ENTRYPOINT ["/usr/bin/traffic-control"]
