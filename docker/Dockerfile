FROM gliderlabs/alpine
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories && \
	apk add --update runit conntrack-tools && \
	rm -rf /var/cache/apk/*
COPY ./app ./probe ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
