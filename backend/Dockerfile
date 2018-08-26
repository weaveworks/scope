FROM golang:1.10.2
ENV SCOPE_SKIP_UI_ASSETS true
RUN set -eux; \
   export arch_val="$(dpkg --print-architecture)"; \
   apt-get update && \
   if [ "$arch_val" = "amd64" ]; then \
     apt-get install -y libpcap-dev time file shellcheck git gcc-arm-linux-gnueabihf curl build-essential python-pip; \
   else \
     apt-get install -y libpcap-dev time file shellcheck git curl build-essential python-pip; \
   fi; \
   \
   rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN go clean -i net && \
	go install -tags netgo std && \
   export arch_val="$(dpkg --print-architecture)"; \
   if [ "$arch_val" != "ppc64el" ]; then \
	go install -race -tags netgo std; \
   fi;
RUN export arch_val="$(dpkg --print-architecture)"; \
   if [ "$arch_val" = "amd64" ]; then \
     curl -fsSL -o shfmt https://github.com/mvdan/sh/releases/download/v1.3.0/shfmt_v1.3.0_linux_amd64 && \
     chmod +x shfmt && \
     mv shfmt /usr/bin; \
   fi;
   # Skipped installing shfmt, as the version v1.3.0 isn't supported for ppc64le
   # and the later version of shfmt doesn't work with the application well
RUN go get -tags netgo \
		github.com/fzipp/gocyclo \
		github.com/golang/lint/golint \
		github.com/kisielk/errcheck \
		github.com/fatih/hclfmt \
		github.com/mjibson/esc \
		github.com/client9/misspell/cmd/misspell && \
	chmod a+wr --recursive /usr/local/go && \
	rm -rf /go/pkg/ /go/src/
RUN pip install yapf==0.16.2 flake8==3.3.0 requests==2.19.1

# Install Docker (client only)
ENV DOCKERVERSION=17.03.2-ce
RUN curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz && \
    tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 -C /usr/local/bin docker/docker && \
    rm docker-${DOCKERVERSION}.tgz

COPY build.sh /
ENTRYPOINT ["/build.sh"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="backend" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/backend" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
