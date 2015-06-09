FROM progrium/busybox
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
ADD searchapp entrypoint.sh /home/weave/
EXPOSE 8080
ENTRYPOINT ["/home/weave/entrypoint.sh"]
