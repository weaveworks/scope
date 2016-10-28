FROM zlim/bcc
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
WORKDIR /home/weave
RUN apt-get update -y && apt-get install -y runit bash conntrack iproute2 util-linux libbcc libpcap0.8 localepurge && printf "#NEEDSCONFIGFIRST\nMANDELETE\nSHOWFREEDSPACE\nC\n" > /etc/locale.nopurge && dpkg-reconfigure localepurge && apt-get autoremove -y python curl localepurge && rm -rf /var/lib/dpkg/info/* && rm -rf /var/lib/apt/lists/*
ADD ./docker.tgz /
ADD ./demo.json /
ADD ./weave /usr/bin/
COPY ./scope ./runsvinit ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
