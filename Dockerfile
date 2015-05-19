FROM gliderlabs/alpine
MAINTAINER Weaveworks Inc <help@weave.works>
WORKDIR /home/weave
RUN apk add --update supervisor
RUN ["sh", "-c", "rm -rf /var/cache/apk/*"]
ADD supervisord.conf /etc/
ADD ./app/app /home/weave/
ADD ./probe/probe /home/weave/
ADD ./entrypoint.sh /home/weave/
EXPOSE 4040
ENTRYPOINT ["/home/weave/entrypoint.sh"]
