FROM ubuntu
LABEL maintainer="Weaveworks Inc <help@weave.works>"
WORKDIR /home/weave
ADD ./qotd /home/weave/
EXPOSE 4446
ENTRYPOINT ["/home/weave/qotd"]
