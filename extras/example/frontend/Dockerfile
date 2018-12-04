FROM ubuntu
RUN apt-get update && \
  apt-get install -y nginx && \
  rm -rf /var/lib/apt/lists/*
RUN rm /etc/nginx/sites-available/default && \
  ln -sf /dev/stdout /var/log/nginx/access.log && \
  ln -sf /dev/stderr /var/log/nginx/error.log
COPY default.conf /etc/nginx/conf.d/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-frontend" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/frontend" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
