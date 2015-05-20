FROM gliderlabs/alpine
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
RUN apk add --update curl
ADD ./client /home/weave/
RUN chmod u+x /home/weave/client
ENTRYPOINT ["/home/weave/client"]
