FROM ubuntu:xenial
MAINTAINER Weaveworks Inc <help@weave.works>
LABEL works.weave.role=system
WORKDIR /home/weave
# we could use the xenial-nightly repo to get a more up to date (and smaller) libbcc, but it is not signed
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4052245BD4284CDD \
    && apt-get update -y \
    && apt-get install -y apt-transport-https \
    && echo "deb https://repo.iovisor.org/apt/xenial xenial main" > /etc/apt/sources.list.d/iovisor.list \
    && apt-get update -y \
    && apt-get install -y \
        runit \
        bash \
        conntrack \
        iproute2 \
        util-linux \
        libbcc \
        libpcap0.8 \
    && apt-get remove -y apt-transport-https \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/dpkg/info/* \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /lib/udev/hwdb.* \
    && rm -rf /usr/share/perl/* \
    && rm -rf /usr/share/locale/* \
    && rm -rf /usr/share/man/* \
    && rm -rf /usr/share/doc/* \
    && rm -rf /usr/lib/perl \
    && rm -rf /usr/lib/python3.* \
    && rm -rf /tmp/* \
    && rm -rf /var/tmp/* \
    && rm -rf /var/cache/* \
    && rm -rf /var/log/*
ADD ./docker.tgz ./demo.json /
ADD ./weave /usr/bin/
COPY ./scope ./runsvinit ./entrypoint.sh /home/weave/
COPY ./run-app /etc/service/app/run
COPY ./run-probe /etc/service/probe/run
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
