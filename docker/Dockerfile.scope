FROM weaveworks/cloud-agent
RUN apk add --update runit && \
	rm -rf /var/cache/apk/*
ADD ./demo.json /
ADD ./weave ./weaveutil /usr/bin/
COPY ./runsvinit ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
