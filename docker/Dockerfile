FROM zlim/bcc
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
WORKDIR /home/weave
RUN apt-get update -y && apt-get install -y bash runit conntrack iproute2 util-linux curl python bcc-tools python-bcc libbcc
ADD ./docker.tgz /
ADD ./demo.json /
ADD ./weave /usr/bin/
COPY ./scope ./runsvinit ./entrypoint.sh ./tcpv4tracer.py /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
