FROM alpine:3.5
WORKDIR /home/weave
RUN apk add --update bash conntrack-tools iproute2 util-linux curl && \
	rm -rf /var/cache/apk/*
ADD ./weave ./weaveutil /usr/bin/
COPY ./scope /home/weave/
ENTRYPOINT ["/home/weave/scope", "--mode=probe", "--no-app", "--probe.docker=true"]

ARG revision
LABEL works.weave.role="system" \
      maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="cloud-agent" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
