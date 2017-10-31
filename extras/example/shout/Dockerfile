FROM alpine:latest
LABEL maintainer="Weaveworks Inc <help@weave.works>"
WORKDIR /home/weave
ADD shout /home/weave/
EXPOSE 8090
ENTRYPOINT ["/home/weave/shout"]
