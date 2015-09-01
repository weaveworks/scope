FROM alpine:latest
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
WORKDIR /home/weave
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories && \
	apk add --update runit conntrack-tools iproute2 util-linux && \
	rm -rf /var/cache/apk/*
ADD ./docker.tgz /
ADD ./weave /usr/bin/
COPY ./scope-app ./scope-probe ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
