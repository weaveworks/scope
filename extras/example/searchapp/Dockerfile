FROM progrium/busybox
LABEL maintainer="Weaveworks Inc <help@weave.works>"
WORKDIR /home/weave
ADD searchapp /home/weave/
EXPOSE 8080
ENTRYPOINT ["/home/weave/searchapp"]
