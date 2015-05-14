FROM gliderlabs/alpine
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
ADD ./client /home/weave/
RUN chmod u+x /home/weave/client
ENTRYPOINT ["/home/weave/client"]
