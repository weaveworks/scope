FROM ubuntu
LABEL maintainer="Weaveworks Inc <help@weave.works>"
RUN apt-get update && \
  apt-get install -y nginx && \
  rm -rf /var/lib/apt/lists/*
RUN rm /etc/nginx/sites-available/default && \
  ln -sf /dev/stdout /var/log/nginx/access.log && \
  ln -sf /dev/stderr /var/log/nginx/error.log
COPY default.conf /etc/nginx/conf.d/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
