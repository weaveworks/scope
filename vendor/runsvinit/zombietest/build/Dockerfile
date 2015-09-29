FROM alpine:latest
RUN apk add --update gcc musl-dev && rm -rf /var/cache/apk/*
COPY zombie.c /
