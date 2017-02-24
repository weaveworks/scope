FROM alpine:3.5
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
WORKDIR /home/weave
RUN apk add --update bash runit conntrack-tools iproute2 util-linux curl && \
	rm -rf /var/cache/apk/*
ADD ./docker /usr/local/bin/docker
ADD ./demo.json /
ADD ./weave ./weaveutil /usr/bin/
COPY ./scope ./runsvinit ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
