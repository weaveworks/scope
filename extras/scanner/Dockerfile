FROM       alpine:3.8
RUN        apk add --no-cache ca-certificates
COPY       scanner /bin/scanner
EXPOSE     6060
ENTRYPOINT [ "/bin/scanner" ]

ARG revision
LABEL org.opencontainers.image.title="scanner" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/app/multitenant/scanner" \
      org.opencontainers.image.revision="${revision}"
