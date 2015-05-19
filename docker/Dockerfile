FROM gliderlabs/alpine
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
ADD ./app /home/weave/
ADD ./probe /home/weave/
ADD ./entrypoint.sh /home/weave/
ADD ./run-app /etc/service/app/run
ADD ./run-probe /etc/service/probe/run

RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories && apk add --update runit

EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
